//go:build windows

package environment

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"devenv/internal/domain/software"
)

func TestInspectNodePathOrderAndUnicode(t *testing.T) {
	root := t.TempDir()
	first := filepath.Join(root, "系统 Node")
	second := filepath.Join(root, "DevEnv commands")
	for _, folder := range []string{first, second} {
		if err := os.MkdirAll(folder, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeNodeFixture(t, first, "@echo off\r\necho v20.13.1\r\n")
	writeNodeFixture(t, second, "@echo off\r\necho v24.22.0\r\n")
	environment := nodeTestEnvironment(first + ";" + second + ";" + first)
	entries, err := inspectNodeFixture(context.Background(), environment)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected two distinct commands: %+v", entries)
	}
	if entries[0].Path != filepath.Join(first, "node.cmd") || strings.Join(entries[0].Versions, ",") != "20.13.1" || !entries[0].Active {
		t.Fatalf("wrong PATH winner: %+v", entries)
	}
	if strings.Join(entries[1].Versions, ",") != "24.22.0" || entries[1].Active {
		t.Fatalf("wrong second command: %+v", entries)
	}
}

func TestInspectNodeMissingAndBrokenCommand(t *testing.T) {
	folder := t.TempDir()
	entries, err := inspectNodeFixture(context.Background(), nodeTestEnvironment(folder))
	if err != nil || entries == nil || len(entries) != 0 {
		t.Fatalf("missing command: %+v, %v", entries, err)
	}
	writeNodeFixture(t, folder, "@echo off\r\nexit /b 7\r\n")
	entries, err = inspectNodeFixture(context.Background(), nodeTestEnvironment(folder))
	if err != nil || len(entries) != 1 {
		t.Fatalf("broken command: %+v, %v", entries, err)
	}
	if entries[0].Error == "" || len(entries[0].Versions) != 0 || !entries[0].Active {
		t.Fatalf("broken winner must not be reported as absent: %+v", entries)
	}
}

func TestProbeNodeRejectsInvalidOutputAndTimesOut(t *testing.T) {
	folder := t.TempDir()
	path := writeNodeFixture(t, folder, "@echo off\r\necho not-node\r\n")
	if _, err := probeCommand(context.Background(), path, software.InspectionProbes(software.KIND_NODE)[0], nodeTestEnvironment(folder)); err == nil {
		t.Fatal("accepted invalid version")
	}
	writeNodeFixture(t, folder, "@echo off\r\n:repeat\r\ngoto repeat\r\n")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if _, err := probeCommand(ctx, path, software.InspectionProbes(software.KIND_NODE)[0], nodeTestEnvironment(folder)); err == nil {
		t.Fatal("expected timeout")
	}
}

func TestCommandInspectorRejectsUnsupportedCommands(t *testing.T) {
	if _, err := NewCommandInspector().Inspect(context.Background(), "node & other", ""); err == nil {
		t.Fatal("expected command allowlist validation")
	}
}

func writeNodeFixture(t *testing.T, folder, content string) string {
	t.Helper()
	path := filepath.Join(folder, "node.cmd")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func nodeTestEnvironment(path string) []string {
	result := make([]string, 0)
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(strings.ToUpper(entry), "PATH=") && !strings.HasPrefix(strings.ToUpper(entry), "PATHEXT=") {
			result = append(result, entry)
		}
	}
	return append(result, "PATH="+path, "PATHEXT=.COM;.EXE;.BAT;.CMD")
}

// Explicit opt-in, read-only diagnostic for the developer's actual Windows PATH.
func TestCommandInspectorLive(t *testing.T) {
	if os.Getenv("DEVENV_TEST_COMMAND_ENVIRONMENT") != "1" {
		t.Skip("live command detection is opt-in")
	}
	inspector := NewCommandInspector()
	for _, kind := range software.AllKinds() {
		scan, err := inspector.Inspect(context.Background(), kind, "")
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range scan.Entries {
			t.Logf("%s active=%t versions=%v source=%s path=%s error=%s", kind, entry.Active, entry.Versions, entry.Source, entry.Path, entry.Error)
		}
		for _, warning := range scan.Warnings {
			t.Logf("%s warning: %s", kind, warning)
		}
	}
}

func inspectNodeFixture(ctx context.Context, environment []string) ([]software.CommandLocation, error) {
	probes := software.InspectionProbes(software.KIND_NODE)
	entries, err := discoverCommand(ctx, probes[0], environment)
	if err != nil {
		return entries, err
	}
	inspector := &commandInspector{workers: make(chan struct{}, 4)}
	inspector.inspectEntries(ctx, entries, probes, environment)
	return entries, nil
}
