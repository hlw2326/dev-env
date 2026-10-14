package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"devenv/internal/domain/software"
	"devenv/internal/infra/installer"
	"devenv/internal/infra/process"
	"devenv/internal/ports"
	appErrors "devenv/internal/shared/errors"
)

type SoftwareView struct {
	software.Software
	Running bool
	PID     int
}

type SoftwareService struct {
	baseDir   string
	repo      ports.SoftwareRepository
	sites     ports.SiteRepository
	catalog   ports.Catalog
	installer ports.Installer
	runner    ports.Runner
	probe     ports.PortChecker
	bus       ports.EventPublisher
	onInstalled func(ctx context.Context, sw software.Software)

	mu      sync.Mutex
	cancels map[string]context.CancelFunc
}

func NewSoftwareService(
	baseDir string,
	repo ports.SoftwareRepository,
	sites ports.SiteRepository,
	catalog ports.Catalog,
	inst ports.Installer,
	runner ports.Runner,
	probe ports.PortChecker,
	bus ports.EventPublisher,
) *SoftwareService {
	return &SoftwareService{
		baseDir:   baseDir,
		repo:      repo,
		sites:     sites,
		catalog:   catalog,
		installer: inst,
		runner:    runner,
		probe:     probe,
		bus:       bus,
		cancels:   make(map[string]context.CancelFunc),
	}
}

func (s *SoftwareService) SetOnInstalled(fn func(ctx context.Context, sw software.Software)) {
	s.onInstalled = fn
}

func (s *SoftwareService) ReconcileStartup(ctx context.Context) {
	list, err := s.repo.List(ctx)
	if err != nil {
		return
	}
	for _, item := range list {
		if item.State == software.StateInstalling {
			item.State = software.StateFailed
			item.Progress = 0
			item.Error = "安装被中断"
			_ = s.repo.Update(ctx, &item)
			_ = os.Remove(s.cachePath(item.Kind, item.Version))
			_ = os.RemoveAll(item.Path + ".tmp")
		}
	}

	s.discoverDiskSoftware(ctx)

	updatedList, err := s.repo.List(ctx)
	if err != nil {
		return
	}
	for _, item := range updatedList {
		if item.State == software.StateInstalled && software.IsDaemon(item.Kind) {
			logPath := s.softwareLogPath(item.Kind, item.Version)
			if s.probe.Listening(item.Port) {
				pid, exePath, _ := s.probe.FindProcess(item.Port)
				if pid > 0 && isPathMatch(exePath, item.Path) {
					s.runner.Adopt(daemonKey(item.ID), pid, logPath)
					s.bus.Emit("service:status-changed", map[string]interface{}{
						"id":      item.ID,
						"running": true,
						"pid":     pid,
						"message": "",
					})
					continue
				}
			}
			if item.Autostart {
				_ = s.Start(ctx, item.ID)
			}
		}
	}
}

func (s *SoftwareService) discoverDiskSoftware(ctx context.Context) {
	softRoot := filepath.Join(s.baseDir, "software")
	for _, kind := range []string{software.KindNginx, software.KindApache, software.KindPHP, software.KindNode, software.KindMySQL, software.KindRedis} {
		kindDir := filepath.Join(softRoot, kind)
		entries, err := os.ReadDir(kindDir)
		if err != nil {
			continue
		}
		for _, ent := range entries {
			if !ent.IsDir() || strings.HasSuffix(ent.Name(), ".tmp") {
				continue
			}
			version := ent.Name()
			existing, _ := s.repo.GetByKindVersion(ctx, kind, version)
			if existing != nil {
				continue
			}
			dirPath := filepath.Join(kindDir, version)
			if !hasEntryBinary(kind, dirPath) {
				continue
			}
			mainSw, _ := s.repo.GetMain(ctx, kind)
			port := 0
			if software.IsDaemon(kind) || kind == software.KindPHP {
				port = s.allocateDaemonPort(ctx, kind)
			}
			sw := &software.Software{
				Kind:     kind,
				Version:  version,
				Path:     dirPath,
				State:    software.StateInstalled,
				Progress: 100,
				Main:     mainSw == nil,
				Port:     port,
			}
			if kind == software.KindPHP {
				s.fixPHPIniPath(*sw)
			}
			if err := s.repo.Create(ctx, sw); err == nil && s.onInstalled != nil {
				s.onInstalled(ctx, *sw)
			}
		}
	}
}

func (s *SoftwareService) fixPHPIniPath(record software.Software) {
	iniPath := filepath.Join(record.Path, "php.ini")
	raw, err := os.ReadFile(iniPath)
	if err != nil {
		return
	}
	extDir := filepath.Join(record.Path, "ext")
	re := regexp.MustCompile(`(?m)^extension_dir\s*=\s*".*"$`)
	text := re.ReplaceAllString(string(raw), fmt.Sprintf(`extension_dir = "%s"`, extDir))
	_ = os.WriteFile(iniPath, []byte(text), 0644)
	s.configurePHPErrorLog(record)
}

func hasEntryBinary(kind, dir string) bool {
	switch kind {
	case software.KindPHP:
		_, err := os.Stat(filepath.Join(dir, "php-cgi.exe"))
		return err == nil
	case software.KindNode:
		_, err := os.Stat(filepath.Join(dir, "node.exe"))
		return err == nil
	case software.KindMySQL:
		_, err := os.Stat(filepath.Join(dir, "bin", "mysqld.exe"))
		return err == nil
	case software.KindRedis:
		return installer.FindRedisBinary(dir) != ""
	case software.KindNginx:
		_, err := os.Stat(filepath.Join(dir, "nginx.exe"))
		return err == nil
	case software.KindApache:
		return installer.FindApacheBinary(dir) != ""
	default:
		return false
	}
}

