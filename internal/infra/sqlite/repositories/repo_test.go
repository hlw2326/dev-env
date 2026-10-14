package repositories_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"devenv/internal/domain/runtime"
	"devenv/internal/domain/service"
	"devenv/internal/domain/site"
	"devenv/internal/domain/task"
	"devenv/internal/infra/sqlite"
	"devenv/internal/infra/sqlite/repositories"
)

func TestSystemTablesAndIntegerIDs(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "devenv_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test_devenv.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// 1. Test system_runtime_versions
	runtimes := repositories.NewRuntimeRepository(db.DB)
	now := time.Now().UTC()
	rtID, err := runtimes.Create(ctx, runtime.RuntimeVersion{
		RuntimeType:  runtime.TypePHP,
		Version:      "8.2.9",
		Architecture: "x64",
		InstallPath:  `D:\phpstudy_pro\Extensions\php\php8.2.9nts`,
		Status:       runtime.StatusInstalled,
		MetadataJSON: "{}",
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		t.Fatalf("failed to create runtime version: %v", err)
	}
	if rtID <= 0 {
		t.Fatalf("expected positive integer autoincrement ID, got %d", rtID)
	}

	rt, err := runtimes.GetByID(ctx, rtID)
	if err != nil || rt == nil {
		t.Fatalf("failed to get runtime version by ID %d: %v", rtID, err)
	}
	if rt.Version != "8.2.9" {
		t.Fatalf("expected version 8.2.9, got %s", rt.Version)
	}

	// 2. Test system_service_instances
	instances := repositories.NewServiceRepository(db.DB)
	port := 9000
	srvID, err := instances.Create(ctx, service.Instance{
		Name:        "PHP FastCGI",
		RuntimeID:   rtID,
		ServiceType: runtime.TypePHP,
		Port:        &port,
		AutoStart:   true,
		Enabled:     true,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}
	if srvID <= 0 {
		t.Fatalf("expected positive integer ID for service, got %d", srvID)
	}

	srv, err := instances.GetByID(ctx, srvID)
	if err != nil || srv == nil {
		t.Fatalf("failed to get service: %v", err)
	}
	if *srv.Port != 9000 {
		t.Fatalf("expected port 9000, got %v", srv.Port)
	}

	// 3. Test system_sites
	sites := repositories.NewSiteRepository(db.DB)
	curSite := &site.Site{
		Name:      "test-site",
		Kind:      site.PHP,
		Domains:   []string{"test.local"},
		Root:      `G:\www\test-site`,
		RuntimeID: &rtID,
		ServiceID: &srvID,
		Port:      80,
		Proxy:     false,
		Upstream:  &port,
		Meta: site.Meta{
			Entry:   "index.php",
			History: make([]site.Snapshot, 0),
			Env:     make(map[string]string),
		},
		Status: site.Stopped,
	}
	if err := sites.Create(ctx, curSite); err != nil {
		t.Fatalf("failed to create site: %v", err)
	}
	if curSite.ID <= 0 {
		t.Fatalf("expected positive integer ID for site, got %d", curSite.ID)
	}

	gotSite, err := sites.Get(ctx, curSite.ID)
	if err != nil || gotSite == nil {
		t.Fatalf("failed to get site by ID %d: %v", curSite.ID, err)
	}
	if gotSite.Name != "test-site" || gotSite.Kind != site.PHP {
		t.Fatalf("unexpected site: %+v", gotSite)
	}

	// 4. Test system_tasks
	tasks := repositories.NewTaskRepository(db.DB)
	taskID, err := tasks.Create(ctx, task.Task{
		TaskType:    "install",
		Status:      task.StatusRunning,
		Progress:    50,
		Message:     "Downloading PHP 8.3...",
		PayloadJSON: "{}",
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}
	if taskID <= 0 {
		t.Fatalf("expected positive integer ID for task, got %d", taskID)
	}
}
