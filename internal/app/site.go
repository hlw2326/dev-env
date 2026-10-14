package app

import (
	"context"
	"fmt"
	"os"
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
	baseDir  string
	sites    ports.SiteRepository
	software ports.SoftwareRepository
	settings ports.SettingRepository
	gateway  ports.Gateway
	runner   ports.Runner
	hosts    ports.HostsWriter
	probe    ports.PortChecker
	bus      ports.EventPublisher

	mu     sync.RWMutex
	active map[int64]bool
	failed map[int64]string
}

func NewSiteService(
	baseDir string,
	sites ports.SiteRepository,
	softwareRepo ports.SoftwareRepository,
	settings ports.SettingRepository,
	gateway ports.Gateway,
	runner ports.Runner,
	hosts ports.HostsWriter,
	probe ports.PortChecker,
	bus ports.EventPublisher,
) *SiteService {
	return &SiteService{
		baseDir:  baseDir,
		sites:    sites,
		software: softwareRepo,
		settings: settings,
		gateway:  gateway,
		runner:   runner,
		hosts:    hosts,
		probe:    probe,
		bus:      bus,
		active:   make(map[int64]bool),
		failed:   make(map[int64]string),
	}
}

func (s *SiteService) ReconcileStartup(ctx context.Context) {
	list, err := s.sites.List(ctx)
	if err != nil {
		return
	}
	for _, item := range list {
		if item.Autostart {
			_ = s.Start(ctx, item.ID)
		}
	}
	s.syncGatewayRoutes(ctx)
	_ = s.syncCustomHosts(ctx)
}

func (s *SiteService) OnSoftwareInstalled(ctx context.Context, sw software.Software) {
	var siteKind string
	switch sw.Kind {
	case software.KindPHP:
		siteKind = site.KindPHP
	case software.KindNode:
		siteKind = site.KindNode
	default:
		return
	}
	missing, err := s.sites.ListMissingSoftware(ctx, siteKind)
	if err != nil {
		return
	}
	for _, st := range missing {
		if err := s.sites.BindSoftware(ctx, st.ID, sw.ID); err == nil {
			_ = s.Start(ctx, st.ID)
		}
	}
	s.syncGatewayRoutes(ctx)
}

func (s *SiteService) List(ctx context.Context) ([]SiteView, error) {
	list, err := s.sites.List(ctx)
	if err != nil {
		return nil, err
	}
	softList, err := s.software.List(ctx)
	if err != nil {
		return nil, err
	}
	softMap := make(map[int64]software.Software, len(softList))
	for _, sw := range softList {
		softMap[sw.ID] = sw
	}

	gwPort := s.gateway.Port()
	if gwPort <= 0 {
		if cfg, err := s.settings.Load(ctx); err == nil {
			gwPort = cfg.Port
		}
	}

	views := make([]SiteView, 0, len(list))
	for _, item := range list {
		views = append(views, s.buildView(item, softMap, gwPort))
	}
	return views, nil
}

func (s *SiteService) Get(ctx context.Context, id int64) (*SiteView, error) {
	item, err := s.sites.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, appErrors.New(appErrors.CodeNotFound, "站点不存在")
	}
	softMap := make(map[int64]software.Software)
	if item.Software != nil {
		if sw, err := s.software.Get(ctx, *item.Software); err == nil && sw != nil {
			softMap[sw.ID] = *sw
		}
	}
	gwPort := s.gateway.Port()
	if gwPort <= 0 {
		if cfg, err := s.settings.Load(ctx); err == nil {
			gwPort = cfg.Port
		}
	}
	v := s.buildView(*item, softMap, gwPort)
	return &v, nil
}

func (s *SiteService) FindRun(root, kind string) string {
	return site.FindRun(root, kind)
}

