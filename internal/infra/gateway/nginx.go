package gateway

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"html"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"devenv/internal/domain/project"
	"devenv/internal/domain/software"
	processinfra "devenv/internal/infra/process"
	"devenv/internal/ports"
	appErrors "devenv/internal/shared/errors"
	"devenv/internal/shared/paths"
)

const (
	NGINX_KEY             = "gateway:nginx"
	FASTCGI_PORT_START    = 9000
	FASTCGI_PORT_COUNT    = 1000
	GATEWAY_START_TIMEOUT = 10 * time.Second
	GATEWAY_POLL_INTERVAL = 100 * time.Millisecond
	PHP_REQUEST_TIMEOUT   = "60s"
	NGINX_PHP_CONCURRENCY = "8"
	NGINX_CONNECTIONS     = "1024"
	NGINX_KEEPALIVE       = "65"
	NGINX_UPLOAD_LIMIT    = "100m"
)

type pool struct {
	key      string
	port     int
	pid      int
	external bool
}

type softwareReader interface {
	GetMain(ctx context.Context, kind string) (*software.Software, error)
}

// Nginx 管理 Web 项目访问入口及项目使用的 PHP FastCGI 进程。
type Nginx struct {
	home     string
	software softwareReader
	runner   ports.Runner
	checker  ports.PortChecker

	mutex     sync.RWMutex
	operation sync.Mutex
	port      int
	listening int
	external  bool
	routes    []ports.Route
	pools     map[string]pool
}

func NewNginx(home string, software softwareReader, runner ports.Runner, checker ports.PortChecker) *Nginx {
	return &Nginx{
		home:     home,
		software: software,
		runner:   runner,
		checker:  checker,
		pools:    make(map[string]pool),
	}
}

func (n *Nginx) Start(port int) error {
	if port < 1 || port > ports.MAX_PORT {
		return appErrors.New(appErrors.INVALID_ARGUMENT, "Nginx 监听端口无效")
	}

	n.operation.Lock()
	defer n.operation.Unlock()
	if n.isExternalNginx(port) {
		n.mutex.Lock()
		n.port = port
		n.listening = port
		n.external = true
		n.mutex.Unlock()
		return nil
	}

	n.mutex.Lock()
	previous := n.port
	n.port = port
	n.mutex.Unlock()

	if err := n.reconcile(); err != nil {
		n.mutex.Lock()
		n.port = previous
		n.mutex.Unlock()
		return err
	}
	return nil
}

func (n *Nginx) Stop() error {
	n.operation.Lock()
	defer n.operation.Unlock()

	n.mutex.RLock()
	external := n.external
	pools := make([]pool, 0, len(n.pools))
	for _, item := range n.pools {
		pools = append(pools, item)
	}
	n.mutex.RUnlock()

	stopErrors := make([]error, 0)
	if !external {
		if err := n.runner.Stop(NGINX_KEY); err != nil {
			stopErrors = append(stopErrors, err)
		}
	}
	if err := n.stopPools(pools); err != nil {
		stopErrors = append(stopErrors, err)
	}
	if len(stopErrors) > 0 {
		return errors.Join(stopErrors...)
	}

	n.mutex.Lock()
	n.port = 0
	n.listening = 0
	n.external = false
	n.pools = make(map[string]pool)
	n.mutex.Unlock()
	return nil
}

// StopManaged stops Nginx from the service management screen. An externally
// adopted Nginx instance is terminated here explicitly; the regular Stop
// method remains lifecycle-safe for application shutdown.
func (n *Nginx) StopManaged() error {
	n.mutex.RLock()
	external := n.external
	port := n.listening
	n.mutex.RUnlock()
	if external {
		if finder, ok := n.checker.(interface{ ListeningPID(int) int }); ok {
			if pid := finder.ListeningPID(port); pid > 0 && processinfra.MatchesImageName(pid, "nginx.exe") {
				if err := processinfra.StopPID(pid); err != nil {
					return err
				}
			}
		}
	}
	return n.Stop()
}

// Restart stops the managed Nginx instance and starts it with the latest
// generated routes. It is used after a project Nginx configuration changes.
func (n *Nginx) Restart() error {
	n.mutex.RLock()
	port := n.listening
	if port == 0 {
		port = n.port
	}
	n.mutex.RUnlock()
	if port == 0 {
		return nil
	}
	if err := n.StopManaged(); err != nil {
		return err
	}
	return n.Start(port)
}

