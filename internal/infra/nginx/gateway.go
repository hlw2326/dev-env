package nginx

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"devenv/internal/domain/site"
	"devenv/internal/ports"
	appErrors "devenv/internal/shared/errors"
)

type Gateway struct {
	mu     sync.RWMutex
	base   string
	port   int
	routes []ports.Route
	runner ports.Runner
	probe  ports.PortChecker
}

func NewGateway(base string, runner ports.Runner, probe ports.PortChecker) *Gateway {
	return &Gateway{
		base:   base,
		runner: runner,
		probe:  probe,
	}
}

func (g *Gateway) Port() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.port <= 0 {
		return 0
	}
	if g.probe != nil && !g.probe.Listening(g.port) {
		return 0
	}
	return g.port
}

func (g *Gateway) Start(port int) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.port = port
	dir := g.findNginx()
	if dir == "" {
		return appErrors.New(appErrors.CodeRuntimeNotInstalled, "Nginx 未安装，请在软件管理中安装 Nginx")
	}

	if err := g.prepareConfig(dir, port); err != nil {
		return err
	}

	if err := g.syncVhosts(dir, g.routes); err != nil {
		return err
	}

	nginxBin := filepath.Join(dir, "nginx.exe")
	if g.probe.Listening(port) {
		g.reload(nginxBin, dir)
		return nil
	}

	logPath := filepath.Join(g.base, "logs", "services", "nginx-gateway.log")
	_ = os.MkdirAll(filepath.Dir(logPath), 0755)

	_, err := g.runner.Start("nginx:gateway", ports.Options{
		Binary:    nginxBin,
		Args:      []string{"-p", dir, "-c", "conf/nginx.conf"},
		Directory: dir,
		Log:       logPath,
	})
	if err != nil {
		return err
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if g.probe.Listening(port) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func (g *Gateway) Stop() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	_ = g.runner.Stop("nginx:gateway")
	dir := g.findNginx()
	if dir != "" {
		nginxBin := filepath.Join(dir, "nginx.exe")
		cmd := exec.Command(nginxBin, "-p", dir, "-s", "stop")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		_ = cmd.Run()
	}
	return nil
}

func (g *Gateway) SetRoutes(routes []ports.Route) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.routes = routes
	dir := g.findNginx()
	if dir == "" {
		return
	}

	_ = g.syncVhosts(dir, routes)
	nginxBin := filepath.Join(dir, "nginx.exe")
	if g.port > 0 && g.probe.Listening(g.port) {
		g.reload(nginxBin, dir)
	}
}

func (g *Gateway) findNginx() string {
	root := filepath.Join(g.base, "software", "nginx")
	entries, err := os.ReadDir(root)
	if err != nil {
		return ""
	}
	var dirs []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		candidate := filepath.Join(root, entry.Name())
		if _, statErr := os.Stat(filepath.Join(candidate, "nginx.exe")); statErr == nil {
			dirs = append(dirs, candidate)
		}
	}
	if len(dirs) == 0 {
		return ""
	}
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i] > dirs[j]
	})
	return dirs[0]
}

func (g *Gateway) prepareConfig(dir string, port int) error {
	_ = os.MkdirAll(filepath.Join(dir, "logs"), 0755)
	_ = os.MkdirAll(filepath.Join(dir, "temp"), 0755)
	vhostsDir := filepath.Join(dir, "conf", "vhosts")
	_ = os.MkdirAll(vhostsDir, 0755)

	g.ensureFastcgiParams(dir)

	confPath := filepath.Join(dir, "conf", "nginx.conf")
	raw, err := os.ReadFile(confPath)
	if err == nil {
		text := string(raw)
		if !strings.Contains(text, "include vhosts/*.conf;") {
			text = strings.Replace(text, "http {", "http {\r\n    include vhosts/*.conf;", 1)
		}
		re := regexp.MustCompile(`(?m)^(\s*listen\s+)\d+(\s*;.*)$`)
		text = re.ReplaceAllString(text, fmt.Sprintf("${1}%d${2}", port))
		return os.WriteFile(confPath, []byte(text), 0644)
	}

	content := fmt.Sprintf(`worker_processes 1;
events {
    worker_connections 1024;
}
http {
    include mime.types;
    default_type application/octet-stream;
    sendfile on;
    keepalive_timeout 65;
    client_max_body_size 100m;

    include vhosts/*.conf;

    server {
        listen %d default_server;
        server_name _;
        return 404 "DevEnv: No site matches this domain\n";
    }
}
`, port)
	return os.WriteFile(confPath, []byte(content), 0644)
}

