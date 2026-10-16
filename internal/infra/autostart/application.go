package autostart

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const applicationName = "DevEnv"

type Application struct{}

func NewApplication() *Application { return &Application{} }

func (a *Application) Enabled() (bool, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.QUERY_VALUE)
	if err == registry.ErrNotExist {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("读取软件开机启动设置失败: %w", err)
	}
	defer key.Close()
	command, _, err := key.GetStringValue(applicationName)
	if err == registry.ErrNotExist {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	arguments, err := windows.DecomposeCommandLine(command)
	if err != nil {
		return false, err
	}
	executable, err := os.Executable()
	if err != nil {
		return false, err
	}
	return len(arguments) == 1 && strings.EqualFold(filepath.Clean(arguments[0]), filepath.Clean(executable)), nil
}

func applicationCommand(executable string) (string, error) {
	if !filepath.IsAbs(executable) {
		return "", fmt.Errorf("启动程序路径必须为绝对路径")
	}
	if strings.HasSuffix(strings.ToLower(filepath.Base(executable)), "-dev.exe") {
		return "", fmt.Errorf("开发调试程序依赖开发服务器，请在安装版 DevEnv 中设置开机启动")
	}
	return windows.ComposeCommandLine([]string{executable}), nil
}

func (a *Application) Set(enabled bool) error {
	if !enabled {
		return Remove(applicationName)
	}
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	command, err := applicationCommand(executable)
	if err != nil {
		return err
	}
	return Set(applicationName, command)
}
