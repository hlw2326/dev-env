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
	"devenv/internal/shared/paths"
)

type SettingsView struct {
	setting.Settings
	Listening bool
	Error     string
}

type SettingService struct {
	base     string
	repo     ports.SettingRepository
	projects ports.ProjectRepository
	nginx    ports.Gateway
	hosts    ports.HostsWriter

	mu        sync.RWMutex
	lastError string
}

func NewSettingService(
	base string,
	repo ports.SettingRepository,
	projects ports.ProjectRepository,
	nginx ports.Gateway,
	hosts ports.HostsWriter,
) *SettingService {
	return &SettingService{
		base:     base,
		repo:     repo,
		projects: projects,
		nginx:    nginx,
		hosts:    hosts,
	}
}

func (s *SettingService) StartNginx(ctx context.Context) {
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
	if err := s.nginx.Start(current.Port); err != nil {
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
	listening := s.nginx.Port() > 0
	if listening {
		errText = ""
	}

	return SettingsView{
		Settings:  current,
		Listening: listening,
		Error:     errText,
	}, nil
}

func (s *SettingService) SaveSettings(ctx context.Context, next setting.Settings) error {
	if next.Theme != "light" && next.Theme != "dark" && next.Theme != "system" {
		return appErrors.New(appErrors.INVALID_ARGUMENT, "主题只能是 light、dark 或 system")
	}
	next.Root = filepath.Clean(strings.TrimSpace(next.Root))
	if !filepath.IsAbs(next.Root) {
		return appErrors.New(appErrors.PATH_INVALID, "项目根目录必须是绝对路径")
	}
	if next.Port <= 0 || next.Port > ports.MAX_PORT {
		return appErrors.New(appErrors.INVALID_ARGUMENT, "Nginx 监听端口必须在 1 到 65535 之间")
	}
	next.DeveloperToolsCLI = strings.TrimSpace(next.DeveloperToolsCLI)
	if next.DeveloperToolsCLI != "" {
		if !filepath.IsAbs(next.DeveloperToolsCLI) {
			return appErrors.New(appErrors.PATH_INVALID, "微信开发者工具 CLI 必须是绝对路径")
		}
		info, statErr := os.Stat(next.DeveloperToolsCLI)
		if statErr != nil || info.IsDir() {
			return appErrors.New(appErrors.PATH_NOT_ACCESSIBLE, "微信开发者工具 CLI 文件不存在或不可用")
		}
	}
	if err := os.MkdirAll(next.Root, 0o755); err != nil {
		return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "创建项目根目录失败")
	}

	prev, err := s.repo.Load(ctx)
	if err != nil {
		return err
	}

	if err := s.repo.Save(ctx, next); err != nil {
		return err
	}

	if prev.Port != next.Port || s.nginx.Port() == 0 {
		if err := s.nginx.Start(next.Port); err != nil {
			s.setError(err.Error())
			return err
		}
		s.setError("")
	}

	if next.Hosts {
		list, err := s.projects.List(ctx)
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

func (s *SettingService) OpenNginxConfig() error {
	configPath := paths.NginxConfig(s.base)
	info, err := os.Stat(configPath)
	if err != nil {
		return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "Nginx 配置文件尚未生成")
	}
	if info.IsDir() {
		return appErrors.New(appErrors.PATH_NOT_ACCESSIBLE, "Nginx 配置文件路径指向了目录")
	}
	command := exec.Command("explorer.exe", configPath)
	return command.Start()
}
