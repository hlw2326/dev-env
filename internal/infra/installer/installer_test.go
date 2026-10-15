package installer

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"devenv/internal/domain/software"
	"devenv/internal/ports"
)

func buildZip(t *testing.T, entries map[string]string) ([]byte, string) {
	t.Helper()
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	for name, content := range entries {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	raw := buf.Bytes()
	sum := sha256.Sum256(raw)
	return raw, hex.EncodeToString(sum[:])
}

func TestInstallCases(t *testing.T) {
	okData, checksum := buildZip(t, map[string]string{
		"node-v24.21.0-win-x64/node.exe": "fake-node-binary",
	})
	slipData, slipSum := buildZip(t, map[string]string{
		"../evil.txt": "pwned",
	})
	emptyEntry, emptySum := buildZip(t, map[string]string{
		"node-v24.21.0-win-x64/readme.txt": "missing-node-exe",
	})

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/ok.zip":
			writer.WriteHeader(http.StatusOK)
			if _, err := writer.Write(okData); err != nil {
				t.Error(err)
			}
		case "/slip.zip":
			writer.WriteHeader(http.StatusOK)
			if _, err := writer.Write(slipData); err != nil {
				t.Error(err)
			}
		case "/missing.zip":
			writer.WriteHeader(http.StatusOK)
			if _, err := writer.Write(emptyEntry); err != nil {
				t.Error(err)
			}
		default:
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	manager := NewWithClient(server.Client())
	folder := t.TempDir()

	// 1. Success
	targetOk := filepath.Join(folder, "software", "nodejs", "24.21.0")
	cacheOk := filepath.Join(folder, "cache", "nodejs-24.21.0.zip.part")
	if err := manager.Install(context.Background(), ports.Request{
		Kind:     software.KIND_NODE,
		Version:  "24.21.0",
		URL:      server.URL + "/ok.zip",
		Checksum: checksum,
		Target:   targetOk,
		Cache:    cacheOk,
	}, func(int, string) {}); err != nil {
		t.Fatalf("expected install success, got: %v", err)
	}
	if !fileExists(filepath.Join(targetOk, "node.exe")) {
		t.Fatalf("expected node.exe in target directory")
	}

	// 2. Non-200 status fails
	if err := manager.Install(context.Background(), ports.Request{
		Kind:     software.KIND_NODE,
		Version:  "24.21.0",
		URL:      server.URL + "/404.zip",
		Checksum: checksum,
		Target:   filepath.Join(folder, "software", "nodejs", "404"),
		Cache:    filepath.Join(folder, "cache", "404.part"),
	}, func(int, string) {}); err == nil {
		t.Fatalf("expected non-200 download to fail")
	}

	// 3. Checksum mismatch fails & cleans up
	badTarget := filepath.Join(folder, "software", "nodejs", "badsum")
	badCache := filepath.Join(folder, "cache", "badsum.part")
	if err := manager.Install(context.Background(), ports.Request{
		Kind:     software.KIND_NODE,
		Version:  "24.21.0",
		URL:      server.URL + "/ok.zip",
		Checksum: "0000000000000000000000000000000000000000000000000000000000000000",
		Target:   badTarget,
		Cache:    badCache,
	}, func(int, string) {}); err == nil {
		t.Fatalf("expected checksum mismatch to fail")
	}
	if _, err := os.Stat(badTarget); !os.IsNotExist(err) {
		t.Fatalf("expected target not to exist on checksum failure")
	}
	if _, err := os.Stat(badCache); !os.IsNotExist(err) {
		t.Fatalf("expected cache not to exist on checksum failure")
	}

	// 4. Zip-slip fails & cleans up
	slipTarget := filepath.Join(folder, "software", "nodejs", "slip")
	if err := manager.Install(context.Background(), ports.Request{
		Kind:     software.KIND_NODE,
		Version:  "24.21.0",
		URL:      server.URL + "/slip.zip",
		Checksum: slipSum,
		Target:   slipTarget,
		Cache:    filepath.Join(folder, "cache", "slip.part"),
	}, func(int, string) {}); err == nil {
		t.Fatalf("expected zip-slip archive to fail")
	}
	if _, err := os.Stat(slipTarget + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("expected temp folder cleaned up after zip-slip")
	}

	// 5. Missing entry binary fails & cleans up
	missingTarget := filepath.Join(folder, "software", "nodejs", "missing")
	if err := manager.Install(context.Background(), ports.Request{
		Kind:     software.KIND_NODE,
		Version:  "24.21.0",
		URL:      server.URL + "/missing.zip",
		Checksum: emptySum,
		Target:   missingTarget,
		Cache:    filepath.Join(folder, "cache", "missing.part"),
	}, func(int, string) {}); err == nil {
		t.Fatalf("expected missing entry binary to fail")
	}
	if _, err := os.Stat(missingTarget + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("expected temp folder cleaned up after missing entry")
	}

	// 6. Canceled context fails & cleans up
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel()
	cancelTarget := filepath.Join(folder, "software", "nodejs", "cancel")
	cancelCache := filepath.Join(folder, "cache", "cancel.part")
	if err := manager.Install(cancelCtx, ports.Request{
		Kind:     software.KIND_NODE,
		Version:  "24.21.0",
		URL:      server.URL + "/ok.zip",
		Checksum: checksum,
		Target:   cancelTarget,
		Cache:    cancelCache,
	}, func(int, string) {}); err == nil {
		t.Fatalf("expected canceled install to fail")
	}
	if _, err := os.Stat(cancelCache); !os.IsNotExist(err) {
		t.Fatalf("expected cache cleaned up after cancel")
	}
}

func TestPortableDeveloperToolEntries(t *testing.T) {
	cases := []struct {
		kind  string
		entry string
	}{
		{software.KIND_JAVA, filepath.Join("bin", "java.exe")},
		{software.KIND_DOTNET, "dotnet.exe"},
		{software.KIND_DENO, "deno.exe"},
		{software.KIND_BUN, "bun.exe"},
		{software.KIND_NGINX, "nginx.exe"},
		{software.KIND_CADDY, "caddy.exe"},
		{software.KIND_SQLITE, "sqlite3.exe"},
		{software.KIND_GRADLE, filepath.Join("bin", "gradle.bat")},
		{software.KIND_MAVEN, filepath.Join("bin", "mvn.cmd")},
	}
	for _, test := range cases {
		t.Run(test.kind, func(t *testing.T) {
			folder := t.TempDir()
			entry := filepath.Join(folder, test.entry)
			if err := os.MkdirAll(filepath.Dir(entry), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(entry, []byte("test"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := checkEntry(test.kind, folder); err != nil {
				t.Fatalf("expected %s entry to pass: %v", test.kind, err)
			}
		})
	}
}
