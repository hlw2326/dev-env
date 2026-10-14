package nginx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devenv/internal/domain/site"
	"devenv/internal/ports"
)

type mockProbe struct {
	listening map[int]bool
}

func (m *mockProbe) Available(port int) bool {
	return !m.listening[port]
}

func (m *mockProbe) Listening(port int) bool {
	return m.listening[port]
}

func (m *mockProbe) FindProcess(port int) (int, string, error) {
	return 0, "", nil
}

func TestNginxGatewayConfigAndVhosts(t *testing.T) {
	tmp := t.TempDir()
	nginxDir := filepath.Join(tmp, "software", "nginx", "1.24.0")
	if err := os.MkdirAll(nginxDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nginxDir, "nginx.exe"), []byte("mock"), 0755); err != nil {
		t.Fatal(err)
	}

	probe := &mockProbe{listening: make(map[int]bool)}
	gw := NewGateway(tmp, nil, probe)

	if dir := gw.findNginx(); dir != nginxDir {
		t.Fatalf("expected findNginx to return %s, got %s", nginxDir, dir)
	}

	if err := gw.prepareConfig(nginxDir, 80); err != nil {
		t.Fatal(err)
	}

	confRaw, err := os.ReadFile(filepath.Join(nginxDir, "conf", "nginx.conf"))
	if err != nil {
		t.Fatal(err)
	}
	confText := string(confRaw)
	if !strings.Contains(confText, "include vhosts/*.conf;") {
		t.Fatalf("expected nginx.conf to include vhosts/*.conf;, got:\n%s", confText)
	}
	if !strings.Contains(confText, "listen 80 default_server;") {
		t.Fatalf("expected nginx.conf to listen 80 default_server, got:\n%s", confText)
	}

	routes := []ports.Route{
		{
			ID:    1,
			Hosts: []string{"static.test"},
			Kind:  site.KindStatic,
			Path:  filepath.Join(tmp, "www", "static"),
			Log:   filepath.Join(tmp, "logs", "static.log"),
		},
		{
			ID:    2,
			Hosts: []string{"php.test"},
			Kind:  site.KindPHP,
			Path:  filepath.Join(tmp, "www", "php"),
			Port:  9074,
			Log:   filepath.Join(tmp, "logs", "php.log"),
		},
		{
			ID:    3,
			Hosts: []string{"node.test"},
			Kind:  site.KindNode,
			Port:  3001,
			Log:   filepath.Join(tmp, "logs", "node.log"),
		},
	}

	gw.port = 80
	if err := gw.syncVhosts(nginxDir, routes); err != nil {
		t.Fatal(err)
	}

	v1, err := os.ReadFile(filepath.Join(nginxDir, "conf", "vhosts", "site_1.conf"))
	if err != nil {
		t.Fatal(err)
	}
	v1Text := string(v1)
	if !strings.Contains(v1Text, "server_name static.test;") || !strings.Contains(v1Text, "try_files $uri $uri/ /index.html =404;") {
		t.Fatalf("unexpected static vhost:\n%s", v1Text)
	}
	if strings.Contains(v1Text, `\`) {
		t.Fatalf("vhost should use forward slashes for paths on Windows:\n%s", v1Text)
	}

	v2, err := os.ReadFile(filepath.Join(nginxDir, "conf", "vhosts", "site_2.conf"))
	if err != nil {
		t.Fatal(err)
	}
	v2Text := string(v2)
	if !strings.Contains(v2Text, "fastcgi_pass   127.0.0.1:9074;") {
		t.Fatalf("unexpected php vhost:\n%s", v2Text)
	}

	v3, err := os.ReadFile(filepath.Join(nginxDir, "conf", "vhosts", "site_3.conf"))
	if err != nil {
		t.Fatal(err)
	}
	v3Text := string(v3)
	if !strings.Contains(v3Text, "proxy_pass http://127.0.0.1:3001;") {
		t.Fatalf("unexpected node vhost:\n%s", v3Text)
	}

	routes2 := []ports.Route{routes[0]}
	if err := gw.syncVhosts(nginxDir, routes2); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(nginxDir, "conf", "vhosts", "site_2.conf")); !os.IsNotExist(err) {
		t.Fatalf("expected site_2.conf to be deleted after sync")
	}
}
