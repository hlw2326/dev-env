//go:build windows

package environment

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"devenv/internal/domain/software"
)

func TestEnvironmentHomeDoesNotBecomePathWinner(t *testing.T) {
	root := t.TempDir()
	pathFolder := filepath.Join(root, "path")
	homeFolder := filepath.Join(root, "home")
	for _, folder := range []string{pathFolder, homeFolder} {
		if err := os.MkdirAll(folder, 0o755); err != nil {
			t.Fatal(err)
		}
		writeNodeFixture(t, folder, "@echo off\r\necho v24.22.0\r\n")
	}
	probe := software.InspectionProbes(software.KIND_NODE)[0]
	probe.Home = "TEST_RUNTIME_HOME"
	probe.Relative = "node.cmd"
	entries, err := discoverCommand(context.Background(), probe, append(nodeTestEnvironment(pathFolder), "TEST_RUNTIME_HOME="+homeFolder))
	if err != nil || len(entries) != 2 {
		t.Fatalf("home detection: %+v, %v", entries, err)
	}
	if !entries[0].Active || entries[1].Active || entries[1].Source != probe.Home {
		t.Fatalf("home treated as PATH: %+v", entries)
	}
	entries, err = discoverCommand(context.Background(), probe, append(nodeTestEnvironment(pathFolder), "TEST_RUNTIME_HOME="+pathFolder))
	if err != nil || len(entries) != 1 {
		t.Fatalf("duplicate command: %+v, %v", entries, err)
	}
	entries, err = discoverCommand(context.Background(), probe, append(nodeTestEnvironment(pathFolder), "TEST_RUNTIME_HOME="+filepath.Join(root, "missing")))
	if err != nil || len(entries) != 2 || entries[1].Error == "" {
		t.Fatalf("missing HOME target hidden: %+v, %v", entries, err)
	}
}

func TestServiceExecutableDiscardsStartupArguments(t *testing.T) {
	for _, commandLine := range []string{
		`"C:\Program Files\PostgreSQL\bin\pg_ctl.exe" runservice -N "Postgres" -D "C:\data"`,
		`C:\Program Files\PostgreSQL\bin\pg_ctl.exe runservice -N Postgres`,
	} {
		path, err := serviceExecutable(commandLine)
		if err != nil || path != `C:\Program Files\PostgreSQL\bin\pg_ctl.exe` {
			t.Fatalf("bad service executable: %q %v", path, err)
		}
	}
	for _, invalid := range []string{`postgres.exe start`, `C:\Windows\system32\driver.sys`, `cmd.exe /c start mysql`} {
		if _, err := serviceExecutable(invalid); err == nil {
			t.Fatalf("invalid service target accepted: %s", invalid)
		}
	}
}

func TestProbeSkipsWindowsStoreAliasesAndClearsHooks(t *testing.T) {
	probe := software.InspectionProbes(software.KIND_PYTHON)[0]
	_, err := probeCommand(context.Background(), `C:\Users\test\AppData\Local\Microsoft\WindowsApps\python.exe`, probe, nodeTestEnvironment(t.TempDir()))
	if err == nil || !strings.Contains(err.Error(), "执行别名") {
		t.Fatalf("store alias was invoked: %v", err)
	}
	result := probeEnvironment([]string{"PATH=C:\\tools", "NODE_OPTIONS=--require bad.js", "JAVA_TOOL_OPTIONS=-javaagent:bad.jar", "GOTOOLCHAIN=auto", "JDK_JAVA_OPTIONS=bad", "DOTNET_STARTUP_HOOKS=bad.dll"})
	expected := []string{"PATH=C:\\tools", "GOTOOLCHAIN=local", "DOTNET_CLI_TELEMETRY_OPTOUT=1", "DOTNET_SKIP_FIRST_TIME_EXPERIENCE=1"}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("unsafe probe environment: %v", result)
	}
}

func TestVersionProbeReadsStderr(t *testing.T) {
	folder := t.TempDir()
	path := writeNodeFixture(t, folder, "@echo off\r\necho nginx version: nginx/1.30.6 1>&2\r\n")
	versions, err := probeCommand(context.Background(), path, software.InspectionProbes(software.KIND_NGINX)[0], nodeTestEnvironment(folder))
	if err != nil || !reflect.DeepEqual(versions, []string{"1.30.6"}) {
		t.Fatalf("stderr version: %v, %v", versions, err)
	}
}

func TestSelectedPortableExecutable(t *testing.T) {
	folder := t.TempDir()
	path := writeNodeFixture(t, folder, "@echo off\r\necho v24.22.0\r\n")
	inspector := NewCommandInspector()
	result, err := inspector.Inspect(context.Background(), software.KIND_NODE, path)
	if err != nil || len(result.Entries) != 1 {
		t.Fatalf("manual program: %+v %v", result, err)
	}
	entry := result.Entries[0]
	if entry.Source != "手动选择" || entry.Active || !reflect.DeepEqual(entry.Versions, []string{"24.22.0"}) {
		t.Fatalf("incorrect manual result: %+v", entry)
	}
	for _, invalid := range []string{folder, "node.cmd", filepath.Join(folder, "missing.exe")} {
		if _, err := inspector.Inspect(context.Background(), software.KIND_NODE, invalid); err == nil {
			t.Fatalf("accepted invalid program: %s", invalid)
		}
	}
	if _, err := inspector.Inspect(context.Background(), software.KIND_PHP, path); err == nil {
		t.Fatal("accepted wrong software's executable")
	}
}