func (s *SoftwareService) Market(ctx context.Context) ([]software.Item, error) {
	installed, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int)
	for _, item := range installed {
		if item.State == software.StateInstalled {
			counts[item.Kind]++
		}
	}

	defs := []software.Item{
		{
			Kind:        software.KindNginx,
			Name:        "Nginx",
			Description: "高性能 HTTP 与反向代理 Web 服务器，占用内存少且并发能力强",
			Category:    "Web服务",
			Count:       counts[software.KindNginx],
		},
		{
			Kind:        software.KindApache,
			Name:        "Apache",
			Description: "经典高效且模块丰富的开源 HTTP Web 服务器，支持多种扩展与配置",
			Category:    "Web服务",
			Count:       counts[software.KindApache],
		},
		{
			Kind:        software.KindPHP,
			Name:        "PHP",
			Description: "广泛使用的通用开源脚本语言，适用于 Web 开发并内置 CGI 支持",
			Category:    "运行环境",
			Count:       counts[software.KindPHP],
		},
		{
			Kind:        software.KindNode,
			Name:        "Node.js",
			Description: "基于 Chrome V8 引擎的 JavaScript 运行环境，包含 npm 包管理器",
			Category:    "运行环境",
			Count:       counts[software.KindNode],
		},
		{
			Kind:        software.KindMySQL,
			Name:        "MySQL",
			Description: "主流开源关系型数据库管理系统，支持事务与高性能存储",
			Category:    "数据库",
			Count:       counts[software.KindMySQL],
		},
		{
			Kind:        software.KindRedis,
			Name:        "Redis",
			Description: "高性能开源内存数据结构存储，常用作数据库、缓存和消息代理",
			Category:    "缓存",
			Count:       counts[software.KindRedis],
		},
	}

	shortCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	for i := range defs {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			vers, err := s.catalog.Versions(shortCtx, defs[idx].Kind)
			if err != nil || len(vers) == 0 {
				return
			}
			for _, v := range vers {
				if v.Recommended {
					defs[idx].Recommended = v.Version
					return
				}
			}
			defs[idx].Recommended = vers[0].Version
		}(i)
	}
	wg.Wait()

	return defs, nil
}

func (s *SoftwareService) List(ctx context.Context) ([]SoftwareView, error) {
	list, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	views := make([]SoftwareView, 0, len(list))
	for _, item := range list {
		v := SoftwareView{Software: item}
		if software.IsDaemon(item.Kind) {
			st := s.runner.Status(daemonKey(item.ID))
			v.Running = st.Running
			v.PID = st.PID
		}
		views = append(views, v)
	}
	return views, nil
}

func (s *SoftwareService) Versions(ctx context.Context, kind string) ([]software.Version, error) {
	if !software.IsValidKind(kind) {
		return nil, appErrors.New(appErrors.CodeInvalidArgument, "不支持的软件类型: "+kind)
	}
	remote, err := s.catalog.Versions(ctx, kind)
	if err != nil {
		return nil, err
	}
	local, err := s.repo.ListByKind(ctx, kind)
	if err != nil {
		return nil, err
	}
	installedMap := make(map[string]bool, len(local))
	for _, item := range local {
		if item.State == software.StateInstalled {
			installedMap[item.Version] = true
		}
	}
	for i := range remote {
		remote[i].Installed = installedMap[remote[i].Version]
	}
	return remote, nil
}

func (s *SoftwareService) Install(ctx context.Context, kind, version string) error {
	kind = strings.TrimSpace(kind)
	version = strings.TrimSpace(version)
	if !software.IsValidKind(kind) || version == "" {
		return appErrors.New(appErrors.CodeInvalidArgument, "软件类型或版本号不能为空")
	}

	taskKey := kind + ":" + version
	s.mu.Lock()
	if _, active := s.cancels[taskKey]; active {
		s.mu.Unlock()
		return appErrors.New(appErrors.CodeAlreadyExists, "该版本正在安装中")
	}
	s.mu.Unlock()

	existing, err := s.repo.GetByKindVersion(ctx, kind, version)
	if err != nil {
		return err
	}
	if existing != nil {
		if existing.State == software.StateInstalled {
			return appErrors.New(appErrors.CodeAlreadyExists, fmt.Sprintf("%s %s 已安装", kind, version))
		}
		_ = os.RemoveAll(existing.Path)
		_ = os.RemoveAll(existing.Path + ".tmp")
		_ = os.Remove(s.cachePath(kind, version))
		if err := s.repo.Delete(ctx, existing.ID); err != nil {
			return err
		}
	}

	versions, err := s.catalog.Versions(ctx, kind)
	if err != nil {
		return err
	}
	var matched *software.Version
	for i := range versions {
		if versions[i].Version == version {
			matched = &versions[i]
			break
		}
	}
	if matched == nil {
		return appErrors.New(appErrors.CodeNotFound, fmt.Sprintf("版本清单中未找到 %s %s", kind, version))
	}

	targetDir := filepath.Join(s.baseDir, "software", kind, version)
	record := &software.Software{
		Kind:     kind,
		Version:  version,
		Path:     targetDir,
		State:    software.StateInstalling,
		Progress: 0,
	}
	if err := s.repo.Create(ctx, record); err != nil {
		return err
	}

	installCtx, cancel := context.WithCancel(context.Background())
	s.mu.Lock()
	s.cancels[taskKey] = cancel
	s.mu.Unlock()

	go s.runInstall(installCtx, taskKey, *record, *matched)
	return nil
}

