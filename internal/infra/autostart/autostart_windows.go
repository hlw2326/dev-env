//go:build windows

package autostart

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const (
	regKey  = `Software\Microsoft\Windows\CurrentVersion\Run`
	appName = "DevEnv"
)

func Set(enable bool) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, regKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	if enable {
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		cmd := fmt.Sprintf(`"%s" --minimized`, exe)
		return k.SetStringValue(appName, cmd)
	}
	_ = k.DeleteValue(appName)
	return nil
}

func Enabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, regKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	_, _, err = k.GetStringValue(appName)
	return err == nil
}

func SetService(name string, enable bool, binary string, args []string, dir string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, regKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	regName := "DevEnv-" + name
	if !enable {
		_ = k.DeleteValue(regName)
		return nil
	}

	appData := os.Getenv("LOCALAPPDATA")
	scriptDir := filepath.Join(appData, "DevEnv", "autostart")
	_ = os.MkdirAll(scriptDir, 0755)
	scriptFile := filepath.Join(scriptDir, name+".vbs")

	cmdLine := `"` + binary + `"`
	if len(args) > 0 {
		for _, arg := range args {
			cmdLine += ` ` + `"` + arg + `"`
		}
	}
	cmdLine = strings.ReplaceAll(cmdLine, `"`, `""`)

	content := fmt.Sprintf("Set ws = CreateObject(\"WScript.Shell\")\r\nws.CurrentDirectory = \"%s\"\r\nws.Run \"%s\", 0, False\r\n", dir, cmdLine)

	if err := os.WriteFile(scriptFile, []byte(content), 0644); err != nil {
		return err
	}

	val := fmt.Sprintf(`wscript.exe "%s"`, scriptFile)
	return k.SetStringValue(regName, val)
}

func ServiceEnabled(name string) bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, regKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	_, _, err = k.GetStringValue("DevEnv-" + name)
	return err == nil
}
