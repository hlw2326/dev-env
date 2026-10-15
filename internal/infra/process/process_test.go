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

func TestLifecycle(t *testing.T) {
	manager := process.NewManager()
	folder := t.TempDir()
	logPath := filepath.Join(folder, "runner.log")

	pid, err := manager.Start("test:ping", ports.Options{
		Binary:    "cmd.exe",
		Args:      []string{"/C", "echo", "hello-runner", "&&", "ping", "-n", "30", "127.0.0.1"},
		Directory: folder,
		Log:       logPath,
	})
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if pid <= 0 {
		t.Fatalf("expected positive pid, got %d", pid)
	}

	status := manager.Status("test:ping")
	if !status.Running || status.PID != pid {
		t.Fatalf("expected running status with pid %d, got %+v", pid, status)
	}

	time.Sleep(250 * time.Millisecond)
	lines := manager.Tail("test:ping", 50)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "hello-runner") {
		t.Fatalf("expected tail to contain 'hello-runner', got: %q", joined)
	}

	if err := manager.Stop("test:ping"); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	stopped := manager.Status("test:ping")
	if stopped.Running || stopped.PID != 0 {
		t.Fatalf("expected stopped status, got %+v", stopped)
	}

	exitLog := filepath.Join(folder, "exit.log")
	if _, err := manager.Start("test:exit", ports.Options{
		Binary:    "cmd.exe",
		Args:      []string{"/C", "exit", "2"},
		Directory: folder,
		Log:       exitLog,
	}); err != nil {
		t.Fatalf("Start exit process failed: %v", err)
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		current := manager.Status("test:exit")
		if !current.Running {
			if current.Message == "" {
				t.Fatalf("expected exit message when process exits on its own, got empty message")
			}
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if manager.Status("test:exit").Running {
		t.Fatalf("expected test:exit to have exited")
	}
}

func TestRedis(t *testing.T) {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		t.Skip("LOCALAPPDATA not set")
	}
	folder := filepath.Join(base, "DevEnv", "software", "redis", "8.10.2")
	binary := filepath.Join(folder, "redis-server.exe")
	if _, err := os.Stat(binary); err != nil {
		t.Skip("redis 8.10.2 not installed locally")
	}

	manager := process.NewManager()
	logPath := filepath.Join(t.TempDir(), "redis.log")
	pid, err := manager.Start("software:redis", ports.Options{
		Binary:    binary,
		Args:      []string{"redis.conf", "--port", "16389"},
		Directory: folder,
		Log:       logPath,
	})
	if err != nil {
		t.Fatalf("Start redis failed: %v", err)
	}
	defer func() {
		if stopErr := manager.Stop("software:redis"); stopErr != nil {
			t.Log(stopErr)
		}
	}()

	time.Sleep(800 * time.Millisecond)
	status := manager.Status("software:redis")
	if !status.Running || status.PID != pid {
		t.Fatalf("expected redis running with pid %d, got %+v", pid, status)
	}
	lines := manager.Tail("software:redis", 20)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "Ready to accept connections") {
		t.Fatalf("expected redis log to contain 'Ready to accept connections', got:\n%s", joined)
	}
}
