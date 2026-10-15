package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"devenv/internal/domain/site"
	"devenv/internal/domain/software"
	"devenv/internal/infra/process"
	"devenv/internal/ports"
	appErrors "devenv/internal/shared/errors"
	"devenv/internal/shared/paths"
)

const (
	NODE_START_PORT = 3000
	NODE_END_PORT   = 4000
	DEFAULT_HTTP    = 80
	START_TIMEOUT   = 30 * time.Second
	POLL_INTERVAL   = 250 * time.Millisecond
	LOG_TAIL_LIMIT  = 20
)

type SiteView struct {
	site.Site
	Path    string
	Version string
	State   string
	URL     string
	Need    string
}

type SiteService struct {
	base     string
	sites    ports.SiteRepository
	software ports.SoftwareRepository
	settings ports.SettingRepository
	nginx    ports.Gateway
	runner   ports.Runner
	hosts    ports.HostsWriter
	probe    ports.PortChecker
	bus      ports.EventPublisher

	mu     sync.RWMutex
	active map[int64]bool
	failed map[int64]string
}

func NewSiteService(
	base string,
	sites ports.SiteRepository,
	softwareRepo ports.SoftwareRepository,
	settings ports.SettingRepository,
	nginx ports.Gateway,
	runner ports.Runner,
	hosts ports.HostsWriter,
	probe ports.PortChecker,
	bus ports.EventPublisher,
) *SiteService {
	return &SiteService{
		base:     base,
		sites:    sites,
		software: softwareRepo,
		settings: settings,
		nginx:    nginx,
		runner:   runner,
		hosts:    hosts,
		probe:    probe,
		bus:      bus,
		active:   make(map[int64]bool),
		failed:   make(map[int64]string),
	}
}

func (s *SiteService) StartAuto(ctx context.Context) {
	list, err := s.sites.List(ctx)
	if err != nil {
		log.Printf("读取自动启动站点失败: %v", err)
		return
	}
	for _, item := range list {
		if item.Autostart {
			if err := s.Start(ctx, item.ID); err != nil {
				log.Printf("自动启动站点 %s 失败: %v", item.Name, err)
				continue
			}
		}
	}
	if err := s.syncRoutes(ctx); err != nil {
		log.Printf("刷新自动启动站点路由失败: %v", err)
	}
	if err := s.syncHosts(ctx); err != nil {
		log.Printf("同步自动启动站点域名失败: %v", err)
		return
	}
}

func (s *SiteService) BindInstalled(ctx context.Context, installed software.Software) {
	var siteKind string
	switch installed.Kind {
	case software.KIND_PHP:
		siteKind = site.KIND_PHP
	case software.KIND_NODE:
		siteKind = site.KIND_NODE
	default:
		return
	}
	missing, err := s.sites.ListMissingSoftware(ctx, siteKind)
	if err != nil {
		log.Printf("读取缺少运行环境的站点失败: %v", err)
		return
	}
	for _, item := range missing {
		if err := s.sites.BindSoftware(ctx, item.ID, installed.ID); err != nil {
			log.Printf("绑定站点 %s 运行环境失败: %v", item.Name, err)
			continue
		}
		if err := s.Start(ctx, item.ID); err != nil {
			log.Printf("启动新绑定环境的站点 %s 失败: %v", item.Name, err)
		}
	}
	if err := s.syncRoutes(ctx); err != nil {
		log.Printf("刷新新绑定环境的站点路由失败: %v", err)
	}
}

func (s *SiteService) List(ctx context.Context) ([]SiteView, error) {
	list, err := s.sites.List(ctx)
	if err != nil {
		return nil, err
	}
	softwareList, err := s.software.List(ctx)
	if err != nil {
		return nil, err
	}
	softwareMap := make(map[int64]software.Software, len(softwareList))
	for _, item := range softwareList {
		softwareMap[item.ID] = item
	}

	current, err := s.settings.Load(ctx)
	if err != nil {
		return nil, err
	}
	port := current.Port

	views := make([]SiteView, 0, len(list))
	for _, item := range list {
		views = append(views, s.buildView(item, softwareMap, port))
	}
	return views, nil
}