func (n *Nginx) SetRoutes(routes []ports.Route) error {
	copied := copyRoutes(routes)

	n.mutex.Lock()
	previous := n.routes
	n.routes = copied
	started := n.port > 0
	n.mutex.Unlock()
	if !started {
		return nil
	}

	n.operation.Lock()
	defer n.operation.Unlock()
	if err := n.reconcile(); err != nil {
		n.mutex.Lock()
		n.routes = previous
		n.mutex.Unlock()
		return err
	}
	return nil
}

func (n *Nginx) Port() int {
	n.mutex.RLock()
	if n.external {
		port := n.listening
		n.mutex.RUnlock()
		if n.checker.Listening(port) {
			return port
		}
		n.mutex.Lock()
		if n.external && n.listening == port {
			n.port = 0
			n.listening = 0
			n.external = false
		}
		n.mutex.Unlock()
		return 0
	}
	defer n.mutex.RUnlock()
	if !n.runner.Status(NGINX_KEY).Running {
		return 0
	}
	return n.listening
}

func (n *Nginx) reconcile() error {
	n.mutex.RLock()
	port := n.port
	routes := copyRoutes(n.routes)
	listening := n.listening
	n.mutex.RUnlock()
	if port == 0 {
		return nil
	}

	installed, err := n.software.GetMain(context.Background(), software.KIND_NGINX)
	if err != nil {
		return err
	}
	if installed == nil || installed.State != software.STATE_INSTALLED {
		return appErrors.New(appErrors.RUNTIME_NOT_INSTALLED, "请先在软件商城安装 Nginx")
	}
	executable := filepath.Join(installed.Path, "nginx.exe")
	info, err := os.Stat(executable)
	if err != nil {
		return appErrors.Wrap(err, appErrors.RUNTIME_NOT_INSTALLED, "读取 nginx.exe 失败")
	}
	if info.IsDir() {
		return appErrors.New(appErrors.RUNTIME_NOT_INSTALLED, "nginx.exe 路径指向了目录")
	}
	if listening != port && !n.checker.Available(port) {
		return appErrors.New(appErrors.PORT_ALREADY_IN_USE, fmt.Sprintf("Nginx 监听端口 %d 已被占用", port))
	}

	pools, started, err := n.startPools(routes, port)
	if err != nil {
		return err
	}
	cleanup := func(source error) error {
		if stopError := n.stopPools(started); stopError != nil {
			return errors.Join(source, stopError)
		}
		return source
	}

	if err := n.writeConfig(installed.Path, port, routes, pools); err != nil {
		return cleanup(err)
	}
	if err := checkNginx(executable, n.home); err != nil {
		return cleanup(err)
	}

	if n.external {
		if err := reloadNginx(executable, n.home); err != nil {
			return cleanup(err)
		}
		if err := n.stopUnused(pools); err != nil {
			return err
		}
		n.mutex.Lock()
		n.listening = port
		n.pools = pools
		n.mutex.Unlock()
		return nil
	}

	status := n.runner.Status(NGINX_KEY)
	if status.Running {
		if err := reloadNginx(executable, n.home); err != nil {
			return cleanup(err)
		}
	} else {
		_, err := n.runner.Start(NGINX_KEY, ports.Options{
			Binary:    executable,
			Args:      []string{"-p", nginxPrefix(n.home), "-c", paths.NGINX_CONFIG},
			Directory: installed.Path,
			Log:       paths.GatewayLog(n.home, "nginx.log"),
			Detached:  true,
		})
		if err != nil {
			return cleanup(err)
		}
	}

	if err := waitPort(n.checker, port); err != nil {
		stopError := n.runner.Stop(NGINX_KEY)
		return cleanup(errors.Join(err, stopError))
	}
	if err := n.stopUnused(pools); err != nil {
		return err
	}

	n.mutex.Lock()
	n.listening = port
	n.external = false
	n.pools = pools
	n.mutex.Unlock()
	return nil
}

func (n *Nginx) isExternalNginx(port int) bool {
	if port <= 0 || n.runner.Status(NGINX_KEY).Running || !n.checker.Listening(port) {
		return false
	}
	finder, ok := n.checker.(interface{ ListeningPID(int) int })
	if !ok {
		return false
	}
	pid := finder.ListeningPID(port)
	return processinfra.MatchesImageName(pid, "nginx.exe")
}

