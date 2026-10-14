package gateway

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"devenv/internal/domain/site"
	"devenv/internal/ports"
	appErrors "devenv/internal/shared/errors"
)

type Server struct {
	mu       sync.RWMutex
	port     int
	listener net.Listener
	server   *http.Server
	routes   map[string]ports.Route
	all      []ports.Route
	phpSem   chan struct{}
}

func NewServer() *Server {
	return &Server{
		routes: make(map[string]ports.Route),
		phpSem: make(chan struct{}, 8),
	}
}

func (s *Server) Start(port int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener != nil && s.port == port {
		return nil
	}
	s.stopLocked()

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodePortAlreadyInUse, fmt.Sprintf("网关端口 %d 已被占用", port))
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: http.HandlerFunc(s.ServeHTTP),
	}
	s.port = port
	s.listener = ln
	s.server = srv

	go func() {
		_ = srv.Serve(ln)
	}()
	return nil
}

func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopLocked()
	return nil
}

func (s *Server) stopLocked() {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_ = s.server.Shutdown(ctx)
		cancel()
		_ = s.server.Close()
		s.server = nil
	}
	if s.listener != nil {
		_ = s.listener.Close()
		s.listener = nil
	}
	s.port = 0
}

func (s *Server) Port() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.port
}

func (s *Server) SetRoutes(routes []ports.Route) {
	byHost := make(map[string]ports.Route)
	cloned := make([]ports.Route, len(routes))
	copy(cloned, routes)
	for _, r := range cloned {
		for _, h := range r.Hosts {
			clean := strings.ToLower(strings.TrimSpace(h))
			if clean != "" {
				byHost[clean] = r
			}
		}
	}
	s.mu.Lock()
	s.routes = byHost
	s.all = cloned
	s.mu.Unlock()
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := strings.ToLower(strings.TrimSpace(r.Host))
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	s.mu.RLock()
	route, matched := s.routes[host]
	allRoutes := s.all
	port := s.port
	s.mu.RUnlock()

	if !matched {
		renderUnmatchedPage(w, allRoutes, port)
		return
	}

	start := time.Now()
	recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
	defer func() {
		appendAccessLog(route.Log, r, recorder.status, time.Since(start))
	}()

	if isHiddenOrUnsafeRequestPath(r.URL.Path, r.URL.RawPath) {
		http.Error(recorder, "403 Forbidden", http.StatusForbidden)
		return
	}

	switch route.Kind {
	case site.KindStatic:
		s.serveStatic(recorder, r, route)
	case site.KindPHP:
		s.servePHP(recorder, r, route)
	case site.KindNode:
		s.serveNodeProxy(recorder, r, route)
	default:
		http.Error(recorder, "502 Unsupported Site Kind", http.StatusBadGateway)
	}
}

func (s *Server) serveStatic(w http.ResponseWriter, r *http.Request, route ports.Route) {
	target, ok := resolveSafeFile(route.Path, r.URL.Path)
	if !ok {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}
	info, err := os.Stat(target)
	if err == nil && info.IsDir() {
		target = filepath.Join(target, "index.html")
		info, err = os.Stat(target)
	}
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, target)
}

func (s *Server) servePHP(w http.ResponseWriter, r *http.Request, route ports.Route) {
	target, ok := resolveSafeFile(route.Path, r.URL.Path)
	if !ok {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}

	info, err := os.Stat(target)
	if err == nil && !info.IsDir() && !strings.HasSuffix(strings.ToLower(target), ".php") {
		http.ServeFile(w, r, target)
		return
	}

	scriptFile := target
	scriptName := r.URL.Path
	pathInfo := ""

	if err == nil && info.IsDir() {
		idxPHP := filepath.Join(target, "index.php")
		if fi, statErr := os.Stat(idxPHP); statErr == nil && !fi.IsDir() {
			scriptFile = idxPHP
			scriptName = strings.TrimRight(r.URL.Path, "/") + "/index.php"
		} else {
			idxHTML := filepath.Join(target, "index.html")
			if fi, statErr := os.Stat(idxHTML); statErr == nil && !fi.IsDir() {
				http.ServeFile(w, r, idxHTML)
				return
			}
		}
	}

	if fi, statErr := os.Stat(scriptFile); statErr != nil || fi.IsDir() || !strings.HasSuffix(strings.ToLower(scriptFile), ".php") {
		fallback := filepath.Join(route.Path, "index.php")
		if fi2, err2 := os.Stat(fallback); err2 == nil && !fi2.IsDir() {
			scriptFile = fallback
			scriptName = "/index.php"
			pathInfo = r.URL.Path
		} else {
			http.NotFound(w, r)
			return
		}
	}

	phpCGI := filepath.Join(route.PHP, "php-cgi.exe")
	if route.PHP == "" {
		http.Error(w, "PHP 运行环境未安装", http.StatusServiceUnavailable)
		return
	}
	if fi, err := os.Stat(phpCGI); err != nil || fi.IsDir() {
		http.Error(w, "找不到 php-cgi.exe: "+phpCGI, http.StatusServiceUnavailable)
		return
	}

	select {
	case s.phpSem <- struct{}{}:
		defer func() { <-s.phpSem }()
	case <-r.Context().Done():
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	var bodyBuf bytes.Buffer
	if r.Body != nil {
		_, _ = io.Copy(&bodyBuf, r.Body)
	}

	cmd := exec.CommandContext(ctx, phpCGI, "-c", route.PHP)
	cmd.Dir = route.Path
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Stdin = &bodyBuf

	env := append(os.Environ(),
		"GATEWAY_INTERFACE=CGI/1.1",
		"SERVER_PROTOCOL="+r.Proto,
		"REDIRECT_STATUS=200",
		"REQUEST_METHOD="+r.Method,
		"SCRIPT_FILENAME="+scriptFile,
		"SCRIPT_NAME="+scriptName,
		"DOCUMENT_ROOT="+route.Path,
		"PATH_INFO="+pathInfo,
		"REQUEST_URI="+r.URL.RequestURI(),
		"QUERY_STRING="+r.URL.RawQuery,
		"HTTP_HOST="+r.Host,
		"SERVER_NAME="+r.Host,
		"REMOTE_ADDR=127.0.0.1",
		"CONTENT_TYPE="+r.Header.Get("Content-Type"),
		"CONTENT_LENGTH="+strconv.Itoa(bodyBuf.Len()),
	)
	for k, vals := range r.Header {
		envKey := "HTTP_" + strings.ToUpper(strings.ReplaceAll(k, "-", "_"))
		if envKey == "HTTP_CONTENT_TYPE" || envKey == "HTTP_CONTENT_LENGTH" {
			continue
		}
		env = append(env, envKey+"="+strings.Join(vals, ", "))
	}
	cmd.Env = env

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil && out.Len() == 0 {
		http.Error(w, "PHP 执行失败: "+stderr.String(), http.StatusBadGateway)
		return
	}

	writeCGIResponse(w, out.Bytes())
}

func (s *Server) serveNodeProxy(w http.ResponseWriter, r *http.Request, route ports.Route) {
	if route.Port <= 0 {
		http.Error(w, "Node 站点端口未配置", http.StatusBadGateway)
		return
	}
	targetURL, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", route.Port))
	if err != nil {
		http.Error(w, "反代目标地址无效", http.StatusBadGateway)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	origDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		origDirector(req)
		req.Host = r.Host
	}
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, proxyErr error) {
		http.Error(rw, "Node 服务未就绪: "+proxyErr.Error(), http.StatusBadGateway)
	}
	proxy.ServeHTTP(w, r)
}

