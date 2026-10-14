package app

import (
	"context"
	"os"
	"path/filepath"

	"devenv/internal/domain/runtime"
	"devenv/internal/domain/service"
	"devenv/internal/infra/events"
	"devenv/internal/infra/network"
	"devenv/internal/infra/process"
	"devenv/internal/infra/sqlite"
	"devenv/internal/infra/sqlite/repositories"
)

type Container struct {
	DB          *sqlite.DB
	Publisher   *events.Publisher
	ProjectSvc  *ProjectService
	PHPSvc      *PHPService
	ServiceSvc  *ServiceAppService
	HostSvc     *HostService
	TaskSvc     *TaskService
	SettingsSvc *SettingsService
}

func Bootstrap() (*Container, error) {
	// Determine database path
	dataDir := os.Getenv("LOCALAPPDATA")
	if dataDir == "" {
		dataDir = "."
	}
	dbPath := filepath.Join(dataDir, "DevEnv", "data", "devenv.db")

	db, err := sqlite.Open(dbPath)
	if err != nil {
		return nil, err
	}

	publisher := events.NewPublisher()
	portChecker := network.NewChecker()
	procMgr := process.NewManager()

	projectRepo := repositories.NewProjectRepository(db.DB)
	runtimeRepo := repositories.NewRuntimeRepository(db.DB)
	serviceRepo := repositories.NewServiceRepository(db.DB)
	hostRepo := repositories.NewHostRepository(db.DB)
	taskRepo := repositories.NewTaskRepository(db.DB)
	settingsRepo := repositories.NewSettingsRepository(db.DB)

	projectSvc := NewProjectService(projectRepo, runtimeRepo, publisher)
	phpSvc := NewPHPService(runtimeRepo, settingsRepo, publisher)
	serviceSvc := NewServiceAppService(serviceRepo, runtimeRepo, procMgr, portChecker, publisher)
	hostSvc := NewHostService(hostRepo, projectRepo, publisher)
	taskSvc := NewTaskService(taskRepo, publisher)
	settingsSvc := NewSettingsService(settingsRepo)

	// Auto scan local PHP if none registered
	ctx := context.Background()
	_ = phpSvc.ScanAndRegisterCommonPaths(ctx)

	// Seed sample services if runtime exists and services empty
	seedServicesIfEmpty(ctx, runtimeRepo, serviceRepo)

	return &Container{
		DB:          db,
		Publisher:   publisher,
		ProjectSvc:  projectSvc,
		PHPSvc:      phpSvc,
		ServiceSvc:  serviceSvc,
		HostSvc:     hostSvc,
		TaskSvc:     taskSvc,
		SettingsSvc: settingsSvc,
	}, nil
}

func seedServicesIfEmpty(ctx context.Context, runtimeRepo *repositories.RuntimeRepository, serviceRepo *repositories.ServiceRepository) {
	services, _ := serviceRepo.ListAll(ctx)
	if len(services) > 0 {
		return
	}

	runtimes, _ := runtimeRepo.ListAll(ctx)
	if len(runtimes) == 0 {
		return
	}

	// If there's an installed PHP runtime, create a default PHP-FPM service
	for _, rt := range runtimes {
		if rt.RuntimeType == runtime.TypePHP {
			port := 9000
			_, _ = serviceRepo.Create(ctx, service.Instance{
				Name:             "PHP FastCGI (" + rt.Version + ")",
				ServiceType:      runtime.TypePHP,
				RuntimeVersionID: rt.ID,
				Port:             &port,
				AutoStart:        false,
				Enabled:          true,
			})
			break
		}
	}
}
