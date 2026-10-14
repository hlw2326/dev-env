package nginx

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"devenv/internal/domain/site"
)

type Manager struct {
	root           string
	mu             sync.RWMutex
	sites          []site.Site
	phpPaths       map[int64]string
	defaultPHPPath string
	servers        map[int]*http.Server
}

func NewManager(root string) *Manager {
	return &Manager{
		root:     root,
		phpPaths: make(map[int64]string),
		servers:  make(map[int]*http.Server),
	}
}

func (m *Manager) Generate(item *site.Site) string {
	var b strings.Builder

	domains := strings.Join(item.Domains, " ")
	rootSlash := filepath.ToSlash(item.Root)

	b.WriteString("server {\n")
	b.WriteString(fmt.Sprintf("    listen %d;\n", item.Port))
	if item.TLS {
		b.WriteString("    listen 443 ssl http2;\n")
		if item.Cert != "" {
			b.WriteString(fmt.Sprintf("    ssl_certificate %s;\n", filepath.ToSlash(item.Cert)))
		}
		if item.Key != "" {
			b.WriteString(fmt.Sprintf("    ssl_certificate_key %s;\n", filepath.ToSlash(item.Key)))
		}
		b.WriteString("    ssl_protocols TLSv1.2 TLSv1.3;\n")
	}
	b.WriteString(fmt.Sprintf("    server_name %s;\n", domains))

	switch item.Kind {
	case site.Static:
		b.WriteString(fmt.Sprintf("    root %s;\n", rootSlash))
		b.WriteString(fmt.Sprintf("    index %s;\n\n", item.Meta.Entry))
		b.WriteString("    location / {\n")
		b.WriteString(fmt.Sprintf("        try_files $uri $uri/ /%s;\n", item.Meta.Entry))
		b.WriteString("    }\n")

	case site.PHP:
		upstream := 9000
		if item.Upstream != nil {
			upstream = *item.Upstream
		}
		b.WriteString(fmt.Sprintf("    root %s;\n", rootSlash))
		b.WriteString(fmt.Sprintf("    index %s index.html;\n\n", item.Meta.Entry))
		b.WriteString("    location / {\n")
		b.WriteString(fmt.Sprintf("        try_files $uri $uri/ /%s?$query_string;\n", item.Meta.Entry))
		b.WriteString("    }\n\n")
		b.WriteString("    location ~ \\.php$ {\n")
		b.WriteString(fmt.Sprintf("        fastcgi_pass 127.0.0.1:%d;\n", upstream))
		b.WriteString(fmt.Sprintf("        fastcgi_index %s;\n", item.Meta.Entry))
		b.WriteString("        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;\n")
		b.WriteString("        include fastcgi_params;\n")
		b.WriteString("    }\n")

	case site.Node, site.Go, site.Python:
		upstream := site.Presets[item.Kind].Port
		if item.Upstream != nil {
			upstream = *item.Upstream
		}
		b.WriteString("\n    location / {\n")
		b.WriteString(fmt.Sprintf("        proxy_pass http://127.0.0.1:%d;\n", upstream))
		b.WriteString("        proxy_http_version 1.1;\n")
		b.WriteString("        proxy_set_header Upgrade $http_upgrade;\n")
		b.WriteString("        proxy_set_header Connection 'upgrade';\n")
		b.WriteString("        proxy_set_header Host $host;\n")
		b.WriteString("        proxy_set_header X-Real-IP $remote_addr;\n")
		b.WriteString("        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n")
		b.WriteString("        proxy_cache_bypass $http_upgrade;\n")
		b.WriteString("    }\n")
	}

	if strings.TrimSpace(item.Meta.Custom) != "" {
		b.WriteString("\n    # 自定义配置\n")
		for _, line := range strings.Split(strings.TrimSpace(item.Meta.Custom), "\n") {
			b.WriteString("    " + line + "\n")
		}
	}

	b.WriteString("\n")
	if item.Meta.AccessLog != "" {
		b.WriteString(fmt.Sprintf("    access_log %s;\n", filepath.ToSlash(item.Meta.AccessLog)))
	}
	if item.Meta.ErrorLog != "" {
		b.WriteString(fmt.Sprintf("    error_log %s;\n", filepath.ToSlash(item.Meta.ErrorLog)))
	}
	b.WriteString("}\n")

	return b.String()
}

