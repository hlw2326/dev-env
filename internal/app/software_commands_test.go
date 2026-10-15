package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devenv/internal/domain/software"
	"devenv/internal/infra/sqlite"
	"devenv/internal/shared/paths"
)

type commandPathRecorder struct {
	path    string
	failure error
}

func (p *commandPathRecorder) Ensure(path string) error {
	p.path = path
	return p.failure
}

func (p *commandPathRecorder) Remove(path string) error {
	p.path = ""
	return p.failure
}

func TestSyncGlobalCommandsRepairsExistingInstallations(t *testing.T) {
	base := t.TempDir()
	ctx := context.Background()
	db, err := sqlite.Open(paths.Database(base), sqlite.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := sqlite.NewSoftwareRepository(db.DB)
	create := func(kind, version, binary string, main bool) software.Software {
		t.Helper()
		item := software.Software{Kind: kind, Version: version, State: software.STATE_INSTALLED, Main: main, Path: paths.Software(base, kind, version)}
		if err := os.MkdirAll(filepath.Dir(filepath.Join(item.Path, binary)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(item.Path, binary), []byte("test fixture"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := repo.Create(ctx, &item); err != nil {
			t.Fatal(err)
		}
		return item
	}
	composer := create(software.KIND_COMPOSER, "2.10.3", "composer.phar", true)
	php := create(software.KIND_PHP, "8.3.35", "php.exe", true)
	newPHP := create(software.KIND_PHP, "8.5.11", "php.exe", false)
	node := create(software.KIND_NODE, "24.22.0", "node.exe", true)
	readOnly, err := sqlite.OpenReadOnly(paths.Database(base))
	if err != nil {
		t.Fatal(err)
	}
	defer readOnly.Close()
	readRepo := sqlite.NewSoftwareRepository(readOnly)
	pathManager := &commandPathRecorder{}
	syncCommands := func() {
		t.Helper()
		if err := SyncGlobalCommands(ctx, base, readRepo, pathManager); err != nil {
			t.Fatal(err)
		}
	}
	assertCommand := func(name, target string) {
		t.Helper()
		raw, err := os.ReadFile(filepath.Join(paths.Commands(base), name+".bat"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(raw), target) {
			t.Fatalf("%s does not target %s: %s", name, target, raw)
		}
	}
	syncCommands()
	if pathManager.path != paths.Commands(base) {
		t.Fatalf("PATH was not bound: %q", pathManager.path)
	}
	assertCommand("composer", filepath.Join(php.Path, "php.exe"))
	assertCommand("node", filepath.Join(node.Path, "node.exe"))
	if err := repo.SetMain(ctx, software.KIND_PHP, newPHP.ID); err != nil {
		t.Fatal(err)
	}
	syncCommands()
	assertCommand("composer", filepath.Join(newPHP.Path, "php.exe"))
	if err := repo.Delete(ctx, composer.ID); err != nil {
		t.Fatal(err)
	}
	syncCommands()
	if _, err := os.Stat(filepath.Join(paths.Commands(base), "composer.bat")); !os.IsNotExist(err) {
		t.Fatalf("stale Composer command: %v", err)
	}
	assertCommand("node", filepath.Join(node.Path, "node.exe"))
	pathManager.failure = errors.New("registry write denied")
	if err := SyncGlobalCommands(ctx, base, readRepo, pathManager); err == nil || !strings.Contains(err.Error(), pathManager.failure.Error()) {
		t.Fatalf("expected PATH failure to be reported, got %v", err)
	}
}
