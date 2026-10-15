package app

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"devenv/internal/domain/setting"
	"devenv/internal/ports"
	appErrors "devenv/internal/shared/errors"
)

type SettingsView struct {
	setting.Settings
	Listening bool
	Error     string
}

type SettingService struct {
	base    string
	repo    ports.SettingRepository
	sites   ports.SiteRepository
	gateway ports.Gateway
	hosts   ports.HostsWriter

	mu        sync.RWMutex
	lastError string
}

func NewSettingService(
	base string,
	repo ports.SettingRepository,
	sites ports.SiteRepository,
	gateway ports.Gateway,
	hosts ports.HostsWriter,
) *SettingService {
	return &SettingService{
		base:    base,
		repo:    repo,
		sites:   sites,
		gateway: gateway,
		hosts:   hosts,
	}
}

func (s *SettingService) StartGateway(ctx context.Context) {
	current, err := s.repo.Load(ctx)
	if err != nil {
		s.setError(err.Error())
		return
	}
	if current.Root != "" {
		if err := os.MkdirAll(current.Root, 0o755); err != nil {
			s.setError(err.Error())
			return
		}
	}
	if err := s.gateway.Start(current.Port); err != nil {
		s.setError(err.Error())
		return
	}
	s.setError("")
}

func (s *SettingService) setError(message string) {
	s.mu.Lock()
	s.lastError = message
	s.mu.Unlock()
}

func (s *SettingService) Settings(ctx context.Context) (SettingsView, error) {
	current, err := s.repo.Load(ctx)
	if err != nil {
		return SettingsView{}, err
	}
	s.mu.RLock()
	errText := s.lastError
	s.mu.RUnlock()

	return SettingsView{
		Settings:  current,
		Listening: s.gateway.Port() > 0 && errText == "",
		Error:     errText,
	}, nil
}

func (s *SettingService) SaveSettings(ctx context.Context, next setting.Settings) error {
	if next.Theme != "light" && next.Theme != "dark" && next.Theme != "system" {
		return appErrors.New(appErrors.INVALID_ARGUMENT, "主题只能是 light、dark 或 system")
	}
	next.Root = filepath.Clean(strings.TrimSpace(next.Root))
	if !filepath.IsAbs(next.Root) {
		return appErrors.New(appErrors.PATH_INVALID, "站点根目录必须是绝对路径")
	}
	if next.Port <= 0 || next.Port > ports.MAX_PORT {
		return appErrors.New(appErrors.INVALID_ARGUMENT, "网关端口必须在 1 到 65535 之间")
	}
	if err := os.MkdirAll(next.Root, 0o755); err != nil {
		return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "创建站点根目录失败")
	}

	prev, err := s.repo.Load(ctx)
	if err != nil {
		return err
	}

	if err := s.repo.Save(ctx, next); err != nil {
		return err
	}

	if prev.Port != next.Port || s.gateway.Port() == 0 {
		if err := s.gateway.Start(next.Port); err != nil {
			s.setError(err.Error())
			return err
		}
		s.setError("")
	}

	if next.Hosts {
		list, err := s.sites.List(ctx)
		if err != nil {
			return err
		}
		var allHosts []string
		for _, item := range list {
			allHosts = append(allHosts, item.Hosts...)
		}
		if err := s.hosts.Sync(allHosts); err != nil {
			return err
		}
	} else if prev.Hosts && !next.Hosts {
		if err := s.hosts.Sync(nil); err != nil {
			return err
		}
	}

	return nil
}

func (s *SettingService) DataFolder() string {
	return s.base
}

func (s *SettingService) OpenFolder(path string) error {
	clean := filepath.Clean(strings.TrimSpace(path))
	if clean == "" {
		return appErrors.New(appErrors.PATH_INVALID, "目录路径不能为空")
	}
	info, err := os.Stat(clean)
	if err != nil || !info.IsDir() {
		return appErrors.New(appErrors.PATH_NOT_ACCESSIBLE, "目录不存在: "+clean)
	}
	command := exec.Command("explorer.exe", clean)
	return command.Start()
}
