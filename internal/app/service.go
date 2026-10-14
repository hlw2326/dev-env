package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"devenv/internal/domain/runtime"
	"devenv/internal/domain/service"
	"devenv/internal/ports"
	appErrors "devenv/internal/shared/errors"
)

type servicePreset struct {
	name      string
	port      int
	autoStart bool
}

var defaultServicePresets = map[runtime.RuntimeType]servicePreset{
	runtime.TypePHP:   {name: "PHP FastCGI", port: 9000, autoStart: true},
	runtime.TypeNginx: {name: "Nginx", port: 80, autoStart: true},
	runtime.TypeMySQL: {name: "MySQL", port: 3306, autoStart: false},
	runtime.TypeRedis: {name: "Redis", port: 6379, autoStart: true},
}

type ServiceService struct {
	instances ports.ServiceRepository
	runtimes  ports.RuntimeRepository
	proc      ports.ProcessManager
	probe     ports.PortChecker
	bus       ports.EventPublisher
}

func NewServiceService(
	instances ports.ServiceRepository,
	runtimes ports.RuntimeRepository,
	proc ports.ProcessManager,
	probe ports.PortChecker,
	bus ports.EventPublisher,
) *ServiceService {
	return &ServiceService{
		instances: instances,
		runtimes:  runtimes,
		proc:      proc,
		probe:     probe,
		bus:       bus,
	}
}

func (s *ServiceService) List(ctx context.Context) ([]service.Instance, error) {
	return s.instances.ListAll(ctx)
}

func (s *ServiceService) Get(ctx context.Context, id int64) (*service.Instance, error) {
	return s.instances.GetByID(ctx, id)
}

func (s *ServiceService) GetRuntime(ctx context.Context, runtimeID int64) (*runtime.RuntimeVersion, error) {
	return s.runtimes.GetByID(ctx, runtimeID)
}

func (s *ServiceService) Status(ctx context.Context, id int64) (service.StatusInfo, error) {
	if _, err := s.instances.GetByID(ctx, id); err != nil {
		return service.StatusInfo{}, err
	}
	return s.proc.GetStatus(ctx, id), nil
}

