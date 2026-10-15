//go:build windows

package app

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"devenv/internal/domain/software"
	"devenv/internal/shared/paths"
)

// A native program returning 1 reproduces the stale ERRORLEVEL expansion in
// the former IF block without requiring an installed PHP runtime.
func TestComposerCommandPreservesExitCode(t *testing.T) {
	folder := t.TempDir()
	content := renderGlobalCommand(globalCommandBinding{
		Target:     "devenv-command-test-file-that-does-not-exist",
		PHPBinary:  filepath.Join(os.Getenv("SystemRoot"), "System32", "where.exe"),
		Definition: software.CommandDefinition{RequiresPHP: true},
	})
	if err := os.WriteFile(filepath.Join(folder, "composer.bat"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, paths.Shell(), "/d", "/c", "composer.bat")
	command.Dir = folder
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := command.CombinedOutput()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Fatalf("wrapper must preserve exit code 1, got %v: %s", err, output)
	}
}
