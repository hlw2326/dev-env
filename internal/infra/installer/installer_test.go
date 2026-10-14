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

func buildTestZip(t *testing.T, entries map[string]string) ([]byte, string) {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	raw := buf.Bytes()
	sum := sha256.Sum256(raw)
	return raw, hex.EncodeToString(sum[:])
}

func TestInstallerSuccessAndFailures(t *testing.T) {
	zipData, checksum := buildTestZip(t, map[string]string{
		"node-v24.21.0-win-x64/node.exe": "fake-node-binary",
	})
	zipSlipData, slipSum := buildTestZip(t, map[string]string{
		"../evil.txt": "pwned",
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok.zip":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(zipData)
		case "/slip.zip":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(zipSlipData)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	mgr := NewManagerWithClient(srv.Client())
	tmp := t.TempDir()

	// 1. Success
	targetOk := filepath.Join(tmp, "software", "nodejs", "24.21.0")
	cacheOk := filepath.Join(tmp, "cache", "nodejs-24.21.0.zip.part")
	err := mgr.Install(context.Background(), ports.Request{
		Kind:     software.KindNode,
		Version:  "24.21.0",
		URL:      srv.URL + "/ok.zip",
		Checksum: checksum,
		Target:   targetOk,
		Cache:    cacheOk,
	}, func(int, string) {})
	if err != nil {
		t.Fatalf("expected install success, got: %v", err)
	}
	if !fileExists(filepath.Join(targetOk, "node.exe")) {
		t.Fatalf("expected node.exe in target directory")
	}

	// 2. Non-200 status fails
	err = mgr.Install(context.Background(), ports.Request{
		Kind:     software.KindNode,
		Version:  "24.21.0",
		URL:      srv.URL + "/404.zip",
		Checksum: checksum,
		Target:   filepath.Join(tmp, "software", "nodejs", "404"),
		Cache:    filepath.Join(tmp, "cache", "404.part"),
	}, func(int, string) {})
	if err == nil {
		t.Fatalf("expected non-200 download to fail")
	}

	// 3. Checksum mismatch fails & cleans up
	badTarget := filepath.Join(tmp, "software", "nodejs", "badsum")
	err = mgr.Install(context.Background(), ports.Request{
		Kind:     software.KindNode,
		Version:  "24.21.0",
		URL:      srv.URL + "/ok.zip",
		Checksum: "0000000000000000000000000000000000000000000000000000000000000000",
		Target:   badTarget,
		Cache:    filepath.Join(tmp, "cache", "badsum.part"),
	}, func(int, string) {})
	if err == nil {
		t.Fatalf("expected checksum mismatch to fail")
	}
	if _, statErr := os.Stat(badTarget); !os.IsNotExist(statErr) {
		t.Fatalf("expected target dir not to exist on checksum failure")
	}

	// 4. Zip-slip fails
	err = mgr.Install(context.Background(), ports.Request{
		Kind:     software.KindNode,
		Version:  "24.21.0",
		URL:      srv.URL + "/slip.zip",
		Checksum: slipSum,
		Target:   filepath.Join(tmp, "software", "nodejs", "slip"),
		Cache:    filepath.Join(tmp, "cache", "slip.part"),
	}, func(int, string) {})
	if err == nil {
		t.Fatalf("expected zip-slip archive to fail")
	}
}
