//go:build windows

package environment

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"devenv/internal/domain/software"
	"devenv/internal/shared/paths"
)

const COMMAND_OUTPUT_LIMIT = 64 * 1024

func probeCommand(ctx context.Context, path string, probe software.CommandProbe, environment []string) ([]string, error) {
	if !filepath.IsAbs(path) {
		return nil, fmt.Errorf("命令路径不是绝对路径")
	}
	if strings.Contains(strings.ToLower(path), `\microsoft\windowsapps\`) {
		return nil, fmt.Errorf("Windows 应用执行别名不作为已安装环境，未执行以避免打开商店")
	}
	probeContext, cancel := context.WithTimeout(ctx, COMMAND_TIMEOUT)
	defer cancel()
	command := exec.CommandContext(probeContext, path, probe.Arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".exe", ".com":
	case ".bat", ".cmd":
		if strings.ContainsAny(path, "\"%\r\n") {
			return nil, fmt.Errorf("命令路径包含不支持的字符")
		}
		command = exec.CommandContext(probeContext, paths.Shell())
		command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true,
			CmdLine: `"` + paths.Shell() + `" /d /v:off /s /c ""` + path + `" ` + strings.Join(probe.Arguments, " ") + `"`,
		}
	default:
		return nil, fmt.Errorf("不支持检测此命令文件类型")
	}
	command.Dir = filepath.Dir(path)
	command.Env = probeEnvironment(environment)
	// 超时只结束本次检测创建的进程树，不触碰已有服务。
	command.Cancel = func() error {
		cleanupContext, stop := context.WithTimeout(context.Background(), time.Second)
		defer stop()
		killer := exec.CommandContext(cleanupContext, paths.WindowsCommand("taskkill.exe"), "/F", "/T", "/PID", strconv.Itoa(command.Process.Pid))
		killer.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		if err := killer.Run(); err != nil {
			return command.Process.Kill()
		}
		return nil
	}
	output, err := readCommandOutput(command)
	if probeContext.Err() != nil {
		return nil, fmt.Errorf("版本检测超时或已取消")
	}
	if err != nil {
		return nil, fmt.Errorf("无法读取版本: %w", err)
	}
	return parseVersions(probe, output)
}

func probeEnvironment(environment []string) []string {
	result := make([]string, 0, len(environment))
	blocked := map[string]bool{"NODE_OPTIONS": true, "JAVA_TOOL_OPTIONS": true, "JDK_JAVA_OPTIONS": true, "_JAVA_OPTIONS": true, "MAVEN_OPTS": true, "GRADLE_OPTS": true, "DOTNET_STARTUP_HOOKS": true, "GOTOOLCHAIN": true, "DOTNET_CLI_TELEMETRY_OPTOUT": true, "DOTNET_SKIP_FIRST_TIME_EXPERIENCE": true}
	for _, entry := range environment {
		name, _, _ := strings.Cut(entry, "=")
		if !blocked[strings.ToUpper(name)] {
			result = append(result, entry)
		}
	}
	return append(result, "GOTOOLCHAIN=local", "DOTNET_CLI_TELEMETRY_OPTOUT=1", "DOTNET_SKIP_FIRST_TIME_EXPERIENCE=1")
}

type commandOutput struct{ bytes.Buffer }

func (output *commandOutput) Write(data []byte) (int, error) {
	if output.Len()+len(data) > COMMAND_OUTPUT_LIMIT {
		return 0, fmt.Errorf("命令输出超过检测限制")
	}
	return output.Buffer.Write(data)
}

func readCommandOutput(command *exec.Cmd) (string, error) {
	output := &commandOutput{}
	command.Stdout = output
	command.Stderr = output
	command.WaitDelay = 200 * time.Millisecond
	err := command.Run()
	return output.String(), err
}
