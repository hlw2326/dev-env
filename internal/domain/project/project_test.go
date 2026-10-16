package project

import (
	"os"
	"path/filepath"
	"testing"

	"devenv/internal/domain/software"
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
	if got := FindCommand(nodeFolder, KIND_NODE); got != "npm start" {
		t.Fatalf("expected npm start, got %q", got)
	}

	miniFolder := filepath.Join(folder, "miniprogram")
	if err := os.MkdirAll(miniFolder, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(miniFolder, "package.json"), []byte(`{"scripts":{"dev:mp-weixin":"uni -p mp-weixin"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := FindCommand(miniFolder, KIND_MINIPROGRAM); got != "npm run dev:mp-weixin" {
		t.Fatalf("expected mini program development command, got %q", got)
	}
	commands := FindCommands(miniFolder, KIND_MINIPROGRAM)
	if len(commands) != 1 || commands[0] != "npm run dev:mp-weixin" {
		t.Fatalf("expected mini program command options, got %#v", commands)
	}

	pythonFolder := filepath.Join(folder, "pythonapp")
	if err := os.MkdirAll(pythonFolder, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pythonFolder, "app.py"), []byte("print('ok')"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := FindCommand(pythonFolder, KIND_PYTHON); got != "python app.py" {
		t.Fatalf("expected python app.py, got %q", got)
	}

	rustFolder := filepath.Join(folder, "rustapp")
	if err := os.MkdirAll(rustFolder, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rustFolder, "Cargo.toml"), []byte("[package]"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := FindCommand(rustFolder, KIND_RUST); got != "cargo run" {
		t.Fatalf("expected cargo run, got %q", got)
	}

	if _, err := NormalizeRun(phpFolder, "../etc"); err == nil {
		t.Fatalf("expected path traversal error")
	}
	if got, err := NormalizeRun(phpFolder, "/public"); err != nil || got != "public" {
		t.Fatalf("expected public, got %q, err=%v", got, err)
	}
}

func TestMiniProgramProjectKind(t *testing.T) {
	if !IsValidKind(KIND_MINIPROGRAM) {
		t.Fatal("mini program kind should be valid")
	}
	if !IsLocal(KIND_MINIPROGRAM) || UsesNginx(KIND_MINIPROGRAM) {
		t.Fatal("mini program should run locally without Nginx")
	}
	if RuntimeSoftwareKind(KIND_MINIPROGRAM) != software.KIND_NODE {
		t.Fatal("mini program should use the Node.js runtime")
	}
}

func TestEnvironmentFilesPreferDevelopmentAndStayInProjectRoot(t *testing.T) {
	folder := t.TempDir()
	for _, name := range []string{".env.prod", ".env.dev", ".env.example"} {
		if err := os.WriteFile(filepath.Join(folder, name), []byte("APP_ENV=test"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	files := FindEnvironmentFiles(folder)
	if len(files) != 2 || DefaultEnvironmentFile(folder) != ".env.dev" {
		t.Fatalf("unexpected environment files: %#v", files)
	}
	if got, err := NormalizeEnvironmentFile(folder, ".env.dev"); err != nil || got != ".env.dev" {
		t.Fatalf("expected normalized environment file, got %q, err=%v", got, err)
	}
	if _, err := NormalizeEnvironmentFile(folder, "../.env.dev"); err == nil {
		t.Fatal("expected environment file traversal error")
	}
}

func TestResolveEnvironmentFileSkipsMissingOptionalFile(t *testing.T) {
	folder := t.TempDir()

	resolved, err := ResolveEnvironmentFile(folder, ".env.dev")
	if err != nil {
		t.Fatalf("missing optional environment file should not stop startup: %v", err)
	}
	if resolved != "" {
		t.Fatalf("expected missing environment file to be skipped, got %q", resolved)
	}
	if _, err := ResolveEnvironmentFile(folder, "../.env.dev"); err == nil {
		t.Fatal("invalid environment file path should still fail")
	}
}

func TestProjectTemplatesExposeSharedCapabilities(t *testing.T) {
	phpTemplate, ok := ProjectTemplateForKind(KIND_PHP)
	if !ok {
		t.Fatal("PHP template should be registered")
	}
	if !HasCapability(KIND_PHP, CAPABILITY_RUN_DIRECTORY) || !HasCapability(KIND_PHP, CAPABILITY_RUNTIME) {
		t.Fatal("PHP should use shared run directory and runtime capabilities")
	}
	if HasCapability(KIND_PHP, CAPABILITY_COMMAND) {
		t.Fatal("PHP should not expose a process command capability")
	}
	if phpTemplate.RuntimeKind == "" {
		t.Fatal("PHP template should declare a runtime")
	}
	if !HasCapability(KIND_MINIPROGRAM, CAPABILITY_COMMAND) || HasCapability(KIND_MINIPROGRAM, CAPABILITY_DOMAIN) {
		t.Fatal("mini program capabilities should be local command-only")
	}
}
