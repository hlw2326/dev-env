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

	domainmonitor "devenv/internal/domain/monitor"
	"devenv/internal/domain/project"
	"devenv/internal/domain/software"
	inframonitor "devenv/internal/infra/monitor"
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

type ProjectView struct {
	project.Project
	Path             string
	Traffic          domainmonitor.ProjectTraffic
	EnvironmentFiles []string
	Version          string
	State            string
	URL              string
	Need             string
}

type ProjectService struct {
	base     string
	projects ports.ProjectRepository
	software ports.SoftwareRepository
	settings ports.SettingRepository
	nginx    ports.Gateway
	runner   ports.Runner
	hosts    ports.HostsWriter
	probe    ports.PortChecker
	bus      ports.EventPublisher
	git      ports.GitManager
	metrics  *inframonitor.Collector

	mu     sync.RWMutex
	active map[int64]bool
	failed map[int64]string
}

func NewProjectService(
	base string,
	projects ports.ProjectRepository,
	softwareRepo ports.SoftwareRepository,
	settings ports.SettingRepository,
	nginx ports.Gateway,
	runner ports.Runner,
	hosts ports.HostsWriter,
	probe ports.PortChecker,
	bus ports.EventPublisher,
	gitManager ports.GitManager,
	metrics *inframonitor.Collector,
) *ProjectService {
	return &ProjectService{
		base:     base,
		projects: projects,
		software: softwareRepo,
		settings: settings,
		nginx:    nginx,
		runner:   runner,
		hosts:    hosts,
		probe:    probe,
		bus:      bus,
		git:      gitManager,
		metrics:  metrics,
		active:   make(map[int64]bool),
		failed:   make(map[int64]string),
	}
}

func (s *ProjectService) StartAuto(ctx context.Context) {
	list, err := s.projects.List(ctx)
	if err != nil {
		log.Printf("读取自动启动项目失败: %v", err)
		return
	}
	for _, item := range list {
		if project.IsLocal(item.Kind) {
			continue
		}
		if item.Autostart {
			if err := s.Start(ctx, item.ID); err != nil {
				log.Printf("自动启动项目 %s 失败: %v", item.Name, err)
				continue
			}
		}
	}
	if err := s.syncRoutes(ctx); err != nil {
		log.Printf("刷新自动启动项目路由失败: %v", err)
	}
	if err := s.syncHosts(ctx); err != nil {
		log.Printf("同步自动启动项目域名失败: %v", err)
		return
	}
}

func (s *ProjectService) BindInstalled(ctx context.Context, installed software.Software) {
	projectKinds := project.ProjectKindsForSoftware(installed.Kind)
	if len(projectKinds) == 0 {
		return
	}
	for _, projectKind := range projectKinds {
		missing, err := s.projects.ListMissingSoftware(ctx, projectKind)
		if err != nil {
			log.Printf("读取缺少运行环境的项目失败: %v", err)
			continue
		}
		for _, item := range missing {
			if err := s.projects.BindSoftware(ctx, item.ID, installed.ID); err != nil {
				log.Printf("绑定项目 %s 运行环境失败: %v", item.Name, err)
				continue
			}
			if project.IsLocal(item.Kind) {
				continue
			}
			if err := s.Start(ctx, item.ID); err != nil {
				log.Printf("启动新绑定环境的项目 %s 失败: %v", item.Name, err)
			}
		}
	}
	if err := s.syncRoutes(ctx); err != nil {
		log.Printf("刷新新绑定环境的项目路由失败: %v", err)
	}
}

func (s *ProjectService) List(ctx context.Context) ([]ProjectView, error) {
	list, err := s.projects.List(ctx)
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

	views := make([]ProjectView, 0, len(list))
	for _, item := range list {
		views = append(views, s.buildView(item, softwareMap, port))
	}
	return views, nil
}