func (s *SiteService) Get(ctx context.Context, id int64) (*SiteView, error) {
	item, err := s.sites.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, appErrors.New(appErrors.NOT_FOUND, "站点不存在")
	}
	softwareMap := make(map[int64]software.Software)
	if item.Software != nil {
		found, err := s.software.Get(ctx, *item.Software)
		if err != nil {
			return nil, err
		}
		if found != nil {
			softwareMap[found.ID] = *found
		}
	}
	settings, err := s.settings.Load(ctx)
	if err != nil {
		return nil, err
	}
	view := s.buildView(*item, softwareMap, settings.Port)
	return &view, nil
}

func (s *SiteService) FindRun(root, kind string) string {
	return site.FindRun(root, kind)
}

func (s *SiteService) requireNginx(ctx context.Context) error {
	installed, err := s.software.GetMain(ctx, software.KIND_NGINX)
	if err != nil {
		return err
	}
	if installed == nil || installed.State != software.STATE_INSTALLED {
		return appErrors.New(appErrors.RUNTIME_NOT_INSTALLED, "创建网站前请先在软件商城安装 Nginx")
	}
	executable := filepath.Join(installed.Path, "nginx.exe")
	info, err := os.Stat(executable)
	if err != nil {
		return appErrors.Wrap(err, appErrors.RUNTIME_NOT_INSTALLED, "读取 nginx.exe 失败")
	}
	if info.IsDir() {
		return appErrors.New(appErrors.RUNTIME_NOT_INSTALLED, "nginx.exe 路径指向了目录")
	}
	if s.nginx.Port() > 0 {
		return nil
	}
	settings, err := s.settings.Load(ctx)
	if err != nil {
		return err
	}
	return s.nginx.Start(settings.Port)
}

func (s *SiteService) deleteCreated(ctx context.Context, id int64, source error) error {
	cleanup := []error{source}
	if err := s.sites.Delete(ctx, id); err != nil {
		cleanup = append(cleanup, err)
	}
	if err := s.syncHosts(ctx); err != nil {
		cleanup = append(cleanup, err)
	}
	return errors.Join(cleanup...)
}

func (s *SiteService) Create(ctx context.Context, in site.CreateInput) (*SiteView, error) {
	kind := strings.TrimSpace(in.Kind)
	name := strings.TrimSpace(in.Name)
	host := strings.ToLower(strings.TrimSpace(in.Host))

	if !site.IsValidKind(kind) {
		return nil, appErrors.New(appErrors.INVALID_ARGUMENT, "不支持的站点类型")
	}
	if name == "" {
		return nil, appErrors.New(appErrors.INVALID_ARGUMENT, "站点名称不能为空")
	}
	if host == "" {
		return nil, appErrors.New(appErrors.INVALID_ARGUMENT, "站点域名不能为空")
	}
	if err := s.requireNginx(ctx); err != nil {
		return nil, err
	}

	root := strings.TrimSpace(in.Root)
	if root == "" {
		current, err := s.settings.Load(ctx)
		if err != nil {
			return nil, err
		}
		root = filepath.Join(current.Root, name)
	}
	root = filepath.Clean(root)
	if !filepath.IsAbs(root) {
		return nil, appErrors.New(appErrors.PATH_INVALID, "站点根目录必须是绝对路径")
	}

	runFolder, err := site.NormalizeRun(root, in.Run)
	if err != nil {
		return nil, appErrors.New(appErrors.PATH_INVALID, err.Error())
	}

	docPath := site.ResolvePath(root, runFolder)
	if err := os.MkdirAll(docPath, 0o755); err != nil {
		return nil, appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "创建站点目录失败")
	}

	if err := writeSample(kind, name, docPath); err != nil {
		return nil, appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "写入默认示例文件失败")
	}

	var softwareID *int64
	switch kind {
	case site.KIND_PHP:
		mainItem, err := s.software.GetMain(ctx, software.KIND_PHP)
		if err != nil {
			return nil, err
		}
		if mainItem != nil && mainItem.State == software.STATE_INSTALLED {
			id := mainItem.ID
			softwareID = &id
		}
	case site.KIND_NODE:
		mainItem, err := s.software.GetMain(ctx, software.KIND_NODE)
		if err != nil {
			return nil, err
		}
		if mainItem != nil && mainItem.State == software.STATE_INSTALLED {
			id := mainItem.ID
			softwareID = &id
		}
	}

	port := 0
	command := strings.TrimSpace(in.Command)
	if kind == site.KIND_NODE {
		port = s.findPort(ctx)
		if command == "" {
			command = site.FindCommand(docPath)
			if command == "" && docPath != root {
				command = site.FindCommand(root)
			}
		}
	}

	created := &site.Site{
		Name:      name,
		Kind:      kind,
		Hosts:     []string{host},
		Root:      root,
		Run:       runFolder,
		Software:  softwareID,
		Command:   command,
		Port:      port,
		Autostart: true,
	}

	if err := s.sites.Create(ctx, created); err != nil {
		return nil, err
	}

	if err := s.syncHosts(ctx); err != nil {
		return nil, s.deleteCreated(ctx, created.ID, err)
	}

	ready := kind == site.KIND_STATIC || (kind == site.KIND_PHP && softwareID != nil) || (kind == site.KIND_NODE && softwareID != nil && command != "")
	if ready {
		if err := s.Start(ctx, created.ID); err != nil {
			return nil, s.deleteCreated(ctx, created.ID, err)
		}
	}

	return s.Get(ctx, created.ID)
}