func (m *Manager) Write(ctx context.Context, id int64, domain, content string) (string, error) {
	dir := filepath.Join(m.root, "config", "nginx", "vhosts")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	name := strings.ReplaceAll(domain, ":", "_")
	if name == "" {
		name = fmt.Sprintf("site-%d", id)
	}
	path := filepath.Join(dir, name+".conf")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}
	return path, nil
}

func (m *Manager) Read(ctx context.Context, path string) (string, error) {
	if path == "" {
		return "", nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (m *Manager) Delete(ctx context.Context, path string) error {
	if path == "" {
		return nil
	}
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return os.Remove(path)
}

func (m *Manager) Check(ctx context.Context, content string) (bool, string) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return false, "nginx: [emerg] configuration file is empty"
	}
	openBraces := strings.Count(trimmed, "{")
	closeBraces := strings.Count(trimmed, "}")
	if openBraces != closeBraces {
		return false, fmt.Sprintf("nginx: [emerg] unexpected \"}\" or unclosed \"{\" (open: %d, close: %d)", openBraces, closeBraces)
	}
	if !strings.Contains(trimmed, "server") || !strings.Contains(trimmed, "listen") {
		return false, "nginx: [emerg] missing 'server' block or 'listen' directive in configuration"
	}
	return true, "nginx: the configuration file syntax is ok\nnginx: configuration file test is successful"
}

func (m *Manager) Reload(ctx context.Context) error {
	return nil
}

// SyncSites 根据当前所有网站状态同步内置虚拟主机 HTTP 监听服务
func (m *Manager) SyncSites(sites []site.Site, phpPaths map[int64]string, defaultPHPPath string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sites = make([]site.Site, len(sites))
	copy(m.sites, sites)

	m.phpPaths = make(map[int64]string, len(phpPaths))
	for k, v := range phpPaths {
		m.phpPaths[k] = v
	}
	m.defaultPHPPath = defaultPHPPath

	neededPorts := make(map[int]bool)
	for _, s := range sites {
		if s.Status == site.Running {
			port := s.Port
			if port <= 0 {
				port = 80
			}
			neededPorts[port] = true
		}
	}

	// 关闭不再需要的端口监听
	for port, srv := range m.servers {
		if !neededPorts[port] {
			_ = srv.Close()
			delete(m.servers, port)
		}
	}

	// 启动新增需要的端口监听
	for port := range neededPorts {
		if _, exists := m.servers[port]; exists {
			continue
		}
		p := port
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", p))
		if err != nil {
			continue
		}
		srv := &http.Server{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				m.serveHTTPForPort(p, w, r)
			}),
		}
		m.servers[p] = srv
		go func() {
			_ = srv.Serve(ln)
		}()
	}
}

func (m *Manager) serveHTTPForPort(port int, w http.ResponseWriter, r *http.Request) {
	host := strings.ToLower(strings.TrimSpace(r.Host))
	if idx := strings.IndexByte(host, ':'); idx >= 0 {
		host = host[:idx]
	}

	m.mu.RLock()
	var matched *site.Site
	var fallback *site.Site
	for i := range m.sites {
		s := &m.sites[i]
		if s.Status != site.Running {
			continue
		}
		sp := s.Port
		if sp <= 0 {
			sp = 80
		}
		if sp != port {
			continue
		}
		if fallback == nil {
			cp := *s
			fallback = &cp
		}
		for _, d := range s.Domains {
			if strings.EqualFold(strings.TrimSpace(d), host) {
				cp := *s
				matched = &cp
				break
			}
		}
		if matched != nil {
			break
		}
	}
	if matched == nil {
		matched = fallback
	}

	var phpDir string
	if matched != nil && matched.RuntimeID != nil {
		phpDir = m.phpPaths[*matched.RuntimeID]
	}
	if phpDir == "" {
		phpDir = m.defaultPHPPath
	}
	m.mu.RUnlock()

	if matched == nil {
		http.Error(w, "404 Site Not Found", http.StatusNotFound)
		return
	}

	m.appendAccessLog(*matched, r)

	switch matched.Kind {
	case site.Static:
		m.serveStaticSite(*matched, w, r)
	case site.PHP:
		m.servePHPSite(*matched, phpDir, port, w, r)
	case site.Node, site.Go, site.Python:
		m.serveReverseProxy(*matched, w, r)
	default:
		m.serveStaticSite(*matched, w, r)
	}
}