func (n *Nginx) startPools(routes []ports.Route, nginxPort int) (map[string]pool, []pool, error) {
	n.mutex.RLock()
	existing := make(map[string]pool, len(n.pools))
	for executable, item := range n.pools {
		existing[executable] = item
	}
	n.mutex.RUnlock()

	executables := make([]string, 0)
	seen := make(map[string]bool)
	for _, route := range routes {
		executable := filepath.Clean(strings.TrimSpace(route.PHP))
		if route.Kind == project.KIND_PHP && executable != "." && executable != "" && !seen[executable] {
			seen[executable] = true
			executables = append(executables, executable)
		}
	}
	sort.Strings(executables)

	used := map[int]bool{nginxPort: true}
	for _, route := range routes {
		if route.Port > 0 {
			used[route.Port] = true
		}
	}
	result := make(map[string]pool, len(executables))
	started := make([]pool, 0)
	cleanup := func(source error) (map[string]pool, []pool, error) {
		if stopError := n.stopPools(started); stopError != nil {
			return nil, nil, errors.Join(source, stopError)
		}
		return nil, nil, source
	}

	for _, executable := range executables {
		info, err := os.Stat(executable)
		if err != nil {
			return cleanup(appErrors.Wrap(err, appErrors.RUNTIME_NOT_INSTALLED, "读取 php-cgi.exe 失败"))
		}
		if info.IsDir() {
			return cleanup(appErrors.New(appErrors.RUNTIME_NOT_INSTALLED, "php-cgi.exe 路径指向了目录"))
		}

		port := poolPort(executable)
		if used[port] {
			return cleanup(appErrors.New(appErrors.PORT_ALREADY_IN_USE, fmt.Sprintf("PHP FastCGI 固定端口 %d 与现有端口冲突", port)))
		}
		used[port] = true

		item, found := existing[executable]
		if found && item.port == port {
			if n.runner.Status(item.key).Running || item.external && n.externalPoolRunning(item, executable) {
				result[executable] = item
				continue
			}
		}
		if !n.checker.Available(port) {
			if adopted, ok := n.adoptPool(executable, port); ok {
				result[executable] = adopted
				continue
			}
			return cleanup(appErrors.New(appErrors.PORT_ALREADY_IN_USE, fmt.Sprintf("PHP FastCGI 固定端口 %d 已被占用", port)))
		}

		item = pool{key: poolKey(executable), port: port}
		pid, startErr := n.runner.Start(item.key, ports.Options{
			Binary:      executable,
			Args:        []string{"-b", paths.LOOPBACK + ":" + strconv.Itoa(port)},
			Directory:   filepath.Dir(executable),
			Environment: []string{"PHP_FCGI_CHILDREN=" + NGINX_PHP_CONCURRENCY, "PHP_FCGI_MAX_REQUESTS=0"},
			Log:         paths.GatewayLog(n.home, fmt.Sprintf("php-%d.log", port)),
			Detached:    true,
		})
		if startErr != nil {
			return cleanup(startErr)
		}
		item.pid = pid
		started = append(started, item)
		if err := waitPort(n.checker, port); err != nil {
			return cleanup(err)
		}
		result[executable] = item
	}
	return result, started, nil
}

func (n *Nginx) stopUnused(active map[string]pool) error {
	n.mutex.RLock()
	existing := make(map[string]pool, len(n.pools))
	for executable, item := range n.pools {
		existing[executable] = item
	}
	n.mutex.RUnlock()

	unused := make([]pool, 0)
	for executable, item := range existing {
		if _, found := active[executable]; !found {
			unused = append(unused, item)
		}
	}
	return n.stopPools(unused)
}

func (n *Nginx) stopPools(items []pool) error {
	stopErrors := make([]error, 0)
	for _, item := range items {
		if item.external {
			if err := processinfra.StopPID(item.pid); err != nil {
				stopErrors = append(stopErrors, err)
			}
			continue
		}
		if err := n.runner.Stop(item.key); err != nil {
			stopErrors = append(stopErrors, err)
		}
	}
	return errors.Join(stopErrors...)
}

