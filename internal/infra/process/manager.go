package process

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	"devenv/internal/domain/service"
	"devenv/internal/ports"
	appErrors "devenv/internal/shared/errors"
)

type runningProcess struct {
	cmd       *exec.Cmd
	startedAt time.Time
	logs      []string
	logMu     sync.RWMutex
	lastErr   *string
}

type Manager struct {
	mu        sync.RWMutex
	processes map[int64]*runningProcess
}

func NewManager() *Manager {
	return &Manager{
		processes: make(map[int64]*runningProcess),
	}
}

func (m *Manager) Start(ctx context.Context, serviceID int64, spec ports.ProcessSpec) (ports.ProcessInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if p, ok := m.processes[serviceID]; ok && p.cmd != nil && p.cmd.Process != nil {
		// Check if already running
		if isProcessAlive(p.cmd) {
			return ports.ProcessInfo{
				PID:     p.cmd.Process.Pid,
				Running: true,
			}, nil
		}
	}

	cmd := exec.CommandContext(ctx, spec.Executable, spec.Args...)
	if spec.WorkingDir != "" {
		cmd.Dir = spec.WorkingDir
	}
	if len(spec.Environment) > 0 {
		cmd.Env = spec.Environment
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return ports.ProcessInfo{}, appErrors.Wrap(err, appErrors.CodeProcessStartFailed, "failed to get stdout pipe")
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return ports.ProcessInfo{}, appErrors.Wrap(err, appErrors.CodeProcessStartFailed, "failed to get stderr pipe")
	}

	if err := cmd.Start(); err != nil {
		return ports.ProcessInfo{}, appErrors.Wrap(err, appErrors.CodeProcessStartFailed, fmt.Sprintf("failed to start %s", spec.Executable))
	}

	rp := &runningProcess{
		cmd:       cmd,
		startedAt: time.Now(),
		logs:      make([]string, 0, 100),
	}
	m.processes[serviceID] = rp

	// Collect logs in background
	go m.streamLogs(rp, stdout)
	go m.streamLogs(rp, stderr)

	// Monitor exit
	go func() {
		_ = cmd.Wait()
	}()

	return ports.ProcessInfo{
		PID:     cmd.Process.Pid,
		Running: true,
	}, nil
}

func (m *Manager) Stop(ctx context.Context, serviceID int64) error {
	m.mu.Lock()
	rp, ok := m.processes[serviceID]
	m.mu.Unlock()

	if !ok || rp.cmd == nil || rp.cmd.Process == nil {
		return nil
	}

	if !isProcessAlive(rp.cmd) {
		return nil
	}

	// Try graceful stop or kill
	err := rp.cmd.Process.Kill()
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeProcessStopFailed, "failed to kill process")
	}

	return nil
}

func (m *Manager) GetStatus(ctx context.Context, serviceID int64) service.StatusInfo {
	m.mu.RLock()
	rp, ok := m.processes[serviceID]
	m.mu.RUnlock()

	if !ok || rp.cmd == nil || rp.cmd.Process == nil {
		return service.StatusInfo{
			ServiceID: serviceID,
			Status:    service.StatusStopped,
		}
	}

	if isProcessAlive(rp.cmd) {
		pid := rp.cmd.Process.Pid
		startedAt := rp.startedAt
		return service.StatusInfo{
			ServiceID: serviceID,
			Status:    service.StatusRunning,
			PID:       &pid,
			StartedAt: &startedAt,
		}
	}

	return service.StatusInfo{
		ServiceID:     serviceID,
		Status:        service.StatusStopped,
		LastErrorCode: rp.lastErr,
	}
}

func (m *Manager) GetLogs(ctx context.Context, serviceID int64, limit int) []string {
	m.mu.RLock()
	rp, ok := m.processes[serviceID]
	m.mu.RUnlock()

	if !ok {
		return []string{}
	}

	rp.logMu.RLock()
	defer rp.logMu.RUnlock()

	if limit <= 0 || limit > len(rp.logs) {
		limit = len(rp.logs)
	}

	start := len(rp.logs) - limit
	result := make([]string, limit)
	copy(result, rp.logs[start:])
	return result
}

func (m *Manager) streamLogs(rp *runningProcess, r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		text := scanner.Text()
		rp.logMu.Lock()
		if len(rp.logs) >= 500 {
			rp.logs = rp.logs[1:]
		}
		rp.logs = append(rp.logs, text)
		rp.logMu.Unlock()
	}
}

func isProcessAlive(cmd *exec.Cmd) bool {
	if cmd == nil || cmd.Process == nil {
		return false
	}
	if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		return false
	}
	return true
}