func (m *Manager) appendAccessLog(cur site.Site, r *http.Request) {
	if cur.Meta.AccessLog == "" {
		return
	}
	logPath := cur.Meta.AccessLog
	if !filepath.IsAbs(logPath) {
		logPath = filepath.Join(m.root, logPath)
	}
	_ = os.MkdirAll(filepath.Dir(logPath), 0755)
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	line := fmt.Sprintf("[%s] %s %s %s \"%s\"\n",
		time.Now().Format("2006-01-02 15:04:05"),
		r.RemoteAddr,
		r.Method,
		r.URL.RequestURI(),
		r.UserAgent(),
	)
	_, _ = f.WriteString(line)
}

func (m *Manager) serveStaticSite(cur site.Site, w http.ResponseWriter, r *http.Request) {
	entry := cur.Meta.Entry
	if entry == "" {
		entry = "index.html"
	}
	reqPath := filepath.Clean(strings.TrimPrefix(r.URL.Path, "/"))
	target := filepath.Join(cur.Root, reqPath)
	if info, err := os.Stat(target); err == nil && !info.IsDir() {
		http.ServeFile(w, r, target)
		return
	}
	if info, err := os.Stat(target); err == nil && info.IsDir() {
		idx := filepath.Join(target, entry)
		if _, err := os.Stat(idx); err == nil {
			http.ServeFile(w, r, idx)
			return
		}
	}
	fallback := filepath.Join(cur.Root, entry)
	if _, err := os.Stat(fallback); err == nil {
		http.ServeFile(w, r, fallback)
		return
	}
	http.NotFound(w, r)
}

func (m *Manager) serveReverseProxy(cur site.Site, w http.ResponseWriter, r *http.Request) {
	upstream := site.Presets[cur.Kind].Port
	if cur.Upstream != nil && *cur.Upstream > 0 {
		upstream = *cur.Upstream
	}
	targetURL, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", upstream))
	if err != nil {
		http.Error(w, "Bad Upstream URL", http.StatusBadGateway)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.ServeHTTP(w, r)
}

func resolvePHPCGIExe(phpDir string) string {
	candidates := []string{}
	if phpDir != "" {
		candidates = append(candidates, filepath.Join(phpDir, "php-cgi.exe"))
	}
	if entries, err := os.ReadDir(filepath.Join("runtimes", "php")); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				candidates = append(candidates, filepath.Join("runtimes", "php", e.Name(), "php-cgi.exe"))
			}
		}
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			if abs, err := filepath.Abs(c); err == nil {
				return abs
			}
			return c
		}
	}
	return ""
}

