package gateway

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"devenv/internal/domain/project"
	"devenv/internal/domain/software"
	"devenv/internal/infra/network"
	processinfra "devenv/internal/infra/process"
	"devenv/internal/ports"
	"devenv/internal/shared/paths"
)

type fixedNginxRepository struct {
	item software.Software
}

func (r fixedNginxRepository) GetMain(_ context.Context, kind string) (*software.Software, error) {
	if kind != software.KIND_NGINX {
		return nil, nil
	}
	item := r.item
	return &item, nil
}

func TestRenderPHPRouteUsesPathInfoFallback(t *testing.T) {
	config, err := renderRoute(80, ports.Route{
		ID:    1,
		Hosts: []string{"demo.localhost"},
		Kind:  project.KIND_PHP,
		Path:  `C:\projects\demo\public`,
		Log:   `C:\logs\demo\access.log`,
	}, 9000)
	if err != nil {
		t.Fatal(err)
	}

	checks := []string{
		"server_name \"demo.localhost\";",
		"try_files $uri $uri/ /index.php$uri$is_args$args;",
		"fastcgi_split_path_info ^(.+?\\.php)(/.*)$;",
		"set $devenv_path_info $fastcgi_path_info;",
		"fastcgi_param PATH_INFO $devenv_path_info;",
		"fastcgi_pass 127.0.0.1:9000;",
	}
	for _, expected := range checks {
		if !strings.Contains(config, expected) {
			t.Fatalf("expected generated config to contain %q\n%s", expected, config)
		}
	}
}

