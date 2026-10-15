package app

import (
	"context"
	"errors"
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
	"devenv/internal/shared/paths"
)

const (
	MYSQL_START_PORT     = 3306
	MARIADB_START_PORT   = 3307
	POSTGRES_START_PORT  = 5432
	MONGODB_START_PORT   = 27017
	REDIS_START_PORT     = 6379
	MEMCACHED_START_PORT = 11211
	PORT_SPAN            = 1000
	MARKET_TIMEOUT       = 10 * time.Second
)

type SoftwareView struct {
	software.Software
	Running bool
	PID     int
}

type SoftwareService struct {
	base      string
	repo      ports.SoftwareRepository
	sites     ports.SiteRepository
	catalog   ports.Catalog
	installer ports.Installer
	runner    ports.Runner
	probe     ports.PortChecker
	bus       ports.EventPublisher
	installed func(ctx context.Context, item software.Software)

	mu      sync.Mutex
	cancels map[string]context.CancelFunc
}

func NewSoftwareService(
	base string,
	repo ports.SoftwareRepository,
	sites ports.SiteRepository,
	catalog ports.Catalog,
	inst ports.Installer,
	runner ports.Runner,
	probe ports.PortChecker,
	bus ports.EventPublisher,
	installed func(ctx context.Context, item software.Software),
) *SoftwareService {
	return &SoftwareService{
		base:      base,
		repo:      repo,
		sites:     sites,
		catalog:   catalog,
		installer: inst,
		runner:    runner,
		probe:     probe,
		bus:       bus,
		installed: installed,
		cancels:   make(map[string]context.CancelFunc),
	}
}

func (s *SoftwareService) StartAuto(ctx context.Context) {
	list, err := s.repo.List(ctx)
	if err != nil {
		return
	}
	for _, item := range list {
		if !software.IsValidKind(item.Kind) {
			if deleteErr := s.repo.Delete(ctx, item.ID); deleteErr != nil {
				continue
			}
			continue
		}
		if item.State == software.STATE_INSTALLING {
			item.State = software.STATE_FAILED
			item.Progress = 0
			item.Error = "安装被中断"
			if updateErr := s.repo.Update(ctx, &item); updateErr != nil {
				continue
			}
			if removeErr := os.Remove(paths.Cache(s.base, item.Kind, item.Version)); removeErr != nil && !os.IsNotExist(removeErr) {
				continue
			}
			if cleanErr := os.RemoveAll(item.Path + ".tmp"); cleanErr != nil {
				continue
			}
		}
	}

	s.scanDisk(ctx)
	go s.warmCatalog()

	updated, err := s.repo.List(ctx)
	if err != nil {
		return
	}
	for _, item := range updated {
		if item.State == software.STATE_INSTALLED && software.IsDaemon(item.Kind) && item.Autostart {
			if startErr := s.Start(ctx, item.ID); startErr != nil {
				continue
			}
		}
	}
}

func (s *SoftwareService) warmCatalog() {
	var group sync.WaitGroup
	for _, kind := range software.AllKinds() {
		group.Add(1)
		go func(target string) {
			defer group.Done()
			s.warmKind(target)
		}(kind)
	}
	group.Wait()
	s.bus.Emit("service:status-changed", map[string]interface{}{
		"id":      0,
		"running": false,
		"pid":     0,
		"message": "",
	})
}

func (s *SoftwareService) warmKind(kind string) {
	if len(s.catalog.Cached(kind)) > 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), MARKET_TIMEOUT)
	defer cancel()
	if _, err := s.catalog.Versions(ctx, kind); err != nil {
		return
	}
}

func (s *SoftwareService) scanDisk(ctx context.Context) {
	for _, kind := range software.AllKinds() {
		folder := filepath.Join(s.base, paths.SOFTWARE_FOLDER, kind)
		entries, err := os.ReadDir(folder)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() || strings.HasSuffix(entry.Name(), ".tmp") {
				continue
			}
			version := entry.Name()
			existing, err := s.repo.GetByKindVersion(ctx, kind, version)
			if err != nil || existing != nil {
				continue
			}
			target := paths.Software(s.base, kind, version)
			if !checkBinary(kind, target) {
				continue
			}
			mainItem, err := s.repo.GetMain(ctx, kind)
			if err != nil {
				continue
			}
			port := 0
			if software.IsDaemon(kind) {
				port = s.findPort(ctx, kind)
			}
			created := &software.Software{
				Kind:     kind,
				Version:  version,
				Path:     target,
				State:    software.STATE_INSTALLED,
				Progress: 100,
				Main:     mainItem == nil,
				Port:     port,
			}
			if kind == software.KIND_PHP {
				s.fixPHP(*created)
			}
			if err := s.repo.Create(ctx, created); err == nil {
				s.installed(ctx, *created)
			}
		}
	}
}

func (s *SoftwareService) fixPHP(record software.Software) {
	iniPath := filepath.Join(record.Path, "php.ini")
	raw, err := os.ReadFile(iniPath)
	if err != nil {
		return
	}
	extensionFolder := filepath.Join(record.Path, "ext")
	re := regexp.MustCompile(`(?m)^extension_dir\s*=\s*".*"$`)
	text := re.ReplaceAllString(string(raw), fmt.Sprintf(`extension_dir = "%s"`, extensionFolder))
	if err := os.WriteFile(iniPath, []byte(text), 0o644); err != nil {
		return
	}
	s.setPHPLog(record)
}