func (s *SoftwareService) runInstall(ctx context.Context, taskKey string, record software.Software, ver software.Version) {
	defer func() {
		s.mu.Lock()
		delete(s.cancels, taskKey)
		s.mu.Unlock()
	}()

	req := ports.Request{
		Kind:     record.Kind,
		Version:  record.Version,
		URL:      ver.URL,
		Checksum: ver.Checksum,
		Target:   record.Path,
		Cache:    s.cachePath(record.Kind, record.Version),
	}

	report := func(progress int, message string) {
		record.Progress = progress
		_ = s.repo.Update(context.Background(), &record)
		s.bus.Emit("task:progress", map[string]interface{}{
			"kind":     record.Kind,
			"version":  record.Version,
			"progress": progress,
			"message":  message,
		})
	}

	err := s.installer.Install(ctx, req, report)
	bgCtx := context.Background()

	if err != nil {
		record.State = software.StateFailed
		record.Progress = 0
		record.Error = err.Error()
		_ = s.repo.Update(bgCtx, &record)
		s.bus.Emit("task:completed", map[string]interface{}{
			"kind":    record.Kind,
			"version": record.Version,
			"ok":      false,
			"error":   err.Error(),
		})
		return
	}

	if record.Kind == software.KindPHP {
		s.configurePHPErrorLog(record)
	}

	mainSw, _ := s.repo.GetMain(bgCtx, record.Kind)
	if mainSw == nil {
		record.Main = true
	}
	if software.IsDaemon(record.Kind) || record.Kind == software.KindPHP {
		record.Port = s.allocateDaemonPort(bgCtx, record.Kind)
	}
	record.State = software.StateInstalled
	record.Progress = 100
	record.Error = ""
	_ = s.repo.Update(bgCtx, &record)

	if s.onInstalled != nil {
		s.onInstalled(bgCtx, record)
	}

	s.bus.Emit("task:completed", map[string]interface{}{
		"kind":    record.Kind,
		"version": record.Version,
		"ok":      true,
		"error":   "",
	})
}

func (s *SoftwareService) configurePHPErrorLog(record software.Software) {
	iniPath := filepath.Join(record.Path, "php.ini")
	raw, err := os.ReadFile(iniPath)
	if err != nil {
		return
	}
	logFile := s.softwareLogPath(record.Kind, record.Version)
	_ = os.MkdirAll(filepath.Dir(logFile), 0755)
	text := string(raw)
	text = strings.ReplaceAll(text, ";error_log = php_errors.log", fmt.Sprintf(`error_log = "%s"`, logFile))
	_ = os.WriteFile(iniPath, []byte(text), 0644)
}

func (s *SoftwareService) Cancel(ctx context.Context, kind, version string) error {
	taskKey := strings.TrimSpace(kind) + ":" + strings.TrimSpace(version)
	s.mu.Lock()
	cancel, ok := s.cancels[taskKey]
	s.mu.Unlock()
	if !ok {
		return appErrors.New(appErrors.CodeTaskNotCancellable, "未找到正在运行的安装任务")
	}
	cancel()
	return nil
}

func (s *SoftwareService) Remove(ctx context.Context, id int64) error {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if item == nil {
		return appErrors.New(appErrors.CodeNotFound, "软件版本不存在")
	}

	if software.IsDaemon(item.Kind) {
		st := s.runner.Status(daemonKey(item.ID))
		if st.Running {
			return appErrors.New(appErrors.CodeServiceRunning, fmt.Sprintf("%s %s 正在运行中，请先停止后再卸载", item.Kind, item.Version))
		}
	}

	boundSites, err := s.sites.ListBySoftware(ctx, item.ID)
	if err != nil {
		return err
	}
	if len(boundSites) > 0 {
		names := make([]string, 0, len(boundSites))
		for _, st := range boundSites {
			names = append(names, st.Name)
		}
		return appErrors.New(
			appErrors.CodeRuntimeInUse,
			fmt.Sprintf("该版本正被以下网站使用，请先修改网站版本后再卸载: %s", strings.Join(names, "、")),
		)
	}

	if item.Path != "" {
		if err := os.RemoveAll(item.Path); err != nil {
			return appErrors.Wrap(err, appErrors.CodePathNotAccessible, "删除软件安装目录失败")
		}
	}

	wasMain := item.Main
	kind := item.Kind
	if err := s.repo.Delete(ctx, item.ID); err != nil {
		return err
	}

	if wasMain {
		remaining, err := s.repo.ListByKind(ctx, kind)
		if err == nil {
			var candidates []software.Software
			for _, rem := range remaining {
				if rem.State == software.StateInstalled {
					candidates = append(candidates, rem)
				}
			}
			if len(candidates) > 0 {
				sort.Slice(candidates, func(i, j int) bool {
					return candidates[i].Version > candidates[j].Version
				})
				_ = s.repo.SetMain(ctx, kind, candidates[0].ID)
			}
		}
	}
	return nil
}

func (s *SoftwareService) SetMain(ctx context.Context, id int64) error {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if item == nil {
		return appErrors.New(appErrors.CodeNotFound, "软件版本不存在")
	}
	if item.State != software.StateInstalled {
		return appErrors.New(appErrors.CodeInvalidArgument, "只有已安装完成的版本才能设为默认")
	}
	return s.repo.SetMain(ctx, item.Kind, item.ID)
}

