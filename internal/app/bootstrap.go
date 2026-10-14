package app

import (
	"context"
	"os"
	"path/filepath"

	"devenv/internal/infra/events"
	"devenv/internal/infra/git"
	"devenv/internal/infra/network"
	"devenv/internal/infra/nginx"
	"devenv/internal/infra/process"
	"devenv/internal/infra/sqlite"
	"devenv/internal/infra/sqlite/repositories"
)

type Container struct {
	DB      *sqlite.DB
	Bus     *events.Publisher
	Site    *SiteService
	Git     *GitService
	Runtime *RuntimeService
	Service *ServiceService
	Task    *TaskService
	Setting *SettingService
}

func Bootstrap() (*Container, error) {
	dataDir := os.Getenv("LOCALAPPDATA")
	if dataDir == "" {
		dataDir = "."
	}
	baseDir := filepath.Join(dataDir, "DevEnv")
	dbPath := filepath.Join(baseDir, "data", "devenv.db")

	db, err := sqlite.Open(dbPath)
	if err != nil {
		return nil, err
	}

	bus := events.NewPublisher()
	probe := network.NewChecker()
	proc := process.NewManager()
	gitMgr := git.NewManager()
	nginxMgr := nginx.NewManager(baseDir)

	sites := repositories.NewSiteRepository(db.DB)
	runtimes := repositories.NewRuntimeRepository(db.DB)
	instances := repositories.NewServiceRepository(db.DB)
	tasks := repositories.NewTaskRepository(db.DB)
	settings := repositories.NewSettingsRepository(db.DB)

	serviceSvc := NewServiceService(instances, runtimes, proc, probe, bus)
	runtimeSvc := NewRuntimeService(runtimes, sites, instances, settings, serviceSvc, bus)
	taskSvc := NewTaskService(tasks, bus)
	settingSvc := NewSettingService(settings)
	siteSvc := NewSiteService(sites, runtimes, instances, serviceSvc, nginxMgr, gitMgr, probe, bus)
	gitSvc := NewGitService(sites, gitMgr, bus)

	ctx := context.Background()
	if err := runtimeSvc.Scan(ctx); err != nil {
		return nil, err
	}
	if err := serviceSvc.Reconcile(ctx); err != nil {
		return nil, err
	}

	return &Container{
		DB:      db,
		Bus:     bus,
		Site:    siteSvc,
		Git:     gitSvc,
		Runtime: runtimeSvc,
		Service: serviceSvc,
		Task:    taskSvc,
		Setting: settingSvc,
	}, nil
}
