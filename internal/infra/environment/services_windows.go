//go:build windows

package environment

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"devenv/internal/domain/software"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

const SERVICE_REGISTRY_KEY = `SYSTEM\CurrentControlSet\Services`

func discoverServices(ctx context.Context, probes []software.CommandProbe) ([]software.CommandLocation, []string) {
	entries := make([]software.CommandLocation, 0)
	warnings := make([]string, 0)
	servers := make(map[string]software.CommandProbe)
	for _, probe := range probes {
		if probe.Role == software.ROLE_SERVER {
			servers[probe.Name+".exe"] = probe
		}
	}
	if len(servers) == 0 {
		return entries, warnings
	}
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, SERVICE_REGISTRY_KEY, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return entries, append(warnings, fmt.Sprintf("读取 Windows 服务登记失败: %v", err))
	}
	defer key.Close()
	names, err := key.ReadSubKeyNames(-1)
	if err != nil {
		return entries, append(warnings, fmt.Sprintf("枚举 Windows 服务失败: %v", err))
	}
	unreadable := 0
	for _, name := range names {
		if ctx.Err() != nil {
			warnings = append(warnings, "服务检测被中断，结果可能不完整")
			break
		}
		serviceKey, err := registry.OpenKey(registry.LOCAL_MACHINE, SERVICE_REGISTRY_KEY+`\`+name, registry.QUERY_VALUE)
		if err != nil {
			unreadable++
			continue
		}
		imagePath, _, err := serviceKey.GetStringValue("ImagePath")
		serviceKey.Close()
		if err == registry.ErrNotExist {
			continue
		}
		if err != nil {
			unreadable++
			continue
		}
		path, err := serviceExecutable(imagePath)
		if err != nil {
			continue
		} // 系统驱动等非 exe 服务不在本次检测范围。
		probe, matched := servers[strings.ToLower(filepath.Base(path))]
		if !matched {
			continue
		}
		entry := newLocation(probe, path, software.SOURCE_SERVICE)
		entry.ServiceName = name
		entry.ServiceState, err = readServiceState(name)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("无法读取服务 %s 状态: %v", name, err))
		}
		entries = append(entries, entry)
	}
	if unreadable > 0 {
		warnings = append(warnings, fmt.Sprintf("有 %d 项服务登记无法读取，检测范围不完整", unreadable))
	}
	return entries, warnings
}

func serviceExecutable(commandLine string) (string, error) {
	expanded, err := registry.ExpandString(commandLine)
	if err != nil {
		return "", err
	}
	expanded = strings.TrimSpace(expanded)
	var path string
	if strings.HasPrefix(expanded, `"`) {
		arguments, err := windows.DecomposeCommandLine(expanded)
		if err != nil || len(arguments) == 0 {
			return "", fmt.Errorf("无法解析服务路径")
		}
		path = arguments[0]
	} else {
		end := strings.Index(strings.ToLower(expanded), ".exe")
		if end < 0 {
			return "", fmt.Errorf("不是 exe 服务")
		}
		path = expanded[:end+4]
	}
	if !filepath.IsAbs(path) || !strings.EqualFold(filepath.Ext(path), ".exe") {
		return "", fmt.Errorf("服务路径不是绝对 exe 路径")
	}
	return filepath.Clean(path), nil
}

func readServiceState(name string) (string, error) {
	manager, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT)
	if err != nil {
		return "无法读取", err
	}
	defer windows.CloseServiceHandle(manager)
	namePointer, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return "无法读取", err
	}
	handle, err := windows.OpenService(manager, namePointer, windows.SERVICE_QUERY_STATUS)
	if err != nil {
		return "无法读取", err
	}
	service := mgr.Service{Name: name, Handle: handle}
	defer service.Close()
	status, err := service.Query()
	if err != nil {
		return "无法读取", err
	}
	switch status.State {
	case svc.Running:
		return "运行中", nil
	case svc.Stopped:
		return "已停止", nil
	case svc.StartPending:
		return "启动中", nil
	case svc.StopPending:
		return "停止中", nil
	case svc.Paused:
		return "已暂停", nil
	default:
		return "状态变更中", nil
	}
}