func (s *SoftwareService) Start(ctx context.Context, id int64) error {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if item == nil {
		return appErrors.New(appErrors.CodeNotFound, "软件版本不存在")
	}
	if !software.IsDaemon(item.Kind) {
		return appErrors.New(appErrors.CodeInvalidArgument, "该软件不是常驻服务")
	}
	if item.State != software.StateInstalled {
		return appErrors.New(appErrors.CodeInvalidArgument, "软件尚未安装完成")
	}

	key := daemonKey(item.ID)
	if st := s.runner.Status(key); st.Running {
		return nil
	}

	if !s.probe.Available(item.Port) {
		pid, exePath, _ := s.probe.FindProcess(item.Port)
		if pid > 0 && isPathMatch(exePath, item.Path) {
			if item.Kind == software.KindNginx {
				nginxBin := filepath.Join(item.Path, "nginx.exe")
				_ = exec.Command(nginxBin, "-p", item.Path, "-s", "stop").Run()
			}
			killCmd := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(pid))
			killCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
			_ = killCmd.Run()

			for i := 0; i < 8; i++ {
				if s.probe.Available(item.Port) {
					break
				}
				time.Sleep(250 * time.Millisecond)
			}
		}
		if !s.probe.Available(item.Port) {
			if pid > 0 && exePath != "" {
				return appErrors.New(appErrors.CodePortAlreadyInUse, fmt.Sprintf("端口 %d 已被占用 (PID: %d, %s)", item.Port, pid, filepath.Base(exePath)))
			}
			return appErrors.New(appErrors.CodePortAlreadyInUse, fmt.Sprintf("端口 %d 已被占用", item.Port))
		}
	}

	logPath := s.softwareLogPath(item.Kind, item.Version)
	var opts ports.Options

	switch item.Kind {
	case software.KindMySQL:
		if err := s.prepareMySQL(ctx, *item, logPath); err != nil {
			return err
		}
		iniPath := filepath.Join(item.Path, "my.ini")
		opts = ports.Options{
			Binary:    filepath.Join(item.Path, "bin", "mysqld.exe"),
			Args:      []string{"--defaults-file=" + iniPath, "--console"},
			Directory: item.Path,
			Log:       logPath,
		}
	case software.KindRedis:
		redisBin := installer.FindRedisBinary(item.Path)
		if redisBin == "" {
			return appErrors.New(appErrors.CodeProcessStartFailed, "未找到 redis-server.exe")
		}
		confPath, err := s.prepareRedis(*item, filepath.Dir(redisBin))
		if err != nil {
			return err
		}
		opts = ports.Options{
			Binary:    redisBin,
			Args:      []string{filepath.Base(confPath)},
			Directory: filepath.Dir(redisBin),
			Log:       logPath,
		}
	case software.KindNginx:
		if _, err := s.prepareNginx(*item); err != nil {
			return err
		}
		opts = ports.Options{
			Binary:    filepath.Join(item.Path, "nginx.exe"),
			Args:      []string{"-p", item.Path, "-c", "conf/nginx.conf"},
			Directory: item.Path,
			Log:       logPath,
		}
	case software.KindApache:
		confPath, err := s.prepareApache(*item)
		if err != nil {
			return err
		}
		httpdBin := installer.FindApacheBinary(item.Path)
		if httpdBin == "" {
			httpdBin = filepath.Join(item.Path, "bin", "httpd.exe")
		}
		opts = ports.Options{
			Binary:    httpdBin,
			Args:      []string{"-d", item.Path, "-f", confPath},
			Directory: item.Path,
			Log:       logPath,
		}
	}

	swID := item.ID
	opts.OnExit = func(_ string, message string) {
		s.bus.Emit("service:status-changed", map[string]interface{}{
			"id":      swID,
			"running": false,
			"pid":     0,
			"message": message,
		})
	}

	pid, err := s.runner.Start(key, opts)
	if err != nil {
		return err
	}

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if s.probe.Listening(item.Port) {
			s.bus.Emit("service:status-changed", map[string]interface{}{
				"id":      item.ID,
				"running": true,
				"pid":     pid,
				"message": "",
			})
			return nil
		}
		st := s.runner.Status(key)
		if !st.Running {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}

	_ = s.runner.Stop(key)
	tail := process.ReadLogTail(logPath, 20)
	detail := strings.Join(tail, "\n")
	return appErrors.New(appErrors.CodeProcessStartFailed, "服务启动超时或退出:\n"+detail)
}

func (s *SoftwareService) Stop(ctx context.Context, id int64) error {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if item == nil {
		return appErrors.New(appErrors.CodeNotFound, "软件版本不存在")
	}
	key := daemonKey(item.ID)
	if err := s.runner.Stop(key); err != nil {
		return err
	}
	s.bus.Emit("service:status-changed", map[string]interface{}{
		"id":      item.ID,
		"running": false,
		"pid":     0,
		"message": "",
	})
	return nil
}

func (s *SoftwareService) SetAutostart(ctx context.Context, id int64, on bool) error {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if item == nil {
		return appErrors.New(appErrors.CodeNotFound, "软件版本不存在")
	}
	item.Autostart = on
	return s.repo.Update(ctx, item)
}

func (s *SoftwareService) Config(ctx context.Context, id int64) (software.Config, error) {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return software.Config{}, err
	}
	if item == nil {
		return software.Config{}, appErrors.New(appErrors.CodeNotFound, "软件版本不存在")
	}

	filePath, err := s.ensureConfigFile(*item)
	if err != nil {
		return software.Config{}, err
	}
	if filePath == "" {
		return software.Config{Items: make([]software.ConfigItem, 0), File: ""}, nil
	}

	raw, err := os.ReadFile(filePath)
	if err != nil {
		return software.Config{}, appErrors.Wrap(err, appErrors.CodePathNotAccessible, "读取配置文件失败")
	}
	text := string(raw)

	defs := configSchemaForKind(item.Kind)
	items := make([]software.ConfigItem, 0, len(defs))
	for _, d := range defs {
		val := readConfigValue(item.Kind, text, d.Key)
		items = append(items, software.ConfigItem{
			Key:         d.Key,
			Label:       d.Label,
			Value:       val,
			Description: d.Description,
		})
	}
	return software.Config{
		Items: items,
		File:  filePath,
	}, nil
}

func (s *SoftwareService) SaveConfig(ctx context.Context, id int64, values map[string]string) (bool, error) {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return false, err
	}
	if item == nil {
		return false, appErrors.New(appErrors.CodeNotFound, "软件版本不存在")
	}

	filePath, err := s.ensureConfigFile(*item)
	if err != nil {
		return false, err
	}
	if filePath == "" {
		return false, appErrors.New(appErrors.CodeInvalidArgument, "该软件没有配置文件")
	}

	raw, err := os.ReadFile(filePath)
	if err != nil {
		return false, appErrors.Wrap(err, appErrors.CodePathNotAccessible, "读取配置文件失败")
	}
	if err := os.WriteFile(filePath+".bak", raw, 0644); err != nil {
		return false, appErrors.Wrap(err, appErrors.CodePathNotAccessible, "备份配置文件失败")
	}

	text := string(raw)
	for k, v := range values {
		text = writeConfigValue(item.Kind, text, k, v)
	}

	if err := os.WriteFile(filePath, []byte(text), 0644); err != nil {
		return false, appErrors.Wrap(err, appErrors.CodePathNotAccessible, "保存配置文件失败")
	}

	restart := software.IsDaemon(item.Kind)
	return restart, nil
}

