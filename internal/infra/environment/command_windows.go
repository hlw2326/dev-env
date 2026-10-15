//go:build windows

package environment

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"devenv/internal/domain/software"
	"devenv/internal/ports"
	"devenv/internal/shared/paths"
	"golang.org/x/sys/windows"
)

const (
	COMMAND_TIMEOUT = 5 * time.Second
	SCAN_TIMEOUT    = 20 * time.Second
)

type commandInspector struct{ workers chan struct{} }

func NewCommandInspector() ports.CommandInspector {
	return &commandInspector{workers: make(chan struct{}, 4)}
}

func (inspector *commandInspector) Inspect(ctx context.Context, kind, executable string) (software.EnvironmentScan, error) {
	result := software.EnvironmentScan{Entries: []software.CommandLocation{}, Warnings: []string{}}
	probes := software.InspectionProbes(kind)
	if len(probes) == 0 {
		return result, fmt.Errorf("不支持检测软件 %q", kind)
	}
	if executable != "" {
		if !filepath.IsAbs(executable) {
			return result, fmt.Errorf("请选择程序的绝对路径")
		}
		switch strings.ToLower(filepath.Ext(executable)) {
		case ".exe", ".com", ".cmd", ".bat":
		default:
			return result, fmt.Errorf("请选择 exe、cmd 或 bat 程序文件")
		}
		info, err := os.Stat(executable)
		if err != nil {
			return result, fmt.Errorf("读取所选程序失败: %w", err)
		}
		if info.IsDir() {
			return result, fmt.Errorf("请选择程序文件，不是安装目录")
		}
		name := strings.TrimSuffix(filepath.Base(executable), filepath.Ext(executable))
		for _, probe := range probes {
			if strings.EqualFold(name, probe.Name) {
				result.Entries = append(result.Entries, newLocation(probe, executable, "手动选择"))
			}
		}
		if len(result.Entries) == 0 {
			return result, fmt.Errorf("所选文件不是 %s 支持的程序入口", software.Name(kind))
		}
	}
	token, err := windows.OpenCurrentProcessToken()
	if err != nil {
		return result, err
	}
	defer token.Close()
	environment, err := token.Environ(false)
	if err != nil {
		return result, err
	}
	ctx, cancel := context.WithTimeout(ctx, SCAN_TIMEOUT)
	defer cancel()
	if executable != "" {
		inspector.inspectEntries(ctx, result.Entries, probes, environment)
		return result, nil
	}
	for _, probe := range probes {
		entries, err := discoverCommand(ctx, probe, environment)
		if err != nil {
			result.Warnings = append(result.Warnings, err.Error())
		}
		result.Entries = append(result.Entries, entries...)
	}
	services, warnings := discoverServices(ctx, probes)
	result.Entries = append(result.Entries, services...)
	result.Warnings = append(result.Warnings, warnings...)
	inspector.inspectEntries(ctx, result.Entries, probes, environment)
	if ctx.Err() != nil {
		result.Warnings = append(result.Warnings, "环境检测超时或已取消，结果可能不完整")
	}
	return result, nil
}

func discoverCommand(ctx context.Context, probe software.CommandProbe, environment []string) ([]software.CommandLocation, error) {
	entries := make([]software.CommandLocation, 0)
	lookupContext, cancel := context.WithTimeout(ctx, COMMAND_TIMEOUT)
	defer cancel()
	lookup := exec.CommandContext(lookupContext, paths.WindowsCommand("where.exe"), "$PATH:"+probe.Name)
	lookup.Env = environment
	lookup.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := readCommandOutput(lookup)
	if lookupContext.Err() != nil {
		return entries, fmt.Errorf("查询 %s PATH 超时", probe.Name)
	}
	if err != nil {
		var exitError *exec.ExitError
		if !errors.As(err, &exitError) || exitError.ExitCode() != 1 {
			return entries, fmt.Errorf("查询 %s PATH 失败: %w", probe.Name, err)
		}
	}
	seen := make(map[string]bool)
	if err == nil {
		for _, line := range strings.Split(output, "\n") {
			path := strings.TrimSpace(line)
			if path == "" {
				continue
			}
			key := strings.ToLower(filepath.Clean(path))
			if seen[key] {
				continue
			}
			seen[key] = true
			entry := newLocation(probe, path, software.SOURCE_PATH)
			entry.Active = len(entries) == 0
			entries = append(entries, entry)
		}
	}
	if probe.Home != "" {
		for _, variable := range environment {
			name, home, found := strings.Cut(variable, "=")
			if !found || !strings.EqualFold(name, probe.Home) || home == "" {
				continue
			}
			path := paths.EnvironmentCommand(home, probe.Relative)
			if seen[strings.ToLower(filepath.Clean(path))] {
				continue
			}
			entry := newLocation(probe, path, probe.Home)
			if !filepath.IsAbs(path) {
				entry.Error = "环境变量中的安装目录不是绝对路径"
			}
			if _, err := os.Stat(path); err != nil {
				entry.Error = fmt.Sprintf("环境变量目标不可访问: %v", err)
			}
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

func newLocation(probe software.CommandProbe, path, source string) software.CommandLocation {
	return software.CommandLocation{Path: path, Command: probe.Name, Role: probe.Role, Source: source, Versions: []string{}}
}

func (inspector *commandInspector) inspectEntries(ctx context.Context, entries []software.CommandLocation, probes []software.CommandProbe, environment []string) {
	var group sync.WaitGroup
	for index := range entries {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			if entries[index].Error != "" {
				return
			}
			select {
			case inspector.workers <- struct{}{}:
				defer func() { <-inspector.workers }()
			case <-ctx.Done():
				entries[index].Error = "版本检测超时或已取消"
				return
			}
			for _, probe := range probes {
				if probe.Name != entries[index].Command || probe.Role != entries[index].Role {
					continue
				}
				versions, err := probeCommand(ctx, entries[index].Path, probe, environment)
				if err != nil {
					entries[index].Error = err.Error()
					return
				}
				entries[index].Versions = versions
				return
			}
		}(index)
	}
	group.Wait()
}
