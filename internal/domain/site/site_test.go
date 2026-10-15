package site

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindRun(t *testing.T) {
	folder := t.TempDir()

	phpFolder := filepath.Join(folder, "tp")
	if err := os.MkdirAll(filepath.Join(phpFolder, "public"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(phpFolder, "public", "index.php"), []byte("<?php"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := FindRun(phpFolder, KIND_PHP); got != "public" {
		t.Fatalf("expected public, got %q", got)
	}

	staticFolder := filepath.Join(folder, "web")
	if err := os.MkdirAll(filepath.Join(staticFolder, "dist"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staticFolder, "dist", "index.html"), []byte("<html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := FindRun(staticFolder, KIND_STATIC); got != "dist" {
		t.Fatalf("expected dist, got %q", got)
	}

	nodeFolder := filepath.Join(folder, "nodeapp")
	if err := os.MkdirAll(nodeFolder, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nodeFolder, "package.json"), []byte(`{"scripts":{"start":"node app.js"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := FindCommand(nodeFolder); got != "npm start" {
		t.Fatalf("expected npm start, got %q", got)
	}

	if _, err := NormalizeRun(phpFolder, "../etc"); err == nil {
		t.Fatalf("expected path traversal error")
	}
	if got, err := NormalizeRun(phpFolder, "/public"); err != nil || got != "public" {
		t.Fatalf("expected public, got %q, err=%v", got, err)
	}
}