func checkBinary(kind, folder string) bool {
	switch kind {
	case software.KIND_PHP:
		_, err := os.Stat(filepath.Join(folder, "php-cgi.exe"))
		return err == nil
	case software.KIND_NODE:
		_, err := os.Stat(filepath.Join(folder, "node.exe"))
		return err == nil
	case software.KIND_GO:
		_, err := os.Stat(filepath.Join(folder, "bin", "go.exe"))
		return err == nil
	case software.KIND_PYTHON:
		_, err := os.Stat(filepath.Join(folder, "python.exe"))
		return err == nil
	case software.KIND_COMPOSER:
		_, err := os.Stat(filepath.Join(folder, "composer.phar"))
		return err == nil
	case software.KIND_MYSQL:
		_, err := os.Stat(filepath.Join(folder, "bin", "mysqld.exe"))
		return err == nil
	case software.KIND_MARIADB:
		if _, err := os.Stat(filepath.Join(folder, "bin", "mariadbd.exe")); err == nil {
			return true
		}
		_, err := os.Stat(filepath.Join(folder, "bin", "mysqld.exe"))
		return err == nil
	case software.KIND_POSTGRESQL:
		_, err := os.Stat(filepath.Join(folder, "bin", "postgres.exe"))
		return err == nil
	case software.KIND_MONGODB:
		_, err := os.Stat(filepath.Join(folder, "bin", "mongod.exe"))
		return err == nil
	case software.KIND_REDIS:
		return installer.FindRedis(folder) != ""
	case software.KIND_MEMCACHED:
		return installer.FindBinary(folder, "memcached.exe") != ""
	case software.KIND_JAVA:
		_, err := os.Stat(filepath.Join(folder, "bin", "java.exe"))
		return err == nil
	case software.KIND_DOTNET:
		_, err := os.Stat(filepath.Join(folder, "dotnet.exe"))
		return err == nil
	case software.KIND_DENO:
		_, err := os.Stat(filepath.Join(folder, "deno.exe"))
		return err == nil
	case software.KIND_BUN:
		_, err := os.Stat(filepath.Join(folder, "bun.exe"))
		return err == nil
	case software.KIND_NGINX:
		_, err := os.Stat(filepath.Join(folder, "nginx.exe"))
		return err == nil
	case software.KIND_CADDY:
		_, err := os.Stat(filepath.Join(folder, "caddy.exe"))
		return err == nil
	case software.KIND_SQLITE:
		_, err := os.Stat(filepath.Join(folder, "sqlite3.exe"))
		return err == nil
	case software.KIND_GRADLE:
		_, err := os.Stat(filepath.Join(folder, "bin", "gradle.bat"))
		return err == nil
	case software.KIND_MAVEN:
		_, err := os.Stat(filepath.Join(folder, "bin", "mvn.cmd"))
		return err == nil
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
		if item.State == software.STATE_INSTALLED {
			counts[item.Kind]++
		}
	}

	items := []software.Item{
		{
			Kind:        software.KIND_PHP,
			Name:        "PHP",
			Description: "广泛使用的通用开源脚本语言，适用于 Web 开发并内置 CGI 支持",
			Category:    "运行环境",
			Count:       counts[software.KIND_PHP],
		},
		{
			Kind:        software.KIND_NODE,
			Name:        "Node.js",
			Description: "基于 Chrome V8 引擎的 JavaScript 运行环境，包含 npm 包管理器",
			Category:    "运行环境",
			Count:       counts[software.KIND_NODE],
		},
		{
			Kind:        software.KIND_GO,
			Name:        "Go",
			Description: "Google 开发的高性能编译型编程语言，内置并发支持与工具链",
			Category:    "运行环境",
			Count:       counts[software.KIND_GO],
		},
		{
			Kind:        software.KIND_PYTHON,
			Name:        "Python",
			Description: "简洁高效的通用解释型编程语言，广泛应用于后端与自动化脚本",
			Category:    "运行环境",
			Count:       counts[software.KIND_PYTHON],
		},
		{
			Kind:        software.KIND_COMPOSER,
			Name:        "Composer",
			Description: "PHP 标准依赖包管理工具，一键安装与管理项目类库依赖",
			Category:    "工具",
			Count:       counts[software.KIND_COMPOSER],
		},
		{
			Kind:        software.KIND_MYSQL,
			Name:        "MySQL",
			Description: "主流开源关系型数据库管理系统，支持事务与高性能存储",
			Category:    "数据库",
			Count:       counts[software.KIND_MYSQL],
		},
		{
			Kind:        software.KIND_MARIADB,
			Name:        "MariaDB",
			Description: "MySQL 开源分支关系型数据库，完全兼容 MySQL 协议与生态",
			Category:    "数据库",
			Count:       counts[software.KIND_MARIADB],
		},
		{
			Kind:        software.KIND_POSTGRESQL,
			Name:        "PostgreSQL",
			Description: "功能强大的企业级开源对象关系型数据库，支持复杂查询与 JSON",
			Category:    "数据库",
			Count:       counts[software.KIND_POSTGRESQL],
		},
		{
			Kind:        software.KIND_MONGODB,
			Name:        "MongoDB",
			Description: "高性能开源文档型 NoSQL 数据库，支持灵活的 BSON 数据模型",
			Category:    "数据库",
			Count:       counts[software.KIND_MONGODB],
		},
		{
			Kind:        software.KIND_REDIS,
			Name:        "Redis",
			Description: "高性能开源内存数据结构存储，常用作数据库、缓存和消息代理",
			Category:    "缓存",
			Count:       counts[software.KIND_REDIS],
		},
		{
			Kind:        software.KIND_MEMCACHED,
			Name:        "Memcached",
			Description: "高性能分布式内存对象缓存系统，有效加速动态 Web 应用",
			Category:    "缓存",
			Count:       counts[software.KIND_MEMCACHED],
		},
		{
			Kind: software.KIND_JAVA, Name: "OpenJDK",
			Description: "Eclipse Temurin OpenJDK，适用于 Java 服务端、桌面与构建工具链",
			Category:    "运行环境", Count: counts[software.KIND_JAVA],
		},
		{
			Kind: software.KIND_DOTNET, Name: ".NET SDK",
			Description: "微软跨平台开发工具包，可构建 ASP.NET Core、云服务和桌面应用",
			Category:    "运行环境", Count: counts[software.KIND_DOTNET],
		},
		{
			Kind: software.KIND_DENO, Name: "Deno",
			Description: "安全优先的 JavaScript 与 TypeScript 运行时，内置工具链",
			Category:    "运行环境", Count: counts[software.KIND_DENO],
		},
		{
			Kind: software.KIND_BUN, Name: "Bun",
			Description: "高性能 JavaScript 运行时、包管理器、测试器与打包器",
			Category:    "运行环境", Count: counts[software.KIND_BUN],
		},
		{
			Kind: software.KIND_NGINX, Name: "Nginx",
			Description: "高性能 Web 服务器与反向代理，适合本地静态资源和代理调试",
			Category:    "Web 服务", Count: counts[software.KIND_NGINX],
		},
		{
			Kind: software.KIND_CADDY, Name: "Caddy",
			Description: "配置简洁的现代 Web 服务器，支持反向代理与自动 HTTPS",
			Category:    "Web 服务", Count: counts[software.KIND_CADDY],
		},
		{
			Kind: software.KIND_SQLITE, Name: "SQLite",
			Description: "轻量级嵌入式关系数据库及官方命令行工具集",
			Category:    "数据库", Count: counts[software.KIND_SQLITE],
		},
		{
			Kind: software.KIND_GRADLE, Name: "Gradle",
			Description: "面向 Java、Kotlin 与 Android 项目的自动化构建工具",
			Category:    "工具", Count: counts[software.KIND_GRADLE],
		},
		{
			Kind: software.KIND_MAVEN, Name: "Maven",
			Description: "Apache Java 项目构建与依赖管理工具，遵循约定优于配置",
			Category:    "工具", Count: counts[software.KIND_MAVEN],
		},
	}

	for i := range items {
		versions := s.catalog.Cached(items[i].Kind)
		if len(versions) == 0 {
			go s.warmKind(items[i].Kind)
			continue
		}
		for _, item := range versions {
			if item.Recommended {
				items[i].Recommended = item.Version
				break
			}
		}
		if items[i].Recommended == "" {
			items[i].Recommended = versions[0].Version
		}
	}

	return items, nil
}

