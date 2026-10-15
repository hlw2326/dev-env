package gateway

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"devenv/internal/domain/site"
	"devenv/internal/ports"
)

func TestHostMatch(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "index.html"), []byte("<h1>Hello Static</h1>"), 0o644); err != nil {
		t.Fatal(err)
	}
	logFile := filepath.Join(t.TempDir(), "access.log")

	gw := New()
	gw.SetRoutes([]ports.Route{
		{
			ID:    1,
			Hosts: []string{"demo.localhost"},
			Kind:  site.KIND_STATIC,
			Path:  root,
			Log:   logFile,
		},
	})
	server := httptest.NewServer(gw)
	defer server.Close()

	request, err := http.NewRequest(http.MethodGet, server.URL+"/", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Host = "demo.localhost"
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	if closeErr := response.Body.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || !strings.Contains(string(body), "Hello Static") {
		t.Fatalf("unexpected static response: status=%d body=%s", response.StatusCode, string(body))
	}

	otherRequest, err := http.NewRequest(http.MethodGet, server.URL+"/", nil)
	if err != nil {
		t.Fatal(err)
	}
	otherRequest.Host = "unknown.localhost"
	otherResponse, err := http.DefaultClient.Do(otherRequest)
	if err != nil {
		t.Fatal(err)
	}
	otherBody, err := io.ReadAll(otherResponse.Body)
	if closeErr := otherResponse.Body.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	if err != nil {
		t.Fatal(err)
	}
	if otherResponse.StatusCode != http.StatusNotFound || !strings.Contains(string(otherBody), "demo.localhost") {
		t.Fatalf("expected overview page with site host, got status=%d body=%s", otherResponse.StatusCode, string(otherBody))
	}

	logBytes, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(logBytes), "demo.localhost GET / 200") {
		t.Fatalf("expected access log entry, got %s", string(logBytes))
	}
}

func TestBlockedPaths(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("SECRET=1"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitFolder := filepath.Join(root, ".git")
	if err := os.MkdirAll(gitFolder, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gitFolder, "config"), []byte("git"), 0o644); err != nil {
		t.Fatal(err)
	}

	gw := New()
	gw.SetRoutes([]ports.Route{
		{
			ID:    1,
			Hosts: []string{"sec.localhost"},
			Kind:  site.KIND_STATIC,
			Path:  root,
		},
	})
	server := httptest.NewServer(gw)
	defer server.Close()

	for _, blocked := range []string{"/.env", "/.git/config", "/.htaccess", "/%2e%2e/secret.txt"} {
		request, err := http.NewRequest(http.MethodGet, server.URL+blocked, nil)
		if err != nil {
			t.Fatal(err)
		}
		request.Host = "sec.localhost"
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		if closeErr := response.Body.Close(); closeErr != nil {
			t.Fatal(closeErr)
		}
		if response.StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403 for %s, got %d", blocked, response.StatusCode)
		}
	}
}

func TestReverseProxy(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if strings.EqualFold(request.Header.Get("Upgrade"), "websocket") {
			hijacker, ok := writer.(http.Hijacker)
			if !ok {
				http.Error(writer, "no hijack", http.StatusInternalServerError)
				return
			}
			conn, rw, err := hijacker.Hijack()
			if err != nil {
				return
			}
			defer func() {
				if closeErr := conn.Close(); closeErr != nil {
					return
				}
			}()
			if _, err := rw.WriteString("HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n\r\n"); err != nil {
				return
			}
			if err := rw.Flush(); err != nil {
				return
			}
			line, err := rw.ReadString('\n')
			if err != nil {
				return
			}
			if _, err := rw.WriteString("echo:" + line); err != nil {
				return
			}
			if err := rw.Flush(); err != nil {
				return
			}
			return
		}
		fmt.Fprintf(writer, "node-ok:%s", request.URL.Path)
	}))
	defer upstream.Close()

	parsed, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	_, portText, err := net.SplitHostPort(parsed.Host)
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatal(err)
	}

	gw := New()
	gw.SetRoutes([]ports.Route{
		{
			ID:    2,
			Hosts: []string{"node.localhost"},
			Kind:  site.KIND_NODE,
			Port:  port,
		},
	})
	server := httptest.NewServer(gw)
	defer server.Close()

	request, err := http.NewRequest(http.MethodGet, server.URL+"/api/ping", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Host = "node.localhost"
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	if closeErr := response.Body.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || string(body) != "node-ok:/api/ping" {
		t.Fatalf("unexpected proxy response: %d %s", response.StatusCode, string(body))
	}

	gwURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	conn, err := net.Dial("tcp", gwURL.Host)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			return
		}
	}()
	handshake := "GET /ws HTTP/1.1\r\nHost: node.localhost\r\nConnection: Upgrade\r\nUpgrade: websocket\r\n\r\n"
	if _, err := conn.Write([]byte(handshake)); err != nil {
		t.Fatal(err)
	}
	reader := bufio.NewReader(conn)
	statusLine, err := reader.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(statusLine, "101") {
		t.Fatalf("expected 101 Switching Protocols, got %s", statusLine)
	}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		if line == "\r\n" {
			break
		}
	}
	if _, err := conn.Write([]byte("hello-ws\n")); err != nil {
		t.Fatal(err)
	}
	echoLine, err := reader.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(echoLine) != "echo:hello-ws" {
		t.Fatalf("expected websocket echo, got %s", echoLine)
	}
}

