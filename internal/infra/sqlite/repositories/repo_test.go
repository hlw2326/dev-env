package repositories_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"devenv/internal/domain/host"
	"devenv/internal/domain/project"
	"devenv/internal/domain/runtime"
	"devenv/internal/domain/service"
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

	// 1. Test system_runtime_versions (autoincrement ID)
	runtimeRepo := repositories.NewRuntimeRepository(db.DB)
	now := time.Now().UTC()
	rtID, err := runtimeRepo.Create(ctx, runtime.RuntimeVersion{
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

	rt, err := runtimeRepo.GetByID(ctx, rtID)
	if err != nil || rt == nil {
		t.Fatalf("failed to get runtime version by ID %d: %v", rtID, err)
	}
	if rt.Version != "8.2.9" {
		t.Fatalf("expected version 8.2.9, got %s", rt.Version)
	}

	// 2. Test system_projects (autoincrement ID, integer foreign key)
	projectRepo := repositories.NewProjectRepository(db.DB)
	projID, err := projectRepo.Create(ctx, project.Project{
		Name:         "test-app",
		Path:         `G:\www\test-app`,
		PathKey:      `g:\www\test-app`,
		PHPVersionID: &rtID,
		Status:       project.StatusReady,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}
	if projID <= 0 {
		t.Fatalf("expected positive integer ID, got %d", projID)
	}

	proj, err := projectRepo.GetByID(ctx, projID)
	if err != nil || proj == nil {
		t.Fatalf("failed to get project by ID %d: %v", projID, err)
	}
	if proj.PHPVersionID == nil || *proj.PHPVersionID != rtID {
		t.Fatalf("expected PHPVersionID %d, got %v", rtID, proj.PHPVersionID)
	}

	// 3. Test system_service_instances
	serviceRepo := repositories.NewServiceRepository(db.DB)
	port := 9000
	srvID, err := serviceRepo.Create(ctx, service.Instance{
		Name:             "PHP FastCGI",
		RuntimeVersionID: rtID,
		ServiceType:      runtime.TypePHP,
		Port:             &port,
		AutoStart:        true,
		Enabled:          true,
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}
	if srvID <= 0 {
		t.Fatalf("expected positive integer ID for service, got %d", srvID)
	}

	srv, err := serviceRepo.GetByID(ctx, srvID)
	if err != nil || srv == nil {
		t.Fatalf("failed to get service: %v", err)
	}
	if *srv.Port != 9000 {
		t.Fatalf("expected port 9000, got %v", srv.Port)
	}

	// 4. Test system_virtual_hosts
	hostRepo := repositories.NewHostRepository(db.DB)
	hostID, err := hostRepo.Create(ctx, host.VirtualHost{
		ProjectID:     projID,
		Domain:        "test-app.local",
		DocumentRoot:  `G:\www\test-app\public`,
		WebServerType: "nginx",
		Port:          80,
		Enabled:       true,
		CreatedAt:     now,
		UpdatedAt:     now,
	})
	if err != nil {
		t.Fatalf("failed to create virtual host: %v", err)
	}
	if hostID <= 0 {
		t.Fatalf("expected positive integer ID for host, got %d", hostID)
	}

	// 5. Test system_tasks
	taskRepo := repositories.NewTaskRepository(db.DB)
	taskID, err := taskRepo.Create(ctx, task.Task{
		TaskType:     "install",
		Status:       task.StatusRunning,
		Progress:     50,
		Message:      "Downloading PHP 8.3...",
		PayloadJSON:  "{}",
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}
	if taskID <= 0 {
		t.Fatalf("expected positive integer ID for task, got %d", taskID)
	}
}
