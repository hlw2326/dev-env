package gateway

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"html"
	"io"
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
	"syscall"
	"time"

	"devenv/internal/domain/site"
	"devenv/internal/ports"
	appErrors "devenv/internal/shared/errors"
	"devenv/internal/shared/paths"
)

const (
	PHP_CONCURRENCY  = 8
	PHP_TIMEOUT      = 60 * time.Second
	SHUTDOWN_TIMEOUT = 5 * time.Second
)

type Server struct {
	mu       sync.RWMutex
	port     int
	listener net.Listener
	http     *http.Server
	routes   []ports.Route
	limiter  chan struct{}
	logMu    sync.Mutex
}

func New() *Server {
	return &Server{
		limiter: make(chan struct{}, PHP_CONCURRENCY),
	}
}

func (s *Server) Start(port int) error {
	if port <= 0 {
		return appErrors.New(appErrors.INVALID_ARGUMENT, "网关端口无效")
	}
	if err := s.Stop(); err != nil {
		return err
	}
	address := fmt.Sprintf("%s:%d", paths.LOOPBACK, port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return appErrors.Wrap(err, appErrors.PORT_ALREADY_IN_USE, fmt.Sprintf("网关端口 %d 已被占用", port))
	}
	server := &http.Server{
		Handler: s,
	}
	s.mu.Lock()
	s.port = port
	s.listener = listener
	s.http = server
	s.mu.Unlock()

	go s.serve(server, listener)
	return nil
}

func (s *Server) serve(server *http.Server, listener net.Listener) {
	err := server.Serve(listener)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.mu.Lock()
		if s.http == server {
			s.port = 0
			s.listener = nil
			s.http = nil
		}
		s.mu.Unlock()
	}
}

func (s *Server) Stop() error {
	s.mu.Lock()
	server := s.http
	s.port = 0
	s.listener = nil
	s.http = nil
	s.mu.Unlock()

	if server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), SHUTDOWN_TIMEOUT)
	defer cancel()
	err := server.Shutdown(ctx)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Port() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.port
}

func (s *Server) SetRoutes(routes []ports.Route) error {
	copied := make([]ports.Route, len(routes))
	for i, item := range routes {
		hosts := make([]string, len(item.Hosts))
		copy(hosts, item.Hosts)
		copied[i] = item
		copied[i].Hosts = hosts
	}
	s.mu.Lock()
	s.routes = copied
	s.mu.Unlock()
	return nil
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mu.RLock()
	routes := s.routes
	s.mu.RUnlock()

	host := parseHost(request.Host)
	route, found := matchRoute(routes, host)
	if !found {
		s.serveOverview(writer, routes)
		return
	}

	wrapped := &recorder{
		ResponseWriter: writer,
		status:         http.StatusOK,
	}
	defer s.appendLog(route.Log, request, wrapped.status)

	if checkBlocked(request.URL.Path) || checkBlocked(request.RequestURI) {
		http.Error(wrapped, "Forbidden", http.StatusForbidden)
		return
	}

	switch route.Kind {
	case site.KIND_STATIC:
		s.serveStatic(wrapped, request, route)
	case site.KIND_PHP:
		s.servePHP(wrapped, request, route)
	case site.KIND_NODE:
		s.serveNode(wrapped, request, route)
	default:
		http.NotFound(wrapped, request)
	}
}

func (s *Server) serveOverview(writer http.ResponseWriter, routes []ports.Route) {
	var builder strings.Builder
	builder.WriteString("<!doctype html><html lang=\"zh-CN\"><head><meta charset=\"utf-8\"><title>DevEnv Gateway</title></head><body>")
	builder.WriteString("<h1>DevEnv 站点列表</h1>")
	if len(routes) == 0 {
		builder.WriteString("<p>当前没有运行中的站点</p>")
	} else {
		builder.WriteString("<ul>")
		for _, route := range routes {
			for _, host := range route.Hosts {
				escaped := html.EscapeString(host)
				builder.WriteString("<li><a href=\"http://")
				builder.WriteString(escaped)
				builder.WriteString("/\">")
				builder.WriteString(escaped)
				builder.WriteString("</a></li>")
			}
		}
		builder.WriteString("</ul>")
	}
	builder.WriteString("</body></html>")
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(http.StatusNotFound)
	if _, err := io.WriteString(writer, builder.String()); err != nil {
		return
	}
}

