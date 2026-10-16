//go:build windows

package gateway

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"devenv/internal/infra/network"
	"devenv/internal/infra/windowsservice"
	"devenv/internal/ports"
	"devenv/internal/shared/paths"
)

func TestPersistentServiceAccessIsNonRecursive(t *testing.T) {
	directory := t.TempDir()
	arguments := persistentServiceAccessArguments(serviceAccess{path: directory, permission: "RX"})
	for _, argument := range arguments {
		if argument == "/T" {
			t.Fatal("persistent service access must not recursively traverse project directories")
		}
	}
	want := "*S-1-5-19:(OI)(CI)RX"
	if len(arguments) != 3 || arguments[2] != want {
		t.Fatalf("expected inheritable directory permission %q, got %v", want, arguments)
	}
}

func TestPersistentServiceAccessState(t *testing.T) {
	base := t.TempDir()
	state := map[string]string{`d:\projects\demo|false`: "RX"}
	if err := savePersistentServiceAccess(base, state); err != nil {
		t.Fatal(err)
	}
	loaded, err := loadPersistentServiceAccess(base)
	if err != nil {
		t.Fatal(err)
	}
	if loaded[`d:\projects\demo|false`] != "RX" {
		t.Fatalf("unexpected access state: %v", loaded)
	}
}

// Explicit opt-in: this test restarts the real locally installed gateway.
func TestInstalledNginxServiceRestart(t *testing.T) {
	if os.Getenv("DEVENV_TEST_NGINX_RESTART") != "1" {
		t.Skip("set DEVENV_TEST_NGINX_RESTART=1 to restart the installed Nginx service")
	}
	checker := network.NewChecker()
	const port = 80
	if !persistentServiceExists(NGINX_SERVICE_NAME) || !checker.Listening(port) {
		t.Fatal("expected installed Nginx service to be running on port 80")
	}
	t.Cleanup(func() {
		if err := startPersistentService(NGINX_SERVICE_NAME); err != nil {
			t.Errorf("restore Nginx service: %v", err)
		}
	})
	for attempt := 0; attempt < 3; attempt++ {
		if err := stopPersistentService(NGINX_SERVICE_NAME); err != nil {
			t.Fatal(err)
		}
		if err := waitPortClosed(checker, port); err != nil {
			t.Fatal(err)
		}
		if err := startPersistentService(NGINX_SERVICE_NAME); err != nil {
			t.Fatal(err)
		}
		if err := waitPort(checker, port); err != nil {
			t.Fatal(err)
		}
	}
	definitionBytes, err := os.ReadFile(paths.ServiceConfig(paths.Base(), NGINX_SERVICE_NAME))
	if err != nil {
		t.Fatal(err)
	}
	var definition windowsservice.Definition
	if err := json.Unmarshal(definitionBytes, &definition); err != nil {
		t.Fatal(err)
	}
	for attempt := 0; attempt < 3; attempt++ {
		if err := ReloadNginxConfig(definition.Binary, paths.Base(), port); err != nil {
			t.Fatalf("apply configuration through Windows service: %v", err)
		}
	}
}

func TestInstalledNginxRestartWithProjectLogs(t *testing.T) {
	if os.Getenv("DEVENV_TEST_NGINX_RESTART") != "1" {
		t.Skip("requires explicit opt-in and installed Nginx service host")
	}
	definitionBytes, err := os.ReadFile(paths.ServiceConfig(paths.Base(), NGINX_SERVICE_NAME))
	if err != nil {
		t.Fatal(err)
	}
	var definition windowsservice.Definition
	if err := json.Unmarshal(definitionBytes, &definition); err != nil {
		t.Fatal(err)
	}
	base := t.TempDir()
	if err := os.MkdirAll(filepath.Join(base, "temp"), 0o755); err != nil {
		t.Fatal(err)
	}
	projectLog := paths.ProjectLog(base, "project", "access.log")
	if err := os.MkdirAll(filepath.Dir(projectLog), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(paths.ServiceHost(base)), 0o755); err != nil {
		t.Fatal(err)
	}
	host, err := os.ReadFile(paths.ServiceHost(paths.Base()))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(paths.ServiceHost(base), host, 0o755); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	config := filepath.Join(base, "nginx.conf")
	content := fmt.Sprintf("events {}\nhttp { server { listen 127.0.0.1:%d; access_log %s; error_log %s; location / { return 200 'project-log-test'; } } }\n", port, nginxQuote(nginxPath(projectLog)), nginxQuote(nginxPath(filepath.Join(filepath.Dir(projectLog), "error.log"))))
	if err := os.WriteFile(config, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	name := fmt.Sprintf("DevEnvTestNginx-%d", os.Getpid())
	service := persistentService{
		name:       name,
		options:    ports.Options{Binary: definition.Binary, Args: []string{"-p", nginxPath(base) + "/", "-c", config}, Directory: filepath.Dir(definition.Binary), Log: filepath.Join(base, "host.log")},
		accessList: []serviceAccess{{path: base, permission: "M"}, {path: filepath.Dir(projectLog), permission: "M"}},
	}
	t.Cleanup(func() {
		if err := stopPersistentService(name); err != nil {
			t.Error(err)
		}
		if output, err := runPersistentServiceCommand("sc.exe", "delete", name); err != nil {
			t.Errorf("delete test service: %v: %s", err, output)
		}
	})
	if err := configurePersistentService(base, service); err != nil {
		t.Fatal(err)
	}
	checker := network.NewChecker()
	client := &http.Client{Timeout: 5 * time.Second}
	for attempt := 0; attempt < 3; attempt++ {
		if err := startPersistentService(name); err != nil {
			t.Fatal(err)
		}
		if err := waitPort(checker, port); err != nil {
			log, _ := os.ReadFile(service.options.Log)
			t.Logf("service log: %s", log)
			t.Fatal(err)
		}
		response, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/", port))
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("HTTP %d", response.StatusCode)
		}
		if err := stopPersistentService(name); err != nil {
			t.Fatal(err)
		}
	}
	if info, err := os.Stat(projectLog); err != nil || info.Size() == 0 {
		t.Fatalf("project access log not written: %v", err)
	}
}