func (s *SoftwareService) ReadConfig(ctx context.Context, id int64) (string, error) {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return "", err
	}
	if item == nil {
		return "", appErrors.New(appErrors.CodeNotFound, "软件版本不存在")
	}
	filePath, err := s.ensureConfigFile(*item)
	if err != nil {
		return "", err
	}
	if filePath == "" {
		return "", appErrors.New(appErrors.CodeInvalidArgument, "该软件没有配置文件")
	}
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return "", appErrors.Wrap(err, appErrors.CodePathNotAccessible, "读取配置文件失败")
	}
	return string(raw), nil
}

func (s *SoftwareService) WriteConfig(ctx context.Context, id int64, text string) error {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if item == nil {
		return appErrors.New(appErrors.CodeNotFound, "软件版本不存在")
	}
	filePath, err := s.ensureConfigFile(*item)
	if err != nil {
		return err
	}
	if filePath == "" {
		return appErrors.New(appErrors.CodeInvalidArgument, "该软件没有配置文件")
	}
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodePathNotAccessible, "读取原配置文件失败")
	}
	if err := os.WriteFile(filePath+".bak", raw, 0644); err != nil {
		return appErrors.Wrap(err, appErrors.CodePathNotAccessible, "备份配置文件失败")
	}
	if err := os.WriteFile(filePath, []byte(text), 0644); err != nil {
		return appErrors.Wrap(err, appErrors.CodePathNotAccessible, "写入配置文件失败")
	}
	return nil
}

func (s *SoftwareService) Extensions(ctx context.Context, id int64) ([]software.Extension, error) {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil || item.Kind != software.KindPHP {
		return nil, appErrors.New(appErrors.CodeInvalidArgument, "仅 PHP 支持扩展管理")
	}
	iniPath, err := s.ensureConfigFile(*item)
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(iniPath)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodePathNotAccessible, "读取 php.ini 失败")
	}
	iniText := string(raw)

	extDir := filepath.Join(item.Path, "ext")
	entries, err := os.ReadDir(extDir)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodePathNotAccessible, "读取 PHP ext 目录失败")
	}

	var list []software.Extension
	for _, ent := range entries {
		if ent.IsDir() {
			continue
		}
		lower := strings.ToLower(ent.Name())
		if !strings.HasPrefix(lower, "php_") || !strings.HasSuffix(lower, ".dll") {
			continue
		}
		name := strings.TrimSuffix(strings.TrimPrefix(lower, "php_"), ".dll")
		enabled := isPHPExtensionEnabled(iniText, name)
		list = append(list, software.Extension{
			Name:    name,
			Enabled: enabled,
		})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
	return list, nil
}

func (s *SoftwareService) SetExtension(ctx context.Context, id int64, name string, on bool) error {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if item == nil || item.Kind != software.KindPHP {
		return appErrors.New(appErrors.CodeInvalidArgument, "仅 PHP 支持扩展管理")
	}
	iniPath, err := s.ensureConfigFile(*item)
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(iniPath)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodePathNotAccessible, "读取 php.ini 失败")
	}
	if err := os.WriteFile(iniPath+".bak", raw, 0644); err != nil {
		return appErrors.Wrap(err, appErrors.CodePathNotAccessible, "备份 php.ini 失败")
	}

	updated := togglePHPExtension(string(raw), strings.TrimSpace(name), on)
	if err := os.WriteFile(iniPath, []byte(updated), 0644); err != nil {
		return appErrors.Wrap(err, appErrors.CodePathNotAccessible, "写入 php.ini 失败")
	}
	return nil
}

func (s *SoftwareService) OpenFolder(ctx context.Context, id int64) error {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if item == nil {
		return appErrors.New(appErrors.CodeNotFound, "软件版本不存在")
	}
	target := item.Path
	if item.Kind == software.KindMySQL {
		dataDir := s.mysqlDataDir(item.Version)
		if info, statErr := os.Stat(dataDir); statErr == nil && info.IsDir() {
			target = dataDir
		}
	}
	if info, statErr := os.Stat(target); statErr != nil || !info.IsDir() {
		return appErrors.New(appErrors.CodePathNotAccessible, "目录不存在: "+target)
	}
	cmd := exec.Command("explorer.exe", target)
	return cmd.Start()
}

func (s *SoftwareService) Logs(ctx context.Context, id int64, limit int) ([]string, error) {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, appErrors.New(appErrors.CodeNotFound, "软件版本不存在")
	}
	if item.Kind == software.KindNginx || item.Kind == software.KindApache {
		errLog := filepath.Join(item.Path, "logs", "error.log")
		if lines := process.ReadLogTail(errLog, limit); len(lines) > 0 {
			return lines, nil
		}
	}
	logPath := s.softwareLogPath(item.Kind, item.Version)
	return process.ReadLogTail(logPath, limit), nil
}

func (s *SoftwareService) StopAll() {
	list, err := s.repo.List(context.Background())
	if err != nil {
		return
	}
	for _, item := range list {
		if software.IsDaemon(item.Kind) {
			_ = s.runner.Stop(daemonKey(item.ID))
		}
	}
}

func (s *SoftwareService) allocateDaemonPort(ctx context.Context, kind string) int {
	basePort := 3306
	if kind == software.KindRedis {
		basePort = 6379
	} else if kind == software.KindNginx {
		basePort = 80
	} else if kind == software.KindApache {
		basePort = 8088
	} else if kind == software.KindPHP {
		basePort = 9000
	}
	existing, _ := s.repo.ListByKind(ctx, kind)
	used := make(map[int]bool, len(existing))
	for _, item := range existing {
		if item.Port > 0 {
			used[item.Port] = true
		}
	}
	for port := basePort; port < basePort+1000; port++ {
		if !used[port] && s.probe.Available(port) {
			return port
		}
	}
	return basePort
}