func (n *Nginx) adoptPool(executable string, port int) (pool, bool) {
	finder, ok := n.checker.(interface{ ListeningPID(int) int })
	if !ok {
		return pool{}, false
	}
	pid := finder.ListeningPID(port)
	if pid <= 0 || !processinfra.MatchesImageName(pid, "php-cgi.exe") {
		return pool{}, false
	}
	if !sameExecutable(executable, processinfra.ExecutablePath(pid)) {
		return pool{}, false
	}
	return pool{key: poolKey(executable), port: port, pid: pid, external: true}, true
}

func (n *Nginx) externalPoolRunning(item pool, executable string) bool {
	adopted, ok := n.adoptPool(executable, item.port)
	return ok && adopted.pid == item.pid
}

func sameExecutable(first, second string) bool {
	if first == "" || second == "" {
		return false
	}
	return strings.EqualFold(filepath.Clean(first), filepath.Clean(second))
}

func (n *Nginx) writeConfig(nginxHome string, port int, routes []ports.Route, pools map[string]pool) error {
	hosts := paths.NginxHosts(n.home)
	folders := []string{
		hosts,
		paths.NginxLog(n.home),
		paths.GatewayLog(n.home, ""),
		paths.NginxTemp(n.home, "client_body_temp"),
		paths.NginxTemp(n.home, "proxy_temp"),
		paths.NginxTemp(n.home, "fastcgi_temp"),
		paths.NginxTemp(n.home, "uwsgi_temp"),
		paths.NginxTemp(n.home, "scgi_temp"),
	}
	for _, folder := range folders {
		if err := os.MkdirAll(folder, 0o755); err != nil {
			return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "创建 Nginx 运行目录失败")
		}
	}
	entries, err := os.ReadDir(hosts)
	if err != nil {
		return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "读取 Nginx 虚拟主机目录失败")
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".conf") {
			if err := os.Remove(filepath.Join(hosts, entry.Name())); err != nil {
				return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "清理旧 Nginx 虚拟主机配置失败")
			}
		}
	}
	fastCGIParameters, err := os.ReadFile(filepath.Join(nginxHome, "conf", "fastcgi.conf"))
	if err != nil {
		return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "读取 Nginx FastCGI 参数失败")
	}
	if err := os.WriteFile(paths.NginxFastCGIConfig(n.home), fastCGIParameters, 0o644); err != nil {
		return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "写入 Nginx FastCGI 参数失败")
	}

	mainConfig := renderMain(nginxHome, n.home, port, routes)
	if err := os.WriteFile(paths.NginxConfig(n.home), []byte(mainConfig), 0o644); err != nil {
		return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "写入 Nginx 主配置失败")
	}
	for _, route := range routes {
		if err := os.MkdirAll(filepath.Dir(route.Log), 0o755); err != nil {
			return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "创建项目日志目录失败")
		}
		fastCGIPort := 0
		if route.Kind == project.KIND_PHP {
			executable := filepath.Clean(strings.TrimSpace(route.PHP))
			item, found := pools[executable]
			if !found {
				return appErrors.New(appErrors.RUNTIME_NOT_INSTALLED, "PHP 项目没有可用的 FastCGI 进程")
			}
			fastCGIPort = item.port
		}
		environmentConfig := ""
		if route.Kind == project.KIND_PHP {
			environmentParameter, environmentErr := renderEnvironmentParameter(route.EnvironmentFile)
			if environmentErr != nil {
				return environmentErr
			}
			if environmentParameter != "" {
				environmentPath := filepath.Join(hosts, paths.NginxEnvironmentConfigName(route.ID, route.Name))
				if writeErr := os.WriteFile(environmentPath, []byte(environmentParameter), 0o600); writeErr != nil {
					return appErrors.Wrap(writeErr, appErrors.PATH_NOT_ACCESSIBLE, "写入 PHP 环境变量配置失败")
				}
				environmentConfig = nginxQuote(nginxPath(environmentPath))
			}
		}
		content, err := renderRoute(port, route, fastCGIPort, environmentConfig)
		if err != nil {
			return err
		}
		name := paths.NginxProjectConfigName(route.ID, route.Name)
		if err := os.WriteFile(filepath.Join(hosts, name), []byte(content), 0o644); err != nil {
			return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "写入 Nginx 项目配置失败")
		}
	}
	return nil
}