func (s *SiteService) Create(ctx context.Context, in site.CreateInput) (*SiteView, error) {
	kind := strings.TrimSpace(in.Kind)
	name := strings.TrimSpace(in.Name)
	host := strings.ToLower(strings.TrimSpace(in.Host))

	if !site.IsValidKind(kind) {
		return nil, appErrors.New(appErrors.CodeInvalidArgument, "不支持的站点类型")
	}
	if name == "" {
		return nil, appErrors.New(appErrors.CodeInvalidArgument, "站点名称不能为空")
	}
	if host == "" {
		return nil, appErrors.New(appErrors.CodeInvalidArgument, "站点域名不能为空")
	}

	root := strings.TrimSpace(in.Root)
	if root == "" {
		cfg, err := s.settings.Load(ctx)
		if err != nil {
			return nil, err
		}
		root = filepath.Join(cfg.Root, name)
	}
	root = filepath.Clean(root)
	if !filepath.IsAbs(root) {
		return nil, appErrors.New(appErrors.CodePathInvalid, "站点根目录必须是绝对路径")
	}

	runDir, err := site.NormalizeRun(root, in.Run)
	if err != nil {
		return nil, appErrors.New(appErrors.CodePathInvalid, err.Error())
	}

	docPath := site.ResolvePath(root, runDir)
	if err := os.MkdirAll(docPath, 0755); err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodePathNotAccessible, "创建站点目录失败")
	}

	if err := writeSampleFilesIfEmpty(kind, name, docPath); err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodePathNotAccessible, "写入默认示例文件失败")
	}

	var softID *int64
	switch kind {
	case site.KindPHP:
		if mainSw, err := s.software.GetMain(ctx, software.KindPHP); err == nil && mainSw != nil {
			id := mainSw.ID
			softID = &id
		}
	case site.KindNode:
		if mainSw, err := s.software.GetMain(ctx, software.KindNode); err == nil && mainSw != nil {
			id := mainSw.ID
			softID = &id
		}
	}

	port := 0
	command := strings.TrimSpace(in.Command)
	if kind == site.KindNode {
		port = s.allocateNodePort(ctx)
		if command == "" {
			command = site.FindCommand(docPath)
			if command == "" && docPath != root {
				command = site.FindCommand(root)
			}
		}
	}

	cur := &site.Site{
		Name:      name,
		Kind:      kind,
		Hosts:     []string{host},
		Root:      root,
		Run:       runDir,
		Software:  softID,
		Command:   command,
		Port:      port,
		Autostart: true,
	}

	if err := s.sites.Create(ctx, cur); err != nil {
		return nil, err
	}

	_ = s.syncCustomHosts(ctx)

	canStart := kind == site.KindStatic || (kind == site.KindPHP && softID != nil) || (kind == site.KindNode && softID != nil && command != "")
	if canStart {
		_ = s.Start(ctx, cur.ID)
	}

	return s.Get(ctx, cur.ID)
}

func (s *SiteService) Update(ctx context.Context, in site.UpdateInput) (*SiteView, error) {
	cur, err := s.sites.Get(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if cur == nil {
		return nil, appErrors.New(appErrors.CodeNotFound, "站点不存在")
	}

	var cleanHosts []string
	seen := make(map[string]bool)
	for _, h := range in.Hosts {
		trimmed := strings.ToLower(strings.TrimSpace(h))
		if trimmed != "" && !seen[trimmed] {
			seen[trimmed] = true
			cleanHosts = append(cleanHosts, trimmed)
		}
	}
	if len(cleanHosts) == 0 {
		return nil, appErrors.New(appErrors.CodeInvalidArgument, "至少需要保留一个域名")
	}

	root := filepath.Clean(strings.TrimSpace(in.Root))
	if !filepath.IsAbs(root) {
		return nil, appErrors.New(appErrors.CodePathInvalid, "根目录必须是绝对路径")
	}

	runDir, err := site.NormalizeRun(root, in.Run)
	if err != nil {
		return nil, appErrors.New(appErrors.CodePathInvalid, err.Error())
	}

	docPath := site.ResolvePath(root, runDir)
	if err := os.MkdirAll(docPath, 0755); err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodePathNotAccessible, "创建运行目录失败")
	}

	cur.Hosts = cleanHosts
	cur.Root = root
	cur.Run = runDir
	cur.Software = in.Software
	cur.Command = strings.TrimSpace(in.Command)
	cur.Autostart = in.Autostart

	if err := s.sites.Update(ctx, cur); err != nil {
		return nil, err
	}

	s.syncGatewayRoutes(ctx)
	_ = s.syncCustomHosts(ctx)

	return s.Get(ctx, cur.ID)
}