func (s *SoftwareService) List(ctx context.Context) ([]SoftwareView, error) {
	list, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	views := make([]SoftwareView, 0, len(list))
	for _, item := range list {
		if !software.IsValidKind(item.Kind) {
			continue
		}
		view := SoftwareView{Software: item}
		if software.IsDaemon(item.Kind) {
			status := s.runner.Status(daemonKey(item.ID))
			view.Running = status.Running
			view.PID = status.PID
		}
		views = append(views, view)
	}
	return views, nil
}

func (s *SoftwareService) Versions(ctx context.Context, kind string) ([]software.Version, error) {
	if !software.IsValidKind(kind) {
		return nil, appErrors.New(appErrors.INVALID_ARGUMENT, "不支持的软件类型: "+kind)
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
		if item.State == software.STATE_INSTALLED {
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
		return appErrors.New(appErrors.INVALID_ARGUMENT, "软件类型或版本号不能为空")
	}

	taskKey := kind + ":" + version
	s.mu.Lock()
	if _, active := s.cancels[taskKey]; active {
		s.mu.Unlock()
		return appErrors.New(appErrors.ALREADY_EXISTS, "该版本正在安装中")
	}
	s.mu.Unlock()

	existing, err := s.repo.GetByKindVersion(ctx, kind, version)
	if err != nil {
		return err
	}
	if existing != nil {
		if existing.State == software.STATE_INSTALLED {
			return appErrors.New(appErrors.ALREADY_EXISTS, fmt.Sprintf("%s %s 已安装", kind, version))
		}
		if err := os.RemoveAll(existing.Path); err != nil {
			return appErrors.Wrap(err, appErrors.INSTALL_FAILED, "清理残留安装目录失败")
		}
		if err := os.RemoveAll(existing.Path + ".tmp"); err != nil {
			return appErrors.Wrap(err, appErrors.INSTALL_FAILED, "清理残留临时目录失败")
		}
		if err := os.Remove(paths.Cache(s.base, kind, version)); err != nil && !os.IsNotExist(err) {
			return appErrors.Wrap(err, appErrors.INSTALL_FAILED, "清理残留下载包失败")
		}
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
		return appErrors.New(appErrors.NOT_FOUND, fmt.Sprintf("版本清单中未找到 %s %s", kind, version))
	}

	target := paths.Software(s.base, kind, version)
	record := &software.Software{
		Kind:     kind,
		Version:  version,
		Path:     target,
		State:    software.STATE_INSTALLING,
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

func (s *SoftwareService) runInstall(ctx context.Context, taskKey string, record software.Software, versionInfo software.Version) {
	defer func() {
		s.mu.Lock()
		delete(s.cancels, taskKey)
		s.mu.Unlock()
	}()

	request := ports.Request{
		Kind:     record.Kind,
		Version:  record.Version,
		URL:      versionInfo.URL,
		Checksum: versionInfo.Checksum,
		Target:   record.Path,
		Cache:    paths.Cache(s.base, record.Kind, record.Version),
	}

	report := func(progress int, message string) {
		record.Progress = progress
		if err := s.repo.Update(context.Background(), &record); err != nil {
			return
		}
		s.bus.Emit("task:progress", map[string]interface{}{
			"kind":     record.Kind,
			"version":  record.Version,
			"progress": progress,
			"message":  message,
		})
	}

	err := s.installer.Install(ctx, request, report)
	bgCtx := context.Background()

	if err != nil {
		record.State = software.STATE_FAILED
		record.Progress = 0
		record.Error = err.Error()
		if updateErr := s.repo.Update(bgCtx, &record); updateErr != nil {
			return
		}
		s.bus.Emit("task:completed", map[string]interface{}{
			"kind":    record.Kind,
			"version": record.Version,
			"ok":      false,
			"error":   err.Error(),
		})
		return
	}

	if record.Kind == software.KIND_PHP {
		s.setPHPLog(record)
	}

	mainItem, err := s.repo.GetMain(bgCtx, record.Kind)
	if err == nil && mainItem == nil {
		record.Main = true
	}
	if software.IsDaemon(record.Kind) {
		record.Port = s.findPort(bgCtx, record.Kind)
	}
	record.State = software.STATE_INSTALLED
	record.Progress = 100
	record.Error = ""
	if updateErr := s.repo.Update(bgCtx, &record); updateErr != nil {
		return
	}

	s.installed(bgCtx, record)

	s.bus.Emit("task:completed", map[string]interface{}{
		"kind":    record.Kind,
		"version": record.Version,
		"ok":      true,
		"error":   "",
	})
}

func (s *SoftwareService) setPHPLog(record software.Software) {
	iniPath := filepath.Join(record.Path, "php.ini")
	raw, err := os.ReadFile(iniPath)
	if err != nil {
		return
	}
	logFile := paths.SoftwareLog(s.base, record.Kind, record.Version)
	if err := os.MkdirAll(filepath.Dir(logFile), 0o755); err != nil {
		return
	}
	text := strings.ReplaceAll(string(raw), ";error_log = php_errors.log", fmt.Sprintf(`error_log = "%s"`, logFile))
	if err := os.WriteFile(iniPath, []byte(text), 0o644); err != nil {
		return
	}
}

func (s *SoftwareService) Cancel(ctx context.Context, kind, version string) error {
	taskKey := strings.TrimSpace(kind) + ":" + strings.TrimSpace(version)
	s.mu.Lock()
	cancel, ok := s.cancels[taskKey]
	s.mu.Unlock()
	if !ok {
		return appErrors.New(appErrors.TASK_NOT_CANCELLABLE, "未找到正在运行的安装任务")
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
		return appErrors.New(appErrors.NOT_FOUND, "软件版本不存在")
	}

	if software.IsDaemon(item.Kind) {
		status := s.runner.Status(daemonKey(item.ID))
		if status.Running {
			return appErrors.New(appErrors.SERVICE_RUNNING, fmt.Sprintf("%s %s 正在运行中，请先停止后再卸载", item.Kind, item.Version))
		}
	}

	boundSites, err := s.sites.ListBySoftware(ctx, item.ID)
	if err != nil {
		return err
	}
	if len(boundSites) > 0 {
		names := make([]string, 0, len(boundSites))
		for _, bound := range boundSites {
			names = append(names, bound.Name)
		}
		return appErrors.New(
			appErrors.RUNTIME_IN_USE,
			fmt.Sprintf("该版本正被以下网站使用，请先修改网站版本后再卸载: %s", strings.Join(names, "、")),
		)
	}

	if item.Path != "" {
		if err := os.RemoveAll(item.Path); err != nil {
			return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "删除软件安装目录失败")
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
			for _, candidate := range remaining {
				if candidate.State == software.STATE_INSTALLED {
					candidates = append(candidates, candidate)
				}
			}
			if len(candidates) > 0 {
				sort.Slice(candidates, func(i, j int) bool {
					return candidates[i].Version > candidates[j].Version
				})
				if setErr := s.repo.SetMain(ctx, kind, candidates[0].ID); setErr != nil {
					return nil
				}
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
		return appErrors.New(appErrors.NOT_FOUND, "软件版本不存在")
	}
	if item.State != software.STATE_INSTALLED {
		return appErrors.New(appErrors.INVALID_ARGUMENT, "只有已安装完成的版本才能设为默认")
	}
	return s.repo.SetMain(ctx, item.Kind, item.ID)
}

func (s *SoftwareService) Start(ctx context.Context, id int64) error {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if item == nil {
		return appErrors.New(appErrors.NOT_FOUND, "软件版本不存在")
	}
	if !software.IsDaemon(item.Kind) {
		return appErrors.New(appErrors.INVALID_ARGUMENT, "该软件不是常驻服务")
	}
	if item.State != software.STATE_INSTALLED {
		return appErrors.New(appErrors.INVALID_ARGUMENT, "软件尚未安装完成")
	}

	key := daemonKey(item.ID)
	if status := s.runner.Status(key); status.Running {
		return nil
	}

	if !s.probe.Available(item.Port) {
		return appErrors.New(appErrors.PORT_ALREADY_IN_USE, fmt.Sprintf("端口 %d 已被占用", item.Port))
	}

	options, err := s.buildOptions(ctx, *item)
	if err != nil {
		return err
	}

	softwareID := item.ID
	options.OnExit = func(_ string, message string) {
		s.bus.Emit("service:status-changed", map[string]interface{}{
			"id":      softwareID,
			"running": false,
			"pid":     0,
			"message": message,
		})
	}

	pid, err := s.runner.Start(key, options)
	if err != nil {
		return err
	}

	deadline := time.Now().Add(START_TIMEOUT)
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
		status := s.runner.Status(key)
		if !status.Running {
			break
		}
		time.Sleep(POLL_INTERVAL)
	}

	if stopErr := s.runner.Stop(key); stopErr != nil {
		return appErrors.Wrap(stopErr, appErrors.PROCESS_START_FAILED, "停止超时服务失败")
	}
	tail := process.ReadTail(options.Log, LOG_TAIL_LIMIT)
	detail := strings.Join(tail, "\n")
	return appErrors.New(appErrors.PROCESS_START_FAILED, "服务启动超时或退出:\n"+detail)
}

func (s *SoftwareService) Stop(ctx context.Context, id int64) error {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if item == nil {
		return appErrors.New(appErrors.NOT_FOUND, "软件版本不存在")
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

func (s *SoftwareService) buildOptions(ctx context.Context, item software.Software) (ports.Options, error) {
	logPath := paths.SoftwareLog(s.base, item.Kind, item.Version)
	switch item.Kind {
	case software.KIND_MYSQL:
		if err := s.prepareMySQL(ctx, item, logPath); err != nil {
			return ports.Options{}, err
		}
		iniPath := filepath.Join(item.Path, "my.ini")
		return ports.Options{
			Binary:    filepath.Join(item.Path, "bin", "mysqld.exe"),
			Args:      []string{"--defaults-file=" + iniPath, "--console"},
			Directory: item.Path,
			Log:       logPath,
		}, nil
	case software.KIND_MARIADB:
		binary, iniPath, err := s.prepareMariaDB(ctx, item, logPath)
		if err != nil {
			return ports.Options{}, err
		}
		return ports.Options{
			Binary:    binary,
			Args:      []string{"--defaults-file=" + iniPath, "--console"},
			Directory: item.Path,
			Log:       logPath,
		}, nil
	case software.KIND_POSTGRESQL:
		args, err := s.preparePostgres(ctx, item, logPath)
		if err != nil {
			return ports.Options{}, err
		}
		return ports.Options{
			Binary:    filepath.Join(item.Path, "bin", "postgres.exe"),
			Args:      args,
			Directory: item.Path,
			Log:       logPath,
		}, nil
	case software.KIND_MONGODB:
		args, err := s.prepareMongoDB(item)
		if err != nil {
			return ports.Options{}, err
		}
		return ports.Options{
			Binary:    filepath.Join(item.Path, "bin", "mongod.exe"),
			Args:      args,
			Directory: item.Path,
			Log:       logPath,
		}, nil
	case software.KIND_REDIS:
		binary := installer.FindRedis(item.Path)
		if binary == "" {
			return ports.Options{}, appErrors.New(appErrors.PROCESS_START_FAILED, "未找到 redis-server.exe")
		}
		configPath, err := s.prepareRedis(item, filepath.Dir(binary))
		if err != nil {
			return ports.Options{}, err
		}
		return ports.Options{
			Binary:    binary,
			Args:      []string{filepath.Base(configPath)},
			Directory: filepath.Dir(binary),
			Log:       logPath,
		}, nil
	case software.KIND_MEMCACHED:
		binary, args, err := s.prepareMemcached(item)
		if err != nil {
			return ports.Options{}, err
		}
		return ports.Options{
			Binary:    binary,
			Args:      args,
			Directory: filepath.Dir(binary),
			Log:       logPath,
		}, nil
	}
	return ports.Options{}, appErrors.New(appErrors.INVALID_ARGUMENT, "不支持的服务类型")
}

func (s *SoftwareService) SetAutostart(ctx context.Context, id int64, on bool) error {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if item == nil {
		return appErrors.New(appErrors.NOT_FOUND, "软件版本不存在")
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
		return software.Config{}, appErrors.New(appErrors.NOT_FOUND, "软件版本不存在")
	}

	filePath, err := s.prepareConfig(*item)
	if err != nil {
		return software.Config{}, err
	}
	if filePath == "" {
		return software.Config{Items: make([]software.ConfigItem, 0), File: ""}, nil
	}

	raw, err := os.ReadFile(filePath)
	if err != nil {
		return software.Config{}, appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "读取配置文件失败")
	}
	text := string(raw)

	schema := configSchema(item.Kind)
	items := make([]software.ConfigItem, 0, len(schema))
	for _, field := range schema {
		val := readValue(item.Kind, text, field.Key)
		items = append(items, software.ConfigItem{
			Key:         field.Key,
			Label:       field.Label,
			Value:       val,
			Description: field.Description,
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
		return false, appErrors.New(appErrors.NOT_FOUND, "软件版本不存在")
	}

	filePath, err := s.prepareConfig(*item)
	if err != nil {
		return false, err
	}
	if filePath == "" {
		return false, appErrors.New(appErrors.INVALID_ARGUMENT, "该软件没有配置文件")
	}

	raw, err := os.ReadFile(filePath)
	if err != nil {
		return false, appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "读取配置文件失败")
	}
	if err := os.WriteFile(filePath+".bak", raw, 0o644); err != nil {
		return false, appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "备份配置文件失败")
	}

	text := string(raw)
	for key, val := range values {
		text = writeValue(item.Kind, text, key, val)
	}

	if err := os.WriteFile(filePath, []byte(text), 0o644); err != nil {
		return false, appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "保存配置文件失败")
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
		return "", appErrors.New(appErrors.NOT_FOUND, "软件版本不存在")
	}
	filePath, err := s.prepareConfig(*item)
	if err != nil {
		return "", err
	}
	if filePath == "" {
		return "", appErrors.New(appErrors.INVALID_ARGUMENT, "该软件没有配置文件")
	}
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return "", appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "读取配置文件失败")
	}
	return string(raw), nil
}

func (s *SoftwareService) WriteConfig(ctx context.Context, id int64, text string) error {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if item == nil {
		return appErrors.New(appErrors.NOT_FOUND, "软件版本不存在")
	}
	filePath, err := s.prepareConfig(*item)
	if err != nil {
		return err
	}
	if filePath == "" {
		return appErrors.New(appErrors.INVALID_ARGUMENT, "该软件没有配置文件")
	}
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "读取原配置文件失败")
	}
	if err := os.WriteFile(filePath+".bak", raw, 0o644); err != nil {
		return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "备份配置文件失败")
	}
	if err := os.WriteFile(filePath, []byte(text), 0o644); err != nil {
		return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "写入配置文件失败")
	}
	return nil
}