func (g *Gateway) ensureFastcgiParams(dir string) {
	fastcgiFile := filepath.Join(dir, "conf", "fastcgi_params")
	if _, err := os.Stat(fastcgiFile); err == nil {
		return
	}
	content := `fastcgi_param  QUERY_STRING       $query_string;
fastcgi_param  REQUEST_METHOD     $request_method;
fastcgi_param  CONTENT_TYPE       $content_type;
fastcgi_param  CONTENT_LENGTH     $content_length;
fastcgi_param  SCRIPT_FILENAME    $document_root$fastcgi_script_name;
fastcgi_param  SCRIPT_NAME        $fastcgi_script_name;
fastcgi_param  REQUEST_URI        $request_uri;
fastcgi_param  DOCUMENT_URI       $document_uri;
fastcgi_param  DOCUMENT_ROOT      $document_root;
fastcgi_param  SERVER_PROTOCOL    $server_protocol;
fastcgi_param  GATEWAY_INTERFACE  CGI/1.1;
fastcgi_param  SERVER_SOFTWARE    nginx/$nginx_version;
fastcgi_param  REMOTE_ADDR        $remote_addr;
fastcgi_param  REMOTE_PORT        $remote_port;
fastcgi_param  SERVER_ADDR        $server_addr;
fastcgi_param  SERVER_PORT        $server_port;
fastcgi_param  SERVER_NAME        $server_name;
fastcgi_param  REDIRECT_STATUS    200;
`
	_ = os.WriteFile(fastcgiFile, []byte(content), 0644)
}

func (g *Gateway) syncVhosts(dir string, routes []ports.Route) error {
	vhostsDir := filepath.Join(dir, "conf", "vhosts")
	_ = os.MkdirAll(vhostsDir, 0755)

	activeFiles := make(map[string]bool)
	port := g.port
	if port <= 0 {
		port = 80
	}

	for _, route := range routes {
		fileName := fmt.Sprintf("site_%d.conf", route.ID)
		activeFiles[fileName] = true
		filePath := filepath.Join(vhostsDir, fileName)
		confContent := g.generateVhost(route, port)
		if confContent != "" {
			_ = os.WriteFile(filePath, []byte(confContent), 0644)
		}
	}

	entries, err := os.ReadDir(vhostsDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".conf") {
				continue
			}
			if !activeFiles[entry.Name()] {
				_ = os.Remove(filepath.Join(vhostsDir, entry.Name()))
			}
		}
	}
	return nil
}

func (g *Gateway) generateVhost(route ports.Route, port int) string {
	if len(route.Hosts) == 0 {
		return ""
	}
	hostStr := strings.Join(route.Hosts, " ")
	cleanPath := filepath.ToSlash(route.Path)
	logPath := filepath.ToSlash(route.Log)

	var logDirective string
	if logPath != "" {
		_ = os.MkdirAll(filepath.Dir(route.Log), 0755)
		logDirective = fmt.Sprintf("\r\n    access_log \"%s\";", logPath)
	}

	switch route.Kind {
	case site.KindStatic:
		return fmt.Sprintf(`server {
    listen %d;
    server_name %s;
    root "%s";
    index index.html index.htm;%s

    location / {
        try_files $uri $uri/ /index.html =404;
    }
}
`, port, hostStr, cleanPath, logDirective)

	case site.KindPHP:
		phpPort := route.Port
		if phpPort <= 0 {
			phpPort = 9000
		}
		return fmt.Sprintf(`server {
    listen %d;
    server_name %s;
    root "%s";
    index index.php index.html index.htm;%s

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        fastcgi_pass   127.0.0.1:%d;
        fastcgi_index  index.php;
        fastcgi_param  SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include        fastcgi_params;
    }
}
`, port, hostStr, cleanPath, logDirective, phpPort)

	case site.KindNode:
		nodePort := route.Port
		return fmt.Sprintf(`server {
    listen %d;
    server_name %s;%s

    location / {
        proxy_pass http://127.0.0.1:%d;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
`, port, hostStr, logDirective, nodePort)
	}

	return ""
}

func (g *Gateway) reload(nginxBin, dir string) {
	cmd := exec.Command(nginxBin, "-p", dir, "-s", "reload")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = cmd.Run()
}