func (m *Manager) servePHPSite(cur site.Site, phpDir string, port int, w http.ResponseWriter, r *http.Request) {
	entry := cur.Meta.Entry
	if entry == "" {
		entry = "index.php"
	}

	urlPath := r.URL.Path
	if urlPath == "" {
		urlPath = "/"
	}

	relPath := filepath.FromSlash(strings.TrimPrefix(urlPath, "/"))
	candidate := filepath.Join(cur.Root, relPath)

	// 1. 静态资源文件直接返回
	if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
		if !strings.HasSuffix(strings.ToLower(candidate), ".php") {
			http.ServeFile(w, r, candidate)
			return
		}
	}

	// 2. 确定要执行的 PHP 脚本与 PATH_INFO
	scriptFile := filepath.Join(cur.Root, entry)
	scriptName := "/" + strings.TrimPrefix(entry, "/")
	pathInfo := ""
	queryString := r.URL.RawQuery

	if info, err := os.Stat(candidate); err == nil && !info.IsDir() && strings.HasSuffix(strings.ToLower(candidate), ".php") {
		scriptFile = candidate
		scriptName = urlPath
	} else if urlPath != "/" && urlPath != scriptName {
		pathInfo = urlPath
		if queryString == "" {
			queryString = "s=" + urlPath
		} else if !strings.Contains(queryString, "s=") {
			queryString = "s=" + urlPath + "&" + queryString
		}
	}

	cgiExe := resolvePHPCGIExe(phpDir)
	if cgiExe == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "<!DOCTYPE html><html><head><meta charset=\"utf-8\"><title>%s</title></head><body style=\"font-family:sans-serif;padding:40px;\"><h2>网站「%s」已启动</h2><p>根目录：<code>%s</code></p><p>请先在「软件商店」安装 PHP 运行环境后刷新此页面。</p></body></html>", cur.Name, cur.Name, cur.Root)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	iniPath := filepath.Join(filepath.Dir(cgiExe), "php.ini")
	args := []string{}
	if _, err := os.Stat(iniPath); err == nil {
		args = append(args, "-c", iniPath)
	}

	cmd := exec.CommandContext(ctx, cgiExe, args...)
	cmd.Dir = cur.Root
	if r.Body != nil {
		cmd.Stdin = r.Body
	}

	hostOnly := r.Host
	if idx := strings.IndexByte(hostOnly, ':'); idx >= 0 {
		hostOnly = hostOnly[:idx]
	}
	if hostOnly == "" && len(cur.Domains) > 0 {
		hostOnly = cur.Domains[0]
	}

	env := append(os.Environ(),
		"REDIRECT_STATUS=200",
		"GATEWAY_INTERFACE=CGI/1.1",
		"SERVER_SOFTWARE=nginx/1.26.3",
		"SERVER_NAME="+hostOnly,
		"SERVER_ADDR=127.0.0.1",
		"SERVER_PORT="+strconv.Itoa(port),
		"SERVER_PROTOCOL="+r.Proto,
		"REQUEST_METHOD="+r.Method,
		"REQUEST_URI="+r.URL.RequestURI(),
		"DOCUMENT_ROOT="+cur.Root,
		"SCRIPT_FILENAME="+scriptFile,
		"SCRIPT_NAME="+scriptName,
		"PATH_INFO="+pathInfo,
		"QUERY_STRING="+queryString,
		"REMOTE_ADDR=127.0.0.1",
		"HTTP_HOST="+r.Host,
	)

	if ct := r.Header.Get("Content-Type"); ct != "" {
		env = append(env, "CONTENT_TYPE="+ct)
	}
	if r.ContentLength > 0 {
		env = append(env, "CONTENT_LENGTH="+strconv.FormatInt(r.ContentLength, 10))
	} else if cl := r.Header.Get("Content-Length"); cl != "" {
		env = append(env, "CONTENT_LENGTH="+cl)
	}

	for k, vals := range r.Header {
		envKey := "HTTP_" + strings.ToUpper(strings.ReplaceAll(k, "-", "_"))
		if envKey == "HTTP_CONTENT_TYPE" || envKey == "HTTP_CONTENT_LENGTH" || envKey == "HTTP_HOST" {
			continue
		}
		env = append(env, envKey+"="+strings.Join(vals, "; "))
	}
	cmd.Env = env

	out, err := cmd.Output()
	if err != nil && len(out) == 0 {
		http.Error(w, fmt.Sprintf("PHP CGI Error: %v", err), http.StatusBadGateway)
		return
	}

	writeCGIResponse(w, out)
}

func writeCGIResponse(w http.ResponseWriter, raw []byte) {
	sep := []byte("\r\n\r\n")
	idx := bytes.Index(raw, sep)
	sepLen := 4
	if idx < 0 {
		sep = []byte("\n\n")
		idx = bytes.Index(raw, sep)
		sepLen = 2
	}

	if idx < 0 {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(raw)
		return
	}

	headerBlock := string(raw[:idx])
	body := raw[idx+sepLen:]

	statusCode := http.StatusOK
	lines := strings.Split(strings.ReplaceAll(headerBlock, "\r\n", "\n"), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if strings.EqualFold(key, "Status") {
			fields := strings.Fields(val)
			if len(fields) > 0 {
				if code, err := strconv.Atoi(fields[0]); err == nil {
					statusCode = code
				}
			}
			continue
		}
		if strings.EqualFold(key, "Location") && statusCode == http.StatusOK {
			statusCode = http.StatusFound
		}
		w.Header().Add(key, val)
	}

	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
	w.WriteHeader(statusCode)
	_, _ = w.Write(body)
}