func (s *SiteService) Update(ctx context.Context, in site.UpdateInput) (*SiteView, error) {
	current, err := s.sites.Get(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, appErrors.New(appErrors.NOT_FOUND, "站点不存在")
	}

	original := *current
	original.Hosts = append([]string(nil), current.Hosts...)

	var cleanHosts []string
	seen := make(map[string]bool)
	for _, host := range in.Hosts {
		trimmed := strings.ToLower(strings.TrimSpace(host))
		if trimmed != "" && !seen[trimmed] {
			seen[trimmed] = true
			cleanHosts = append(cleanHosts, trimmed)
		}
	}
	if len(cleanHosts) == 0 {
		return nil, appErrors.New(appErrors.INVALID_ARGUMENT, "至少需要保留一个域名")
	}

	root := filepath.Clean(strings.TrimSpace(in.Root))
	if !filepath.IsAbs(root) {
		return nil, appErrors.New(appErrors.PATH_INVALID, "根目录必须是绝对路径")
	}

	runFolder, err := site.NormalizeRun(root, in.Run)
	if err != nil {
		return nil, appErrors.New(appErrors.PATH_INVALID, err.Error())
	}

	docPath := site.ResolvePath(root, runFolder)
	if err := os.MkdirAll(docPath, 0o755); err != nil {
		return nil, appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "创建运行目录失败")
	}
	nginxConfig := strings.TrimSpace(in.Nginx)
	if err := site.ValidateNginxConfig(nginxConfig); err != nil {
		return nil, appErrors.New(appErrors.INVALID_ARGUMENT, err.Error())
	}

	current.Hosts = cleanHosts
	current.Root = root
	current.Run = runFolder
	current.Software = in.Software
	current.Command = strings.TrimSpace(in.Command)
	current.Nginx = nginxConfig
	current.Autostart = in.Autostart

	if err := s.sites.Update(ctx, current); err != nil {
		return nil, err
	}

	if err := s.syncRoutes(ctx); err != nil {
		if restoreErr := s.sites.Update(ctx, &original); restoreErr != nil {
			return nil, errors.Join(err, restoreErr)
		}
		return nil, err
	}
	if err := s.syncHosts(ctx); err != nil {
		if restoreErr := s.sites.Update(ctx, &original); restoreErr != nil {
			return nil, errors.Join(err, restoreErr)
		}
		if restoreErr := s.syncRoutes(ctx); restoreErr != nil {
			return nil, errors.Join(err, restoreErr)
		}
		return nil, err
	}

	return s.Get(ctx, current.ID)
}

