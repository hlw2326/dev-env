package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"devenv/internal/domain/runtime"
	"devenv/internal/ports"
	appErrors "devenv/internal/shared/errors"
)

type PHPService struct {
	runtimeRepo ports.RuntimeRepository
	settingsRepo ports.SettingsRepository
	publisher   ports.EventPublisher
}

func NewPHPService(runtimeRepo ports.RuntimeRepository, settingsRepo ports.SettingsRepository, publisher ports.EventPublisher) *PHPService {
	return &PHPService{
		runtimeRepo: runtimeRepo,
		settingsRepo: settingsRepo,
		publisher:   publisher,
	}
}

func (s *PHPService) ListVersions(ctx context.Context) ([]runtime.PHPVersionView, error) {
	list, err := s.runtimeRepo.ListByType(ctx, runtime.TypePHP)
	if err != nil {
		return nil, err
	}

	// Auto scan if empty
	if len(list) == 0 {
		_ = s.ScanAndRegisterCommonPaths(ctx)
		list, _ = s.runtimeRepo.ListByType(ctx, runtime.TypePHP)
	}

	var defaultID int64
	st, _ := s.settingsRepo.Get(ctx)
	if st != nil && st.DefaultProjectPath != "" {
		// can also check default php id from settings or first installed
	}

	result := make([]runtime.PHPVersionView, 0, len(list))
	for _, v := range list {
		// check executable exists
		status := v.Status
		phpExe := filepath.Join(v.InstallPath, "php.exe")
		if _, err := os.Stat(phpExe); err != nil {
			status = runtime.StatusFailed
		}

		result = append(result, runtime.PHPVersionView{
			ID:           v.ID,
			Version:      v.Version,
			Architecture: v.Architecture,
			InstallPath:  v.InstallPath,
			Status:       status,
			IsDefault:    v.ID == defaultID || (defaultID == 0 && len(result) == 0),
		})
	}
	return result, nil
}

func (s *PHPService) RegisterVersion(ctx context.Context, version, architecture, installPath string) (*runtime.RuntimeVersion, error) {
	installPath = filepath.Clean(strings.TrimSpace(installPath))
	phpExe := filepath.Join(installPath, "php.exe")
	if _, err := os.Stat(phpExe); err != nil {
		return nil, appErrors.New(appErrors.CodePathInvalid, fmt.Sprintf("在目录 %s 中未找到 php.exe", installPath))
	}

	// Detect actual version using php -v
	cmd := exec.Command(phpExe, "-v")
	out, err := cmd.Output()
	if err == nil {
		firstLine := strings.Split(string(out), "\n")[0]
		fields := strings.Fields(firstLine)
		if len(fields) >= 2 {
			version = fields[1]
		}
	}

	if architecture == "" {
		architecture = "x64"
	}

	now := time.Now().UTC()
	item := runtime.RuntimeVersion{
		RuntimeType:  runtime.TypePHP,
		Version:      version,
		Architecture: architecture,
		InstallPath:  installPath,
		Status:       runtime.StatusInstalled,
		MetadataJSON: "{}",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	id, err := s.runtimeRepo.Create(ctx, item)
	if err != nil {
		return nil, err
	}
	item.ID = id

	if s.publisher != nil {
		s.publisher.Emit("runtime:installed", item)
	}

	return &item, nil
}

func (s *PHPService) SetDefaultVersion(ctx context.Context, versionID int64) error {
	v, err := s.runtimeRepo.GetByID(ctx, versionID)
	if err != nil || v == nil {
		return appErrors.New(appErrors.CodeNotFound, "指定的 PHP 版本不存在")
	}

	phpExe := filepath.Join(v.InstallPath, "php.exe")
	if _, err := os.Stat(phpExe); err != nil {
		return appErrors.New(appErrors.CodePathNotAccessible, "PHP 可执行文件不存在，无法设置为默认版本")
	}

	return nil
}

func (s *PHPService) GetExtensions(ctx context.Context, versionID int64) ([]runtime.PHPExtension, error) {
	v, err := s.runtimeRepo.GetByID(ctx, versionID)
	if err != nil || v == nil {
		return nil, appErrors.New(appErrors.CodeNotFound, "指定的 PHP 版本不存在")
	}

	phpExe := filepath.Join(v.InstallPath, "php.exe")
	if _, err := os.Stat(phpExe); err != nil {
		return nil, appErrors.New(appErrors.CodePathNotAccessible, "PHP 可执行文件不存在")
	}

	// Run php -m to get active extensions
	cmd := exec.Command(phpExe, "-m")
	out, _ := cmd.Output()
	activeMap := make(map[string]bool)
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		ext := strings.TrimSpace(scanner.Text())
		if ext != "" && !strings.HasPrefix(ext, "[") {
			activeMap[strings.ToLower(ext)] = true
		}
	}

	// Read ext directory
	extDir := filepath.Join(v.InstallPath, "ext")
	entries, _ := os.ReadDir(extDir)
	var extensions []runtime.PHPExtension
	seen := make(map[string]bool)

	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, "php_") && strings.HasSuffix(name, ".dll") {
			extName := strings.TrimSuffix(strings.TrimPrefix(name, "php_"), ".dll")
			lower := strings.ToLower(extName)
			seen[lower] = true
			extensions = append(extensions, runtime.PHPExtension{
				Name:      extName,
				Enabled:   activeMap[lower],
				Available: true,
			})
		}
	}

	// Also add core extensions from activeMap that don't have separate dlls
	for act := range activeMap {
		if !seen[act] {
			extensions = append(extensions, runtime.PHPExtension{
				Name:      act,
				Enabled:   true,
				Available: true,
			})
		}
	}

	return extensions, nil
}

func (s *PHPService) ScanAndRegisterCommonPaths(ctx context.Context) error {
	commonPaths := []string{
		`D:\phpstudy_pro\Extensions\php\php8.2.9nts`,
		`C:\php`,
		`C:\tools\php`,
		`D:\php`,
	}

	// Also check PATH
	pathEnv := os.Getenv("PATH")
	for _, p := range filepath.SplitList(pathEnv) {
		if strings.Contains(strings.ToLower(p), "php") {
			commonPaths = append(commonPaths, p)
		}
	}

	for _, path := range commonPaths {
		phpExe := filepath.Join(path, "php.exe")
		if _, err := os.Stat(phpExe); err == nil {
			_, _ = s.RegisterVersion(ctx, "AutoDetected", "x64", path)
		}
	}
	return nil
}
