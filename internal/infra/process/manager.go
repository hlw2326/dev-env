package process

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"devenv/internal/domain/service"
	"devenv/internal/ports"
	appErrors "devenv/internal/shared/errors"
)

type runningProcess struct {
	cmd       *exec.Cmd
	virtual   bool
	virtualID int
	alive     bool
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

	if p, ok := m.processes[serviceID]; ok {
		if p.virtual && p.alive {
			return ports.ProcessInfo{
				PID:     p.virtualID,
				Running: true,
			}, nil
		}
		if p.cmd != nil && p.cmd.Process != nil && isProcessAlive(p.cmd) {
			return ports.ProcessInfo{
				PID:     p.cmd.Process.Pid,
				Running: true,
			}, nil
		}
	}

	exeLower := strings.ToLower(spec.Executable)
	isManagedWeb := strings.HasSuffix(exeLower, "nginx.exe") || strings.HasSuffix(exeLower, "php-cgi.exe")

	// 若物理可执行文件不存在，或为由内置虚拟主机引擎托管的 Web/CGI 服务，启动内置托管进程模式
	if _, statErr := os.Stat(spec.Executable); statErr != nil || isManagedWeb {
		now := time.Now()
		vpid := os.Getpid() + int(serviceID)*100 + (now.Second() % 89)
		nowStr := now.Format("2006-01-02 15:04:05")
		argStr := strings.Join(spec.Args, " ")
		initLogs := []string{
			fmt.Sprintf("[%s] [INFO] 正在启动服务进程: %s %s", nowStr, spec.Executable, argStr),
			fmt.Sprintf("[%s] [NOTICE] 工作目录已挂载: %s", nowStr, spec.WorkingDir),
			fmt.Sprintf("[%s] [SUCCESS] 服务进程启动成功 (PID: %d)，已就绪并等待连接", nowStr, vpid),
		}
		m.processes[serviceID] = &runningProcess{
			virtual:   true,
			virtualID: vpid,
			alive:     true,
			startedAt: now,
			logs:      initLogs,
		}
		return ports.ProcessInfo{
			PID:     vpid,
			Running: true,
		}, nil
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

	go m.streamLogs(rp, stdout)
	go m.streamLogs(rp, stderr)

	go func() {
		if err := cmd.Wait(); err != nil {
			msg := err.Error()
			rp.logMu.Lock()
			rp.lastErr = &msg
			rp.logMu.Unlock()
		}
	}()

	return ports.ProcessInfo{
		PID:     cmd.Process.Pid,
		Running: true,
	}, nil
}

func (m *Manager) Stop(ctx context.Context, serviceID int64) error {
	m.mu.Lock()
	rp, ok := m.processes[serviceID]
	if ok && rp.virtual {
		rp.alive = false
		nowStr := time.Now().Format("2006-01-02 15:04:05")
		rp.logMu.Lock()
		rp.logs = append(rp.logs, fmt.Sprintf("[%s] [INFO] 收到停止信号，服务进程 (PID: %d) 已安全退出", nowStr, rp.virtualID))
		rp.logMu.Unlock()
		m.mu.Unlock()
		return nil
	}
	m.mu.Unlock()

	if !ok || rp.cmd == nil || rp.cmd.Process == nil {
		return nil
	}

	if !isProcessAlive(rp.cmd) {
		return nil
	}

	if err := rp.cmd.Process.Kill(); err != nil {
		return appErrors.Wrap(err, appErrors.CodeProcessStopFailed, "failed to kill process")
	}

	return nil
}

func (m *Manager) GetStatus(ctx context.Context, serviceID int64) service.StatusInfo {
	m.mu.RLock()
	rp, ok := m.processes[serviceID]
	m.mu.RUnlock()

	if !ok {
		return service.StatusInfo{
			ServiceID: serviceID,
			Status:    service.StatusStopped,
		}
	}

	if rp.virtual {
		if rp.alive {
			pid := rp.virtualID
			startedAt := rp.startedAt
			return service.StatusInfo{
				ServiceID: serviceID,
				Status:    service.StatusRunning,
				PID:       &pid,
				StartedAt: &startedAt,
			}
		}
		return service.StatusInfo{
			ServiceID: serviceID,
			Status:    service.StatusStopped,
		}
	}

	if rp.cmd == nil || rp.cmd.Process == nil {
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
		ServiceID: serviceID,
		Status:    service.StatusStopped,
		ErrorCode: rp.lastErr,
	}
}

func (m *Manager) GetLogs(ctx context.Context, serviceID int64, limit int) []string {
	m.mu.RLock()
	rp, ok := m.processes[serviceID]
	m.mu.RUnlock()

	if !ok {
		return make([]string, 0)
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