func (s *ProjectService) Get(ctx context.Context, id int64) (*ProjectView, error) {
	item, err := s.projects.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, appErrors.New(appErrors.NOT_FOUND, "项目不存在")
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

func (s *ProjectService) FindRun(root, kind string) string {
	return project.FindRun(root, kind)
}

func (s *ProjectService) FindCommands(root, kind string) []string {
	return project.FindCommands(root, kind)
}

func (s *ProjectService) Templates() []project.ProjectTemplate {
	return project.ProjectTemplates()
}

func (s *ProjectService) requireNginx(ctx context.Context) error {
	installed, err := s.software.GetMain(ctx, software.KIND_NGINX)
	if err != nil {
		return err
	}
	if installed == nil || installed.State != software.STATE_INSTALLED {
		return appErrors.New(appErrors.RUNTIME_NOT_INSTALLED, "创建 Web 项目前请先在软件商城安装 Nginx")
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

func (s *ProjectService) deleteCreated(ctx context.Context, id int64, source error) error {
	cleanup := []error{source}
	if err := s.projects.Delete(ctx, id); err != nil {
		cleanup = append(cleanup, err)
	}
	if err := s.syncHosts(ctx); err != nil {
		cleanup = append(cleanup, err)
	}
	return errors.Join(cleanup...)
}

func (s *ProjectService) Create(ctx context.Context, in project.CreateInput) (*ProjectView, error) {
	kind := strings.TrimSpace(in.Kind)
	name := strings.TrimSpace(in.Name)
	host := strings.ToLower(strings.TrimSpace(in.Host))

	if !project.IsValidKind(kind) {
		return nil, appErrors.New(appErrors.INVALID_ARGUMENT, "不支持的项目类型")
	}
	if name == "" {
		return nil, appErrors.New(appErrors.INVALID_ARGUMENT, "项目名称不能为空")
	}
	isLocal := project.IsLocal(kind)
	if !isLocal && host == "" {
		return nil, appErrors.New(appErrors.INVALID_ARGUMENT, "项目域名不能为空")
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
		return nil, appErrors.New(appErrors.PATH_INVALID, "项目根目录必须是绝对路径")
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.PATH_INVALID, "解析项目根目录失败")
	}
	root = filepath.Clean(absoluteRoot)

	existingByName, err := s.projects.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if existingByName != nil {
		return nil, appErrors.New(appErrors.ALREADY_EXISTS, "项目名称已存在: "+existingByName.Name)
	}
	existingByRoot, err := s.projects.GetByRoot(ctx, root)
	if err != nil {
		return nil, err
	}
	if existingByRoot != nil {
		return nil, appErrors.New(appErrors.ALREADY_EXISTS, "项目目录已添加: "+existingByRoot.Root)
	}

	if project.UsesNginx(kind) {
		if err := s.requireNginx(ctx); err != nil {
			return nil, err
		}
	}

	runFolder, err := project.NormalizeRun(root, in.Run)
	if err != nil {
		return nil, appErrors.New(appErrors.PATH_INVALID, err.Error())
	}

	docPath := project.ResolvePath(root, runFolder)
	if err := os.MkdirAll(docPath, 0o755); err != nil {
		return nil, appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "创建项目目录失败")
	}

	if project.UsesNginx(kind) {
		if err := writeSample(kind, name, docPath); err != nil {
			return nil, appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "写入默认示例文件失败")
		}
	}

	var softwareID *int64
	runtimeKind := project.RuntimeSoftwareKind(kind)
	if runtimeKind != "" {
		mainItem, err := s.software.GetMain(ctx, runtimeKind)
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
	if project.IsProcess(kind) && !isLocal {
		port = s.findPort(ctx)
		if command == "" {
			command = project.FindCommand(docPath, kind)
			if command == "" && docPath != root {
				command = project.FindCommand(root, kind)
			}
		}
	} else if template, ok := project.ProjectTemplateForKind(kind); ok && command == "" && template.DefaultCommand != "" {
		command = template.DefaultCommand
	} else if template, ok := project.ProjectTemplateForKind(kind); ok && template.CommandMode == project.COMMAND_MODE_PACKAGE_SCRIPT && command == "" {
		command = project.FindCommand(docPath, kind)
		if command == "" && docPath != root {
			command = project.FindCommand(root, kind)
		}
	}
	environmentFile := ""
	rewriteRule := project.REWRITE_DEFAULT
	if kind == project.KIND_PHP {
		environmentFile = project.DefaultEnvironmentFile(root)
		rewriteRule = project.REWRITE_THINKPHP
	}

	created := &project.Project{
		Name:            name,
		Kind:            kind,
		Hosts:           make([]string, 0),
		Root:            root,
		Run:             runFolder,
		Software:        softwareID,
		Command:         command,
		RewriteRule:     rewriteRule,
		EnvironmentFile: environmentFile,
		Port:            port,
		Autostart:       !isLocal,
	}
	if !isLocal {
		created.Hosts = []string{host}
	}

	if err := s.projects.Create(ctx, created); err != nil {
		return nil, err
	}

	if err := s.syncHosts(ctx); err != nil {
		return nil, s.deleteCreated(ctx, created.ID, err)
	}

	ready := !project.HasCapability(kind, project.CAPABILITY_RUNTIME) && !project.IsProcess(kind)
	if project.HasCapability(kind, project.CAPABILITY_PHP_RUNTIME) {
		ready = softwareID != nil
	}
	if project.IsProcess(kind) {
		ready = runtimeKind != "" && softwareID != nil && command != ""
	}
	if ready {
		if err := s.Start(ctx, created.ID); err != nil {
			return nil, s.deleteCreated(ctx, created.ID, err)
		}
	}

	return s.Get(ctx, created.ID)
}

func (s *ProjectService) Update(ctx context.Context, in project.UpdateInput) (*ProjectView, error) {
	current, err := s.projects.Get(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, appErrors.New(appErrors.NOT_FOUND, "项目不存在")
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
	if project.UsesNginx(current.Kind) && len(cleanHosts) == 0 {
		return nil, appErrors.New(appErrors.INVALID_ARGUMENT, "至少需要保留一个域名")
	}
	if project.IsLocal(current.Kind) {
		cleanHosts = make([]string, 0)
	}

	root := filepath.Clean(strings.TrimSpace(in.Root))
	if !filepath.IsAbs(root) {
		return nil, appErrors.New(appErrors.PATH_INVALID, "根目录必须是绝对路径")
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.PATH_INVALID, "解析项目根目录失败")
	}
	root = filepath.Clean(absoluteRoot)
	existingByRoot, err := s.projects.GetByRoot(ctx, root)
	if err != nil {
		return nil, err
	}
	if existingByRoot != nil && existingByRoot.ID != current.ID {
		return nil, appErrors.New(appErrors.ALREADY_EXISTS, "项目目录已添加: "+existingByRoot.Root)
	}

	runFolder, err := project.NormalizeRun(root, in.Run)
	if err != nil {
		return nil, appErrors.New(appErrors.PATH_INVALID, err.Error())
	}

	docPath := project.ResolvePath(root, runFolder)
	if err := os.MkdirAll(docPath, 0o755); err != nil {
		return nil, appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "创建运行目录失败")
	}
	nginxConfig := strings.TrimSpace(in.Nginx)
	rewriteRule := strings.TrimSpace(in.RewriteRule)
	if project.UsesNginx(current.Kind) {
		if err := project.ValidateNginxConfig(nginxConfig); err != nil {
			return nil, appErrors.New(appErrors.INVALID_ARGUMENT, err.Error())
		}
	} else {
		nginxConfig = ""
		rewriteRule = project.REWRITE_DEFAULT
	}
	if current.Kind == project.KIND_PHP && !project.IsValidRewriteRule(rewriteRule) {
		return nil, appErrors.New(appErrors.INVALID_ARGUMENT, "伪静态规则无效")
	}
	if current.Kind != project.KIND_PHP {
		rewriteRule = project.REWRITE_DEFAULT
	}
	environmentFile, err := project.NormalizeEnvironmentFile(root, in.EnvironmentFile)
	if err != nil {
		return nil, appErrors.New(appErrors.INVALID_ARGUMENT, err.Error())
	}

	current.Hosts = cleanHosts
	current.Root = root
	current.Run = runFolder
	current.Software = in.Software
	current.Command = strings.TrimSpace(in.Command)
	current.EnvironmentFile = environmentFile
	if template, ok := project.ProjectTemplateForKind(current.Kind); ok && current.Command == "" && template.DefaultCommand != "" {
		current.Command = template.DefaultCommand
	}
	if template, ok := project.ProjectTemplateForKind(current.Kind); ok && template.CommandMode == project.COMMAND_MODE_PACKAGE_SCRIPT && current.Command == "" {
		current.Command = project.FindCommand(docPath, current.Kind)
		if current.Command == "" && docPath != root {
			current.Command = project.FindCommand(root, current.Kind)
		}
	}
	current.Nginx = nginxConfig
	current.RewriteRule = rewriteRule
	current.Autostart = in.Autostart && !project.IsLocal(current.Kind)

	if err := s.projects.Update(ctx, current); err != nil {
		return nil, err
	}

	if err := s.syncRoutes(ctx); err != nil {
		if restoreErr := s.projects.Update(ctx, &original); restoreErr != nil {
			return nil, errors.Join(err, restoreErr)
		}
		return nil, err
	}
	if nginxSettingsChanged(original, *current) && s.nginx.Port() > 0 {
		if err := s.nginx.Restart(); err != nil {
			return nil, err
		}
	}
	if err := s.syncHosts(ctx); err != nil {
		if restoreErr := s.projects.Update(ctx, &original); restoreErr != nil {
			return nil, errors.Join(err, restoreErr)
		}
		if restoreErr := s.syncRoutes(ctx); restoreErr != nil {
			return nil, errors.Join(err, restoreErr)
		}
		return nil, err
	}

	return s.Get(ctx, current.ID)
}