func writeCGIResponse(w http.ResponseWriter, raw []byte) {
	sep := []byte("\r\n\r\n")
	idx := bytes.Index(raw, sep)
	if idx < 0 {
		sep = []byte("\n\n")
		idx = bytes.Index(raw, sep)
	}
	if idx < 0 {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(raw)
		return
	}

	headerBlock := string(raw[:idx])
	body := raw[idx+len(sep):]
	statusCode := http.StatusOK

	lines := strings.Split(strings.ReplaceAll(headerBlock, "\r\n", "\n"), "\n")
	for _, line := range lines {
		colon := strings.IndexByte(line, ':')
		if colon <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:colon])
		val := strings.TrimSpace(line[colon+1:])
		if strings.EqualFold(key, "Status") {
			fields := strings.Fields(val)
			if len(fields) > 0 {
				if code, err := strconv.Atoi(fields[0]); err == nil {
					statusCode = code
				}
			}
			continue
		}
		w.Header().Add(key, val)
	}
	if w.Header().Get("Location") != "" && statusCode == http.StatusOK {
		statusCode = http.StatusFound
	}
	w.WriteHeader(statusCode)
	_, _ = w.Write(body)
}

func isHiddenOrUnsafeRequestPath(urlPath, rawPath string) bool {
	for _, candidate := range []string{urlPath, rawPath} {
		decoded, err := url.PathUnescape(candidate)
		if err != nil {
			return true
		}
		norm := strings.ReplaceAll(decoded, `\`, "/")
		segments := strings.Split(norm, "/")
		for _, seg := range segments {
			if seg == "" || seg == "." {
				continue
			}
			if seg == ".." {
				return true
			}
			if strings.HasPrefix(seg, ".") {
				return true
			}
		}
	}
	return false
}

func resolveSafeFile(docRoot, reqPath string) (string, bool) {
	absRoot, err := filepath.Abs(docRoot)
	if err != nil {
		return "", false
	}
	rel := strings.TrimPrefix(strings.ReplaceAll(reqPath, `\`, "/"), "/")
	cleaned := filepath.Clean(filepath.FromSlash(rel))
	if cleaned == ".." || strings.HasPrefix(filepath.ToSlash(cleaned), "../") {
		return "", false
	}
	target := filepath.Join(absRoot, cleaned)
	relCheck, err := filepath.Rel(absRoot, target)
	if err != nil || relCheck == ".." || strings.HasPrefix(filepath.ToSlash(relCheck), "../") {
		return "", false
	}
	return target, true
}

func renderUnmatchedPage(w http.ResponseWriter, routes []ports.Route, port int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	var b strings.Builder
	b.WriteString("<!DOCTYPE html><html><head><meta charset=\"utf-8\"><title>DevEnv 站点网关</title></head><body>")
	b.WriteString("<h2>未匹配到站点</h2><ul>")
	sorted := make([]ports.Route, len(routes))
	copy(sorted, routes)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ID < sorted[j].ID })
	for _, r := range sorted {
		for _, h := range r.Hosts {
			href := "http://" + h
			if port != 80 && port > 0 {
				href = fmt.Sprintf("http://%s:%d", h, port)
			}
			b.WriteString(fmt.Sprintf("<li><a href=\"%s\">%s</a></li>", href, h))
		}
	}
	b.WriteString("</ul></body></html>")
	_, _ = w.Write([]byte(b.String()))
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func appendAccessLog(logPath string, r *http.Request, status int, elapsed time.Duration) {
	if logPath == "" {
		return
	}
	_ = os.MkdirAll(filepath.Dir(logPath), 0755)
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	line := fmt.Sprintf("[%s] %s %s %s -> %d (%dms)\n",
		time.Now().Format("2006-01-02 15:04:05"),
		r.Method,
		r.Host,
		r.URL.RequestURI(),
		status,
		elapsed.Milliseconds(),
	)
	_, _ = f.WriteString(line)
}
