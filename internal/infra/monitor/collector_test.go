package monitor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCollectorTraffic(t *testing.T) {
	path := filepath.Join(t.TempDir(), "access.log")
	content := "2026-10-15T12:00:00+08:00\t200\t1024\t0.010\n" +
		"2026-10-15T12:00:01+08:00\t500\t256\t0.030\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write access log: %v", err)
	}

	traffic := NewCollector().Traffic(path)
	if traffic.RequestCount != 2 {
		t.Fatalf("request count = %d, want 2", traffic.RequestCount)
	}
	if traffic.ResponseBytes != 1280 {
		t.Fatalf("response bytes = %d, want 1280", traffic.ResponseBytes)
	}
	if traffic.ErrorCount != 1 {
		t.Fatalf("error count = %d, want 1", traffic.ErrorCount)
	}
	if traffic.AverageResponseMs != 20 {
		t.Fatalf("average response = %v, want 20", traffic.AverageResponseMs)
	}
	if traffic.LastAccess != "2026-10-15T12:00:01+08:00" {
		t.Fatalf("last access = %q", traffic.LastAccess)
	}
}