func (s *SoftwareService) prepareMySQL(ctx context.Context, item software.Software, logPath string) error {
	dataDir := s.mysqlDataDir(item.Version)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return appErrors.Wrap(err, appErrors.CodeProcessStartFailed, "创建 MySQL 数据目录失败")
	}
	if _, err := s.ensureConfigFile(item); err != nil {
		return err
	}

	mysqlSysDir := filepath.Join(dataDir, "mysql")
	if info, err := os.Stat(mysqlSysDir); err != nil || !info.IsDir() {
		iniPath := filepath.Join(item.Path, "my.ini")
		mysqld := filepath.Join(item.Path, "bin", "mysqld.exe")
		cmd := exec.CommandContext(ctx, mysqld, "--defaults-file="+iniPath, "--initialize-insecure", "--console")
		cmd.Dir = item.Path
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		out, initErr := cmd.CombinedOutput()
		if len(out) > 0 {
			_ = os.MkdirAll(filepath.Dir(logPath), 0755)
			f, fErr := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if fErr == nil {
				_, _ = f.Write(out)
				_ = f.Close()
			}
		}
		if initErr != nil {
			return appErrors.Wrap(initErr, appErrors.CodeProcessStartFailed, "初始化 MySQL 数据目录失败: "+string(out))
		}
	}
	return nil
}

func (s *SoftwareService) prepareRedis(item software.Software, binDir string) (string, error) {
	confPath := filepath.Join(binDir, "redis.conf")
	port := item.Port
	if port <= 0 {
		port = 6379
	}
	if raw, err := os.ReadFile(confPath); err == nil {
		updated := writeConfigValue(software.KindRedis, string(raw), "port", strconv.Itoa(port))
		updated = writeConfigValue(software.KindRedis, updated, "daemonize", "no")
		updated = writeConfigValue(software.KindRedis, updated, "logfile", `""`)
		if err := os.WriteFile(confPath, []byte(updated), 0644); err != nil {
			return "", appErrors.Wrap(err, appErrors.CodeProcessStartFailed, "更新 redis.conf 失败")
		}
		return confPath, nil
	}
	content := fmt.Sprintf("bind 127.0.0.1\r\nport %d\r\ndaemonize no\r\nlogfile \"\"\r\nmaxmemory 0\r\nappendonly no\r\n", port)
	if err := os.WriteFile(confPath, []byte(content), 0644); err != nil {
		return "", appErrors.Wrap(err, appErrors.CodeProcessStartFailed, "生成 redis.conf 失败")
	}
	return confPath, nil
}

func (s *SoftwareService) prepareNginx(item software.Software) (string, error) {
	_ = os.MkdirAll(filepath.Join(item.Path, "logs"), 0755)
	_ = os.MkdirAll(filepath.Join(item.Path, "temp"), 0755)
	confDir := filepath.Join(item.Path, "conf")
	_ = os.MkdirAll(confDir, 0755)
	_ = os.MkdirAll(filepath.Join(confDir, "vhosts"), 0755)
	confPath := filepath.Join(confDir, "nginx.conf")
	port := item.Port
	if port <= 0 {
		port = 80
	}
	if raw, err := os.ReadFile(confPath); err == nil {
		text := string(raw)
		if !strings.Contains(text, "include vhosts/*.conf;") {
			text = strings.Replace(text, "http {", "http {\r\n    include vhosts/*.conf;", 1)
		}
		re := regexp.MustCompile(`(?m)^(\s*listen\s+)\d+(\s*;.*)$`)
		updated := re.ReplaceAllString(text, fmt.Sprintf("${1}%d${2}", port))
		_ = os.WriteFile(confPath, []byte(updated), 0644)
		return confPath, nil
	}
	content := fmt.Sprintf("worker_processes 1;\r\nevents {\r\n    worker_connections 1024;\r\n}\r\nhttp {\r\n    include mime.types;\r\n    default_type application/octet-stream;\r\n    sendfile on;\r\n    keepalive_timeout 65;\r\n    client_max_body_size 100m;\r\n    include vhosts/*.conf;\r\n    server {\r\n        listen %d default_server;\r\n        server_name _;\r\n        return 404 \"DevEnv: No site matches this domain\\n\";\r\n    }\r\n}\r\n", port)
	if err := os.WriteFile(confPath, []byte(content), 0644); err != nil {
		return "", appErrors.Wrap(err, appErrors.CodeProcessStartFailed, "生成 nginx.conf 失败")
	}
	return confPath, nil
}

func (s *SoftwareService) ensureConfigFile(item software.Software) (string, error) {
	switch item.Kind {
	case software.KindPHP:
		iniPath := filepath.Join(item.Path, "php.ini")
		if _, err := os.Stat(iniPath); os.IsNotExist(err) {
			devIni := filepath.Join(item.Path, "php.ini-development")
			if raw, readErr := os.ReadFile(devIni); readErr == nil {
				_ = os.WriteFile(iniPath, raw, 0644)
			}
		}
		return iniPath, nil
	case software.KindMySQL:
		iniPath := filepath.Join(item.Path, "my.ini")
		if _, err := os.Stat(iniPath); os.IsNotExist(err) {
			dataDir := filepath.ToSlash(s.mysqlDataDir(item.Version))
			baseDir := filepath.ToSlash(item.Path)
			content := fmt.Sprintf("[mysqld]\r\nport=%d\r\nbind-address=127.0.0.1\r\nbasedir=\"%s\"\r\ndatadir=\"%s\"\r\ncharacter-set-server=utf8mb4\r\nmax_connections=200\r\ninnodb_buffer_pool_size=128M\r\n",
				item.Port, baseDir, dataDir,
			)
			if err := os.WriteFile(iniPath, []byte(content), 0644); err != nil {
				return "", appErrors.Wrap(err, appErrors.CodePathNotAccessible, "生成 my.ini 失败")
			}
		}
		return iniPath, nil
	case software.KindRedis:
		redisBin := installer.FindRedisBinary(item.Path)
		binDir := item.Path
		if redisBin != "" {
			binDir = filepath.Dir(redisBin)
		}
		return s.prepareRedis(item, binDir)
	case software.KindNginx:
		return s.prepareNginx(item)
	case software.KindApache:
		return s.prepareApache(item)
	default:
		return "", nil
	}
}