func (s *ServiceService) EnsureForRuntime(ctx context.Context, rt runtime.RuntimeVersion) (*service.Instance, bool, error) {
	preset, ok := defaultServicePresets[rt.RuntimeType]
	if !ok {
		return nil, false, nil
	}

	existing, err := s.instances.ListByType(ctx, rt.RuntimeType)
	if err != nil {
		return nil, false, err
	}
	if len(existing) > 0 {
		return &existing[0], false, nil
	}

	port := preset.port
	for i := 0; i < 50; i++ {
		if s.probe.IsPortAvailable(port) {
			break
		}
		port++
	}

	now := time.Now().UTC()
	inst := service.Instance{
		Name:        preset.name,
		ServiceType: rt.RuntimeType,
		RuntimeID:   rt.ID,
		Port:        &port,
		AutoStart:   false,
		Enabled:     true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	id, err := s.instances.Create(ctx, inst)
	if err != nil {
		return nil, false, err
	}
	inst.ID = id
	s.bus.Emit("service:created", id)

	if preset.autoStart {
		_ = s.Start(ctx, id)
	}

	return &inst, true, nil
}

func (s *ServiceService) SwitchRuntime(ctx context.Context, id, runtimeID int64) error {
	cur, err := s.instances.GetByID(ctx, id)
	if err != nil {
		return err
	}
	target, err := s.runtimes.GetByID(ctx, runtimeID)
	if err != nil {
		return err
	}

	if cur.ServiceType != target.RuntimeType {
		return appErrors.New(appErrors.CodeInvalidArgument, "目标版本类型与当前服务实例不匹配")
	}

	st := s.proc.GetStatus(ctx, id)
	if st.Status == service.StatusRunning || st.Status == service.StatusStarting {
		return appErrors.New(appErrors.CodeServiceRunning, "服务实例正在运行中，请先停止服务再切换版本")
	}

	if target.Status != runtime.StatusInstalled && target.Status != runtime.StatusAvailable {
		return appErrors.New(appErrors.CodeInvalidArgument, "目标运行环境版本尚未就绪")
	}

	oldRt, _ := s.runtimes.GetByID(ctx, cur.RuntimeID)
	if oldRt != nil && cur.ConfigPath != nil && *cur.ConfigPath != "" {
		if strings.HasPrefix(strings.ToLower(*cur.ConfigPath), strings.ToLower(oldRt.InstallPath)) {
			cur.ConfigPath = nil
		}
	}

	cur.RuntimeID = target.ID
	if preset, ok := defaultServicePresets[cur.ServiceType]; ok {
		cur.Name = preset.name
	}
	cur.UpdatedAt = time.Now().UTC()

	if err := s.instances.Update(ctx, *cur); err != nil {
		return err
	}

	s.bus.Emit("service:updated", id)
	return nil
}

func (s *ServiceService) RuntimeOptions(ctx context.Context, id int64) ([]runtime.RuntimeVersion, error) {
	cur, err := s.instances.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.runtimes.ListByType(ctx, cur.ServiceType)
}

func (s *ServiceService) IsRuntimeRunning(ctx context.Context, runtimeID int64) (bool, error) {
	all, err := s.instances.ListAll(ctx)
	if err != nil {
		return false, err
	}
	for _, srv := range all {
		if srv.RuntimeID == runtimeID {
			st := s.proc.GetStatus(ctx, srv.ID)
			if st.Status == service.StatusRunning || st.Status == service.StatusStarting {
				return true, nil
			}
		}
	}
	return false, nil
}

func (s *ServiceService) OnRuntimeRemoved(ctx context.Context, rt runtime.RuntimeVersion) error {
	list, err := s.instances.ListByType(ctx, rt.RuntimeType)
	if err != nil {
		return err
	}
	for _, srv := range list {
		if srv.RuntimeID == rt.ID {
			st := s.proc.GetStatus(ctx, srv.ID)
			if st.Status == service.StatusRunning {
				_ = s.proc.Stop(ctx, srv.ID)
			}
			if err := s.instances.Delete(ctx, srv.ID); err != nil {
				return err
			}
			s.bus.Emit("service:deleted", srv.ID)
		}
	}
	return nil
}

func (s *ServiceService) Reconcile(ctx context.Context) error {
	rts, err := s.runtimes.ListAll(ctx)
	if err != nil {
		return err
	}
	seen := make(map[runtime.RuntimeType]bool)
	for _, rt := range rts {
		if seen[rt.RuntimeType] {
			continue
		}
		seen[rt.RuntimeType] = true
		if _, _, err := s.EnsureForRuntime(ctx, rt); err != nil {
			return err
		}
	}
	return nil
}

func (s *ServiceService) Start(ctx context.Context, id int64) error {
	cur, err := s.instances.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !cur.Enabled {
		return appErrors.New(appErrors.CodeInvalidArgument, "服务已禁用")
	}

	st := s.proc.GetStatus(ctx, id)
	if st.Status == service.StatusRunning {
		return nil
	}

	if cur.Port != nil && cur.ServiceType != runtime.TypeNginx && !s.probe.IsPortAvailable(*cur.Port) {
		return appErrors.New(appErrors.CodePortAlreadyInUse, fmt.Sprintf("端口 %d 已被占用", *cur.Port))
	}

	rt, err := s.runtimes.GetByID(ctx, cur.RuntimeID)
	if err != nil {
		return err
	}

	spec, err := buildSpec(cur, rt)
	if err != nil {
		return err
	}

	if _, err := s.proc.Start(ctx, id, spec); err != nil {
		return err
	}

	s.bus.Emit("service:started", id)
	s.bus.Emit("service:status-changed", id)
	return nil
}

func (s *ServiceService) Stop(ctx context.Context, id int64) error {
	if _, err := s.instances.GetByID(ctx, id); err != nil {
		return err
	}
	if err := s.proc.Stop(ctx, id); err != nil {
		return err
	}
	s.bus.Emit("service:stopped", id)
	s.bus.Emit("service:status-changed", id)
	return nil
}

func (s *ServiceService) Restart(ctx context.Context, id int64) error {
	if err := s.Stop(ctx, id); err != nil {
		return err
	}
	return s.Start(ctx, id)
}

func (s *ServiceService) Logs(ctx context.Context, id int64, limit int) []string {
	return s.proc.GetLogs(ctx, id, limit)
}

func (s *ServiceService) Delete(ctx context.Context, id int64) error {
	if _, err := s.instances.GetByID(ctx, id); err != nil {
		return err
	}
	st := s.proc.GetStatus(ctx, id)
	if st.Status == service.StatusRunning {
		_ = s.proc.Stop(ctx, id)
	}
	if err := s.instances.Delete(ctx, id); err != nil {
		return err
	}
	s.bus.Emit("service:deleted", id)
	return nil
}

func buildSpec(cur *service.Instance, rt *runtime.RuntimeVersion) (ports.ProcessSpec, error) {
	port := 9000
	if cur.Port != nil {
		port = *cur.Port
	}

	switch cur.ServiceType {
	case runtime.TypePHP:
		exe := filepath.Join(rt.InstallPath, "php-cgi.exe")
		args := []string{"-b", fmt.Sprintf("127.0.0.1:%d", port)}
		if cur.ConfigPath != nil && *cur.ConfigPath != "" {
			args = append(args, "-c", *cur.ConfigPath)
		}
		return ports.ProcessSpec{
			Executable: exe,
			Args:       args,
			WorkingDir: rt.InstallPath,
		}, nil

	case runtime.TypeNginx:
		exe := filepath.Join(rt.InstallPath, "nginx.exe")
		args := make([]string, 0)
		if cur.ConfigPath != nil && *cur.ConfigPath != "" {
			args = append(args, "-c", *cur.ConfigPath)
		}
		return ports.ProcessSpec{
			Executable: exe,
			Args:       args,
			WorkingDir: rt.InstallPath,
		}, nil

	case runtime.TypeMySQL:
		exe := filepath.Join(rt.InstallPath, "bin", "mysqld.exe")
		args := []string{fmt.Sprintf("--port=%d", port)}
		if cur.ConfigPath != nil && *cur.ConfigPath != "" {
			args = append([]string{fmt.Sprintf("--defaults-file=%s", *cur.ConfigPath)}, args...)
		}
		return ports.ProcessSpec{
			Executable: exe,
			Args:       args,
			WorkingDir: rt.InstallPath,
		}, nil

	case runtime.TypeRedis:
		exe := filepath.Join(rt.InstallPath, "redis-server.exe")
		args := make([]string, 0)
		if cur.ConfigPath != nil && *cur.ConfigPath != "" {
			args = append(args, *cur.ConfigPath)
		}
		args = append(args, "--port", fmt.Sprintf("%d", port))
		return ports.ProcessSpec{
			Executable: exe,
			Args:       args,
			WorkingDir: rt.InstallPath,
		}, nil

	default:
		return ports.ProcessSpec{}, appErrors.New(appErrors.CodeInvalidArgument, "不支持的服务类型")
	}
}