func renderMain(nginxHome, home string, port int, routes []ports.Route) string {
	mime := nginxPath(filepath.Join(nginxHome, "conf", "mime.types"))
	errorLog := nginxPath(paths.GatewayLog(home, "error.log"))
	pid := nginxPath(paths.NginxPID(home))
	hosts := nginxPath(filepath.Join(paths.NginxHosts(home), "*.conf"))
	overview := renderOverview(routes)
	return fmt.Sprintf(`worker_processes 1;
error_log %s;
pid %s;

events {
    worker_connections %s;
}

http {
    server_names_hash_max_size 4096;
    server_names_hash_bucket_size 128;
    include %s;
    default_type application/octet-stream;
    log_format devenv '$time_iso8601\t$status\t$body_bytes_sent\t$request_time';
    sendfile on;
    keepalive_timeout %s;
    client_max_body_size %s;
    limit_conn_zone $server_name zone=php:1m;

    map $http_upgrade $connection_upgrade {
        default upgrade;
        '' close;
    }

    server {
        listen %s:%d default_server;
        server_name _;
        default_type text/html;
        return 404 %s;
    }

    include %s;
}
`, nginxQuote(errorLog), nginxQuote(pid), NGINX_CONNECTIONS, nginxQuote(mime), NGINX_KEEPALIVE, NGINX_UPLOAD_LIMIT, paths.LOOPBACK, port, nginxQuote(overview), nginxQuote(hosts))
}

func renderRoute(port int, route ports.Route, fastCGIPort int, environmentConfig string) (string, error) {
	hosts := make([]string, 0, len(route.Hosts))
	for _, host := range route.Hosts {
		clean := strings.TrimSpace(host)
		if clean != "" {
			hosts = append(hosts, nginxQuote(clean))
		}
	}
	if len(hosts) == 0 {
		return "", appErrors.New(appErrors.INVALID_ARGUMENT, "项目域名不能为空")
	}
	root := nginxQuote(nginxPath(route.Path))
	accessLog := nginxQuote(nginxPath(route.Log))
	errorLog := nginxQuote(nginxPath(filepath.Join(filepath.Dir(route.Log), "error.log")))

	var body string
	switch route.Kind {
	case project.KIND_PHP:
		environmentInclude := ""
		if environmentConfig != "" {
			environmentInclude = "        include " + environmentConfig + ";\n"
		}
		body = fmt.Sprintf(`%s

    location ~ ^(.+?\.php)(/.*)?$ {
        try_files $fastcgi_script_name =404;
        include %s;
        fastcgi_split_path_info ^(.+?\.php)(/.*)$;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        fastcgi_param PATH_INFO $fastcgi_path_info;
        fastcgi_pass %s:%d;
%s    }
`, renderPHPRewrite(route.RewriteRule), nginxQuote(paths.NGINX_FASTCGI_FILE), paths.LOOPBACK, fastCGIPort, environmentInclude)
	case project.KIND_NODE, project.KIND_PYTHON, project.KIND_GO, project.KIND_RUST, project.KIND_JAVA:
		body = fmt.Sprintf(`    location / {
        proxy_pass http://%s:%d;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
    }
`, paths.LOOPBACK, route.Port)
	case project.KIND_STATIC:
		body = fmt.Sprintf(`    location / {
        try_files $uri $uri/ =404;
    }
`)
	default:
		return "", appErrors.New(appErrors.INVALID_ARGUMENT, "项目类型无效: "+route.Kind)
	}

	advanced := strings.TrimSpace(route.Nginx)
	if advanced != "" {
		body += "\n    # DevEnv advanced configuration\n" + indentNginx(advanced) + "\n"
	}

	return fmt.Sprintf(`server {
    listen %s:%d;
    server_name %s;
    root %s;
    index index.php index.html index.htm;
    access_log %s devenv;
    error_log %s;

%s}
`, paths.LOOPBACK, port, strings.Join(hosts, " "), root, accessLog, errorLog, body), nil
}

func renderPHPRewrite(rule string) string {
	switch rule {
	case project.REWRITE_THINKPHP:
		return "    location / {\n        try_files $uri $uri/ /index.php?s=$uri&$args;\n    }"
	case project.REWRITE_LARAVEL:
		return "    location / {\n        try_files $uri $uri/ /index.php?$query_string;\n    }"
	case project.REWRITE_WORDPRESS:
		return "    location / {\n        try_files $uri $uri/ /index.php?$args;\n    }"
	case project.REWRITE_TYPECHO:
		return "    location / {\n        try_files $uri $uri/ /index.php$uri?$args;\n    }"
	default:
		return "    location / {\n        try_files $uri $uri/ /index.php$uri$is_args$args;\n    }"
	}
}