func (s *SiteService) Start(ctx context.Context, id int64) error {
	cur, err := s.sites.Get(ctx, id)
	if err != nil {
		return err
	}
	if cur == nil {
		return appErrors.New(appErrors.CodeNotFound, "站点不存在")
	}

	switch cur.Kind {
	case site.KindStatic:
		s.mu.Lock()
		s.active[cur.ID] = true
		delete(s.failed, cur.ID)
		s.mu.Unlock()

	case site.KindPHP:
		if cur.Software == nil {
			return appErrors.New(appErrors.CodeRuntimeNotInstalled, "站点未绑定 PHP 运行环境")
		}
		sw, err := s.software.Get(ctx, *cur.Software)
		if err != nil {
			return err
		}
		if sw == nil || sw.State != software.StateInstalled {
			return appErrors.New(appErrors.CodeRuntimeNotInstalled, "PHP 运行环境未安装")
		}
		phpCGI := filepath.Join(sw.Path, "php-cgi.exe")
		if info, err := os.Stat(phpCGI); err != nil || info.IsDir() {
			return appErrors.New(appErrors.CodeRuntimeNotInstalled, "找不到 php-cgi.exe: "+phpCGI)
		}
		phpPort := sw.Port
		if phpPort <= 0 {
			phpPort = 9000
		}
		phpKey := fmt.Sprintf("php:%d", sw.ID)
		if !s.runner.Status(phpKey).Running && !s.probe.Listening(phpPort) {
			logPath := filepath.Join(s.baseDir, "logs", "services", fmt.Sprintf("php-%s.log", sw.Version))
			_ = os.MkdirAll(filepath.Dir(logPath), 0755)
			_, err := s.runner.Start(phpKey, ports.Options{
				Binary:    phpCGI,
				Args:      []string{"-b", fmt.Sprintf("127.0.0.1:%d", phpPort), "-c", filepath.Join(sw.Path, "php.ini")},
				Directory: sw.Path,
				Log:       logPath,
			})
			if err != nil {
				return err
			}
		}
		s.mu.Lock()
		s.active[cur.ID] = true
		delete(s.failed, cur.ID)
		s.mu.Unlock()

	case site.KindNode:
		if cur.Software == nil {
			return appErrors.New(appErrors.CodeRuntimeNotInstalled, "站点未绑定 Node.js 运行环境")
		}
		sw, err := s.software.Get(ctx, *cur.Software)
		if err != nil {
			return err
		}
		if sw == nil || sw.State != software.StateInstalled {
			return appErrors.New(appErrors.CodeRuntimeNotInstalled, "Node.js 运行环境未安装")
		}
		if strings.TrimSpace(cur.Command) == "" {
			return appErrors.New(appErrors.CodeInvalidArgument, "未配置 Node 启动命令")
		}
		if err := s.startNodeProcess(*cur, *sw); err != nil {
			s.mu.Lock()
			delete(s.active, cur.ID)
			s.failed[cur.ID] = err.Error()
			s.mu.Unlock()
			return err
		}
		s.mu.Lock()
		s.active[cur.ID] = true
		delete(s.failed, cur.ID)
		s.mu.Unlock()
	}

	s.syncGatewayRoutes(ctx)
	s.bus.Emit("site:status-changed", map[string]interface{}{
		"id":      cur.ID,
		"state":   site.StateRunning,
		"message": "",
	})
	return nil
}