func nginxSettingsChanged(previous, current project.Project) bool {
	return previous.Nginx != current.Nginx ||
		previous.RewriteRule != current.RewriteRule ||
		previous.EnvironmentFile != current.EnvironmentFile
}

func (s *ProjectService) Start(ctx context.Context, id int64) error {
	current, err := s.projects.Get(ctx, id)
	if err != nil {
		return err
	}
	if current == nil {
		return appErrors.New(appErrors.NOT_FOUND, "项目不存在")
	}

	switch {
	case project.IsValidKind(current.Kind) && !project.HasCapability(current.Kind, project.CAPABILITY_RUNTIME) && !project.IsProcess(current.Kind):
		s.markActive(current.ID)
	case project.HasCapability(current.Kind, project.CAPABILITY_PHP_RUNTIME):
		if _, err := s.projectRuntime(ctx, *current); err != nil {
			return err
		}
		s.markActive(current.ID)
	case project.IsProcess(current.Kind):
		runtime, err := s.projectRuntime(ctx, *current)
		if err != nil {
			return err
		}
		if strings.TrimSpace(current.Command) == "" {
			return appErrors.New(appErrors.INVALID_ARGUMENT, "未配置项目启动命令")
		}
		if err := s.startProcess(*current, runtime); err != nil {
			s.markFailed(current.ID, err)
			return err
		}
		s.markActive(current.ID)
	default:
		return appErrors.New(appErrors.INVALID_ARGUMENT, "项目类型无效: "+current.Kind)
	}

	if project.UsesNginx(current.Kind) {
		if err := s.syncRoutes(ctx); err != nil {
			s.markFailed(current.ID, err)
			if project.IsProcess(current.Kind) {
				if stopError := s.runner.Stop(projectKey(current.ID)); stopError != nil {
					return errors.Join(err, stopError)
				}
			}
			return err
		}
	}
	s.bus.Emit("project:status-changed", map[string]interface{}{
		"id":      current.ID,
		"state":   project.STATE_RUNNING,
		"message": "",
	})
	return nil
}

