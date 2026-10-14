package process_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"devenv/internal/infra/process"
	"devenv/internal/ports"
)

func TestRunnerLifecycleAndSelfExit(t *testing.T) {
	mgr := process.NewManager()
	tmp := t.TempDir()
	logPath := filepath.Join(tmp, "runner.log")

	// Start a long-running ping process
	pid, err := mgr.Start("test:ping", ports.Options{
		Binary:    "cmd.exe",
		Args:      []string{"/C", "echo", "hello-runner", "&&", "ping", "-n", "30", "127.0.0.1"},
		Directory: tmp,
		Log:       logPath,
	})
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if pid <= 0 {
		t.Fatalf("expected positive pid, got %d", pid)
	}

	st := mgr.Status("test:ping")
	if !st.Running || st.PID != pid {
		t.Fatalf("expected running status with pid %d, got %+v", pid, st)
	}

	// Wait briefly for echo output
	time.Sleep(250 * time.Millisecond)
	lines := mgr.Tail("test:ping", 50)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "hello-runner") {
		t.Fatalf("expected tail to contain 'hello-runner', got: %q", joined)
	}

	// Stop the process tree
	if err := mgr.Stop("test:ping"); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	stAfter := mgr.Status("test:ping")
	if stAfter.Running || stAfter.PID != 0 {
		t.Fatalf("expected stopped status, got %+v", stAfter)
	}

	// Start a process that exits immediately with non-zero code
	selfExitLog := filepath.Join(tmp, "exit.log")
	_, err = mgr.Start("test:exit", ports.Options{
		Binary:    "cmd.exe",
		Args:      []string{"/C", "exit", "2"},
		Directory: tmp,
		Log:       selfExitLog,
	})
	if err != nil {
		t.Fatalf("Start exit process failed: %v", err)
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		s := mgr.Status("test:exit")
		if !s.Running {
			if s.Message == "" {
				t.Fatalf("expected exit message when process exits on its own, got empty message")
			}
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if mgr.Status("test:exit").Running {
		_ = os.Remove(selfExitLog)
		t.Fatalf("expected test:exit to have exited")
	}
}

func TestRedisStartupIfInstalled(t *testing.T) {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		t.Skip("LOCALAPPDATA not set")
	}
	redisDir := filepath.Join(localAppData, "DevEnv", "software", "redis", "8.10.2")
	redisBin := filepath.Join(redisDir, "redis-server.exe")
	if _, err := os.Stat(redisBin); err != nil {
		t.Skip("redis 8.10.2 not installed locally")
	}

	mgr := process.NewManager()
	logPath := filepath.Join(t.TempDir(), "redis.log")
	pid, err := mgr.Start("software:redis", ports.Options{
		Binary:    redisBin,
		Args:      []string{"redis.conf", "--port", "16389"},
		Directory: redisDir,
		Log:       logPath,
	})
	if err != nil {
		t.Fatalf("Start redis failed: %v", err)
	}
	defer func() {
		_ = mgr.Stop("software:redis")
	}()

	time.Sleep(800 * time.Millisecond)
	st := mgr.Status("software:redis")
	if !st.Running || st.PID != pid {
		t.Fatalf("expected redis running with pid %d, got %+v", pid, st)
	}
	lines := mgr.Tail("software:redis", 20)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "Ready to accept connections") {
		t.Fatalf("expected redis log to contain 'Ready to accept connections', got:\n%s", joined)
	}
}

