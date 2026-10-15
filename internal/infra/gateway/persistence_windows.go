//go:build windows

package gateway

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"devenv/internal/infra/windowsservice"
	"devenv/internal/ports"
	"devenv/internal/shared/paths"
)

const gatewayServiceAccount = `NT AUTHORITY\LocalService`

type persistentService struct {
	name       string
	options    ports.Options
	autoStart  bool
	accessList []serviceAccess
}

type serviceAccess struct {
	path       string
	permission string
}

func configurePersistentService(base string, service persistentService) error {
	configPath := paths.ServiceConfig(base, service.name)
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return err
	}
	definition, err := windowsservice.Encode(windowsservice.Definition{
		ServiceName: service.name,
		Binary:      service.options.Binary,
		Arguments:   service.options.Args,
		Environment: service.options.Environment,
		Directory:   service.options.Directory,
		Log:         service.options.Log,
	})
	if err != nil {
		return err
	}
	temporaryPath := configPath + ".tmp"
	if err := os.WriteFile(temporaryPath, definition, 0o644); err != nil {
		return err
	}
	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		_ = os.Remove(temporaryPath)
		return err
	}
	if err := os.Rename(temporaryPath, configPath); err != nil {
		_ = os.Remove(temporaryPath)
		return err
	}

	service.accessList = append(service.accessList,
		serviceAccess{path: base, permission: "RX"},
		serviceAccess{path: filepath.Dir(configPath), permission: "RX"},
		serviceAccess{path: service.options.Binary, permission: "RX"},
	)
	for _, access := range service.accessList {
		if strings.TrimSpace(access.path) == "" {
			continue
		}
		if err := grantPersistentServiceAccess(access); err != nil {
			return err
		}
	}

	executable, err := windowsservice.ServiceExecutable(base)
	if err != nil {
		return err
	}
	startType := "demand"
	if service.autoStart {
		startType = "auto"
	}
	commandLine := fmt.Sprintf(`"%s" --devenv-service-config "%s"`, executable, configPath)
	if persistentServiceExists(service.name) {
		_, err = runPersistentServiceCommand("sc.exe", "config", service.name, "binPath=", commandLine, "start=", startType, "obj=", gatewayServiceAccount)
		return err
	}
	_, err = runPersistentServiceCommand("sc.exe", "create", service.name, "binPath=", commandLine, "start=", startType, "obj=", gatewayServiceAccount, "DisplayName=", service.name)
	return err
}

func startPersistentService(name string) error {
	output, err := runPersistentServiceCommand("sc.exe", "start", name)
	if err != nil && !strings.Contains(string(output), "1056") {
		return fmt.Errorf("start Windows service %s: %w: %s", name, err, strings.TrimSpace(string(output)))
	}
	return nil
}

func stopPersistentService(name string) error {
	output, err := runPersistentServiceCommand("sc.exe", "stop", name)
	if err != nil && !strings.Contains(string(output), "1062") && !strings.Contains(string(output), "1061") {
		return fmt.Errorf("stop Windows service %s: %w: %s", name, err, strings.TrimSpace(string(output)))
	}
	return nil
}

func persistentServiceExists(name string) bool {
	output, err := runPersistentServiceCommand("sc.exe", "query", name)
	return err == nil || !strings.Contains(string(output), "1060")
}

func grantPersistentServiceAccess(access serviceAccess) error {
	permission := gatewayServiceAccount + ":(OI)(CI)" + access.permission
	_, err := runPersistentServiceCommand("icacls", access.path, "/grant", permission, "/T", "/C")
	return err
}

func runPersistentServiceCommand(name string, arguments ...string) ([]byte, error) {
	command := exec.Command(name, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return command.CombinedOutput()
}