func (s *SiteService) Start(ctx context.Context, id int64) error {
	current, err := s.sites.Get(ctx, id)
	if err != nil {
		return err
	}
	if current == nil {
		return appErrors.New(appErrors.NOT_FOUND, "站点不存在")
	}

	switch current.Kind {
	case site.KIND_STATIC:
		s.mu.Lock()
		s.active[current.ID] = true
		delete(s.failed, current.ID)
		s.mu.Unlock()

	case site.KIND_PHP:
		if current.Software == nil {
			return appErrors.New(appErrors.RUNTIME_NOT_INSTALLED, "站点未绑定 PHP 运行环境")
		}
		found, err := s.software.Get(ctx, *current.Software)
		if err != nil {
			return err
		}
		if found == nil || found.State != software.STATE_INSTALLED {
			return appErrors.New(appErrors.RUNTIME_NOT_INSTALLED, "PHP 运行环境未安装")
		}
		phpBinary := filepath.Join(found.Path, "php-cgi.exe")
		if info, err := os.Stat(phpBinary); err != nil || info.IsDir() {
			return appErrors.New(appErrors.RUNTIME_NOT_INSTALLED, "找不到 php-cgi.exe: "+phpBinary)
		}
		s.mu.Lock()
		s.active[current.ID] = true
		delete(s.failed, current.ID)
		s.mu.Unlock()

	case site.KIND_NODE:
		if current.Software == nil {
			return appErrors.New(appErrors.RUNTIME_NOT_INSTALLED, "站点未绑定 Node.js 运行环境")
		}
		found, err := s.software.Get(ctx, *current.Software)
		if err != nil {
			return err
		}
		if found == nil || found.State != software.STATE_INSTALLED {
			return appErrors.New(appErrors.RUNTIME_NOT_INSTALLED, "Node.js 运行环境未安装")
		}
		if strings.TrimSpace(current.Command) == "" {
			return appErrors.New(appErrors.INVALID_ARGUMENT, "未配置 Node 启动命令")
		}
		if err := s.startNode(*current, *found); err != nil {
			s.mu.Lock()
			delete(s.active, current.ID)
			s.failed[current.ID] = err.Error()
			s.mu.Unlock()
			return err
		}
		s.mu.Lock()
		s.active[current.ID] = true
		delete(s.failed, current.ID)
		s.mu.Unlock()
	}

	if err := s.syncRoutes(ctx); err != nil {
		s.mu.Lock()
		delete(s.active, current.ID)
		s.failed[current.ID] = err.Error()
		s.mu.Unlock()
		if current.Kind == site.KIND_NODE {
			if stopError := s.runner.Stop(siteKey(current.ID)); stopError != nil {
				return errors.Join(err, stopError)
			}
		}
		return err
	}
	s.bus.Emit("site:status-changed", map[string]interface{}{
		"id":      current.ID,
		"state":   site.STATE_RUNNING,
		"message": "",
	})
	return nil
}

func (s *SiteService) Stop(ctx context.Context, id int64) error {
	current, err := s.sites.Get(ctx, id)
	if err != nil {
		return err
	}
	if current == nil {
		return appErrors.New(appErrors.NOT_FOUND, "站点不存在")
	}

	if current.Kind == site.KIND_NODE {
		if err := s.runner.Stop(siteKey(current.ID)); err != nil {
			return err
		}
	}

	s.mu.Lock()
	delete(s.active, current.ID)
	delete(s.failed, current.ID)
	s.mu.Unlock()

	if err := s.syncRoutes(ctx); err != nil {
		return err
	}
	s.bus.Emit("site:status-changed", map[string]interface{}{
		"id":      current.ID,
		"state":   site.STATE_STOPPED,
		"message": "",
	})
	return nil
}

func (s *SiteService) Restart(ctx context.Context, id int64) error {
	if err := s.Stop(ctx, id); err != nil {
		return err
	}
	return s.Start(ctx, id)
}

func (s *SiteService) Delete(ctx context.Context, id int64) error {
	current, err := s.sites.Get(ctx, id)
	if err != nil {
		return err
	}
	if current == nil {
		return appErrors.New(appErrors.NOT_FOUND, "站点不存在")
	}

	if err := s.Stop(ctx, id); err != nil {
		return err
	}
	if err := s.sites.Delete(ctx, id); err != nil {
		return err
	}

	if err := s.syncRoutes(ctx); err != nil {
		return err
	}
	if err := s.syncHosts(ctx); err != nil {
		return err
	}
	return nil
}

func (s *SiteService) Logs(ctx context.Context, id int64, kind string, limit int) ([]string, error) {
	current, err := s.sites.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, appErrors.New(appErrors.NOT_FOUND, "站点不存在")
	}
	logFile := paths.SiteLog(s.base, current.Name, "access.log")
	if kind == "run" {
		logFile = paths.SiteLog(s.base, current.Name, "run.log")
	}
	return process.ReadTail(logFile, limit), nil
}

func (s *SiteService) OpenNginxConfig(ctx context.Context, id int64) error {
	item, err := s.sites.Get(ctx, id)
	if err != nil {
		return err
	}
	if item == nil {
		return appErrors.New(appErrors.NOT_FOUND, "网站不存在")
	}
	configPath := paths.NginxSiteConfig(s.base, item.ID, item.Name)
	info, err := os.Stat(configPath)
	if err != nil {
		return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "该网站的 Nginx 配置文件尚未生成，请先启动网站")
	}
	if info.IsDir() {
		return appErrors.New(appErrors.PATH_NOT_ACCESSIBLE, "网站 Nginx 配置文件路径指向了目录")
	}
	return exec.Command("explorer.exe", configPath).Start()
}

