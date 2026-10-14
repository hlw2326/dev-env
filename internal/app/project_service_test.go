package app_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"devenv/internal/app"
	"devenv/internal/domain/project"
	"devenv/internal/infra/events"
	"devenv/internal/infra/sqlite"
	"devenv/internal/infra/sqlite/repositories"
)

func TestProjectService(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "devenv_proj_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "devenv.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	publisher := events.NewPublisher()
	projRepo := repositories.NewProjectRepository(db.DB)
	runtimeRepo := repositories.NewRuntimeRepository(db.DB)
	projSvc := app.NewProjectService(projRepo, runtimeRepo, publisher)

	// Test validate path
	val := projSvc.ValidateProjectPath(ctx, filepath.Join(tempDir, "new-project"))
	if !val.Valid {
		t.Fatalf("expected valid new project path, got %v", val.Errors)
	}

	// Test create project
	targetPath := filepath.Join(tempDir, "my-app")
	created, err := projSvc.CreateProject(ctx, project.CreateInput{
		Name: "my-app",
		Path: targetPath,
	})
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}
	if created.ID <= 0 {
		t.Fatalf("expected autoincrement ID > 0, got %d", created.ID)
	}

	// Test list projects
	list, err := projSvc.ListProjects(ctx)
	if err != nil {
		t.Fatalf("failed to list projects: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 project, got %d", len(list))
	}
	if list[0].Name != "my-app" {
		t.Fatalf("expected name my-app, got %s", list[0].Name)
	}
}
