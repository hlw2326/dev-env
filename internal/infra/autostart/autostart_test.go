package autostart

import (
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