func (s *Server) serveStatic(writer http.ResponseWriter, request *http.Request, route ports.Route) {
	target, ok := resolveTarget(route.Path, request.URL.Path)
	if !ok {
		http.Error(writer, "Forbidden", http.StatusForbidden)
		return
	}
	info, err := os.Stat(target)
	if err != nil {
		http.NotFound(writer, request)
		return
	}
	if info.IsDir() {
		index := filepath.Join(target, "index.html")
		indexInfo, statErr := os.Stat(index)
		if statErr != nil || indexInfo.IsDir() {
			http.NotFound(writer, request)
			return
		}
		serveFile(writer, request, index, indexInfo)
		return
	}
	serveFile(writer, request, target, info)
}

func (s *Server) servePHP(writer http.ResponseWriter, request *http.Request, route ports.Route) {
	target, ok := resolveTarget(route.Path, request.URL.Path)
	if !ok {
		http.Error(writer, "Forbidden", http.StatusForbidden)
		return
	}
	info, err := os.Stat(target)
	if err == nil && !info.IsDir() {
		if !strings.EqualFold(filepath.Ext(target), ".php") {
			serveFile(writer, request, target, info)
			return
		}
		s.runPHP(writer, request, route, target)
		return
	}

	var script string
	if err == nil && info.IsDir() {
		candidate := filepath.Join(target, "index.php")
		if fileInfo, statErr := os.Stat(candidate); statErr == nil && !fileInfo.IsDir() {
			script = candidate
		}
	}
	if script == "" {
		entry := filepath.Join(route.Path, "index.php")
		if fileInfo, statErr := os.Stat(entry); statErr == nil && !fileInfo.IsDir() {
			script = entry
		}
	}
	if script == "" {
		http.NotFound(writer, request)
		return
	}
	s.runPHP(writer, request, route, script)
}

func (s *Server) runPHP(writer http.ResponseWriter, request *http.Request, route ports.Route, script string) {
	if strings.TrimSpace(route.PHP) == "" {
		http.Error(writer, "PHP runtime not configured", http.StatusBadGateway)
		return
	}
	if _, err := os.Stat(route.PHP); err != nil {
		http.Error(writer, "PHP runtime missing", http.StatusBadGateway)
		return
	}

	select {
	case s.limiter <- struct{}{}:
		defer func() { <-s.limiter }()
	case <-request.Context().Done():
		http.Error(writer, "Request cancelled", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(request.Context(), PHP_TIMEOUT)
	defer cancel()

	rel, err := filepath.Rel(route.Path, script)
	if err != nil {
		http.Error(writer, "Invalid script path", http.StatusInternalServerError)
		return
	}
	name := "/" + strings.TrimPrefix(filepath.ToSlash(rel), "/")
	uri := request.URL.RequestURI()
	if request.RequestURI != "" {
		uri = request.RequestURI
	}

	command := exec.CommandContext(ctx, route.PHP, "-d", "cgi.force_redirect=0")
	command.Dir = filepath.Dir(script)
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	command.Env = buildEnv(request, route.Path, script, name, uri)
	if request.Body != nil {
		command.Stdin = request.Body
	}

	output, err := command.Output()
	if err != nil && len(output) == 0 {
		http.Error(writer, "PHP execution failed", http.StatusBadGateway)
		return
	}
	writeCGI(writer, output)
}

func (s *Server) serveNode(writer http.ResponseWriter, request *http.Request, route ports.Route) {
	if route.Port <= 0 {
		http.Error(writer, "Node upstream port not configured", http.StatusBadGateway)
		return
	}
	upstream := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", paths.LOOPBACK, route.Port),
	}
	proxy := httputil.NewSingleHostReverseProxy(upstream)
	proxy.ErrorHandler = func(rw http.ResponseWriter, r *http.Request, err error) {
		http.Error(rw, "Bad Gateway", http.StatusBadGateway)
	}
	proxy.ServeHTTP(writer, request)
}

func (s *Server) appendLog(logPath string, request *http.Request, status int) {
	if strings.TrimSpace(logPath) == "" {
		return
	}
	s.logMu.Lock()
	defer s.logMu.Unlock()

	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return
	}
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	line := fmt.Sprintf("%s %s %s %s %d\n",
		time.Now().Format(time.RFC3339),
		request.Host,
		request.Method,
		request.URL.RequestURI(),
		status,
	)
	if _, err := file.WriteString(line); err != nil {
		if closeErr := file.Close(); closeErr != nil {
			return
		}
		return
	}
	if err := file.Close(); err != nil {
		return
	}
}

func parseHost(raw string) string {
	trimmed := strings.TrimSpace(strings.ToLower(raw))
	if host, _, err := net.SplitHostPort(trimmed); err == nil {
		return host
	}
	return trimmed
}

func matchRoute(routes []ports.Route, host string) (ports.Route, bool) {
	if host == "" {
		return ports.Route{}, false
	}
	for _, route := range routes {
		for _, item := range route.Hosts {
			if strings.EqualFold(strings.TrimSpace(item), host) {
				return route, true
			}
		}
	}
	return ports.Route{}, false
}

