package process

import (
	"encoding/csv"
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

const EXIT_TAIL_LINES = 5

type entry struct {
	command  *exec.Cmd
	pid      int
	running  bool
	stopping bool
	message  string
	log      string
}

type Manager struct {
	mu      sync.RWMutex
	entries map[string]*entry
	job     *jobObject
}

func NewManager() *Manager {
	return &Manager{
		entries: make(map[string]*entry),
		job:     newJob(),
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
		return 0, appErrors.New(appErrors.PROCESS_START_FAILED, "可执行文件路径为空")
	}
	if filepath.IsAbs(options.Binary) {
		if info, err := os.Stat(options.Binary); err != nil || info.IsDir() {
			return 0, appErrors.New(appErrors.PROCESS_START_FAILED, "可执行文件不存在: "+options.Binary)
		}
	} else {
		if _, err := exec.LookPath(options.Binary); err != nil {
			return 0, appErrors.Wrap(err, appErrors.PROCESS_START_FAILED, "找不到可执行文件: "+options.Binary)
		}
	}

	var logFile *os.File
	if options.Log != "" {
		if err := os.MkdirAll(filepath.Dir(options.Log), 0o755); err != nil {
			return 0, appErrors.Wrap(err, appErrors.PROCESS_START_FAILED, "创建日志目录失败")
		}
		file, err := os.OpenFile(options.Log, os.O_CREATE|os.O_RDWR, 0o644)
		if err != nil {
			return 0, appErrors.Wrap(err, appErrors.PROCESS_START_FAILED, "打开日志文件失败")
		}
		if _, err := file.Seek(0, io.SeekEnd); err != nil {
			if closeErr := file.Close(); closeErr != nil {
				return 0, appErrors.Wrap(closeErr, appErrors.PROCESS_START_FAILED, "关闭日志文件失败")
			}
			return 0, appErrors.Wrap(err, appErrors.PROCESS_START_FAILED, "定位日志文件失败")
		}
		logFile = file
	}

	command := exec.Command(options.Binary, options.Args...)
	if options.Directory != "" {
		command.Dir = options.Directory
	}
	if len(options.Environment) > 0 {
		command.Env = append(os.Environ(), options.Environment...)
	}
	if logFile != nil {
		command.Stdout = logFile
		command.Stderr = logFile
	}
	command.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	if err := command.Start(); err != nil {
		if logFile != nil {
			if closeErr := logFile.Close(); closeErr != nil {
				return 0, appErrors.Wrap(err, appErrors.PROCESS_START_FAILED, "进程启动失败")
			}
		}
		return 0, appErrors.Wrap(err, appErrors.PROCESS_START_FAILED, "进程启动失败")
	}

	m.job.assign(command.Process)

	pid := command.Process.Pid
	item := &entry{
		command: command,
		pid:     pid,
		running: true,
		log:     options.Log,
	}

	m.mu.Lock()
	m.entries[key] = item
	m.mu.Unlock()

	go func() {
		waitErr := command.Wait()
		if logFile != nil {
			if closeErr := logFile.Close(); closeErr != nil {
				logFile = nil
			}
		}

		m.mu.Lock()
		item.running = false
		stopping := item.stopping
		if waitErr != nil && !stopping {
			lines := readTail(item.log, EXIT_TAIL_LINES)
			if len(lines) > 0 {
				item.message = strings.Join(lines, " | ")
			} else {
				item.message = waitErr.Error()
			}
		} else if !stopping {
			item.message = "进程已退出"
		} else {
			item.message = ""
		}
		message := item.message
		m.mu.Unlock()

		if !stopping && options.OnExit != nil {
			options.OnExit(key, message)
		}
	}()

	return pid, nil
}

func (m *Manager) Stop(key string) error {
	m.mu.Lock()
	item, ok := m.entries[key]
	if !ok {
		m.mu.Unlock()
		return nil
	}
	if !item.running || item.command == nil || item.command.Process == nil {
		item.running = false
		m.mu.Unlock()
		return nil
	}
	item.stopping = true
	pid := item.pid
	proc := item.command.Process
	m.mu.Unlock()

	killer := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(pid))
	killer.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := killer.Run(); err != nil {
		if killErr := proc.Kill(); killErr != nil {
			m.mu.Lock()
			item.running = false
			item.message = ""
			m.mu.Unlock()
			return nil
		}
	}

	m.mu.Lock()
	item.running = false
	item.message = ""
	m.mu.Unlock()
	return nil
}

func (m *Manager) Status(key string) ports.Status {
	m.mu.RLock()
	defer m.mu.RUnlock()
	item, ok := m.entries[key]
	if !ok {
		return ports.Status{}
	}
	pid := 0
	if item.running {
		pid = item.pid
	}
	return ports.Status{
		Running: item.running,
		PID:     pid,
		Message: item.message,
	}
}

func (m *Manager) Tail(key string, limit int) []string {
	m.mu.RLock()
	item, ok := m.entries[key]
	logPath := ""
	if ok {
		logPath = item.log
	}
	m.mu.RUnlock()

	if logPath == "" {
		return make([]string, 0)
	}
	return readTail(logPath, limit)
}

func ReadTail(path string, limit int) []string {
	return readTail(path, limit)
}

// ImageName returns the executable name for a Windows process ID.
func ImageName(pid int) string {
	if pid <= 0 {
		return ""
	}
	command := exec.Command("tasklist", "/FI", "PID eq "+strconv.Itoa(pid), "/FO", "CSV", "/NH")
	output, err := command.Output()
	if err != nil {
		return ""
	}
	reader := csv.NewReader(strings.NewReader(string(output)))
	record, err := reader.Read()
	if err != nil || len(record) == 0 {
		return ""
	}
	return strings.TrimSpace(record[0])
}

// MatchesImageName reports whether a process belongs to one of the expected executables.
func MatchesImageName(pid int, names ...string) bool {
	imageName := ImageName(pid)
	for _, name := range names {
		if strings.EqualFold(imageName, name) {
			return true
		}
	}
	return false
}

// StopPID terminates an externally started Windows process tree.
func StopPID(pid int) error {
	if pid <= 0 {
		return nil
	}
	killer := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(pid))
	killer.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := killer.Run(); err != nil {
		return appErrors.Wrap(err, appErrors.PROCESS_START_FAILED, "停止外部服务进程失败")
	}
	return nil
}

func readTail(path string, limit int) []string {
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
			lines = append(lines, line)
		}
	}
	if limit > 0 && len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}
	return lines
}
