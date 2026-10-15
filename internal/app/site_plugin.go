package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"devenv/internal/domain/plugin"
	"devenv/internal/domain/site"
	"devenv/internal/domain/software"
	"devenv/internal/ports"
	appErrors "devenv/internal/shared/errors"
	"devenv/internal/shared/paths"
)

const pluginExitMessage = "进程已退出"

type pluginDefinition struct {
	ID          string
	Name        string
	Description string
	Actions     []plugin.Action
}

var projectPluginDefinitions = []pluginDefinition{
	{
		ID:          plugin.ID_COMPOSER,
		Name:        "Composer",
		Description: "管理 PHP 项目的依赖包和自动加载文件",
		Actions: []plugin.Action{
			{ID: plugin.ACTION_INSTALL, Name: "安装依赖"},
			{ID: plugin.ACTION_UPDATE, Name: "更新依赖"},
			{ID: plugin.ACTION_DUMP_AUTOLOAD, Name: "生成自动加载"},
		},
	},
	{
		ID:          plugin.ID_NODE_PACKAGE,
		Name:        "Node 包管理",
		Description: "根据项目锁文件选择包管理器并安装依赖",
		Actions: []plugin.Action{
			{ID: plugin.ACTION_INSTALL, Name: "安装依赖"},
		},
	},
	{
		ID:          plugin.ID_PYTHON_PACKAGE,
		Name:        "Python 依赖",
		Description: "根据 requirements.txt 或 pyproject.toml 安装 Python 项目依赖",
		Actions: []plugin.Action{
			{ID: plugin.ACTION_INSTALL, Name: "安装依赖"},
		},
	},
	{
		ID:          plugin.ID_GO_MODULE,
		Name:        "Go 模块",
		Description: "整理 Go 模块依赖并生成校验信息",
		Actions: []plugin.Action{
			{ID: plugin.ACTION_UPDATE, Name: "整理依赖"},
		},
	},
	{
		ID:          plugin.ID_RUST_PACKAGE,
		Name:        "Cargo",
		Description: "获取 Rust 项目的 Cargo 依赖",
		Actions: []plugin.Action{
			{ID: plugin.ACTION_INSTALL, Name: "获取依赖"},
		},
	},
	{
		ID:          plugin.ID_JAVA_BUILD,
		Name:        "Java 构建",
		Description: "解析 Maven 或 Gradle 项目的依赖",
		Actions: []plugin.Action{
			{ID: plugin.ACTION_INSTALL, Name: "解析依赖"},
		},
	},
}

func (s *SiteService) Plugins(ctx context.Context, id int64) ([]plugin.ProjectPlugin, error) {
	item, err := s.sites.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, appErrors.New(appErrors.NOT_FOUND, "项目不存在")
	}

	result := make([]plugin.ProjectPlugin, 0)
	for _, definition := range projectPluginDefinitions {
		if !s.pluginMatches(*item, definition.ID) {
			continue
		}
		current := plugin.ProjectPlugin{
			ID:          definition.ID,
			Name:        definition.Name,
			Description: definition.Description,
			Available:   true,
			Actions:     append([]plugin.Action(nil), definition.Actions...),
		}
		if definition.ID == plugin.ID_COMPOSER {
			current = s.composerPlugin(ctx, current)
		}
		if definition.ID == plugin.ID_NODE_PACKAGE {
			current = s.nodePackagePlugin(*item, current)
		}
		if definition.ID == plugin.ID_PYTHON_PACKAGE || definition.ID == plugin.ID_GO_MODULE || definition.ID == plugin.ID_RUST_PACKAGE || definition.ID == plugin.ID_JAVA_BUILD {
			current = s.runtimePlugin(ctx, *item, current, definition.ID)
		}
		if current.Actions == nil {
			current.Actions = make([]plugin.Action, 0)
		}
		result = append(result, current)
	}
	return result, nil
}

func (s *SiteService) RunPlugin(ctx context.Context, input plugin.ActionInput) error {
	item, err := s.sites.Get(ctx, input.ProjectID)
	if err != nil {
		return err
	}
	if item == nil {
		return appErrors.New(appErrors.NOT_FOUND, "项目不存在")
	}
	available, command, err := s.pluginCommand(ctx, *item, input.PluginID, input.ActionID)
	if err != nil {
		return err
	}
	if !available {
		return appErrors.New(appErrors.RUNTIME_NOT_INSTALLED, "项目插件运行环境不可用")
	}

	key := fmt.Sprintf("project-plugin:%d:%s", item.ID, input.PluginID)
	if status := s.runner.Status(key); status.Running {
		return appErrors.New(appErrors.PROCESS_START_FAILED, "该项目插件任务正在运行")
	}
	workFolder := filepath.Clean(item.Root)
	logPath := paths.SiteLog(s.base, item.Name, "plugin-"+input.PluginID+".log")
	options := ports.Options{
		Binary:      paths.Shell(),
		Args:        []string{"/C", command},
		Directory:   workFolder,
		Environment: s.pluginEnvironment(ctx, *item),
		Log:         logPath,
		OnExit: func(_ string, message string) {
			s.emitPluginStatus(item.ID, input.PluginID, input.ActionID, false, message == pluginExitMessage, message)
		},
	}
	s.emitPluginStatus(item.ID, input.PluginID, input.ActionID, true, true, "")
	if _, err := s.runner.Start(key, options); err != nil {
		s.emitPluginStatus(item.ID, input.PluginID, input.ActionID, false, false, err.Error())
		return err
	}
	return nil
}