func (s *SoftwareService) prepareApache(item software.Software) (string, error) {
	_ = os.MkdirAll(filepath.Join(item.Path, "logs"), 0755)
	confDir := filepath.Join(item.Path, "conf")
	_ = os.MkdirAll(confDir, 0755)
	confPath := filepath.Join(confDir, "httpd.conf")
	port := item.Port
	if port <= 0 {
		port = 8088
	}
	targetRoot := filepath.ToSlash(item.Path)
	if raw, err := os.ReadFile(confPath); err == nil {
		text := string(raw)
		reSrvRoot := regexp.MustCompile(`(?mi)^\s*Define\s+SRVROOT\s+.*$`)
		if reSrvRoot.MatchString(text) {
			text = reSrvRoot.ReplaceAllString(text, fmt.Sprintf(`Define SRVROOT "%s"`, targetRoot))
		} else {
			text = fmt.Sprintf("Define SRVROOT \"%s\"\r\n%s", targetRoot, text)
		}

		reListen := regexp.MustCompile(`(?mi)^\s*Listen\s+\d+.*$`)
		if reListen.MatchString(text) {
			text = reListen.ReplaceAllString(text, fmt.Sprintf("Listen %d", port))
		} else {
			text = fmt.Sprintf("Listen %d\r\n%s", port, text)
		}

		reServerName := regexp.MustCompile(`(?mi)^\s*#?\s*ServerName\s+.*$`)
		if reServerName.MatchString(text) {
			text = reServerName.ReplaceAllString(text, fmt.Sprintf("ServerName localhost:%d", port))
		}

		_ = os.WriteFile(confPath, []byte(text), 0644)
		return confPath, nil
	}
	content := fmt.Sprintf("Define SRVROOT \"%s\"\r\nServerRoot \"${SRVROOT}\"\r\nListen %d\r\nServerName localhost:%d\r\nDocumentRoot \"${SRVROOT}/htdocs\"\r\n<Directory \"${SRVROOT}/htdocs\">\r\n    Options Indexes FollowSymLinks\r\n    AllowOverride None\r\n    Require all granted\r\n</Directory>\r\n", targetRoot, port, port)
	if err := os.WriteFile(confPath, []byte(content), 0644); err != nil {
		return "", appErrors.Wrap(err, appErrors.CodeProcessStartFailed, "生成 httpd.conf 失败")
	}
	return confPath, nil
}

func (s *SoftwareService) cachePath(kind, version string) string {
	return filepath.Join(s.baseDir, "cache", fmt.Sprintf("%s-%s.zip.part", kind, version))
}

func (s *SoftwareService) softwareLogPath(kind, version string) string {
	return filepath.Join(s.baseDir, "logs", "software", fmt.Sprintf("%s-%s.log", kind, version))
}

func (s *SoftwareService) mysqlDataDir(version string) string {
	return filepath.Join(s.baseDir, "dbdata", "mysql-"+version)
}

func daemonKey(id int64) string {
	return "software:" + strconv.FormatInt(id, 10)
}

func configSchemaForKind(kind string) []software.ConfigItem {
	switch kind {
	case software.KindPHP:
		return []software.ConfigItem{
			{Key: "upload_max_filesize", Label: "最大上传文件", Description: "允许上传的单个文件最大体积，如 50M"},
			{Key: "post_max_size", Label: "POST 最大体积", Description: "POST 请求体最大体积，如 50M"},
			{Key: "memory_limit", Label: "内存限制", Description: "单脚本最大内存消耗，如 256M"},
			{Key: "max_execution_time", Label: "最大执行时间", Description: "脚本最长执行秒数，如 300"},
			{Key: "date.timezone", Label: "默认时区", Description: "PHP 日期与时间函数默认时区"},
			{Key: "display_errors", Label: "显示错误", Description: "是否输出错误信息 (On / Off)"},
		}
	case software.KindMySQL:
		return []software.ConfigItem{
			{Key: "max_connections", Label: "最大连接数", Description: "允许同时连接的客户端总数"},
			{Key: "innodb_buffer_pool_size", Label: "缓冲池大小", Description: "InnoDB 缓冲池内存大小，如 128M"},
			{Key: "character-set-server", Label: "服务端字符集", Description: "默认字符集，推荐 utf8mb4"},
		}
	case software.KindRedis:
		return []software.ConfigItem{
			{Key: "maxmemory", Label: "最大内存", Description: "最大内存上限（0 表示不限制，如 256mb）"},
			{Key: "requirepass", Label: "访问密码", Description: "客户端连接密码，留空表示无密码"},
			{Key: "appendonly", Label: "AOF 持久化", Description: "是否开启 AOF 日志持久化 (yes / no)"},
		}
	case software.KindNginx:
		return []software.ConfigItem{
			{Key: "worker_processes", Label: "工作进程数", Description: "Nginx worker 进程数量，如 1 或 auto"},
			{Key: "worker_connections", Label: "单进程最大连接", Description: "单个 worker 最大并发连接数，如 1024"},
			{Key: "keepalive_timeout", Label: "长连接超时", Description: "HTTP keepalive 超时秒数，如 65"},
		}
	case software.KindApache:
		return []software.ConfigItem{
			{Key: "ServerName", Label: "服务器名称", Description: "服务器用于自我标识的主机名与端口，如 localhost:8088"},
			{Key: "ServerAdmin", Label: "管理员邮箱", Description: "发生错误时展示的管理员联络邮箱，如 admin@localhost"},
			{Key: "Timeout", Label: "超时时间", Description: "接收或发送请求前等待的最长秒数，如 60"},
			{Key: "KeepAlive", Label: "保持长连接", Description: "是否开启 HTTP 持续连接支持 (On / Off)"},
			{Key: "MaxKeepAliveRequests", Label: "最大长连接请求数", Description: "单个长连接允许的最大请求数，如 100"},
			{Key: "KeepAliveTimeout", Label: "长连接超时", Description: "同一连接上等待后续请求的秒数，如 5"},
		}
	default:
		return nil
	}
}

