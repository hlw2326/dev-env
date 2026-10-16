//go:build windows

package gateway

import (
	"os"
	"path/filepath"
	"testing"

	"devenv/internal/domain/project"
	"devenv/internal/infra/network"
	processinfra "devenv/internal/infra/process"
	"devenv/internal/ports"
	"devenv/internal/shared/paths"
)

func TestInstalledPHPAdoption(t *testing.T) {
	if os.Getenv("DEVENV_TEST_PHP_ADOPTION") != "1" {
		t.Skip("requires installed, running PHP FastCGI service")
	}
	executable := filepath.Join(paths.Software(paths.Base(), "php", "8.3.35"), "php-cgi.exe")
	port := poolPort(executable)
	checker := network.NewChecker()
	pid := checker.ListeningPID(port)
	t.Logf("port=%d pid=%d image=%q path=%q expected=%q", port, pid, processinfra.ImageName(pid), processinfra.ExecutablePath(pid), executable)
	gateway := &Nginx{checker: checker, runner: processinfra.NewManager()}
	adopted, ok := gateway.adoptPool(executable, port)
	if !ok {
		t.Fatal("failed to adopt existing PHP FastCGI process")
	}
	if adopted.service != phpServiceName(executable) {
		t.Fatal("existing FastCGI was not identified by Windows service ownership")
	}
	if persistentServiceOwnsProcess("DevEnvMissingPHPService", pid, executable) {
		t.Fatal("unrelated service was accepted as process owner")
	}
	for attempt := 0; attempt < 3; attempt++ {
		pools, started, err := gateway.startPools([]ports.Route{{Kind: project.KIND_PHP, PHP: executable}}, 80)
		if err != nil {
			t.Fatal(err)
		}
		if len(started) != 0 || pools[executable].pid != pid {
			t.Fatal("starting PHP again should reuse the existing service without a new process")
		}
		gateway.pools = pools
	}
}