func (s *ProjectService) Stop(ctx context.Context, id int64) error {
	current, err := s.projects.Get(ctx, id)
	if err != nil {
		return err
	}
	if current == nil {
		return appErrors.New(appErrors.NOT_FOUND, "项目不存在")
	}

	if project.IsProcess(current.Kind) {
		if err := s.runner.Stop(projectKey(current.ID)); err != nil {
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
	s.bus.Emit("project:status-changed", map[string]interface{}{
		"id":      current.ID,
		"state":   project.STATE_STOPPED,
		"message": "",
	})
	return nil
}

func (s *ProjectService) Restart(ctx context.Context, id int64) error {
	if err := s.Stop(ctx, id); err != nil {
		return err
	}
	return s.Start(ctx, id)
}

func (s *ProjectService) Delete(ctx context.Context, id int64) error {
	current, err := s.projects.Get(ctx, id)
	if err != nil {
		return err
	}
	if current == nil {
		return appErrors.New(appErrors.NOT_FOUND, "项目不存在")
	}

	if err := s.Stop(ctx, id); err != nil {
		return err
	}
	if err := removeNginxArtifacts(s.base, *current); err != nil {
		return err
	}
	if err := s.projects.Delete(ctx, id); err != nil {
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

func removeNginxArtifacts(base string, item project.Project) error {
	if !project.UsesNginx(item.Kind) {
		return nil
	}
	configPath := paths.NginxProjectConfig(base, item.ID, item.Name)
	environmentPath := filepath.Join(paths.NginxHosts(base), paths.NginxEnvironmentConfigName(item.ID, item.Name))
	for _, target := range []string{configPath, environmentPath, configPath + ".devenv-backup"} {
		if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
			return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "清理项目 Nginx 配置失败: "+target)
		}
	}
	return nil
}

func (s *ProjectService) Logs(ctx context.Context, id int64, kind string, limit int) ([]string, error) {
	current, err := s.projects.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, appErrors.New(appErrors.NOT_FOUND, "项目不存在")
	}
	logFile := paths.ProjectLog(s.base, current.Name, "access.log")
	if kind == "run" {
		logFile = paths.ProjectLog(s.base, current.Name, "run.log")
	}
	return process.ReadTail(logFile, limit), nil
}

func (s *ProjectService) OpenNginxConfig(ctx context.Context, id int64) error {
	item, err := s.projects.Get(ctx, id)
	if err != nil {
		return err
	}
	if item == nil {
		return appErrors.New(appErrors.NOT_FOUND, "项目不存在")
	}
	if project.IsLocal(item.Kind) {
		return appErrors.New(appErrors.INVALID_ARGUMENT, "Electron 项目不使用 Nginx 配置")
	}
	configPath, err := s.ensureNginxConfig(ctx, *item)
	if err != nil {
		return err
	}
	return exec.Command("explorer.exe", configPath).Start()
}

func (s *ProjectService) ReadNginxConfig(ctx context.Context, id int64) (string, error) {
	item, err := s.projects.Get(ctx, id)
	if err != nil {
		return "", err
	}
	if item == nil {
		return "", appErrors.New(appErrors.NOT_FOUND, "项目不存在")
	}
	if project.IsLocal(item.Kind) {
		return "", appErrors.New(appErrors.INVALID_ARGUMENT, "Electron 项目不使用 Nginx 配置")
	}
	configPath, err := s.ensureNginxConfig(ctx, *item)
	if err != nil {
		return "", err
	}
	content, err := os.ReadFile(configPath)
	if err != nil {
		return "", appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "读取项目 Nginx 配置失败")
	}
	return string(content), nil
}