func TestRenderNodeRouteSupportsWebSocket(t *testing.T) {
	config, err := renderRoute(8080, ports.Route{
		ID:    2,
		Hosts: []string{"node.localhost"},
		Kind:  project.KIND_NODE,
		Port:  3100,
		Log:   `C:\logs\node\access.log`,
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(config, "proxy_pass http://127.0.0.1:3100;") {
		t.Fatalf("expected Node upstream in generated config\n%s", config)
	}
	if !strings.Contains(config, "proxy_set_header Upgrade $http_upgrade;") {
		t.Fatalf("expected WebSocket upgrade headers in generated config\n%s", config)
	}
}

func TestRenderStaticRouteReturnsNotFoundWithoutIndexFallback(t *testing.T) {
	config, err := renderRoute(8080, ports.Route{
		ID:    3,
		Hosts: []string{"static.localhost"},
		Kind:  project.KIND_STATIC,
		Path:  `C:\projects\static`,
		Log:   `C:\logs\static\access.log`,
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(config, "try_files $uri $uri/ =404;") {
		t.Fatalf("expected missing static files to return 404\n%s", config)
	}
	if strings.Contains(config, "/index.html;") {
		t.Fatalf("did not expect an SPA fallback\n%s", config)
	}
}

func TestRenderRouteIncludesAdvancedConfiguration(t *testing.T) {
	config, err := renderRoute(8080, ports.Route{
		ID:    4,
		Hosts: []string{"advanced.localhost"},
		Kind:  project.KIND_STATIC,
		Path:  `C:\projects\advanced`,
		Log:   `C:\logs\advanced\access.log`,
		Nginx: "client_max_body_size 20m;\nadd_header X-DevEnv advanced;",
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(config, "client_max_body_size 20m;") || !strings.Contains(config, "add_header X-DevEnv advanced;") {
		t.Fatalf("expected advanced configuration in generated server block\n%s", config)
	}
}

func TestPoolPortIsStable(t *testing.T) {
	executable := `C:\DevEnv\software\php\8.3.35\php-cgi.exe`
	first := poolPort(executable)
	second := poolPort(executable)
	if first != second {
		t.Fatalf("expected stable FastCGI port, got %d and %d", first, second)
	}
	if first < FASTCGI_PORT_START || first >= FASTCGI_PORT_START+FASTCGI_PORT_COUNT {
		t.Fatalf("expected FastCGI port in configured range, got %d", first)
	}
}

func TestWriteConfigCreatesNginxLogFolder(t *testing.T) {
	base := t.TempDir()
	gateway := &Nginx{home: base}

	if err := gateway.writeConfig(t.TempDir(), 18080, []ports.Route{}, map[string]pool{}); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(paths.NginxLog(base))
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Fatalf("expected Nginx log path to be a directory: %s", paths.NginxLog(base))
	}
}

func TestInstalledNginxAcceptsGeneratedConfig(t *testing.T) {
	binary, err := exec.LookPath("nginx.exe")
	if err != nil {
		t.Skip("nginx.exe is not available in PATH")
	}
	home := filepath.Dir(binary)
	base := t.TempDir()
	projectRoot := filepath.Join(base, "project")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectRoot, "index.php"), []byte("<?php echo 'ok';"), 0o644); err != nil {
		t.Fatal(err)
	}

	gateway := &Nginx{home: base}
	routes := []ports.Route{{
		ID:    3,
		Hosts: []string{"php.localhost"},
		Kind:  project.KIND_PHP,
		Path:  projectRoot,
		PHP:   `C:\php\php-cgi.exe`,
		Log:   filepath.Join(base, "logs", "project", "access.log"),
	}}
	pools := map[string]pool{filepath.Clean(routes[0].PHP): {port: 9000}}
	if err := gateway.writeConfig(home, 18080, routes, pools); err != nil {
		t.Fatal(err)
	}
	if err := checkNginx(binary, base); err != nil {
		t.Fatal(err)
	}
}

func TestInstalledNginxForwardsPrettyURLPathInfo(t *testing.T) {
	nginxBinary, err := exec.LookPath("nginx.exe")
	if err != nil {
		t.Skip("nginx.exe is not available in PATH")
	}
	phpBinary, err := exec.LookPath("php-cgi.exe")
	if err != nil {
		t.Skip("php-cgi.exe is not available in PATH")
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}

	base := t.TempDir()
	documentRoot := filepath.Join(base, "project")
	if err := os.MkdirAll(documentRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	script := `<?php echo 'PATH_INFO=' . ($_SERVER['PATH_INFO'] ?? '') . '|SCRIPT_NAME=' . ($_SERVER['SCRIPT_NAME'] ?? '');`
	if err := os.WriteFile(filepath.Join(documentRoot, "index.php"), []byte(script), 0o644); err != nil {
		t.Fatal(err)
	}

	runner := processinfra.NewManager()
	probe := network.NewChecker()
	repo := fixedNginxRepository{item: software.Software{
		Kind:  software.KIND_NGINX,
		Path:  filepath.Dir(nginxBinary),
		State: software.STATE_INSTALLED,
		Main:  true,
	}}
	gateway := NewNginx(base, repo, runner, probe)
	if err := gateway.SetRoutes([]ports.Route{{
		ID:    4,
		Hosts: []string{"pretty.localhost"},
		Kind:  project.KIND_PHP,
		Path:  documentRoot,
		PHP:   phpBinary,
		Log:   filepath.Join(base, "logs", "pretty", "access.log"),
	}}); err != nil {
		t.Fatal(err)
	}
	if err := gateway.Start(port); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := gateway.Stop(); err != nil {
			t.Logf("stop temporary gateway: %v", err)
		}
	})

	client := &http.Client{Timeout: 5 * time.Second}
	request, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:"+strconv.Itoa(port)+"/admin/login.html", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Host = "pretty.localhost"
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	body, readErr := io.ReadAll(response.Body)
	if closeErr := response.Body.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	if readErr != nil {
		t.Fatal(readErr)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d: %s", response.StatusCode, body)
	}
	expected := "PATH_INFO=/admin/login.html|SCRIPT_NAME=/index.php"
	if string(body) != expected {
		t.Fatalf("expected %s, got %s", expected, body)
	}
	t.Log(fmt.Sprintf("Nginx FastCGI pretty URL verified on port %d", port))
}
