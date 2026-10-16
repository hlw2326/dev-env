//go:build windows

package windowsservice

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

// ExecutablePath reads the registered DevEnv service definition without running it.
func ExecutablePath(name string) (string, error) {
	service, err := openService(name, windows.SERVICE_QUERY_CONFIG)
	if err != nil {
		return "", err
	}
	defer service.Close()
	config, err := service.Config()
	if err != nil {
		return "", err
	}
	arguments, err := windows.DecomposeCommandLine(config.BinaryPathName)
	if err != nil {
		return "", err
	}
	path, ok := ConfigPathFromArguments(arguments)
	if !ok {
		return "", fmt.Errorf("服务 %s 未使用 DevEnv 服务定义", name)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var definition Definition
	if err := json.Unmarshal(raw, &definition); err != nil {
		return "", err
	}
	if definition.ServiceName != name || !filepath.IsAbs(definition.Binary) {
		return "", fmt.Errorf("服务 %s 的定义无效", name)
	}
	return definition.Binary, nil
}

func openService(name string, access uint32) (*mgr.Service, error) {
	manager, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT)
	if err != nil {
		return nil, err
	}
	defer windows.CloseServiceHandle(manager)
	serviceName, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return nil, err
	}
	handle, err := windows.OpenService(manager, serviceName, access)
	if err != nil {
		return nil, err
	}
	return &mgr.Service{Name: name, Handle: handle}, nil
}

func Exists(name string) (bool, error) {
	service, err := openService(name, windows.SERVICE_QUERY_STATUS)
	if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	defer service.Close()
	return true, nil
}

func IsStopped(name string) (bool, error) {
	service, err := openService(name, windows.SERVICE_QUERY_STATUS)
	if err != nil {
		return false, err
	}
	defer service.Close()
	status, err := service.Query()
	return status.State == svc.Stopped, err
}

func Start(name string) error {
	service, err := openService(name, windows.SERVICE_START)
	if err != nil {
		return err
	}
	defer service.Close()
	err = service.Start()
	if errors.Is(err, windows.ERROR_SERVICE_ALREADY_RUNNING) {
		return nil
	}
	return err
}

// Stop waits for SCM to confirm completion, including pending start/stop states.
func Stop(name string, timeout time.Duration) error {
	service, err := openService(name, windows.SERVICE_QUERY_STATUS|windows.SERVICE_STOP)
	if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
		return nil
	}
	if err != nil {
		return err
	}
	defer service.Close()
	deadline := time.Now().Add(timeout)
	for {
		status, err := service.Query()
		if err != nil {
			return fmt.Errorf("查询 Windows 服务 %s: %w", name, err)
		}
		if status.State == svc.Stopped {
			return nil
		}
		if status.State == svc.Running || status.State == svc.Paused {
			if _, err := service.Control(svc.Stop); err != nil && !errors.Is(err, windows.ERROR_SERVICE_NOT_ACTIVE) {
				return fmt.Errorf("停止 Windows 服务 %s: %w", name, err)
			}
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("等待 Windows 服务 %s 停止超时，当前状态: %d", name, status.State)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// OwnsProcess verifies both the executable name and its ancestry under the
// registered service host without requiring cross-account process query access.
func OwnsProcess(name string, pid int, executable string) bool {
	service, err := openService(name, windows.SERVICE_QUERY_STATUS)
	if err != nil {
		return false
	}
	defer service.Close()
	status, err := service.Query()
	if err != nil || status.State != svc.Running || status.ProcessId == 0 {
		return false
	}
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return false
	}
	defer windows.CloseHandle(snapshot)
	entry := windows.ProcessEntry32{Size: uint32(unsafe.Sizeof(windows.ProcessEntry32{}))}
	parents := make(map[uint32]uint32)
	matched := false
	for err := windows.Process32First(snapshot, &entry); err == nil; err = windows.Process32Next(snapshot, &entry) {
		parents[entry.ProcessID] = entry.ParentProcessID
		if entry.ProcessID == uint32(pid) {
			matched = strings.EqualFold(windows.UTF16ToString(entry.ExeFile[:]), filepath.Base(executable))
		}
	}
	if !matched {
		return false
	}
	visited := make(map[uint32]bool)
	for current := uint32(pid); current != 0 && !visited[current]; current = parents[current] {
		if current == status.ProcessId {
			return true
		}
		visited[current] = true
	}
	return false
}