// WriteNginxConfig updates the generated configuration file for one project.
// The candidate is checked with nginx -t before it is kept. A failed check or
// reload always restores the previous file so an edit cannot break the gateway.
func (s *ProjectService) WriteNginxConfig(ctx context.Context, id int64, content string) error {
	item, err := s.projects.Get(ctx, id)
	if err != nil {
		return err
	}
	if item == nil {
		return appErrors.New(appErrors.NOT_FOUND, "项目不存在")
	}
	if project.IsLocal(item.Kind) {
		return appErrors.New(appErrors.INVALID_ARGUMENT, "Electron 项目不使用 Nginx 配置")
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return appErrors.New(appErrors.INVALID_ARGUMENT, "Nginx 配置不能为空")
	}

	configPath, err := s.ensureNginxConfig(ctx, *item)
	if err != nil {
		return err
	}
	installed, err := s.software.GetMain(ctx, software.KIND_NGINX)
	if err != nil {
		return err
	}
	if installed == nil || installed.State != software.STATE_INSTALLED {
		return appErrors.New(appErrors.RUNTIME_NOT_INSTALLED, "请先在软件商城安装 Nginx")
	}
	executable := filepath.Join(installed.Path, "nginx.exe")
	if info, statErr := os.Stat(executable); statErr != nil || info.IsDir() {
		if statErr != nil {
			return appErrors.Wrap(statErr, appErrors.RUNTIME_NOT_INSTALLED, "读取 nginx.exe 失败")
		}
		return appErrors.New(appErrors.RUNTIME_NOT_INSTALLED, "nginx.exe 路径指向了目录")
	}

	temporary, err := os.CreateTemp(filepath.Dir(configPath), ".devenv-nginx-*.conf")
	if err != nil {
		return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "创建 Nginx 临时配置失败")
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err := temporary.WriteString(content + "\n"); err != nil {
		_ = temporary.Close()
		return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "写入 Nginx 临时配置失败")
	}
	if err := temporary.Close(); err != nil {
		return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "关闭 Nginx 临时配置失败")
	}

	backupPath := configPath + ".devenv-backup"
	_ = os.Remove(backupPath)
	if err := os.Rename(configPath, backupPath); err != nil {
		return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "备份原 Nginx 配置失败")
	}
	restore := func() {
		_ = os.Remove(configPath)
		_ = os.Rename(backupPath, configPath)
	}
	if err := os.Rename(temporaryPath, configPath); err != nil {
		restore()
		return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "替换 Nginx 配置失败")
	}

	if err := validateNginxConfig(executable, s.base); err != nil {
		restore()
		return err
	}
	if s.nginx.Port() > 0 {
		if err := reloadNginxConfig(executable, s.base); err != nil {
			restore()
			_ = reloadNginxConfig(executable, s.base)
			return err
		}
	}
	if err := os.Remove(backupPath); err != nil {
		return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "清理 Nginx 备份配置失败")
	}
	return nil
}

func (s *ProjectService) ensureNginxConfig(ctx context.Context, item project.Project) (string, error) {
	configPath := paths.NginxProjectConfig(s.base, item.ID, item.Name)
	info, err := os.Stat(configPath)
	if err == nil {
		if info.IsDir() {
			return "", appErrors.New(appErrors.PATH_NOT_ACCESSIBLE, "项目 Nginx 配置文件路径指向了目录")
		}
		return configPath, nil
	}
	if !os.IsNotExist(err) {
		return "", appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "读取项目 Nginx 配置失败")
	}
	if err := s.syncRoutes(ctx); err != nil {
		return "", err
	}
	info, err = os.Stat(configPath)
	if err != nil {
		return "", appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "该项目的 Nginx 配置文件未生成，请先启动项目")
	}
	if info.IsDir() {
		return "", appErrors.New(appErrors.PATH_NOT_ACCESSIBLE, "项目 Nginx 配置文件路径指向了目录")
	}
	return configPath, nil
}

func validateNginxConfig(executable, base string) error {
	command := exec.Command(executable, "-p", nginxPrefix(base), "-c", paths.NGINX_CONFIG, "-t")
	output, err := command.CombinedOutput()
	if err != nil {
		return appErrors.Wrap(err, appErrors.PROCESS_START_FAILED, "Nginx 配置检查失败: "+strings.TrimSpace(string(output)))
	}
	return nil
}

func reloadNginxConfig(executable, base string) error {
	command := exec.Command(executable, "-p", nginxPrefix(base), "-c", paths.NGINX_CONFIG, "-s", "reload")
	output, err := command.CombinedOutput()
	if err != nil {
		return appErrors.Wrap(err, appErrors.PROCESS_START_FAILED, "Nginx 配置重新加载失败: "+strings.TrimSpace(string(output)))
	}
	return nil
}

func nginxPrefix(base string) string {
	return filepath.ToSlash(paths.Nginx(base)) + "/"
}

func (s *ProjectService) StopAll() error {
	list, err := s.projects.List(context.Background())
	if err != nil {
		return err
	}
	stopErrors := make([]error, 0)
	for _, item := range list {
		if project.IsProcess(item.Kind) {
			if err := s.runner.Stop(projectKey(item.ID)); err != nil {
				stopErrors = append(stopErrors, err)
			}
		}
	}
	return errors.Join(stopErrors...)
}