func TestPHPFallback(t *testing.T) {
	if os.Getenv("DEVENV_FAKE_PHP_CGI") == "1" {
		fmt.Print("Content-Type: text/plain\r\n\r\n")
		fmt.Printf("SCRIPT=%s|URI=%s|QUERY=%s",
			filepath.Base(os.Getenv("SCRIPT_FILENAME")),
			os.Getenv("REQUEST_URI"),
			os.Getenv("QUERY_STRING"),
		)
		return
	}

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "index.php"), []byte("<?php ?>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "asset.txt"), []byte("static-asset"), 0o644); err != nil {
		t.Fatal(err)
	}

	selfExe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	wrapper := filepath.Join(t.TempDir(), "fake-php-cgi.cmd")
	wrapperContent := fmt.Sprintf("@echo off\r\nset DEVENV_FAKE_PHP_CGI=1\r\n\"%s\" -test.run=^TestPHPFallback$\r\n", selfExe)
	if err := os.WriteFile(wrapper, []byte(wrapperContent), 0o755); err != nil {
		t.Fatal(err)
	}

	gw := New()
	gw.SetRoutes([]ports.Route{
		{
			ID:    3,
			Hosts: []string{"php.localhost"},
			Kind:  site.KIND_PHP,
			Path:  root,
			PHP:   wrapper,
		},
	})
	server := httptest.NewServer(gw)
	defer server.Close()

	assetRequest, err := http.NewRequest(http.MethodGet, server.URL+"/asset.txt", nil)
	if err != nil {
		t.Fatal(err)
	}
	assetRequest.Host = "php.localhost"
	assetResponse, err := http.DefaultClient.Do(assetRequest)
	if err != nil {
		t.Fatal(err)
	}
	assetBody, err := io.ReadAll(assetResponse.Body)
	if closeErr := assetResponse.Body.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	if err != nil {
		t.Fatal(err)
	}
	if string(assetBody) != "static-asset" {
		t.Fatalf("expected direct static asset, got %s", string(assetBody))
	}

	fallbackRequest, err := http.NewRequest(http.MethodGet, server.URL+"/posts/42?page=2", nil)
	if err != nil {
		t.Fatal(err)
	}
	fallbackRequest.Host = "php.localhost"
	fallbackResponse, err := http.DefaultClient.Do(fallbackRequest)
	if err != nil {
		t.Fatal(err)
	}
	fallbackBody, err := io.ReadAll(fallbackResponse.Body)
	if closeErr := fallbackResponse.Body.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	if err != nil {
		t.Fatal(err)
	}
	expected := "SCRIPT=index.php|URI=/posts/42?page=2|QUERY=page=2"
	if !strings.Contains(string(fallbackBody), expected) {
		t.Fatalf("expected %s, got %s", expected, string(fallbackBody))
	}

	if _, err := exec.LookPath("php-cgi.exe"); err != nil {
		t.Log("skipping real php-cgi.exe verification: php-cgi.exe not found in PATH")
	}
}