func (s *SoftwareService) Extensions(ctx context.Context, id int64) ([]software.Extension, error) {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil || item.Kind != software.KIND_PHP {
		return nil, appErrors.New(appErrors.INVALID_ARGUMENT, "仅 PHP 支持扩展管理")
	}
	iniPath, err := s.prepareConfig(*item)
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(iniPath)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "读取 php.ini 失败")
	}
	iniText := string(raw)

	extensionFolder := filepath.Join(item.Path, "ext")
	entries, err := os.ReadDir(extensionFolder)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "读取 PHP ext 目录失败")
	}

	var list []software.Extension
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		lower := strings.ToLower(entry.Name())
		if !strings.HasPrefix(lower, "php_") || !strings.HasSuffix(lower, ".dll") {
			continue
		}
		name := strings.TrimSuffix(strings.TrimPrefix(lower, "php_"), ".dll")
		enabled := checkExtension(iniText, name)
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
	if item == nil || item.Kind != software.KIND_PHP {
		return appErrors.New(appErrors.INVALID_ARGUMENT, "仅 PHP 支持扩展管理")
	}
	iniPath, err := s.prepareConfig(*item)
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(iniPath)
	if err != nil {
		return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "读取 php.ini 失败")
	}
	if err := os.WriteFile(iniPath+".bak", raw, 0o644); err != nil {
		return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "备份 php.ini 失败")
	}

	updated := switchExtension(string(raw), strings.TrimSpace(name), on)
	if err := os.WriteFile(iniPath, []byte(updated), 0o644); err != nil {
		return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "写入 php.ini 失败")
	}
	return nil
}

