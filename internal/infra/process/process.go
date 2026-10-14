package process

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"devenv/internal/ports"
	appErrors "devenv/internal/shared/errors"
)

type entry struct {
	cmd      *exec.Cmd
	pid      int
	running  bool
	stopping bool
	message  string
	log      string
}

type Manager struct {
	mu      sync.RWMutex
	entries map[string]*entry
}

func NewManager() *Manager {
	return &Manager{
		entries: make(map[string]*entry),
	}
}

func (m *Manager) Start(key string, options ports.Options) (int, error) {
	m.mu.Lock()
	if existing, ok := m.entries[key]; ok && existing.running {
		pid := existing.pid
		m.mu.Unlock()
		return pid, nil
	}
	m.mu.Unlock()

	if options.Binary == "" {
		return 0, appErrors.New(appErrors.CodeProcessStartFailed, "可执行文件路径为空")
	}
	if filepath.IsAbs(options.Binary) {
		if info, err := os.Stat(options.Binary); err != nil || info.IsDir() {
			return 0, appErrors.New(appErrors.CodeProcessStartFailed, "可执行文件不存在: "+options.Binary)
		}
	} else {
		if _, err := exec.LookPath(options.Binary); err != nil {
			return 0, appErrors.Wrap(err, appErrors.CodeProcessStartFailed, "找不到可执行文件: "+options.Binary)
		}
	}

	var logFile *os.File
	if options.Log != "" {
		if err := os.MkdirAll(filepath.Dir(options.Log), 0755); err != nil {
			return 0, appErrors.Wrap(err, appErrors.CodeProcessStartFailed, "创建日志目录失败")
		}
		f, err := os.OpenFile(options.Log, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return 0, appErrors.Wrap(err, appErrors.CodeProcessStartFailed, "打开日志文件失败")
		}
		_, _ = f.Seek(0, io.SeekEnd)
		logFile = f
	}

	cmd := exec.Command(options.Binary, options.Args...)
	if options.Directory != "" {
		cmd.Dir = options.Directory
	}
	if len(options.Environment) > 0 {
		cmd.Env = append(os.Environ(), options.Environment...)
	}
	if logFile != nil {
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	if err := cmd.Start(); err != nil {
		if logFile != nil {
			_ = logFile.Close()
		}
		return 0, appErrors.Wrap(err, appErrors.CodeProcessStartFailed, "进程启动失败")
	}

	pid := cmd.Process.Pid
	ent := &entry{
		cmd:     cmd,
		pid:     pid,
		running: true,
		log:     options.Log,
	}

	m.mu.Lock()
	m.entries[key] = ent
	m.mu.Unlock()

	go func() {
		waitErr := cmd.Wait()
		if logFile != nil {
			_ = logFile.Close()
		}

		m.mu.Lock()
		ent.running = false
		wasStopping := ent.stopping
		if waitErr != nil && !wasStopping {
			tailLines := readTailLines(ent.log, 5)
			if len(tailLines) > 0 {
				ent.message = strings.Join(tailLines, " | ")
			} else {
				ent.message = waitErr.Error()
			}
		} else if !wasStopping {
			ent.message = "进程已退出"
		} else {
			ent.message = ""
		}
		msg := ent.message
		m.mu.Unlock()

		if !wasStopping && options.OnExit != nil {
			options.OnExit(key, msg)
		}
	}()

	return pid, nil
}

func (m *Manager) Stop(key string) error {
	m.mu.Lock()
	ent, ok := m.entries[key]
	if !ok {
		m.mu.Unlock()
		return nil
	}
	if !ent.running || ent.cmd == nil || ent.cmd.Process == nil {
		ent.running = false
		m.mu.Unlock()
		return nil
	}
	ent.stopping = true
	pid := ent.pid
	proc := ent.cmd.Process
	m.mu.Unlock()

	killCmd := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(pid))
	killCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = killCmd.Run()
	_ = proc.Kill()

	m.mu.Lock()
	ent.running = false
	ent.message = ""
	m.mu.Unlock()
	return nil
}

func (m *Manager) Status(key string) ports.Status {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ent, ok := m.entries[key]
	if !ok {
		return ports.Status{}
	}
	pid := 0
	if ent.running {
		pid = ent.pid
	}
	return ports.Status{
		Running: ent.running,
		PID:     pid,
		Message: ent.message,
	}
}

func (m *Manager) Tail(key string, limit int) []string {
	m.mu.RLock()
	ent, ok := m.entries[key]
	logPath := ""
	if ok {
		logPath = ent.log
	}
	m.mu.RUnlock()

	if logPath == "" {
		return make([]string, 0)
	}
	return readTailLines(logPath, limit)
}

func ReadLogTail(path string, limit int) []string {
	return readTailLines(path, limit)
}

func readTailLines(path string, limit int) []string {
	if path == "" {
		return make([]string, 0)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return make([]string, 0)
	}
	text := strings.ReplaceAll(string(raw), "\r\n", "\n")
	parts := strings.Split(text, "\n")
	lines := make([]string, 0, len(parts))
	for _, line := range parts {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, fmt.Sprintf("%s", line))
		}
	}
	if limit > 0 && len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}
	return lines
}
