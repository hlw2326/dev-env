package app

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"devenv/internal/domain/software"
	"devenv/internal/infra/sqlite"
	"devenv/internal/shared/paths"
)

type scanInspector struct {
	active  atomic.Int32
	maximum atomic.Int32
	failAll bool
}

func (inspector *scanInspector) Inspect(ctx context.Context, kind, executable string) (software.EnvironmentScan, error) {
	active := inspector.active.Add(1)
	defer inspector.active.Add(-1)
	for {
		previous := inspector.maximum.Load()
		if active <= previous || inspector.maximum.CompareAndSwap(previous, active) {
			break
		}
	}
	select {
	case <-time.After(time.Millisecond):
	case <-ctx.Done():
		return software.EnvironmentScan{}, ctx.Err()
	}
	if inspector.failAll || kind == software.KIND_MYSQL {
		return software.EnvironmentScan{}, errors.New("registry unavailable")
	}
	return software.EnvironmentScan{Entries: []software.CommandLocation{}, Warnings: []string{}}, nil
}

func TestScanEnvironmentsIsolatesFailuresAndBoundsConcurrency(t *testing.T) {
	base := t.TempDir()
	db, err := sqlite.Open(paths.Database(base), sqlite.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	inspector := &scanInspector{}
	service := &SoftwareService{base: base, repo: sqlite.NewSoftwareRepository(db.DB), commands: inspector}
	items, err := service.ScanEnvironments(context.Background(), "", "")
	if err != nil || len(items) != len(software.AllKinds()) {
		t.Fatalf("scan result: %v %v", items, err)
	}
	for _, item := range items {
		if item.Kind == software.KIND_MYSQL {
			if !strings.Contains(item.Error, "registry unavailable") {
				t.Fatalf("lost real error: %+v", item)
			}
		} else if item.Error != "" {
			t.Fatalf("failure leaked to other kind: %+v", item)
		}
	}
	if inspector.maximum.Load() > ENVIRONMENT_SCAN_WORKERS {
		t.Fatalf("unbounded concurrency: %d", inspector.maximum.Load())
	}
	inspector.failAll = true
	items, err = service.ScanEnvironments(context.Background(), "", "")
	if err == nil || len(items) != 0 || strings.Count(err.Error(), "registry unavailable") != 1 {
		t.Fatalf("common error repeated: %+v %v", items, err)
	}
	if _, err := service.ScanEnvironments(context.Background(), "", `C:\tools\node.exe`); err == nil {
		t.Fatal("accepted executable without kind")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := service.ScanEnvironments(ctx, "", ""); err == nil {
		t.Fatal("cancelled scan succeeded")
	}
}