func (s *SoftwareService) OpenFolder(ctx context.Context, id int64) error {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if item == nil {
		return appErrors.New(appErrors.NOT_FOUND, "软件版本不存在")
	}
	target := item.Path
	if software.HasData(item.Kind) {
		dataFolder := paths.DatabaseData(s.base, item.Kind, item.Version)
		if info, statErr := os.Stat(dataFolder); statErr == nil && info.IsDir() {
			target = dataFolder
		}
	}
	if info, statErr := os.Stat(target); statErr != nil || !info.IsDir() {
		return appErrors.New(appErrors.PATH_NOT_ACCESSIBLE, "目录不存在: "+target)
	}
	command := exec.Command("explorer.exe", target)
	return command.Start()
}

func (s *SoftwareService) Logs(ctx context.Context, id int64, limit int) ([]string, error) {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, appErrors.New(appErrors.NOT_FOUND, "软件版本不存在")
	}
	logPath := paths.SoftwareLog(s.base, item.Kind, item.Version)
	return process.ReadTail(logPath, limit), nil
}

func (s *SoftwareService) StopAll() error {
	list, err := s.repo.List(context.Background())
	if err != nil {
		return err
	}
	stopErrors := make([]error, 0)
	for _, item := range list {
		if software.IsDaemon(item.Kind) {
			if err := s.runner.Stop(daemonKey(item.ID)); err != nil {
				stopErrors = append(stopErrors, err)
			}
		}
	}
	return errors.Join(stopErrors...)
}

