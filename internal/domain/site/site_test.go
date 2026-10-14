package site

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindRunAndCommand(t *testing.T) {
	tmp := t.TempDir()

	// PHP in public/index.php
	phpDir := filepath.Join(tmp, "tp")
	if err := os.MkdirAll(filepath.Join(phpDir, "public"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(phpDir, "public", "index.php"), []byte("<?php"), 0644); err != nil {
		t.Fatal(err)
	}
	if got := FindRun(phpDir, KindPHP); got != "public" {
		t.Fatalf("expected public, got %q", got)
	}

	// Static in dist/index.html
	staticDir := filepath.Join(tmp, "web")
	if err := os.MkdirAll(filepath.Join(staticDir, "dist"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staticDir, "dist", "index.html"), []byte("<html>"), 0644); err != nil {
		t.Fatal(err)
	}
	if got := FindRun(staticDir, KindStatic); got != "dist" {
		t.Fatalf("expected dist, got %q", got)
	}

	// Node command detection
	nodeDir := filepath.Join(tmp, "nodeapp")
	if err := os.MkdirAll(nodeDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nodeDir, "package.json"), []byte(`{"scripts":{"start":"node app.js"}}`), 0644); err != nil {
		t.Fatal(err)
	}
	if got := FindCommand(nodeDir); got != "npm start" {
		t.Fatalf("expected npm start, got %q", got)
	}

	// Path traversal rejected
	if _, err := NormalizeRun(phpDir, "../etc"); err == nil {
		t.Fatalf("expected path traversal error")
	}
	if got, err := NormalizeRun(phpDir, "/public"); err != nil || got != "public" {
		t.Fatalf("expected public, got %q, err=%v", got, err)
	}
}
