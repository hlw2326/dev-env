package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"devenv/internal/domain/software"
	"devenv/internal/infra/installer"
	"devenv/internal/ports"
	appErrors "devenv/internal/shared/errors"
	"devenv/internal/shared/paths"
)

var commandSyncMutex sync.Mutex

// SyncGlobalCommands refreshes command entries without starting the application
// or its services. The repository is only read by this operation.
func SyncGlobalCommands(ctx context.Context, base string, repo ports.SoftwareRepository, userPath ports.UserPathManager) error {
	commandSyncMutex.Lock()
	defer commandSyncMutex.Unlock()
	if userPath == nil {
		return appErrors.New(appErrors.PATH_NOT_ACCESSIBLE, "当前平台不支持全局命令绑定")
	}

	commandFolder := paths.Commands(base)

	definitions := make(map[string]software.CommandDefinition)
	for _, kind := range software.AllKinds() {
		for _, definition := range software.GlobalCommandDefinitions(kind) {
			definitions[definition.Name] = definition
		}
	}

	bindings := make([]globalCommandBinding, 0)
	usedNames := make(map[string]bool)
	phpBinary := ""
	if item, err := repo.GetMain(ctx, software.KIND_PHP); err != nil {
		return err
	} else if item != nil && item.State == software.STATE_INSTALLED {
		phpBinary = existingFile(filepath.Join(item.Path, "php.exe"))
	}
	javaHome := ""
	if item, err := repo.GetMain(ctx, software.KIND_JAVA); err != nil {
		return err
	} else if item != nil && item.State == software.STATE_INSTALLED {
		javaHome = existingDirectory(item.Path)
	}

	for _, kind := range software.AllKinds() {
		kindDefinitions := software.GlobalCommandDefinitions(kind)
		if len(kindDefinitions) == 0 {
			continue
		}
		item, err := repo.GetMain(ctx, kind)
		if err != nil {
			return err
		}
		if item == nil || item.State != software.STATE_INSTALLED {
			continue
		}
		for _, definition := range kindDefinitions {
			name := strings.ToLower(strings.TrimSpace(definition.Name))
			if name == "" || usedNames[name] {
				continue
			}
			target := resolveGlobalCommandPath(*item, definition)
			if target == "" {
				continue
			}
			usedNames[name] = true
			bindings = append(bindings, globalCommandBinding{
				Name:       name,
				Target:     target,
				Definition: definition,
				PHPBinary:  phpBinary,
				JavaHome:   javaHome,
			})
		}
	}

	if len(bindings) == 0 {
		if err := userPath.Remove(commandFolder); err != nil {
			return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "移除 DevEnv 全局命令路径失败")
		}
		return removeGlobalCommandFiles(commandFolder, definitions)
	}

	if err := os.MkdirAll(commandFolder, 0o755); err != nil {
		return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "创建 DevEnv 全局命令目录失败")
	}
	if err := removeGlobalCommandFiles(commandFolder, definitions); err != nil {
		return err
	}
	for _, binding := range bindings {
		commandPath := filepath.Join(commandFolder, binding.Name+".bat")
		content := renderGlobalCommand(binding)
		if err := os.WriteFile(commandPath, []byte(content), 0o644); err != nil {
			return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "写入全局命令失败")
		}
	}
	if err := userPath.Ensure(commandFolder); err != nil {
		return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "绑定 DevEnv 全局命令到用户 PATH 失败")
	}
	return nil
}

type globalCommandBinding struct {
	Name       string
	Target     string
	Definition software.CommandDefinition
	PHPBinary  string
	JavaHome   string
}

func resolveGlobalCommandPath(item software.Software, definition software.CommandDefinition) string {
	if definition.Name == "memcached" {
		return installer.FindBinary(item.Path, "memcached.exe")
	}
	target := filepath.Join(item.Path, filepath.FromSlash(definition.RelativePath))
	return existingFile(target)
}

func existingFile(path string) string {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return ""
	}
	return path
}

func existingDirectory(path string) string {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return ""
	}
	return path
}

func removeGlobalCommandFiles(commandFolder string, definitions map[string]software.CommandDefinition) error {
	for name := range definitions {
		path := filepath.Join(commandFolder, name+".bat")
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "清理 DevEnv 全局命令失败")
		}
	}
	return nil
}

func renderGlobalCommand(binding globalCommandBinding) string {
	target := quoteWindowsArgument(binding.Target)
	var builder strings.Builder
	builder.WriteString("@echo off\r\n")
	if binding.Definition.RequiresPHP {
		if binding.PHPBinary != "" {
			builder.WriteString(quoteWindowsArgument(binding.PHPBinary))
			builder.WriteByte(' ')
		} else {
			builder.WriteString("php ")
		}
		builder.WriteString(target)
		builder.WriteString(" %*\r\nexit /b %ERRORLEVEL%\r\n")
		return builder.String()
	}

	if binding.Definition.RequiresJava && binding.JavaHome != "" {
		builder.WriteString("setlocal\r\nset \"JAVA_HOME=")
		builder.WriteString(binding.JavaHome)
		builder.WriteString("\"\r\n")
	}
	if binding.Definition.UseCommandShell {
		builder.WriteString("call ")
	}
	builder.WriteString(target)
	builder.WriteString(" %*\r\n")
	if binding.Definition.RequiresJava && binding.JavaHome != "" {
		builder.WriteString("set exitCode=%ERRORLEVEL%\r\nendlocal & exit /b %exitCode%\r\n")
	} else {
		builder.WriteString("exit /b %ERRORLEVEL%\r\n")
	}
	return builder.String()
}