func (s *SiteService) StopAll() error {
	list, err := s.sites.List(context.Background())
	if err != nil {
		return err
	}
	stopErrors := make([]error, 0)
	for _, item := range list {
		if item.Kind == site.KIND_NODE {
			if err := s.runner.Stop(siteKey(item.ID)); err != nil {
				stopErrors = append(stopErrors, err)
			}
		}
	}
	return errors.Join(stopErrors...)
}

func (s *SiteService) startNode(current site.Site, runtimeItem software.Software) error {
	key := siteKey(current.ID)
	if status := s.runner.Status(key); status.Running {
		return nil
	}
	if !s.probe.Available(current.Port) {
		return appErrors.New(appErrors.PORT_ALREADY_IN_USE, fmt.Sprintf("Node 站点端口 %d 已被占用", current.Port))
	}

	portText := strconv.Itoa(current.Port)
	commandText := strings.ReplaceAll(current.Command, "%PORT%", portText)
	workFolder := site.ResolvePath(current.Root, current.Run)
	logPath := paths.SiteLog(s.base, current.Name, "run.log")
	envPath := runtimeItem.Path + ";" + os.Getenv("PATH")
	siteID := current.ID

	options := ports.Options{
		Binary:    paths.Shell(),
		Args:      []string{"/C", commandText},
		Directory: workFolder,
		Environment: []string{
			"PORT=" + portText,
			"PATH=" + envPath,
		},
		Log: logPath,
		OnExit: func(_ string, message string) {
			s.mu.Lock()
			delete(s.active, siteID)
			s.failed[siteID] = message
			s.mu.Unlock()
			if err := s.syncRoutes(context.Background()); err != nil {
				log.Printf("移除已退出 Node 站点路由失败: %v", err)
			}
			s.bus.Emit("site:status-changed", map[string]interface{}{
				"id":      siteID,
				"state":   site.STATE_ERROR,
				"message": message,
			})
		},
	}

	if _, err := s.runner.Start(key, options); err != nil {
		return err
	}

	deadline := time.Now().Add(START_TIMEOUT)
	for time.Now().Before(deadline) {
		if s.probe.Listening(current.Port) {
			return nil
		}
		status := s.runner.Status(key)
		if !status.Running {
			break
		}
		time.Sleep(POLL_INTERVAL)
	}

	if stopErr := s.runner.Stop(key); stopErr != nil {
		return appErrors.Wrap(stopErr, appErrors.PROCESS_START_FAILED, "停止超时 Node 站点失败")
	}
	tail := process.ReadTail(logPath, LOG_TAIL_LIMIT)
	return appErrors.New(appErrors.PROCESS_START_FAILED, "Node 站点启动超时或退出:\n"+strings.Join(tail, "\n"))
}

func (s *SiteService) syncRoutes(ctx context.Context) error {
	list, err := s.sites.List(ctx)
	if err != nil {
		return err
	}
	softwareList, err := s.software.List(ctx)
	if err != nil {
		return err
	}
	softwareMap := make(map[int64]software.Software, len(softwareList))
	for _, item := range softwareList {
		softwareMap[item.ID] = item
	}

	s.mu.RLock()
	activeCopy := make(map[int64]bool, len(s.active))
	for key, val := range s.active {
		activeCopy[key] = val
	}
	s.mu.RUnlock()

	routes := make([]ports.Route, 0)
	for _, item := range list {
		if !activeCopy[item.ID] {
			continue
		}
		phpBinary := ""
		if item.Kind == site.KIND_PHP && item.Software != nil {
			if found, ok := softwareMap[*item.Software]; ok && found.State == software.STATE_INSTALLED {
				phpBinary = filepath.Join(found.Path, "php-cgi.exe")
			}
		}
		routes = append(routes, ports.Route{
			ID:    item.ID,
			Name:  item.Name,
			Hosts: item.Hosts,
			Kind:  item.Kind,
			Path:  site.ResolvePath(item.Root, item.Run),
			PHP:   phpBinary,
			Nginx: item.Nginx,
			Port:  item.Port,
			Log:   paths.SiteLog(s.base, item.Name, "access.log"),
		})
	}
	return s.nginx.SetRoutes(routes)
}