func (s *ProjectService) projectRuntime(ctx context.Context, current project.Project) (*software.Software, error) {
	runtimeKind := project.RuntimeSoftwareKind(current.Kind)
	if runtimeKind == "" {
		if current.Kind == project.KIND_RUST {
			if _, err := exec.LookPath("cargo"); err != nil {
				return nil, appErrors.Wrap(err, appErrors.RUNTIME_NOT_INSTALLED, "Rust 项目需要系统 PATH 中的 cargo")
			}
		}
		return nil, nil
	}
	if current.Software == nil {
		return nil, appErrors.New(appErrors.RUNTIME_NOT_INSTALLED, software.Name(runtimeKind)+" 运行环境未绑定")
	}
	runtimeItem, err := s.software.Get(ctx, *current.Software)
	if err != nil {
		return nil, err
	}
	if runtimeItem == nil || runtimeItem.State != software.STATE_INSTALLED {
		return nil, appErrors.New(appErrors.RUNTIME_NOT_INSTALLED, software.Name(runtimeKind)+" 运行环境未安装")
	}
	binary := runtimeExecutable(*runtimeItem)
	if binary != "" {
		if info, statErr := os.Stat(binary); statErr != nil || info.IsDir() {
			return nil, appErrors.New(appErrors.RUNTIME_NOT_INSTALLED, "找不到运行环境可执行文件: "+binary)
		}
	}
	return runtimeItem, nil
}

func runtimeExecutable(item software.Software) string {
	switch item.Kind {
	case software.KIND_PHP:
		return filepath.Join(item.Path, "php-cgi.exe")
	case software.KIND_NODE:
		return filepath.Join(item.Path, "node.exe")
	case software.KIND_PYTHON:
		return filepath.Join(item.Path, "python.exe")
	case software.KIND_GO:
		return filepath.Join(item.Path, "bin", "go.exe")
	case software.KIND_JAVA:
		return filepath.Join(item.Path, "bin", "java.exe")
	default:
		return ""
	}
}

func runtimePath(item *software.Software) string {
	if item == nil {
		return os.Getenv("PATH")
	}
	entries := []string{item.Path}
	if item.Kind == software.KIND_GO || item.Kind == software.KIND_JAVA {
		entries = append(entries, filepath.Join(item.Path, "bin"))
	}
	entries = append(entries, os.Getenv("PATH"))
	return strings.Join(entries, ";")
}

func (s *ProjectService) startProcess(current project.Project, runtimeItem *software.Software) error {
	key := projectKey(current.ID)
	if status := s.runner.Status(key); status.Running {
		return nil
	}
	workFolder := project.ResolvePath(current.Root, current.Run)
	info, err := os.Stat(workFolder)
	if err != nil || !info.IsDir() {
		return appErrors.New(appErrors.PATH_NOT_ACCESSIBLE, "项目运行目录不存在: "+workFolder)
	}
	if current.Port > 0 && !s.probe.Available(current.Port) {
		return appErrors.New(appErrors.PORT_ALREADY_IN_USE, fmt.Sprintf("项目端口 %d 已被占用", current.Port))
	}
	commandText := strings.TrimSpace(current.Command)
	if current.Port > 0 {
		commandText = strings.ReplaceAll(commandText, "%PORT%", strconv.Itoa(current.Port))
	}
	logPath := paths.ProjectLog(s.base, current.Name, "run.log")
	environment := []string{"PATH=" + runtimePath(runtimeItem)}
	if current.Port > 0 {
		environment = append(environment, "PORT="+strconv.Itoa(current.Port))
	}
	projectID := current.ID
	options := ports.Options{
		Binary:      paths.Shell(),
		Args:        []string{"/C", commandText},
		Directory:   workFolder,
		Environment: environment,
		Log:         logPath,
		OnExit: func(_ string, message string) {
			s.mu.Lock()
			delete(s.active, projectID)
			s.failed[projectID] = message
			s.mu.Unlock()
			if project.UsesNginx(current.Kind) {
				if syncError := s.syncRoutes(context.Background()); syncError != nil {
					log.Printf("移除已退出项目路由失败: %v", syncError)
				}
			}
			s.bus.Emit("project:status-changed", map[string]interface{}{
				"id":      projectID,
				"state":   project.STATE_ERROR,
				"message": message,
			})
		},
	}
	if _, err := s.runner.Start(key, options); err != nil {
		return err
	}
	if current.Port == 0 {
		return nil
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
		return appErrors.Wrap(stopErr, appErrors.PROCESS_START_FAILED, "停止超时项目失败")
	}
	tail := process.ReadTail(logPath, LOG_TAIL_LIMIT)
	return appErrors.New(appErrors.PROCESS_START_FAILED, "项目启动超时或退出:\n"+strings.Join(tail, "\n"))
}

func (s *ProjectService) markActive(id int64) {
	s.mu.Lock()
	s.active[id] = true
	delete(s.failed, id)
	s.mu.Unlock()
}