func checkBlocked(raw string) bool {
	cleaned := raw
	if idx := strings.IndexByte(cleaned, '?'); idx >= 0 {
		cleaned = cleaned[:idx]
	}
	decoded, err := url.PathUnescape(cleaned)
	if err != nil {
		return true
	}
	for _, candidate := range []string{cleaned, decoded} {
		normalized := strings.ReplaceAll(candidate, `\`, "/")
		segments := strings.Split(normalized, "/")
		for _, segment := range segments {
			trimmed := strings.TrimSpace(segment)
			if trimmed == "" || trimmed == "." {
				continue
			}
			if trimmed == ".." || strings.HasPrefix(trimmed, ".") {
				return true
			}
		}
	}
	return false
}

func resolveTarget(root, requestPath string) (string, bool) {
	if strings.TrimSpace(root) == "" {
		return "", false
	}
	rel := strings.TrimPrefix(strings.ReplaceAll(requestPath, `\`, "/"), "/")
	target := filepath.Join(root, filepath.FromSlash(rel))
	if err := paths.CheckWithin(root, target); err != nil {
		return "", false
	}
	return target, true
}

func serveFile(writer http.ResponseWriter, request *http.Request, filePath string, info os.FileInfo) {
	file, err := os.Open(filePath)
	if err != nil {
		http.NotFound(writer, request)
		return
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			return
		}
	}()
	http.ServeContent(writer, request, info.Name(), info.ModTime(), file)
}

func buildEnv(request *http.Request, root, script, name, uri string) []string {
	length := ""
	if request.ContentLength > 0 {
		length = strconv.FormatInt(request.ContentLength, 10)
	}
	env := []string{
		"GATEWAY_INTERFACE=CGI/1.1",
		"REDIRECT_STATUS=200",
		"SERVER_SOFTWARE=DevEnv",
		"SERVER_PROTOCOL=" + request.Proto,
		"REQUEST_METHOD=" + request.Method,
		"DOCUMENT_ROOT=" + root,
		"SCRIPT_FILENAME=" + script,
		"SCRIPT_NAME=" + name,
		"REQUEST_URI=" + uri,
		"QUERY_STRING=" + request.URL.RawQuery,
		"CONTENT_TYPE=" + request.Header.Get("Content-Type"),
		"CONTENT_LENGTH=" + length,
		"HTTP_HOST=" + request.Host,
		"REMOTE_ADDR=" + paths.LOOPBACK,
		"SERVER_NAME=" + parseHost(request.Host),
	}
	for _, key := range []string{"SystemRoot", "SYSTEMROOT", "WINDIR", "TEMP", "TMP", "PATH", "PATHEXT", "COMSPEC"} {
		if val := os.Getenv(key); val != "" {
			env = append(env, key+"="+val)
		}
	}
	for key, values := range request.Header {
		upper := strings.ToUpper(strings.ReplaceAll(key, "-", "_"))
		if upper == "CONTENT_TYPE" || upper == "CONTENT_LENGTH" || upper == "HOST" {
			continue
		}
		env = append(env, "HTTP_"+upper+"="+strings.Join(values, ", "))
	}
	return env
}

func writeCGI(writer http.ResponseWriter, output []byte) {
	headers := output
	body := []byte(nil)
	if idx := bytes.Index(output, []byte("\r\n\r\n")); idx >= 0 {
		headers = output[:idx]
		body = output[idx+4:]
	} else if idx := bytes.Index(output, []byte("\n\n")); idx >= 0 {
		headers = output[:idx]
		body = output[idx+2:]
	} else {
		body = output
		headers = nil
	}

	status := http.StatusOK
	lines := strings.Split(strings.ReplaceAll(string(headers), "\r\n", "\n"), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if strings.EqualFold(key, "Status") {
			fields := strings.Fields(value)
			if len(fields) > 0 {
				if parsed, err := strconv.Atoi(fields[0]); err == nil {
					status = parsed
				}
			}
			continue
		}
		writer.Header().Add(key, value)
	}
	if writer.Header().Get("Content-Type") == "" {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
	writer.WriteHeader(status)
	if _, err := writer.Write(body); err != nil {
		return
	}
}

type recorder struct {
	http.ResponseWriter
	status  int
	written bool
}

func (r *recorder) WriteHeader(code int) {
	if !r.written {
		r.status = code
		r.written = true
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *recorder) Write(data []byte) (int, error) {
	if !r.written {
		r.status = http.StatusOK
		r.written = true
	}
	return r.ResponseWriter.Write(data)
}

func (r *recorder) Flush() {
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (r *recorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijacker not supported")
	}
	return hijacker.Hijack()
}