func (s *SiteService) Stop(ctx context.Context, id int64) error {
	cur, err := s.sites.Get(ctx, id)
	if err != nil {
		return err
	}
	if cur == nil {
		return appErrors.New(appErrors.CodeNotFound, "站点不存在")
	}

	if cur.Kind == site.KindNode {
		_ = s.runner.Stop(siteKey(cur.ID))
	} else if cur.Kind == site.KindPHP && cur.Software != nil {
		otherUsing := false
		s.mu.RLock()
		for activeID := range s.active {
			if activeID != cur.ID {
				if otherSite, err := s.sites.Get(ctx, activeID); err == nil && otherSite != nil {
					if otherSite.Kind == site.KindPHP && otherSite.Software != nil && *otherSite.Software == *cur.Software {
						otherUsing = true
						break
					}
				}
			}
		}
		s.mu.RUnlock()
		if !otherUsing {
			_ = s.runner.Stop(fmt.Sprintf("php:%d", *cur.Software))
		}
	}

	s.mu.Lock()
	delete(s.active, cur.ID)
	delete(s.failed, cur.ID)
	s.mu.Unlock()

	s.syncGatewayRoutes(ctx)
	s.bus.Emit("site:status-changed", map[string]interface{}{
		"id":      cur.ID,
		"state":   site.StateStopped,
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
	cur, err := s.sites.Get(ctx, id)
	if err != nil {
		return err
	}
	if cur == nil {
		return appErrors.New(appErrors.CodeNotFound, "站点不存在")
	}

	_ = s.Stop(ctx, id)
	if err := s.sites.Delete(ctx, id); err != nil {
		return err
	}

	s.syncGatewayRoutes(ctx)
	_ = s.syncCustomHosts(ctx)
	return nil
}

func (s *SiteService) Logs(ctx context.Context, id int64, kind string, limit int) ([]string, error) {
	cur, err := s.sites.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if cur == nil {
		return nil, appErrors.New(appErrors.CodeNotFound, "站点不存在")
	}
	var logFile string
	if kind == "run" {
		logFile = s.siteLogPath(cur.Name, "run.log")
	} else {
		logFile = s.siteLogPath(cur.Name, "access.log")
	}
	return process.ReadLogTail(logFile, limit), nil
}

func (s *SiteService) StopAll() {
	list, err := s.sites.List(context.Background())
	if err != nil {
		return
	}
	for _, item := range list {
		if item.Kind == site.KindNode {
			_ = s.runner.Stop(siteKey(item.ID))
		}
	}
}

func (s *SiteService) startNodeProcess(cur site.Site, sw software.Software) error {
	key := siteKey(cur.ID)
	if st := s.runner.Status(key); st.Running {
		return nil
	}
	if !s.probe.Available(cur.Port) {
		return appErrors.New(appErrors.CodePortAlreadyInUse, fmt.Sprintf("Node 站点端口 %d 已被占用", cur.Port))
	}

	portStr := strconv.Itoa(cur.Port)
	expandedCmd := strings.ReplaceAll(cur.Command, "%PORT%", portStr)
	workDir := site.ResolvePath(cur.Root, cur.Run)
	logPath := s.siteLogPath(cur.Name, "run.log")

	sysRoot := os.Getenv("SystemRoot")
	if sysRoot == "" {
		sysRoot = `C:\Windows`
	}
	cmdExe := filepath.Join(sysRoot, "System32", "cmd.exe")

	envPath := sw.Path + ";" + os.Getenv("PATH")
	siteID := cur.ID

	opts := ports.Options{
		Binary:    cmdExe,
		Args:      []string{"/C", expandedCmd},
		Directory: workDir,
		Environment: []string{
			"PORT=" + portStr,
			"PATH=" + envPath,
		},
		Log: logPath,
		OnExit: func(_ string, message string) {
			s.mu.Lock()
			delete(s.active, siteID)
			s.failed[siteID] = message
			s.mu.Unlock()
			s.syncGatewayRoutes(context.Background())
			s.bus.Emit("site:status-changed", map[string]interface{}{
				"id":      siteID,
				"state":   site.StateError,
				"message": message,
			})
		},
	}

	if _, err := s.runner.Start(key, opts); err != nil {
		return err
	}

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if s.probe.Listening(cur.Port) {
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
	return appErrors.New(appErrors.CodeProcessStartFailed, "Node 站点启动超时或退出:\n"+strings.Join(tail, "\n"))
}

func (s *SiteService) syncGatewayRoutes(ctx context.Context) {
	list, err := s.sites.List(ctx)
	if err != nil {
		return
	}
	softList, err := s.software.List(ctx)
	if err != nil {
		return
	}
	softMap := make(map[int64]software.Software, len(softList))
	for _, sw := range softList {
		softMap[sw.ID] = sw
	}

	s.mu.RLock()
	activeCopy := make(map[int64]bool, len(s.active))
	for k, v := range s.active {
		activeCopy[k] = v
	}
	s.mu.RUnlock()

	var routes []ports.Route
	for _, item := range list {
		if !activeCopy[item.ID] {
			continue
		}
		phpDir := ""
		port := item.Port
		if item.Kind == site.KindPHP && item.Software != nil {
			if sw, ok := softMap[*item.Software]; ok && sw.State == software.StateInstalled {
				phpDir = sw.Path
				if sw.Port > 0 {
					port = sw.Port
				} else {
					port = 9000
				}
			}
		}
		routes = append(routes, ports.Route{
			ID:    item.ID,
			Hosts: item.Hosts,
			Kind:  item.Kind,
			Path:  site.ResolvePath(item.Root, item.Run),
			PHP:   phpDir,
			Port:  port,
			Log:   s.siteLogPath(item.Name, "access.log"),
		})
	}
	s.gateway.SetRoutes(routes)
}

func (s *SiteService) syncCustomHosts(ctx context.Context) error {
	cfg, err := s.settings.Load(ctx)
	if err != nil {
		return err
	}
	if !cfg.Hosts {
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

func (s *SiteService) buildView(item site.Site, softMap map[int64]software.Software, gwPort int) SiteView {
	docPath := site.ResolvePath(item.Root, item.Run)
	version := ""
	need := ""

	switch item.Kind {
	case site.KindPHP:
		if item.Software != nil {
			if sw, ok := softMap[*item.Software]; ok && sw.State == software.StateInstalled {
				version = sw.Version
			} else {
				need = software.KindPHP
			}
		} else {
			need = software.KindPHP
		}
	case site.KindNode:
		if item.Software != nil {
			if sw, ok := softMap[*item.Software]; ok && sw.State == software.StateInstalled {
				version = sw.Version
			} else {
				need = software.KindNode
			}
		} else {
			need = software.KindNode
		}
	}

	s.mu.RLock()
	isActive := s.active[item.ID]
	failMsg, hasFail := s.failed[item.ID]
	s.mu.RUnlock()
	_ = failMsg

	state := site.StateStopped
	if hasFail {
		state = site.StateError
	} else if isActive {
		if item.Kind == site.KindNode {
			st := s.runner.Status(siteKey(item.ID))
			if st.Running {
				state = site.StateRunning
			} else if st.Message != "" {
				state = site.StateError
			}
		} else {
			state = site.StateRunning
		}
	}

	siteURL := ""
	if len(item.Hosts) > 0 {
		if gwPort > 0 && gwPort != 80 {
			siteURL = fmt.Sprintf("http://%s:%d", item.Hosts[0], gwPort)
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

func (s *SiteService) allocateNodePort(ctx context.Context) int {
	list, _ := s.sites.List(ctx)
	used := make(map[int]bool, len(list))
	for _, item := range list {
		if item.Port > 0 {
			used[item.Port] = true
		}
	}
	for port := 3000; port < 4000; port++ {
		if !used[port] && s.probe.Available(port) {
			return port
		}
	}
	return 3000
}

func (s *SiteService) siteLogPath(name, file string) string {
	return filepath.Join(s.baseDir, "logs", "sites", name, file)
}

func siteKey(id int64) string {
	return "site:" + strconv.FormatInt(id, 10)
}

func writeSampleFilesIfEmpty(kind, name, docPath string) error {
	entries, err := os.ReadDir(docPath)
	if err != nil {
		return err
	}
	if len(entries) > 0 {
		return nil
	}

	switch kind {
	case site.KindStatic:
		html := fmt.Sprintf("<!DOCTYPE html>\n<html lang=\"zh-CN\">\n<head><meta charset=\"utf-8\"><title>%s</title></head>\n<body><h1>%s 运行正常</h1></body>\n</html>\n", name, name)
		return os.WriteFile(filepath.Join(docPath, "index.html"), []byte(html), 0644)
	case site.KindPHP:
		phpCode := fmt.Sprintf("<?php\necho '<h1>%s (PHP ' . PHP_VERSION . ')</h1>';\n", name)
		return os.WriteFile(filepath.Join(docPath, "index.php"), []byte(phpCode), 0644)
	case site.KindNode:
		pkg := fmt.Sprintf("{\n  \"name\": %q,\n  \"version\": \"1.0.0\",\n  \"main\": \"server.js\",\n  \"scripts\": {\n    \"start\": \"node server.js\"\n  }\n}\n", name)
		if err := os.WriteFile(filepath.Join(docPath, "package.json"), []byte(pkg), 0644); err != nil {
			return err
		}
		srv := fmt.Sprintf("const http = require('http');\nconst port = process.env.PORT || 3000;\nhttp.createServer((req, res) => {\n  res.writeHead(200, {'Content-Type': 'text/html; charset=utf-8'});\n  res.end('<h1>%s (Node.js ' + process.version + ')</h1>');\n}).listen(port, '127.0.0.1');\n", name)
		return os.WriteFile(filepath.Join(docPath, "server.js"), []byte(srv), 0644)
	}
	return nil
}
