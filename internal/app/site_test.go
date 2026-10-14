package app_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"devenv/internal/app"
	"devenv/internal/domain/site"
	"devenv/internal/infra/git"
	"devenv/internal/infra/network"
	"devenv/internal/infra/nginx"
	"devenv/internal/infra/process"
	"devenv/internal/infra/sqlite"
	"devenv/internal/infra/sqlite/repositories"
)

type noopBus struct{}

func (n *noopBus) Emit(name string, data interface{}) {}

func TestSiteServiceLifecycle(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "devenv_site_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	defer db.Close()

	sites := repositories.NewSiteRepository(db.DB)
	runtimes := repositories.NewRuntimeRepository(db.DB)
	instances := repositories.NewServiceRepository(db.DB)
	proc := process.NewManager()
	probe := network.NewChecker()
	bus := &noopBus{}
	nginxMgr := nginx.NewManager(tempDir)
	gitMgr := git.NewManager()

	serviceSvc := app.NewServiceService(instances, runtimes, proc, probe, bus)
	siteSvc := app.NewSiteService(sites, runtimes, instances, serviceSvc, nginxMgr, gitMgr, probe, bus)

	ctx := context.Background()
	siteRoot := filepath.Join(tempDir, "www", "demo")

	created, err := siteSvc.Create(ctx, site.CreateInput{
		Name:    "Demo Site",
		Kind:    site.Static,
		Domains: []string{"demo.local"},
		Root:    siteRoot,
		Port:    8080,
		Entry:   "index.html",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.ID <= 0 {
		t.Fatalf("expected positive site ID, got %d", created.ID)
	}
	if len(created.Meta.History) != 1 {
		t.Fatalf("expected 1 snapshot in Meta.History, got %d", len(created.Meta.History))
	}

	if err := siteSvc.Start(ctx, created.ID); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	started, err := siteSvc.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if started.Status != site.Running {
		t.Fatalf("expected status running, got %s", started.Status)
	}

	if err := siteSvc.Delete(ctx, created.ID, true); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}