func (s *SiteService) pluginMatches(item site.Site, pluginID string) bool {
	root := site.ResolvePath(item.Root, "")
	switch pluginID {
	case plugin.ID_COMPOSER:
		return item.Kind == site.KIND_PHP && pluginFileExists(filepath.Join(root, "composer.json"))
	case plugin.ID_NODE_PACKAGE:
		return (item.Kind == site.KIND_NODE || item.Kind == site.KIND_ELECTRON) && pluginFileExists(filepath.Join(root, "package.json"))
	case plugin.ID_PYTHON_PACKAGE:
		return item.Kind == site.KIND_PYTHON && (pluginFileExists(filepath.Join(root, "requirements.txt")) || pluginFileExists(filepath.Join(root, "pyproject.toml")))
	case plugin.ID_GO_MODULE:
		return item.Kind == site.KIND_GO && pluginFileExists(filepath.Join(root, "go.mod"))
	case plugin.ID_RUST_PACKAGE:
		return item.Kind == site.KIND_RUST && pluginFileExists(filepath.Join(root, "Cargo.toml"))
	case plugin.ID_JAVA_BUILD:
		return item.Kind == site.KIND_JAVA && (pluginFileExists(filepath.Join(root, "pom.xml")) || pluginFileExists(filepath.Join(root, "build.gradle")) || pluginFileExists(filepath.Join(root, "build.gradle.kts")))
	default:
		return false
	}
}

func (s *SiteService) composerPlugin(ctx context.Context, item plugin.ProjectPlugin) plugin.ProjectPlugin {
	composer, err := s.software.GetMain(ctx, software.KIND_COMPOSER)
	if err != nil || composer == nil || composer.State != software.STATE_INSTALLED {
		item.Available = false
		item.Reason = "请先在软件商城安装 Composer"
		return item
	}
	if !pluginFileExists(filepath.Join(composer.Path, "composer.bat")) {
		item.Available = false
		item.Reason = "Composer 可执行文件不存在"
		return item
	}
	item.Version = composer.Version
	return item
}

func (s *SiteService) nodePackagePlugin(item site.Site, current plugin.ProjectPlugin) plugin.ProjectPlugin {
	manager := nodePackageManager(item.Root)
	if manager == "" {
		current.Available = false
		current.Reason = "未找到 npm、pnpm 或 yarn"
		return current
	}
	if _, err := exec.LookPath(manager); err != nil {
		current.Available = false
		current.Reason = manager + " 不在系统 PATH 中"
	}
	current.Version = manager
	return current
}

func (s *SiteService) runtimePlugin(ctx context.Context, item site.Site, current plugin.ProjectPlugin, pluginID string) plugin.ProjectPlugin {
	command := runtimePluginCommand(item, pluginID)
	if command == "" {
		current.Available = false
		current.Reason = "未找到项目依赖配置文件"
		return current
	}
	tool := strings.Fields(command)[0]
	if available, reason := s.pluginToolAvailable(ctx, item, tool); !available {
		current.Available = false
		current.Reason = reason
		return current
	}
	current.Version = tool
	return current
}

