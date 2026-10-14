package app

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"devenv/internal/domain/runtime"
	"devenv/internal/domain/service"
	"devenv/internal/ports"
	appErrors "devenv/internal/shared/errors"
)

type ServiceAppService struct {
	repo         ports.ServiceRepository
	runtimeRepo  ports.RuntimeRepository
	procMgr      ports.ProcessManager
	portChecker  ports.PortChecker
	publisher    ports.EventPublisher
}

func NewServiceAppService(
	repo ports.ServiceRepository,
	runtimeRepo ports.RuntimeRepository,
	procMgr ports.ProcessManager,
	portChecker ports.PortChecker,
	publisher ports.EventPublisher,
) *ServiceAppService {
	return &ServiceAppService{
		repo:        repo,
		runtimeRepo: runtimeRepo,
		procMgr:     procMgr,
		portChecker: portChecker,
		publisher:   publisher,
	}
}

func (s *ServiceAppService) ListServices(ctx context.Context) ([]service.Instance, error) {
	list, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (s *ServiceAppService) GetService(ctx context.Context, id int64) (*service.Instance, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ServiceAppService) GetServiceStatus(ctx context.Context, id int64) (service.StatusInfo, error) {
	return s.procMgr.GetStatus(ctx, id), nil
}

func (s *ServiceAppService) StartService(ctx context.Context, id int64) error {
	inst, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	currentStatus := s.procMgr.GetStatus(ctx, id)
	if currentStatus.Status == service.StatusRunning {
		return nil // already running, idempotent
	}

	// Check port conflict
	if inst.Port != nil {
		if !s.portChecker.IsPortAvailable(*inst.Port) {
			return appErrors.New(appErrors.CodePortAlreadyInUse, fmt.Sprintf("端口 %d 已被其他程序占用", *inst.Port))
		}
	}

	// Fetch runtime version info
	rt, err := s.runtimeRepo.GetByID(ctx, inst.RuntimeVersionID)
	if err != nil || rt == nil {
		return appErrors.New(appErrors.CodeRuntimeNotInstalled, "关联的运行环境未安装")
	}

	// Build process spec
	spec := ports.ProcessSpec{
		WorkingDir: rt.InstallPath,
	}

	switch inst.ServiceType {
	case runtime.TypePHP:
		spec.Executable = filepath.Join(rt.InstallPath, "php-cgi.exe")
		port := 9000
		if inst.Port != nil {
			port = *inst.Port
		}
		spec.Args = []string{"-b", fmt.Sprintf("127.0.0.1:%d", port)}
	case runtime.TypeNginx:
		spec.Executable = filepath.Join(rt.InstallPath, "nginx.exe")
	case runtime.TypeMySQL:
		spec.Executable = filepath.Join(rt.InstallPath, "bin", "mysqld.exe")
		if inst.ConfigPath != nil {
			spec.Args = []string{"--defaults-file=" + *inst.ConfigPath}
		}
	default:
		return appErrors.New(appErrors.CodeInvalidArgument, fmt.Sprintf("暂不支持启动该类型的服务: %s", inst.ServiceType))
	}

	_, err = s.procMgr.Start(ctx, id, spec)
	if err != nil {
		return err
	}

	if s.publisher != nil {
		s.publisher.Emit("service:status-changed", map[string]interface{}{
			"serviceId": id,
			"status":    service.StatusRunning,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}

	return nil
}

func (s *ServiceAppService) StopService(ctx context.Context, id int64) error {
	err := s.procMgr.Stop(ctx, id)
	if err != nil {
		return err
	}

	if s.publisher != nil {
		s.publisher.Emit("service:status-changed", map[string]interface{}{
			"serviceId": id,
			"status":    service.StatusStopped,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}
	return nil
}

func (s *ServiceAppService) RestartService(ctx context.Context, id int64) error {
	_ = s.StopService(ctx, id)
	time.Sleep(300 * time.Millisecond)
	return s.StartService(ctx, id)
}

func (s *ServiceAppService) GetServiceLogs(ctx context.Context, id int64, limit int) []string {
	return s.procMgr.GetLogs(ctx, id, limit)
}
