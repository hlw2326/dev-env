package gateway

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"devenv/internal/domain/site"
	"devenv/internal/ports"
)

func TestGatewayStaticProxyAndSecurity(t *testing.T) {
	tmp := t.TempDir()
	staticRoot := filepath.Join(tmp, "static")
	if err := os.MkdirAll(staticRoot, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staticRoot, "index.html"), []byte("<h1>Hello Static</h1>"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staticRoot, ".env"), []byte("SECRET=1"), 0644); err != nil {
		t.Fatal(err)
	}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("proxied:" + r.URL.Path))
	}))
	defer upstream.Close()

	u, _ := url.Parse(upstream.URL)
	_, portStr, _ := strings.Cut(u.Host, ":")
	nodePort, _ := strconv.Atoi(portStr)

	gw := NewServer()
	gw.SetRoutes([]ports.Route{
		{
			ID:    1,
			Hosts: []string{"static.localhost"},
			Kind:  site.KindStatic,
			Path:  staticRoot,
			Log:   filepath.Join(tmp, "access.log"),
		},
		{
			ID:    2,
			Hosts: []string{"node.localhost"},
			Kind:  site.KindNode,
			Port:  nodePort,
		},
	})

	// 1. Static index
	req := httptest.NewRequest(http.MethodGet, "http://static.localhost/", nil)
	rec := httptest.NewRecorder()
	gw.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Hello Static") {
		t.Fatalf("expected 200 Hello Static, got %d %s", rec.Code, rec.Body.String())
	}

	// 2. Hidden file blocked (.env)
	reqEnv := httptest.NewRequest(http.MethodGet, "http://static.localhost/.env", nil)
	recEnv := httptest.NewRecorder()
	gw.ServeHTTP(recEnv, reqEnv)
	if recEnv.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for .env, got %d", recEnv.Code)
	}

	// 3. Traversal blocked
	reqTrav := httptest.NewRequest(http.MethodGet, "http://static.localhost/..%2fsecret", nil)
	recTrav := httptest.NewRecorder()
	gw.ServeHTTP(recTrav, reqTrav)
	if recTrav.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for path traversal, got %d", recTrav.Code)
	}

	// 4. Node reverse proxy
	reqNode := httptest.NewRequest(http.MethodGet, "http://node.localhost/api/ping", nil)
	recNode := httptest.NewRecorder()
	gw.ServeHTTP(recNode, reqNode)
	if recNode.Code != http.StatusOK || recNode.Body.String() != "proxied:/api/ping" {
		t.Fatalf("expected proxied:/api/ping, got %d %s", recNode.Code, recNode.Body.String())
	}

	// 5. Unmatched host lists sites
	reqUnmatched := httptest.NewRequest(http.MethodGet, "http://unknown.localhost/", nil)
	recUnmatched := httptest.NewRecorder()
	gw.ServeHTTP(recUnmatched, reqUnmatched)
	if recUnmatched.Code != http.StatusNotFound || !strings.Contains(recUnmatched.Body.String(), "static.localhost") {
		t.Fatalf("expected unmatched host page, got %d %s", recUnmatched.Code, recUnmatched.Body.String())
	}
}

func TestGatewayRealPHPCGI(t *testing.T) {
	localAppData := os.Getenv("LOCALAPPDATA")
	phpDir := filepath.Join(localAppData, "DevEnv", "software", "php", "8.3.35")
	phpCgi := filepath.Join(phpDir, "php-cgi.exe")
	if _, err := os.Stat(phpCgi); err != nil {
		t.Skipf("skipping real PHP-CGI integration test: %s not found", phpCgi)
	}

	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "index.php"), []byte(`<?php header("X-Custom: DevEnv"); echo "PHP-" . PHP_MAJOR_VERSION . ":" . $_SERVER['REQUEST_URI'];`), 0644); err != nil {
		t.Fatal(err)
	}

	gw := NewServer()
	gw.SetRoutes([]ports.Route{
		{
			ID:    10,
			Hosts: []string{"php.localhost"},
			Kind:  site.KindPHP,
			Path:  tmp,
			PHP:   phpDir,
		},
	})

	req := httptest.NewRequest(http.MethodGet, "http://php.localhost/hello/world?foo=bar", nil)
	rec := httptest.NewRecorder()
	gw.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 from php-cgi, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("X-Custom") != "DevEnv" {
		t.Fatalf("expected X-Custom header 'DevEnv', got %q", rec.Header().Get("X-Custom"))
	}
	if !strings.Contains(rec.Body.String(), "PHP-8:/hello/world?foo=bar") {
		t.Fatalf("unexpected PHP output: %q", rec.Body.String())
	}
}