func (s *SiteService) syncHosts(ctx context.Context) error {
	current, err := s.settings.Load(ctx)
	if err != nil {
		return err
	}
	if !current.Hosts {
		return nil
	}
	list, err := s.sites.List(ctx)
	if err != nil {
		return err
	}
	var allHosts []string
	for _, item := range list {
		allHosts = append(allHosts, item.Hosts...)
	}
	return s.hosts.Sync(allHosts)
}

func (s *SiteService) buildView(item site.Site, softwareMap map[int64]software.Software, port int) SiteView {
	docPath := site.ResolvePath(item.Root, item.Run)
	version := ""
	need := ""

	switch item.Kind {
	case site.KIND_PHP:
		if item.Software != nil {
			if found, ok := softwareMap[*item.Software]; ok && found.State == software.STATE_INSTALLED {
				version = found.Version
			} else {
				need = software.KIND_PHP
			}
		} else {
			need = software.KIND_PHP
		}
	case site.KIND_NODE:
		if item.Software != nil {
			if found, ok := softwareMap[*item.Software]; ok && found.State == software.STATE_INSTALLED {
				version = found.Version
			} else {
				need = software.KIND_NODE
			}
		} else {
			need = software.KIND_NODE
		}
	}

	s.mu.RLock()
	active := s.active[item.ID]
	_, hasFailed := s.failed[item.ID]
	s.mu.RUnlock()

	state := site.STATE_STOPPED
	if hasFailed {
		state = site.STATE_ERROR
	} else if active {
		if item.Kind == site.KIND_NODE {
			status := s.runner.Status(siteKey(item.ID))
			if status.Running {
				state = site.STATE_RUNNING
			} else if status.Message != "" {
				state = site.STATE_ERROR
			}
		} else {
			state = site.STATE_RUNNING
		}
	}

	siteURL := ""
	if len(item.Hosts) > 0 {
		if port > 0 && port != DEFAULT_HTTP {
			siteURL = fmt.Sprintf("http://%s:%d", item.Hosts[0], port)
		} else {
			siteURL = "http://" + item.Hosts[0]
		}
	}

	return SiteView{
		Site:    item,
		Path:    docPath,
		Version: version,
		State:   state,
		URL:     siteURL,
		Need:    need,
	}
}

func (s *SiteService) findPort(ctx context.Context) int {
	list, err := s.sites.List(ctx)
	if err != nil {
		list = nil
	}
	used := make(map[int]bool, len(list))
	for _, item := range list {
		if item.Port > 0 {
			used[item.Port] = true
		}
	}
	for port := NODE_START_PORT; port < NODE_END_PORT; port++ {
		if !used[port] && s.probe.Available(port) {
			return port
		}
	}
	return NODE_START_PORT
}

func siteKey(id int64) string {
	return "site:" + strconv.FormatInt(id, 10)
}

func writeSample(kind, name, folder string) error {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return err
	}
	if len(entries) > 0 {
		return nil
	}

	switch kind {
	case site.KIND_STATIC:
		content := fmt.Sprintf("<!DOCTYPE html>\n<html lang=\"zh-CN\">\n<head><meta charset=\"utf-8\"><title>%s</title></head>\n<body><h1>%s 运行正常</h1></body>\n</html>\n", name, name)
		return os.WriteFile(filepath.Join(folder, "index.html"), []byte(content), 0o644)
	case site.KIND_PHP:
		content := fmt.Sprintf("<?php\necho '<h1>%s (PHP ' . PHP_VERSION . ')</h1>';\n", name)
		return os.WriteFile(filepath.Join(folder, "index.php"), []byte(content), 0o644)
	case site.KIND_NODE:
		pkg := fmt.Sprintf("{\n  \"name\": %q,\n  \"version\": \"1.0.0\",\n  \"main\": \"server.js\",\n  \"scripts\": {\n    \"start\": \"node server.js\"\n  }\n}\n", name)
		if err := os.WriteFile(filepath.Join(folder, "package.json"), []byte(pkg), 0o644); err != nil {
			return err
		}
		script := fmt.Sprintf("const http = require('http');\nconst port = Number(process.env.PORT);\nhttp.createServer((request, response) => {\n  response.writeHead(200, {'Content-Type': 'text/html; charset=utf-8'});\n  response.end('<h1>%s (Node.js ' + process.version + ')</h1>');\n}).listen(port, '%s');\n", name, paths.LOOPBACK)
		return os.WriteFile(filepath.Join(folder, "server.js"), []byte(script), 0o644)
	}
	return nil
}
