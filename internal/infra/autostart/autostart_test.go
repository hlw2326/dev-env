package autostart

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAutostart(t *testing.T) {
	orig := Enabled()
	defer func() {
		_ = Set(orig)
	}()

	if err := Set(true); err != nil {
		t.Fatalf("Set(true) failed: %v", err)
	}
	if !Enabled() {
		t.Errorf("expected Enabled() to be true")
	}

	if err := Set(false); err != nil {
		t.Fatalf("Set(false) failed: %v", err)
	}
	if Enabled() {
		t.Errorf("expected Enabled() to be false")
	}
}

func TestServiceAutostart(t *testing.T) {
	svcName := "test-nginx-service"
	defer func() {
		_ = SetService(svcName, false, "", nil, "")
	}()

	tmp := t.TempDir()
	bin := filepath.Join(tmp, "nginx.exe")
	_ = os.WriteFile(bin, []byte("mock"), 0755)

	if err := SetService(svcName, true, bin, []string{"-p", tmp}, tmp); err != nil {
		t.Fatalf("SetService failed: %v", err)
	}
	if !ServiceEnabled(svcName) {
		t.Errorf("expected ServiceEnabled() to be true")
	}

	if err := SetService(svcName, false, "", nil, ""); err != nil {
		t.Fatalf("SetService(false) failed: %v", err)
	}
	if ServiceEnabled(svcName) {
		t.Errorf("expected ServiceEnabled() to be false")
	}
}