func readConfigValue(kind, text, key string) string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if kind == software.KindRedis {
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			fields := strings.Fields(trimmed)
			if len(fields) >= 2 && strings.EqualFold(fields[0], key) {
				return strings.Trim(strings.Join(fields[1:], " "), `"`)
			}
		}
		return ""
	}
	if kind == software.KindNginx {
		re := regexp.MustCompile(`(?i)^\s*` + regexp.QuoteMeta(key) + `\s+([^;]+);`)
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			if m := re.FindStringSubmatch(trimmed); m != nil {
				return strings.TrimSpace(m[1])
			}
		}
		return ""
	}
	if kind == software.KindApache {
		re := regexp.MustCompile(`(?im)^\s*` + regexp.QuoteMeta(key) + `\s+([^\r\n#]+)`)
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			if m := re.FindStringSubmatch(trimmed); m != nil {
				return strings.TrimSpace(m[1])
			}
		}
		return ""
	}

	pattern := regexp.MustCompile(`(?i)^\s*` + regexp.QuoteMeta(key) + `\s*=\s*(.*)$`)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, ";") || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if m := pattern.FindStringSubmatch(trimmed); m != nil {
			val := strings.TrimSpace(m[1])
			return strings.Trim(val, `"`)
		}
	}
	return ""
}

func writeConfigValue(kind, text, key, value string) string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	replaced := false

	if kind == software.KindRedis {
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "#") {
				continue
			}
			fields := strings.Fields(trimmed)
			if len(fields) >= 1 && strings.EqualFold(fields[0], key) {
				if strings.TrimSpace(value) == "" && key == "requirepass" {
					lines[i] = "# requirepass"
				} else {
					lines[i] = fmt.Sprintf("%s %s", key, strings.TrimSpace(value))
				}
				replaced = true
				break
			}
		}
		if !replaced && strings.TrimSpace(value) != "" {
			lines = append(lines, fmt.Sprintf("%s %s", key, strings.TrimSpace(value)))
		}
		return strings.Join(lines, "\r\n")
	}

	if kind == software.KindNginx {
		re := regexp.MustCompile(`(?i)^(\s*` + regexp.QuoteMeta(key) + `\s+)[^;]+(;.*)$`)
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "#") {
				continue
			}
			if re.MatchString(line) {
				lines[i] = re.ReplaceAllString(line, fmt.Sprintf("${1}%s${2}", strings.TrimSpace(value)))
				break
			}
		}
		return strings.Join(lines, "\r\n")
	}

	if kind == software.KindApache {
		re := regexp.MustCompile(`(?im)^(\s*` + regexp.QuoteMeta(key) + `\s+)[^\r\n#]+(.*)$`)
		commentRe := regexp.MustCompile(`(?im)^(\s*#\s*` + regexp.QuoteMeta(key) + `\s+)[^\r\n#]+(.*)$`)
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "#") {
				continue
			}
			if re.MatchString(line) {
				lines[i] = fmt.Sprintf("%s %s", key, strings.TrimSpace(value))
				replaced = true
				break
			}
		}
		if !replaced {
			for i, line := range lines {
				if commentRe.MatchString(line) {
					lines[i] = fmt.Sprintf("%s %s", key, strings.TrimSpace(value))
					replaced = true
					break
				}
			}
		}
		if !replaced && strings.TrimSpace(value) != "" {
			lines = append(lines, fmt.Sprintf("%s %s", key, strings.TrimSpace(value)))
		}
		return strings.Join(lines, "\r\n")
	}

	activePattern := regexp.MustCompile(`(?i)^\s*` + regexp.QuoteMeta(key) + `\s*=`)
	commentPattern := regexp.MustCompile(`(?i)^\s*[;#]\s*` + regexp.QuoteMeta(key) + `\s*=`)

	for i, line := range lines {
		if activePattern.MatchString(line) {
			lines[i] = fmt.Sprintf("%s = %s", key, value)
			replaced = true
			break
		}
	}
	if !replaced {
		for i, line := range lines {
			if commentPattern.MatchString(line) {
				lines[i] = fmt.Sprintf("%s = %s", key, value)
				replaced = true
				break
			}
		}
	}
	if !replaced {
		lines = append(lines, fmt.Sprintf("%s = %s", key, value))
	}
	return strings.Join(lines, "\r\n")
}

func isPHPExtensionEnabled(iniText, name string) bool {
	pattern := regexp.MustCompile(`(?im)^\s*extension\s*=\s*(?:php_)?` + regexp.QuoteMeta(name) + `(?:\.dll)?\s*(?:;.*)?$`)
	return pattern.MatchString(iniText)
}

func togglePHPExtension(iniText, name string, on bool) string {
	lines := strings.Split(strings.ReplaceAll(iniText, "\r\n", "\n"), "\n")
	activeRe := regexp.MustCompile(`(?i)^\s*extension\s*=\s*(?:php_)?` + regexp.QuoteMeta(name) + `(?:\.dll)?\s*(?:;.*)?$`)
	commentRe := regexp.MustCompile(`(?i)^\s*;\s*extension\s*=\s*(?:php_)?` + regexp.QuoteMeta(name) + `(?:\.dll)?\s*(?:;.*)?$`)

	if on {
		for i, line := range lines {
			if activeRe.MatchString(line) {
				return strings.Join(lines, "\r\n")
			}
			if commentRe.MatchString(line) {
				lines[i] = "extension=" + name
				return strings.Join(lines, "\r\n")
			}
		}
		lines = append(lines, "extension="+name)
		return strings.Join(lines, "\r\n")
	}

	for i, line := range lines {
		if activeRe.MatchString(line) {
			lines[i] = ";extension=" + name
		}
	}
	return strings.Join(lines, "\r\n")
}

func isPathMatch(exePath, softwarePath string) bool {
	if exePath == "" || softwarePath == "" {
		return false
	}
	cleanExe := strings.ToLower(filepath.Clean(exePath))
	cleanSoft := strings.ToLower(filepath.Clean(softwarePath))
	return strings.HasPrefix(cleanExe, cleanSoft)
}
