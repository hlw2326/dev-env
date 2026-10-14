package sqlite

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"devenv/internal/domain/site"
	"devenv/internal/domain/software"
	appErrors "devenv/internal/shared/errors"
)

func TestMigrateAndRepositories(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "data", "devenv.db")

	db, err := Open(dbPath, Options{})
	if err != nil {
		t.Fatalf("first Open failed: %v", err)
	}
	defer db.Close()

	// Reopening same DB should succeed without drift
	db2, err := Open(dbPath, Options{})
	if err != nil {
		t.Fatalf("second Open failed: %v", err)
	}
	_ = db2.Close()

	ctx := context.Background()
	softRepo := NewSoftwareRepository(db.DB)
	siteRepo := NewSiteRepository(db.DB)
	settingRepo := NewSettingRepository(db.DB)

	// Verify seed settings
	settings, err := settingRepo.Load(ctx)
	if err != nil {
		t.Fatalf("load settings: %v", err)
	}
	if settings.Theme != "light" || settings.Port != 80 || settings.Hosts != false {
		t.Fatalf("unexpected default settings: %+v", settings)
	}

	// Software CRUD
	php := &software.Software{
		Kind:    software.KindPHP,
		Version: "8.3.35",
		Path:    `C:\DevEnv\software\php\8.3.35`,
		State:   software.StateInstalled,
		Main:    true,
	}
	if err := softRepo.Create(ctx, php); err != nil {
		t.Fatalf("create software: %v", err)
	}

	// Site CRUD + unique domain + FK RESTRICT
	s1 := &site.Site{
		Name:      "blog",
		Kind:      site.KindPHP,
		Hosts:     []string{"blog.localhost"},
		Root:      `C:\www\blog`,
		Run:       "public",
		Software:  &php.ID,
		Autostart: true,
	}
	if err := siteRepo.Create(ctx, s1); err != nil {
		t.Fatalf("create site: %v", err)
	}

	// Duplicate domain must fail
	s2 := &site.Site{
		Name:      "blog2",
		Kind:      site.KindStatic,
		Hosts:     []string{"blog.localhost"},
		Root:      `C:\www\blog2`,
		Autostart: true,
	}
	if err := siteRepo.Create(ctx, s2); err == nil {
		t.Fatalf("expected duplicate domain error")
	}

	// Deleting software while referenced by site must fail due to ON DELETE RESTRICT
	if err := softRepo.Delete(ctx, php.ID); err == nil {
		t.Fatalf("expected ON DELETE RESTRICT error when deleting bound software")
	}

	// Deleting site cascades domains and allows deleting software
	if err := siteRepo.Delete(ctx, s1.ID); err != nil {
		t.Fatalf("delete site: %v", err)
	}
	if err := softRepo.Delete(ctx, php.ID); err != nil {
		t.Fatalf("delete software after site deletion: %v", err)
	}
}

func TestMigrationDriftAndReset(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "data", "devenv.db")

	fs1 := fstest.MapFS{
		"000001_init.sql": &fstest.MapFile{
			Data: []byte(`CREATE TABLE system_test (id INTEGER PRIMARY KEY, name TEXT NOT NULL);`),
		},
	}

	db, err := Open(dbPath, Options{Migrations: fs1})
	if err != nil {
		t.Fatalf("initial open: %v", err)
	}
	_ = db.Close()

	// Modify checksum of 000001_init.sql -> should trigger SchemaDriftError
	fsModified := fstest.MapFS{
		"000001_init.sql": &fstest.MapFile{
			Data: []byte(`CREATE TABLE system_test (id INTEGER PRIMARY KEY, title TEXT NOT NULL);`),
		},
	}

	_, err = Open(dbPath, Options{Migrations: fsModified, ResetOnDrift: false})
	if err == nil {
		t.Fatalf("expected drift error when checksum changes")
	}
	var appErr *appErrors.AppError
	if !os.IsExist(err) {
		appErr, _ = err.(*appErrors.AppError)
		if appErr == nil || appErr.Code != appErrors.CodeDatabaseSchemaMismatch {
			t.Fatalf("expected CodeDatabaseSchemaMismatch, got: %v", err)
		}
	}

	// ResetOnDrift=true should backup and rebuild
	dbReset, err := Open(dbPath, Options{Migrations: fsModified, ResetOnDrift: true})
	if err != nil {
		t.Fatalf("open with ResetOnDrift=true failed: %v", err)
	}
	_ = dbReset.Close()

	backups, err := os.ReadDir(filepath.Join(tmp, "data", "backups"))
	if err != nil || len(backups) != 1 {
		t.Fatalf("expected 1 backup file after reset, got %d (err=%v)", len(backups), err)
	}
}
