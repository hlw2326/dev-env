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
	baseDir string
	repo    ports.SettingRepository
	sites   ports.SiteRepository
	gateway ports.Gateway
	hosts   ports.HostsWriter

	mu           sync.RWMutex
	gatewayError string
}

func NewSettingService(
	baseDir string,
	repo ports.SettingRepository,
	sites ports.SiteRepository,
	gateway ports.Gateway,
	hosts ports.HostsWriter,
) *SettingService {
	return &SettingService{
		baseDir: baseDir,
		repo:    repo,
		sites:   sites,
		gateway: gateway,
		hosts:   hosts,
	}
}

func (s *SettingService) StartGateway(ctx context.Context) {
	cfg, err := s.repo.Load(ctx)
	if err != nil {
		s.setGatewayError(err.Error())
		return
	}
	if cfg.Root != "" {
		_ = os.MkdirAll(cfg.Root, 0755)
	}
	if err := s.gateway.Start(cfg.Port); err != nil {
		s.setGatewayError(err.Error())
		return
	}
	s.setGatewayError("")
}

func (s *SettingService) setGatewayError(msg string) {
	s.mu.Lock()
	s.gatewayError = msg
	s.mu.Unlock()
}

func (s *SettingService) Settings(ctx context.Context) (SettingsView, error) {
	cfg, err := s.repo.Load(ctx)
	if err != nil {
		return SettingsView{}, err
	}
	s.mu.RLock()
	gwErr := s.gatewayError
	s.mu.RUnlock()

	return SettingsView{
		Settings:  cfg,
		Listening: s.gateway.Port() > 0 && gwErr == "",
		Error:     gwErr,
	}, nil
}

func (s *SettingService) SaveSettings(ctx context.Context, next setting.Settings) error {
	if next.Theme != "light" && next.Theme != "dark" && next.Theme != "system" {
		return appErrors.New(appErrors.CodeInvalidArgument, "主题只能是 light、dark 或 system")
	}
	next.Root = filepath.Clean(strings.TrimSpace(next.Root))
	if !filepath.IsAbs(next.Root) {
		return appErrors.New(appErrors.CodePathInvalid, "站点根目录必须是绝对路径")
	}
	if next.Port <= 0 || next.Port > 65535 {
		return appErrors.New(appErrors.CodeInvalidArgument, "网关端口必须在 1 到 65535 之间")
	}
	if err := os.MkdirAll(next.Root, 0755); err != nil {
		return appErrors.Wrap(err, appErrors.CodePathNotAccessible, "创建站点根目录失败")
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
			s.setGatewayError(err.Error())
			return err
		}
		s.setGatewayError("")
	}

	if next.Hosts {
		list, err := s.sites.List(ctx)
		if err != nil {
			return err
		}
		var allHosts []string
		for _, st := range list {
			allHosts = append(allHosts, st.Hosts...)
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
	return s.baseDir
}

func (s *SettingService) OpenFolder(path string) error {
	clean := filepath.Clean(strings.TrimSpace(path))
	if clean == "" {
		return appErrors.New(appErrors.CodePathInvalid, "目录路径不能为空")
	}
	info, err := os.Stat(clean)
	if err != nil || !info.IsDir() {
		return appErrors.New(appErrors.CodePathNotAccessible, "目录不存在: "+clean)
	}
	cmd := exec.Command("explorer.exe", clean)
	return cmd.Start()
}