func (s *SiteService) pluginCommand(ctx context.Context, item site.Site, pluginID, actionID string) (bool, string, error) {
	if !s.pluginMatches(item, pluginID) {
		return false, "", appErrors.New(appErrors.INVALID_ARGUMENT, "当前项目不支持该插件")
	}
	switch pluginID {
	case plugin.ID_COMPOSER:
		composer, err := s.software.GetMain(ctx, software.KIND_COMPOSER)
		if err != nil {
			return false, "", err
		}
		if composer == nil || composer.State != software.STATE_INSTALLED {
			return false, "", appErrors.New(appErrors.RUNTIME_NOT_INSTALLED, "请先在软件商城安装 Composer")
		}
		binary := filepath.Join(composer.Path, "composer.bat")
		if !pluginFileExists(binary) {
			return false, "", appErrors.New(appErrors.PATH_NOT_ACCESSIBLE, "Composer 可执行文件不存在")
		}
		if actionID != plugin.ACTION_INSTALL && actionID != plugin.ACTION_UPDATE && actionID != plugin.ACTION_DUMP_AUTOLOAD {
			return false, "", appErrors.New(appErrors.INVALID_ARGUMENT, "不支持的 Composer 操作")
		}
		return true, fmt.Sprintf("\"%s\" %s", binary, actionID), nil
	case plugin.ID_NODE_PACKAGE:
		manager := nodePackageManager(item.Root)
		if manager == "" {
			return false, "", appErrors.New(appErrors.RUNTIME_NOT_INSTALLED, "未找到 npm、pnpm 或 yarn")
		}
		if _, err := exec.LookPath(manager); err != nil {
			return false, "", appErrors.Wrap(err, appErrors.RUNTIME_NOT_INSTALLED, manager+" 不在系统 PATH 中")
		}
		if actionID != plugin.ACTION_INSTALL {
			return false, "", appErrors.New(appErrors.INVALID_ARGUMENT, "不支持的 Node 包管理操作")
		}
		return true, manager + " install", nil
	case plugin.ID_PYTHON_PACKAGE, plugin.ID_GO_MODULE, plugin.ID_RUST_PACKAGE, plugin.ID_JAVA_BUILD:
		command := runtimePluginCommand(item, pluginID)
		if command == "" {
			return false, "", appErrors.New(appErrors.INVALID_ARGUMENT, "未找到项目依赖配置文件")
		}
		tool := strings.Fields(command)[0]
		if available, reason := s.pluginToolAvailable(ctx, item, tool); !available {
			return false, "", appErrors.New(appErrors.RUNTIME_NOT_INSTALLED, reason)
		}
		if actionID != plugin.ACTION_INSTALL && actionID != plugin.ACTION_UPDATE {
			return false, "", appErrors.New(appErrors.INVALID_ARGUMENT, "不支持的项目依赖操作")
		}
		return true, command, nil
	default:
		return false, "", appErrors.New(appErrors.INVALID_ARGUMENT, "不支持的项目插件")
	}
}

func (s *SiteService) pluginToolAvailable(ctx context.Context, item site.Site, tool string) (bool, string) {
	if strings.HasSuffix(strings.ToLower(tool), ".bat") && pluginFileExists(filepath.Join(site.ResolvePath(item.Root, ""), tool)) {
		return true, ""
	}
	if (item.Kind == site.KIND_PYTHON || item.Kind == site.KIND_GO) && item.Software != nil {
		runtime, err := s.software.Get(ctx, *item.Software)
		if err != nil {
			return false, err.Error()
		}
		if runtime == nil || runtime.State != software.STATE_INSTALLED {
			return false, software.Name(site.RuntimeSoftwareKind(item.Kind)) + " 运行环境未安装"
		}
		binary := runtimeExecutable(*runtime)
		if binary == "" {
			return false, tool + " 运行环境不可用"
		}
		if _, err := os.Stat(binary); err != nil {
			return false, binary + " 不存在"
		}
		return true, ""
	}
	if _, err := exec.LookPath(tool); err != nil {
		return false, tool + " 不在系统 PATH 中"
	}
	return true, ""
}

func (s *SiteService) pluginEnvironment(ctx context.Context, item site.Site) []string {
	if item.Software == nil {
		return nil
	}
	runtime, err := s.software.Get(ctx, *item.Software)
	if err != nil || runtime == nil || runtime.Path == "" {
		return nil
	}
	return []string{"PATH=" + runtimePath(runtime)}
}

func runtimePluginCommand(item site.Site, pluginID string) string {
	root := site.ResolvePath(item.Root, "")
	switch pluginID {
	case plugin.ID_PYTHON_PACKAGE:
		if pluginFileExists(filepath.Join(root, "requirements.txt")) {
			return "python -m pip install -r requirements.txt"
		}
		if pluginFileExists(filepath.Join(root, "pyproject.toml")) {
			return "python -m pip install -e ."
		}
	case plugin.ID_GO_MODULE:
		if pluginFileExists(filepath.Join(root, "go.mod")) {
			return "go mod tidy"
		}
	case plugin.ID_RUST_PACKAGE:
		if pluginFileExists(filepath.Join(root, "Cargo.toml")) {
			return "cargo fetch"
		}
	case plugin.ID_JAVA_BUILD:
		if pluginFileExists(filepath.Join(root, "pom.xml")) {
			return "mvn dependency:resolve"
		}
		if pluginFileExists(filepath.Join(root, "build.gradle")) || pluginFileExists(filepath.Join(root, "build.gradle.kts")) {
			if pluginFileExists(filepath.Join(root, "gradlew.bat")) {
				return "gradlew.bat dependencies"
			}
			return "gradle dependencies"
		}
	}
	return ""
}

func nodePackageManager(root string) string {
	for _, candidate := range []struct {
		file    string
		command string
	}{
		{file: "pnpm-lock.yaml", command: "pnpm"},
		{file: "yarn.lock", command: "yarn"},
		{file: "package-lock.json", command: "npm"},
	} {
		if pluginFileExists(filepath.Join(root, candidate.file)) {
			return candidate.command
		}
	}
	return "npm"
}

func pluginFileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func (s *SiteService) emitPluginStatus(id int64, pluginID, actionID string, running, ok bool, message string) {
	s.bus.Emit("project:plugin-status-changed", map[string]interface{}{
		"id":      id,
		"plugin":  pluginID,
		"action":  actionID,
		"running": running,
		"ok":      ok,
		"message": strings.TrimSpace(message),
	})
}