func (s *SoftwareService) findPort(ctx context.Context, kind string) int {
	basePort := MYSQL_START_PORT
	switch kind {
	case software.KIND_MARIADB:
		basePort = MARIADB_START_PORT
	case software.KIND_POSTGRESQL:
		basePort = POSTGRES_START_PORT
	case software.KIND_MONGODB:
		basePort = MONGODB_START_PORT
	case software.KIND_REDIS:
		basePort = REDIS_START_PORT
	case software.KIND_MEMCACHED:
		basePort = MEMCACHED_START_PORT
	}
	existing, err := s.repo.ListByKind(ctx, kind)
	if err != nil {
		existing = nil
	}
	used := make(map[int]bool, len(existing))
	for _, item := range existing {
		if item.Port > 0 {
			used[item.Port] = true
		}
	}
	for port := basePort; port < basePort+PORT_SPAN; port++ {
		if !used[port] && s.probe.Available(port) {
			return port
		}
	}
	return basePort
}

func (s *SoftwareService) prepareMySQL(ctx context.Context, item software.Software, logPath string) error {
	dataFolder := paths.DatabaseData(s.base, item.Kind, item.Version)
	if err := os.MkdirAll(dataFolder, 0o755); err != nil {
		return appErrors.Wrap(err, appErrors.PROCESS_START_FAILED, "创建 MySQL 数据目录失败")
	}
	if _, err := s.prepareConfig(item); err != nil {
		return err
	}

	systemFolder := filepath.Join(dataFolder, "mysql")
	if info, err := os.Stat(systemFolder); err != nil || !info.IsDir() {
		iniPath := filepath.Join(item.Path, "my.ini")
		mysqld := filepath.Join(item.Path, "bin", "mysqld.exe")
		if err := s.runInit(ctx, item.Path, logPath, "MySQL", mysqld, "--defaults-file="+iniPath, "--initialize-insecure", "--console"); err != nil {
			return err
		}
	}
	return nil
}

func (s *SoftwareService) prepareMariaDB(ctx context.Context, item software.Software, logPath string) (string, string, error) {
	dataFolder := paths.DatabaseData(s.base, item.Kind, item.Version)
	if err := os.MkdirAll(dataFolder, 0o755); err != nil {
		return "", "", appErrors.Wrap(err, appErrors.PROCESS_START_FAILED, "创建 MariaDB 数据目录失败")
	}
	iniPath, err := s.prepareConfig(item)
	if err != nil {
		return "", "", err
	}

	binary := filepath.Join(item.Path, "bin", "mariadbd.exe")
	if _, statErr := os.Stat(binary); statErr != nil {
		binary = filepath.Join(item.Path, "bin", "mysqld.exe")
	}

	systemFolder := filepath.Join(dataFolder, "mysql")
	if info, statErr := os.Stat(systemFolder); statErr != nil || !info.IsDir() {
		installBin := filepath.Join(item.Path, "bin", "mysql_install_db.exe")
		if err := s.runInit(ctx, item.Path, logPath, "MariaDB", installBin, "--datadir="+dataFolder); err != nil {
			return "", "", err
		}
	}
	return binary, iniPath, nil
}

func (s *SoftwareService) preparePostgres(ctx context.Context, item software.Software, logPath string) ([]string, error) {
	dataFolder := paths.DatabaseData(s.base, item.Kind, item.Version)
	versionFile := filepath.Join(dataFolder, "PG_VERSION")
	if _, statErr := os.Stat(versionFile); statErr != nil {
		if err := os.RemoveAll(dataFolder); err != nil {
			return nil, appErrors.Wrap(err, appErrors.PROCESS_START_FAILED, "清理 PostgreSQL 数据目录失败")
		}
		if err := os.MkdirAll(filepath.Dir(dataFolder), 0o755); err != nil {
			return nil, appErrors.Wrap(err, appErrors.PROCESS_START_FAILED, "创建 PostgreSQL 数据父目录失败")
		}
		initBin := filepath.Join(item.Path, "bin", "initdb.exe")
		if err := s.runInit(ctx, item.Path, logPath, "PostgreSQL", initBin, "-D", dataFolder, "-U", "postgres", "-A", "trust", "-E", "UTF8", "--no-locale"); err != nil {
			return nil, err
		}
	}

	confPath, err := s.prepareConfig(item)
	if err != nil {
		return nil, err
	}
	port := item.Port
	if port <= 0 {
		port = POSTGRES_START_PORT
	}
	args := []string{"-D", dataFolder, "-p", strconv.Itoa(port), "-h", paths.LOOPBACK}
	if raw, readErr := os.ReadFile(confPath); readErr == nil {
		text := string(raw)
		for _, key := range []string{"max_connections", "shared_buffers", "work_mem"} {
			val := strings.TrimSpace(readValue(software.KIND_POSTGRESQL, text, key))
			if val != "" {
				args = append(args, "-c", fmt.Sprintf("%s=%s", key, val))
			}
		}
	}
	return args, nil
}