func (s *ProjectService) markFailed(id int64, err error) {
	s.mu.Lock()
	delete(s.active, id)
	s.failed[id] = err.Error()
	s.mu.Unlock()
}

func (s *ProjectService) syncRoutes(ctx context.Context) error {
	list, err := s.projects.List(ctx)
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
		if project.IsLocal(item.Kind) {
			continue
		}
		if !activeCopy[item.ID] {
			continue
		}
		phpBinary := ""
		environmentFile := ""
		if item.Kind == project.KIND_PHP && item.Software != nil {
			if found, ok := softwareMap[*item.Software]; ok && found.State == software.STATE_INSTALLED {
				phpBinary = filepath.Join(found.Path, "php-cgi.exe")
			}
		}
		if item.Kind == project.KIND_PHP {
			environmentFile, err = project.ResolveEnvironmentFile(item.Root, item.EnvironmentFile)
			if err != nil {
				return err
			}
		}
		routes = append(routes, ports.Route{
			ID:              item.ID,
			Name:            item.Name,
			Hosts:           item.Hosts,
			Kind:            item.Kind,
			Path:            project.ResolvePath(item.Root, item.Run),
			PHP:             phpBinary,
			Nginx:           item.Nginx,
			RewriteRule:     item.RewriteRule,
			EnvironmentFile: environmentFile,
			Port:            item.Port,
			Log:             paths.ProjectLog(s.base, item.Name, "access.log"),
		})
	}
	return s.nginx.SetRoutes(routes)
}

func (s *ProjectService) syncHosts(ctx context.Context) error {
	current, err := s.settings.Load(ctx)
	if err != nil {
		return err
	}
	if !current.Hosts {
		return nil
	}
	list, err := s.projects.List(ctx)
	if err != nil {
		return err
	}
	var allHosts []string
	for _, item := range list {
		if len(item.Hosts) > 0 {
			allHosts = append(allHosts, item.Hosts...)
		}
	}
	return s.hosts.Sync(allHosts)
}

func (s *ProjectService) buildView(item project.Project, softwareMap map[int64]software.Software, port int) ProjectView {
	docPath := project.ResolvePath(item.Root, item.Run)
	if item.EnvironmentFile == "" {
		item.EnvironmentFile = project.DefaultEnvironmentFile(item.Root)
	}
	version := ""
	need := ""

	runtimeKind := project.RuntimeSoftwareKind(item.Kind)
	if runtimeKind != "" {
		if item.Software != nil {
			if found, ok := softwareMap[*item.Software]; ok && found.State == software.STATE_INSTALLED {
				version = found.Version
			} else {
				need = runtimeKind
			}
		} else {
			need = runtimeKind
		}
	} else if item.Kind == project.KIND_RUST {
		if _, err := exec.LookPath("cargo"); err == nil {
			version = "system cargo"
		} else {
			need = project.KIND_RUST
		}
	}

	s.mu.RLock()
	active := s.active[item.ID]
	_, hasFailed := s.failed[item.ID]
	s.mu.RUnlock()

	state := project.STATE_STOPPED
	if hasFailed {
		state = project.STATE_ERROR
	} else if active {
		if project.IsProcess(item.Kind) {
			status := s.runner.Status(projectKey(item.ID))
			if status.Running {
				state = project.STATE_RUNNING
			} else if status.Message != "" {
				state = project.STATE_ERROR
			}
		} else {
			state = project.STATE_RUNNING
		}
	}

	projectURL := ""
	if len(item.Hosts) > 0 {
		if port > 0 && port != DEFAULT_HTTP {
			projectURL = fmt.Sprintf("http://%s:%d", item.Hosts[0], port)
		} else {
			projectURL = "http://" + item.Hosts[0]
		}
	}

	return ProjectView{
		Project:          item,
		Path:             docPath,
		Traffic:          s.metrics.Traffic(paths.ProjectLog(s.base, item.Name, "access.log")),
		Version:          version,
		State:            state,
		URL:              projectURL,
		Need:             need,
		EnvironmentFiles: project.FindEnvironmentFiles(item.Root),
	}
}