func indentNginx(config string) string {
	lines := strings.Split(config, "\n")
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			lines[index] = ""
			continue
		}
		lines[index] = "    " + trimmed
	}
	return strings.Join(lines, "\n")
}

func renderEnvironmentParameter(file string) (string, error) {
	if strings.TrimSpace(file) == "" {
		return "", nil
	}
	fileName := filepath.Base(filepath.Clean(file))
	if fileName == ".env" {
		return "", nil
	}
	if !strings.HasPrefix(fileName, ".env.") {
		return "", appErrors.New(appErrors.INVALID_ARGUMENT, "PHP 环境文件名称必须是 .env 或 .env.{环境名}")
	}
	environmentName := strings.TrimPrefix(fileName, ".env.")
	if environmentName == "" {
		return "", appErrors.New(appErrors.INVALID_ARGUMENT, "PHP 环境文件缺少环境名")
	}
	return "fastcgi_param ENV_NAME " + nginxQuote(environmentName) + ";\n", nil
}

func renderOverview(routes []ports.Route) string {
	var builder strings.Builder
	builder.WriteString("<!doctype html><html lang='zh-CN'><head><meta charset='utf-8'><title>DevEnv</title></head><body><h1>DevEnv 项目列表</h1>")
	if len(routes) == 0 {
		builder.WriteString("<p>当前没有运行中的项目</p>")
	} else {
		builder.WriteString("<ul>")
		for _, route := range routes {
			for _, host := range route.Hosts {
				clean := html.EscapeString(strings.TrimSpace(host))
				builder.WriteString("<li><a href='http://")
				builder.WriteString(clean)
				builder.WriteString("'>")
				builder.WriteString(clean)
				builder.WriteString("</a></li>")
			}
		}
		builder.WriteString("</ul>")
	}
	builder.WriteString("</body></html>")
	return builder.String()
}

func checkNginx(executable, home string) error {
	command := exec.Command(executable, "-p", nginxPrefix(home), "-c", paths.NGINX_CONFIG, "-t")
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := command.CombinedOutput()
	if err != nil {
		return appErrors.Wrap(err, appErrors.PROCESS_START_FAILED, "Nginx 配置检查失败: "+strings.TrimSpace(string(output)))
	}
	return nil
}

func reloadNginx(executable, home string) error {
	command := exec.Command(executable, "-p", nginxPrefix(home), "-c", paths.NGINX_CONFIG, "-s", "reload")
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := command.CombinedOutput()
	if err != nil {
		return appErrors.Wrap(err, appErrors.PROCESS_START_FAILED, "Nginx 热重载失败: "+strings.TrimSpace(string(output)))
	}
	return nil
}

func waitPort(checker ports.PortChecker, port int) error {
	deadline := time.Now().Add(GATEWAY_START_TIMEOUT)
	for time.Now().Before(deadline) {
		if checker.Listening(port) {
			return nil
		}
		time.Sleep(GATEWAY_POLL_INTERVAL)
	}
	return appErrors.New(appErrors.PROCESS_START_FAILED, fmt.Sprintf("端口 %d 未在限定时间内开始监听", port))
}

func copyRoutes(routes []ports.Route) []ports.Route {
	copied := make([]ports.Route, len(routes))
	for index, route := range routes {
		copied[index] = route
		copied[index].Hosts = append(make([]string, 0, len(route.Hosts)), route.Hosts...)
	}
	return copied
}

func nginxPrefix(home string) string {
	return nginxPath(paths.Nginx(home)) + "/"
}

func nginxPath(value string) string {
	return filepath.ToSlash(filepath.Clean(value))
}

func nginxQuote(value string) string {
	escaped := strings.ReplaceAll(value, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	escaped = strings.ReplaceAll(escaped, `$`, `\$`)
	return `"` + escaped + `"`
}

func poolKey(executable string) string {
	digest := sha256.Sum256([]byte(strings.ToLower(filepath.Clean(executable))))
	return "gateway:php:" + hex.EncodeToString(digest[:])
}

func poolPort(executable string) int {
	digest := sha256.Sum256([]byte(strings.ToLower(filepath.Clean(executable))))
	offset := int(binary.BigEndian.Uint16(digest[:2])) % FASTCGI_PORT_COUNT
	return FASTCGI_PORT_START + offset
}