func (s *SoftwareService) prepareMongoDB(item software.Software) ([]string, error) {
	dataFolder := paths.DatabaseData(s.base, item.Kind, item.Version)
	if err := os.MkdirAll(dataFolder, 0o755); err != nil {
		return nil, appErrors.Wrap(err, appErrors.PROCESS_START_FAILED, "创建 MongoDB 数据目录失败")
	}
	confPath, err := s.prepareConfig(item)
	if err != nil {
		return nil, err
	}
	port := item.Port
	if port <= 0 {
		port = MONGODB_START_PORT
	}
	bindIP := paths.LOOPBACK
	cacheSize := "1"
	raw, err := os.ReadFile(confPath)
	if err == nil {
		text := string(raw)
		if val := strings.TrimSpace(readValue(software.KIND_MONGODB, text, "bind_ip")); val != "" {
			bindIP = val
		}
		if val := strings.TrimSpace(readValue(software.KIND_MONGODB, text, "wiredTigerCacheSizeGB")); val != "" {
			cacheSize = val
		}
	}
	return []string{
		"--dbpath", dataFolder,
		"--bind_ip", bindIP,
		"--port", strconv.Itoa(port),
		"--wiredTigerCacheSizeGB", cacheSize,
	}, nil
}

func (s *SoftwareService) prepareMemcached(item software.Software) (string, []string, error) {
	binary := installer.FindBinary(item.Path, "memcached.exe")
	if binary == "" {
		return "", nil, appErrors.New(appErrors.PROCESS_START_FAILED, "未找到 memcached.exe")
	}
	confPath, err := s.prepareConfig(item)
	if err != nil {
		return "", nil, err
	}
	port := item.Port
	if port <= 0 {
		port = MEMCACHED_START_PORT
	}
	memory := "64"
	connections := "1024"
	threads := "4"
	if raw, readErr := os.ReadFile(confPath); readErr == nil {
		text := string(raw)
		if val := strings.TrimSpace(readValue(software.KIND_MEMCACHED, text, "memory")); val != "" {
			memory = val
		}
		if val := strings.TrimSpace(readValue(software.KIND_MEMCACHED, text, "max_connections")); val != "" {
			connections = val
		}
		if val := strings.TrimSpace(readValue(software.KIND_MEMCACHED, text, "threads")); val != "" {
			threads = val
		}
	}
	args := []string{"-l", paths.LOOPBACK, "-p", strconv.Itoa(port), "-m", memory, "-c", connections, "-t", threads}
	return binary, args, nil
}

func (s *SoftwareService) runInit(ctx context.Context, folder, logPath, label, binary string, args ...string) error {
	command := exec.CommandContext(ctx, binary, args...)
	command.Dir = folder
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, initErr := command.CombinedOutput()
	if len(output) > 0 {
		if mkErr := os.MkdirAll(filepath.Dir(logPath), 0o755); mkErr == nil {
			if file, openErr := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644); openErr == nil {
				if _, writeErr := file.Write(output); writeErr != nil {
					if closeErr := file.Close(); closeErr != nil {
						return appErrors.Wrap(closeErr, appErrors.PROCESS_START_FAILED, "关闭日志文件失败")
					}
				} else if closeErr := file.Close(); closeErr != nil {
					return appErrors.Wrap(closeErr, appErrors.PROCESS_START_FAILED, "关闭日志文件失败")
				}
			}
		}
	}
	if initErr != nil {
		return appErrors.Wrap(initErr, appErrors.PROCESS_START_FAILED, "初始化 "+label+" 数据目录失败: "+string(output))
	}
	return nil
}

func (s *SoftwareService) prepareRedis(item software.Software, folder string) (string, error) {
	configPath := filepath.Join(folder, "redis.conf")
	if item.Port <= 0 {
		return "", appErrors.New(appErrors.INVALID_ARGUMENT, "Redis 端口无效")
	}
	raw, err := os.ReadFile(configPath)
	if err == nil {
		updated := writeValue(software.KIND_REDIS, string(raw), "port", strconv.Itoa(item.Port))
		updated = writeValue(software.KIND_REDIS, updated, "daemonize", "no")
		updated = writeValue(software.KIND_REDIS, updated, "logfile", `""`)
		if err := os.WriteFile(configPath, []byte(updated), 0o644); err != nil {
			return "", appErrors.Wrap(err, appErrors.PROCESS_START_FAILED, "更新 redis.conf 失败")
		}
		return configPath, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "读取 redis.conf 失败")
	}
	content := fmt.Sprintf("bind %s\r\nport %d\r\ndaemonize no\r\nlogfile \"\"\r\nmaxmemory 0\r\nappendonly no\r\n", paths.LOOPBACK, item.Port)
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		return "", appErrors.Wrap(err, appErrors.PROCESS_START_FAILED, "生成 redis.conf 失败")
	}
	return configPath, nil
}

