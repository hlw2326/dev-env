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
	direct     bool
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
	// Nginx resolves long Windows paths by querying each ancestor directory.
	// Traverse-only access is insufficient for paths inside a user profile.
	for parent := filepath.Dir(base); parent != filepath.Dir(parent); parent = filepath.Dir(parent) {
		service.accessList = append(service.accessList, serviceAccess{path: parent, permission: "RX", direct: true})
	}
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
	return windowsservice.Start(name)
}

func stopPersistentService(name string) error {
	return windowsservice.Stop(name, GATEWAY_START_TIMEOUT)
}

func persistentServiceExists(name string) bool {
	exists, err := windowsservice.Exists(name)
	return err == nil && exists
}

func persistentServiceOwnsProcess(name string, pid int, executable string) bool {
	return windowsservice.OwnsProcess(name, pid, executable)
}

func grantPersistentServiceAccess(access serviceAccess) error {
	if access.direct {
		return windowsservice.GrantDirectoryRead(access.path)
	}
	permission := "*S-1-5-19:" + access.permission
	arguments := []string{access.path, "/grant", permission}
	if info, err := os.Stat(access.path); err == nil && info.IsDir() && !access.direct {
		arguments[2] = "*S-1-5-19:(OI)(CI)" + access.permission
		arguments = append(arguments, "/T", "/C")
	}
	output, err := runPersistentServiceCommand("icacls", arguments...)
	if err != nil {
		return fmt.Errorf("grant service access to %s: %w: %s", access.path, err, strings.TrimSpace(string(output)))
	}
	return nil
}

func runPersistentServiceCommand(name string, arguments ...string) ([]byte, error) {
	command := exec.Command(name, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return command.CombinedOutput()
}
