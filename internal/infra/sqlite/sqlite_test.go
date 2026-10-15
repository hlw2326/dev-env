package sqlite

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"devenv/internal/domain/project"
	"devenv/internal/domain/software"
	appErrors "devenv/internal/shared/errors"
)

func TestRepositories(t *testing.T) {
	folder := t.TempDir()
	dbPath := filepath.Join(folder, "data", "devenv.db")

	db, err := Open(dbPath, Options{})
	if err != nil {
		t.Fatalf("first Open failed: %v", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Log(closeErr)
		}
	}()

	second, err := Open(dbPath, Options{})
	if err != nil {
		t.Fatalf("second Open failed: %v", err)
	}
	if err := second.Close(); err != nil {
		t.Fatalf("close second db: %v", err)
	}

	ctx := context.Background()
	softwareRepo := NewSoftwareRepository(db.DB)
	projectRepo := NewProjectRepository(db.DB)
	settingRepo := NewSettingRepository(db.DB)

	settings, err := settingRepo.Load(ctx)
	if err != nil {
		t.Fatalf("load settings: %v", err)
	}
	if settings.Theme != "light" || settings.Port != 80 || settings.Hosts != false {
		t.Fatalf("unexpected default settings: %+v", settings)
	}

	php := &software.Software{
		Kind:    software.KIND_PHP,
		Version: "8.3.35",
		Path:    `C:\DevEnv\software\php\8.3.35`,
		State:   software.STATE_INSTALLED,
		Main:    true,
	}
	if err := softwareRepo.Create(ctx, php); err != nil {
		t.Fatalf("create software: %v", err)
	}

	firstProject := &project.Project{
		Name:      "blog",
		Kind:      project.KIND_PHP,
		Hosts:     []string{"blog.localhost"},
		Root:      `C:\www\blog`,
		Run:       "public",
		Software:  &php.ID,
		Autostart: true,
	}
	if err := projectRepo.Create(ctx, firstProject); err != nil {
		t.Fatalf("create project: %v", err)
	}

	secondProject := &project.Project{
		Name:      "blog2",
		Kind:      project.KIND_STATIC,
		Hosts:     []string{"blog.localhost"},
		Root:      `C:\www\blog2`,
		Autostart: true,
	}
	if err := projectRepo.Create(ctx, secondProject); err == nil {
		t.Fatalf("expected duplicate domain error")
	}

	if err := softwareRepo.Delete(ctx, php.ID); err == nil {
		t.Fatalf("expected ON DELETE RESTRICT error when deleting bound software")
	}

	if err := projectRepo.Delete(ctx, firstProject.ID); err != nil {
		t.Fatalf("delete project: %v", err)
	}
	if err := softwareRepo.Delete(ctx, php.ID); err != nil {
		t.Fatalf("delete software after project deletion: %v", err)
	}
}

func TestMigrationDrift(t *testing.T) {
	folder := t.TempDir()
	dbPath := filepath.Join(folder, "data", "devenv.db")

	initialFS := fstest.MapFS{
		"000001_init.sql": &fstest.MapFile{
			Data: []byte(`CREATE TABLE system_test (id INTEGER PRIMARY KEY, name TEXT NOT NULL);`),
		},
	}

	db, err := Open(dbPath, Options{Migrations: initialFS})
	if err != nil {
		t.Fatalf("initial open: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close initial db: %v", err)
	}

	modifiedFS := fstest.MapFS{
		"000001_init.sql": &fstest.MapFile{
			Data: []byte(`CREATE TABLE system_test (id INTEGER PRIMARY KEY, title TEXT NOT NULL);`),
		},
	}

	_, err = Open(dbPath, Options{Migrations: modifiedFS, ResetOnDrift: false})
	if err == nil {
		t.Fatalf("expected drift error when checksum changes")
	}
	var appErr *appErrors.AppError
	if !errors.As(err, &appErr) || appErr.Code != appErrors.DATABASE_SCHEMA_MISMATCH {
		t.Fatalf("expected DATABASE_SCHEMA_MISMATCH, got: %v", err)
	}

	resetDB, err := Open(dbPath, Options{Migrations: modifiedFS, ResetOnDrift: true})
	if err != nil {
		t.Fatalf("open with ResetOnDrift=true failed: %v", err)
	}
	if err := resetDB.Close(); err != nil {
		t.Fatalf("close reset db: %v", err)
	}

	backups, err := os.ReadDir(filepath.Join(folder, "data", "backups"))
	if err != nil || len(backups) != 1 {
		t.Fatalf("expected 1 backup file after reset, got %d (err=%v)", len(backups), err)
	}
}
