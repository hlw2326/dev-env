//go:build windows

package gateway

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"devenv/internal/infra/windowsservice"
	"devenv/internal/ports"
	"devenv/internal/shared/paths"
)

const gatewayServiceAccount = `NT AUTHORITY\LocalService`

var serviceAccessMutex sync.Mutex

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
		if err := ensurePersistentServiceAccess(base, access); err != nil {
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
	arguments := persistentServiceAccessArguments(access)
	output, err := runPersistentServiceCommand("icacls", arguments...)
	if err != nil {
		return fmt.Errorf("grant service access to %s: %w: %s", access.path, err, strings.TrimSpace(string(output)))
	}
	return nil
}

func persistentServiceAccessArguments(access serviceAccess) []string {
	permission := "*S-1-5-19:" + access.permission
	arguments := []string{access.path, "/grant", permission}
	if info, err := os.Stat(access.path); err == nil && info.IsDir() {
		arguments[2] = "*S-1-5-19:(OI)(CI)" + access.permission
	}
	return arguments
}

func ensurePersistentServiceAccess(base string, access serviceAccess) error {
	serviceAccessMutex.Lock()
	defer serviceAccessMutex.Unlock()

	state, err := loadPersistentServiceAccess(base)
	if err != nil {
		return err
	}
	key, err := persistentServiceAccessKey(access)
	if err != nil {
		return err
	}
	if servicePermissionRank(state[key]) >= servicePermissionRank(access.permission) {
		return nil
	}
	if err := grantPersistentServiceAccess(access); err != nil {
		return err
	}
	state[key] = access.permission
	return savePersistentServiceAccess(base, state)
}

func persistentServiceAccessKey(access serviceAccess) (string, error) {
	absolute, err := filepath.Abs(strings.TrimSpace(access.path))
	if err != nil {
		return "", err
	}
	return strings.ToLower(filepath.Clean(absolute)) + "|" + strconv.FormatBool(access.direct), nil
}

func servicePermissionRank(permission string) int {
	switch strings.ToUpper(strings.TrimSpace(permission)) {
	case "M":
		return 2
	case "RX":
		return 1
	default:
		return 0
	}
}

func loadPersistentServiceAccess(base string) (map[string]string, error) {
	state := make(map[string]string)
	content, err := os.ReadFile(paths.ServiceAccess(base))
	if os.IsNotExist(err) {
		return state, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(content, &state); err != nil {
		return nil, fmt.Errorf("decode persistent service access state: %w", err)
	}
	return state, nil
}

func savePersistentServiceAccess(base string, state map[string]string) error {
	path := paths.ServiceAccess(base)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	content, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	temporaryPath := path + ".tmp"
	if err := os.WriteFile(temporaryPath, content, 0o600); err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		_ = os.Remove(temporaryPath)
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		_ = os.Remove(temporaryPath)
		return err
	}
	return nil
}

func runPersistentServiceCommand(name string, arguments ...string) ([]byte, error) {
	command := exec.Command(name, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return command.CombinedOutput()
}