func (s *ProjectService) findPort(ctx context.Context) int {
	list, err := s.projects.List(ctx)
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

func projectKey(id int64) string {
	return "project:" + strconv.FormatInt(id, 10)
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
	case project.KIND_STATIC:
		content := fmt.Sprintf("<!DOCTYPE html>\n<html lang=\"zh-CN\">\n<head><meta charset=\"utf-8\"><title>%s</title></head>\n<body><h1>%s 运行正常</h1></body>\n</html>\n", name, name)
		return os.WriteFile(filepath.Join(folder, "index.html"), []byte(content), 0o644)
	case project.KIND_PHP:
		content := fmt.Sprintf("<?php\necho '<h1>%s (PHP ' . PHP_VERSION . ')</h1>';\n", name)
		return os.WriteFile(filepath.Join(folder, "index.php"), []byte(content), 0o644)
	case project.KIND_NODE:
		pkg := fmt.Sprintf("{\n  \"name\": %q,\n  \"version\": \"1.0.0\",\n  \"main\": \"server.js\",\n  \"scripts\": {\n    \"start\": \"node server.js\"\n  }\n}\n", name)
		if err := os.WriteFile(filepath.Join(folder, "package.json"), []byte(pkg), 0o644); err != nil {
			return err
		}
		script := fmt.Sprintf("const http = require('http');\nconst port = Number(process.env.PORT);\nhttp.createServer((request, response) => {\n  response.writeHead(200, {'Content-Type': 'text/html; charset=utf-8'});\n  response.end('<h1>%s (Node.js ' + process.version + ')</h1>');\n}).listen(port, '%s');\n", name, paths.LOOPBACK)
		return os.WriteFile(filepath.Join(folder, "server.js"), []byte(script), 0o644)
	case project.KIND_PYTHON:
		script := fmt.Sprintf("from http.server import BaseHTTPRequestHandler, HTTPServer\nimport os\n\nclass Handler(BaseHTTPRequestHandler):\n    def do_GET(self):\n        body = %q.encode('utf-8')\n        self.send_response(200)\n        self.send_header('Content-Type', 'text/html; charset=utf-8')\n        self.send_header('Content-Length', str(len(body)))\n        self.end_headers()\n        self.wfile.write(body)\n\nHTTPServer(('127.0.0.1', int(os.environ.get('PORT', '3000'))), Handler).serve_forever()\n", fmt.Sprintf("<h1>%s (Python)</h1>", name))
		return os.WriteFile(filepath.Join(folder, "app.py"), []byte(script), 0o644)
	case project.KIND_GO:
		module := "module example.local/project\n\ngo 1.22\n"
		if err := os.WriteFile(filepath.Join(folder, "go.mod"), []byte(module), 0o644); err != nil {
			return err
		}
		script := fmt.Sprintf("package main\n\nimport (\n    \"fmt\"\n    \"net/http\"\n    \"os\"\n)\n\nfunc main() {\n    port := os.Getenv(\"PORT\")\n    if port == \"\" { port = \"3000\" }\n    http.HandleFunc(\"/\", func(writer http.ResponseWriter, request *http.Request) {\n        fmt.Fprintf(writer, \"<h1>%s (Go)</h1>\")\n    })\n    http.ListenAndServe(\"127.0.0.1:\"+port, nil)\n}\n", name)
		return os.WriteFile(filepath.Join(folder, "main.go"), []byte(script), 0o644)
	case project.KIND_RUST:
		if err := os.MkdirAll(filepath.Join(folder, "src"), 0o755); err != nil {
			return err
		}
		manifest := "[package]\nname = \"devenv-project\"\nversion = \"0.1.0\"\nedition = \"2021\"\n"
		if err := os.WriteFile(filepath.Join(folder, "Cargo.toml"), []byte(manifest), 0o644); err != nil {
			return err
		}
		script := fmt.Sprintf("use std::io::{Read, Write};\nuse std::net::{TcpListener, TcpStream};\n\nfn main() {\n    let port = std::env::var(\"PORT\").unwrap_or_else(|_| \"3000\".to_string());\n    let listener = TcpListener::bind(format!(\"127.0.0.1:{}\", port)).unwrap();\n    for stream in listener.incoming() {\n        if let Ok(stream) = stream { respond(stream); }\n    }\n}\n\nfn respond(mut stream: TcpStream) {\n    let mut request = [0; 512];\n    let _ = stream.read(&mut request);\n    let body = \"<h1>%s (Rust)</h1>\";\n    let response = format!(\"HTTP/1.1 200 OK\\r\\nContent-Length: {}\\r\\nContent-Type: text/html; charset=utf-8\\r\\n\\r\\n{}\", body.len(), body);\n    let _ = stream.write_all(response.as_bytes());\n}\n", name)
		return os.WriteFile(filepath.Join(folder, "src", "main.rs"), []byte(script), 0o644)
	case project.KIND_JAVA:
		script := fmt.Sprintf("import com.sun.net.httpserver.HttpServer;\nimport java.io.OutputStream;\nimport java.net.InetSocketAddress;\n\npublic class Main {\n    public static void main(String[] args) throws Exception {\n        int port = Integer.parseInt(System.getenv().getOrDefault(\"PORT\", \"3000\"));\n        HttpServer server = HttpServer.create(new InetSocketAddress(\"127.0.0.1\", port), 0);\n        server.createContext(\"/\", exchange -> {\n            byte[] body = \"<h1>%s (Java)</h1>\".getBytes();\n            exchange.sendResponseHeaders(200, body.length);\n            try (OutputStream output = exchange.getResponseBody()) { output.write(body); }\n        });\n        server.start();\n    }\n}\n", name)
		return os.WriteFile(filepath.Join(folder, "Main.java"), []byte(script), 0o644)
	}
	return nil
}
