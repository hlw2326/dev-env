package autostart

import (
	"testing"

	"golang.org/x/sys/windows"
)

func TestApplicationCommand(t *testing.T) {
	path := `D:\Program Files\DevEnv\devenv.exe`
	command, err := applicationCommand(path)
	if err != nil {
		t.Fatal(err)
	}
	arguments, err := windows.DecomposeCommandLine(command)
	if err != nil {
		t.Fatal(err)
	}
	if len(arguments) != 1 || arguments[0] != path {
		t.Fatalf("invalid startup command: %q", command)
	}
	for _, path := range []string{`devenv.exe`, `D:\code\build\bin\devenv-dev.exe`} {
		if _, err := applicationCommand(path); err == nil {
			t.Fatalf("accepted unsupported startup path %q", path)
		}
	}
}