func (s *SoftwareService) prepareConfig(item software.Software) (string, error) {
	switch item.Kind {
	case software.KIND_PHP:
		iniPath := filepath.Join(item.Path, "php.ini")
		missing, err := checkMissing(iniPath)
		if err != nil {
			return "", appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "检查 php.ini 失败")
		}
		if missing {
			sourceIni := filepath.Join(item.Path, "php.ini-development")
			raw, err := os.ReadFile(sourceIni)
			if err != nil {
				return "", appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "读取 php.ini-development 失败")
			}
			if err := os.WriteFile(iniPath, raw, 0o644); err != nil {
				return "", appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "生成 php.ini 失败")
			}
		}
		return iniPath, nil
	case software.KIND_MYSQL, software.KIND_MARIADB:
		iniPath := filepath.Join(item.Path, "my.ini")
		missing, err := checkMissing(iniPath)
		if err != nil {
			return "", appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "检查 my.ini 失败")
		}
		if missing {
			dataFolder := filepath.ToSlash(paths.DatabaseData(s.base, item.Kind, item.Version))
			baseFolder := filepath.ToSlash(item.Path)
			content := fmt.Sprintf("[mysqld]\r\nport=%d\r\nbind-address=%s\r\nbasedir=\"%s\"\r\ndatadir=\"%s\"\r\ncharacter-set-server=utf8mb4\r\nmax_connections=200\r\ninnodb_buffer_pool_size=128M\r\n",
				item.Port, paths.LOOPBACK, baseFolder, dataFolder,
			)
			if err := os.WriteFile(iniPath, []byte(content), 0o644); err != nil {
				return "", appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "生成 my.ini 失败")
			}
		}
		return iniPath, nil
	case software.KIND_POSTGRESQL:
		confPath := filepath.Join(item.Path, "postgresql.conf")
		missing, err := checkMissing(confPath)
		if err != nil {
			return "", appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "检查 postgresql.conf 失败")
		}
		if missing {
			content := "max_connections = 100\r\nshared_buffers = 128MB\r\nwork_mem = 4MB\r\n"
			if err := os.WriteFile(confPath, []byte(content), 0o644); err != nil {
				return "", appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "生成 postgresql.conf 失败")
			}
		}
		return confPath, nil
	case software.KIND_MONGODB:
		confPath := filepath.Join(item.Path, "mongod.conf")
		missing, err := checkMissing(confPath)
		if err != nil {
			return "", appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "检查 mongod.conf 失败")
		}
		if missing {
			content := fmt.Sprintf("bind_ip = %s\r\nmaxConns = 1000\r\nwiredTigerCacheSizeGB = 1\r\n", paths.LOOPBACK)
			if err := os.WriteFile(confPath, []byte(content), 0o644); err != nil {
				return "", appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "生成 mongod.conf 失败")
			}
		}
		return confPath, nil
	case software.KIND_REDIS:
		binary := installer.FindRedis(item.Path)
		folder := item.Path
		if binary != "" {
			folder = filepath.Dir(binary)
		}
		return s.prepareRedis(item, folder)
	case software.KIND_MEMCACHED:
		confPath := filepath.Join(item.Path, "memcached.conf")
		missing, err := checkMissing(confPath)
		if err != nil {
			return "", appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "检查 memcached.conf 失败")
		}
		if missing {
			content := "memory = 64\r\nmax_connections = 1024\r\nthreads = 4\r\n"
			if err := os.WriteFile(confPath, []byte(content), 0o644); err != nil {
				return "", appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "生成 memcached.conf 失败")
			}
		}
		return confPath, nil
	default:
		return "", nil
	}
}

func checkMissing(path string) (bool, error) {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	if info.IsDir() {
		return false, fmt.Errorf("配置文件路径指向目录: %s", path)
	}
	return false, nil
}

func daemonKey(id int64) string {
	return "software:" + strconv.FormatInt(id, 10)
}

func configSchema(kind string) []software.ConfigItem {
	switch kind {
	case software.KIND_PHP:
		return []software.ConfigItem{
			{Key: "upload_max_filesize", Label: "最大上传文件", Description: "允许上传的单个文件最大体积，如 50M"},
			{Key: "post_max_size", Label: "POST 最大体积", Description: "POST 请求体最大体积，如 50M"},
			{Key: "memory_limit", Label: "内存限制", Description: "单脚本最大内存消耗，如 256M"},
			{Key: "max_execution_time", Label: "最大执行时间", Description: "脚本最长执行秒数，如 300"},
			{Key: "date.timezone", Label: "默认时区", Description: "PHP 日期与时间函数默认时区"},
			{Key: "display_errors", Label: "显示错误", Description: "是否输出错误信息 (On / Off)"},
		}
	case software.KIND_MYSQL, software.KIND_MARIADB:
		return []software.ConfigItem{
			{Key: "max_connections", Label: "最大连接数", Description: "允许同时连接的客户端总数"},
			{Key: "innodb_buffer_pool_size", Label: "缓冲池大小", Description: "InnoDB 缓冲池内存大小，如 128M"},
			{Key: "character-set-server", Label: "服务端字符集", Description: "默认字符集，推荐 utf8mb4"},
		}
	case software.KIND_POSTGRESQL:
		return []software.ConfigItem{
			{Key: "max_connections", Label: "最大连接数", Description: "允许同时连接的客户端总数，如 100"},
			{Key: "shared_buffers", Label: "共享缓冲区", Description: "共享内存缓冲区大小，如 128MB"},
			{Key: "work_mem", Label: "工作内存", Description: "排序与哈希操作内存上限，如 4MB"},
		}
	case software.KIND_MONGODB:
		return []software.ConfigItem{
			{Key: "maxConns", Label: "最大连接数", Description: "允许同时连接的客户端总数，如 1000"},
			{Key: "wiredTigerCacheSizeGB", Label: "缓存大小 (GB)", Description: "WiredTiger 存储引擎缓存大小，如 1"},
			{Key: "bind_ip", Label: "绑定地址", Description: "监听网卡地址，默认 " + paths.LOOPBACK},
		}
	case software.KIND_REDIS:
		return []software.ConfigItem{
			{Key: "maxmemory", Label: "最大内存", Description: "最大内存上限（0 表示不限制，如 256mb）"},
			{Key: "requirepass", Label: "访问密码", Description: "客户端连接密码，留空表示无密码"},
			{Key: "appendonly", Label: "AOF 持久化", Description: "是否开启 AOF 日志持久化 (yes / no)"},
		}
	case software.KIND_MEMCACHED:
		return []software.ConfigItem{
			{Key: "memory", Label: "最大内存 (MB)", Description: "分配给 Memcached 的最大内存兆数，如 64"},
			{Key: "max_connections", Label: "最大连接数", Description: "允许同时连接的客户端总数，如 1024"},
			{Key: "threads", Label: "工作线程数", Description: "处理请求的工作线程数量，如 4"},
		}
	default:
		return make([]software.ConfigItem, 0)
	}
}

func readValue(kind, text, key string) string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if kind == software.KIND_REDIS {
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

	pattern := regexp.MustCompile(`(?i)^\s*` + regexp.QuoteMeta(key) + `\s*=\s*(.*)$`)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, ";") || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if matched := pattern.FindStringSubmatch(trimmed); matched != nil {
			val := matched[1]
			if commentPos := strings.IndexAny(val, "#;"); commentPos >= 0 {
				val = val[:commentPos]
			}
			val = strings.TrimSpace(val)
			return strings.Trim(val, `"'`)
		}
	}
	return ""
}

func writeValue(kind, text, key, value string) string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	replaced := false

	if kind == software.KIND_REDIS {
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

func checkExtension(iniText, name string) bool {
	pattern := regexp.MustCompile(`(?im)^\s*extension\s*=\s*(?:php_)?` + regexp.QuoteMeta(name) + `(?:\.dll)?\s*(?:;.*)?$`)
	return pattern.MatchString(iniText)
}

func switchExtension(iniText, name string, on bool) string {
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
