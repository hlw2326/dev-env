package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"devenv/internal/domain/plugin"
	"devenv/internal/domain/project"
	"devenv/internal/domain/software"
	"devenv/internal/ports"
	appErrors "devenv/internal/shared/errors"
	"devenv/internal/shared/paths"
)

const pluginExitMessage = "进程已退出"

func (s *ProjectService) Plugins(ctx context.Context, id int64) ([]plugin.ProjectPlugin, error) {
	item, err := s.projects.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, appErrors.New(appErrors.NOT_FOUND, "项目不存在")
	}

	result := make([]plugin.ProjectPlugin, 0)
	for _, definition := range plugin.Definitions() {
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
		if definition.Handler == plugin.HANDLER_COMPOSER {
			current = s.composerPlugin(ctx, current)
		}
		if definition.Handler == plugin.HANDLER_NODE {
			current = s.nodePackagePlugin(*item, current)
		}
		if definition.Handler == plugin.HANDLER_RUNTIME {
			current = s.runtimePlugin(ctx, *item, current, definition.ID)
		}
		if current.Actions == nil {
			current.Actions = make([]plugin.Action, 0)
		}
		result = append(result, current)
	}
	return result, nil
}

func (s *ProjectService) RunPlugin(ctx context.Context, input plugin.ActionInput) error {
	item, err := s.projects.Get(ctx, input.ProjectID)
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
	logPath := paths.ProjectLog(s.base, item.Name, "plugin-"+input.PluginID+".log")
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

func (s *ProjectService) pluginMatches(item project.Project, pluginID string) bool {
	definition, ok := plugin.DefinitionFor(pluginID)
	if !ok {
		return false
	}
	root := project.ResolvePath(item.Root, "")
	if !definitionSupportsKind(definition.ProjectKinds, item.Kind) {
		return false
	}
	for _, file := range definition.RequiredFiles {
		if !pluginFileExists(filepath.Join(root, file)) {
			return false
		}
	}
	if len(definition.RequiredAnyFiles) == 0 {
		return true
	}
	for _, file := range definition.RequiredAnyFiles {
		if pluginFileExists(filepath.Join(root, file)) {
			return true
		}
	}
	return false
}

func definitionSupportsKind(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func (s *ProjectService) composerPlugin(ctx context.Context, item plugin.ProjectPlugin) plugin.ProjectPlugin {
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

func (s *ProjectService) nodePackagePlugin(item project.Project, current plugin.ProjectPlugin) plugin.ProjectPlugin {
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

func (s *ProjectService) runtimePlugin(ctx context.Context, item project.Project, current plugin.ProjectPlugin, pluginID string) plugin.ProjectPlugin {
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

func (s *ProjectService) pluginCommand(ctx context.Context, item project.Project, pluginID, actionID string) (bool, string, error) {
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

func (s *ProjectService) pluginToolAvailable(ctx context.Context, item project.Project, tool string) (bool, string) {
	if strings.HasSuffix(strings.ToLower(tool), ".bat") && pluginFileExists(filepath.Join(project.ResolvePath(item.Root, ""), tool)) {
		return true, ""
	}
	if (item.Kind == project.KIND_PYTHON || item.Kind == project.KIND_GO) && item.Software != nil {
		runtime, err := s.software.Get(ctx, *item.Software)
		if err != nil {
			return false, err.Error()
		}
		if runtime == nil || runtime.State != software.STATE_INSTALLED {
			return false, software.Name(project.RuntimeSoftwareKind(item.Kind)) + " 运行环境未安装"
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

func (s *ProjectService) pluginEnvironment(ctx context.Context, item project.Project) []string {
	if item.Software == nil {
		return nil
	}
	runtime, err := s.software.Get(ctx, *item.Software)
	if err != nil || runtime == nil || runtime.Path == "" {
		return nil
	}
	return []string{"PATH=" + runtimePath(runtime)}
}

func runtimePluginCommand(item project.Project, pluginID string) string {
	root := project.ResolvePath(item.Root, "")
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

func (s *ProjectService) emitPluginStatus(id int64, pluginID, actionID string, running, ok bool, message string) {
	s.bus.Emit("project:plugin-status-changed", map[string]interface{}{
		"id":      id,
		"plugin":  pluginID,
		"action":  actionID,
		"running": running,
		"ok":      ok,
		"message": strings.TrimSpace(message),
	})
}
