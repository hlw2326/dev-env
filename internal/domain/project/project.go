package project

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	KIND_STATIC      = "static"
	KIND_PHP         = "php"
	KIND_NODE        = "node"
	KIND_PYTHON      = "python"
	KIND_GO          = "go"
	KIND_RUST        = "rust"
	KIND_JAVA        = "java"
	KIND_ELECTRON    = "electron"
	KIND_MINIPROGRAM = "miniprogram"
)

const (
	REWRITE_DEFAULT   = "default"
	REWRITE_THINKPHP  = "thinkphp"
	REWRITE_LARAVEL   = "laravel"
	REWRITE_WORDPRESS = "wordpress"
	REWRITE_TYPECHO   = "typecho"
)

const (
	STATE_STARTING = "starting"
	STATE_RUNNING  = "running"
	STATE_STOPPED  = "stopped"
	STATE_ERROR    = "error"
)

type Project struct {
	ID              int64    `json:"id"`
	Name            string   `json:"name"`
	Kind            string   `json:"kind"`
	Hosts           []string `json:"hosts"`
	Root            string   `json:"root"`
	Run             string   `json:"run"`
	Software        *int64   `json:"software"`
	Command         string   `json:"command"`
	Nginx           string   `json:"nginx"`
	RewriteRule     string   `json:"rewriteRule"`
	EnvironmentFile string   `json:"environmentFile"`
	Port            int      `json:"port"`
	Autostart       bool     `json:"autostart"`
	Created         string   `json:"created"`
	Updated         string   `json:"updated"`
}

type CreateInput struct {
	Kind    string `json:"kind"`
	Name    string `json:"name"`
	Host    string `json:"host"`
	Root    string `json:"root"`
	Run     string `json:"run"`
	Command string `json:"command"`
}

type UpdateInput struct {
	ID              int64    `json:"id"`
	Hosts           []string `json:"hosts"`
	Root            string   `json:"root"`
	Run             string   `json:"run"`
	Software        *int64   `json:"software"`
	Command         string   `json:"command"`
	Nginx           string   `json:"nginx"`
	RewriteRule     string   `json:"rewriteRule"`
	EnvironmentFile string   `json:"environmentFile"`
	Autostart       bool     `json:"autostart"`
}

// RewriteRules returns the built-in PHP pseudo-static rules. The generated
// route owns these rules so they cannot be overwritten by advanced settings.
func RewriteRules() []string {
	return []string{REWRITE_DEFAULT, REWRITE_THINKPHP, REWRITE_LARAVEL, REWRITE_WORDPRESS, REWRITE_TYPECHO}
}

func IsValidRewriteRule(value string) bool {
	for _, item := range RewriteRules() {
		if value == item {
			return true
		}
	}
	return false
}

func IsValidKind(kind string) bool {
	_, ok := ProjectTemplateForKind(kind)
	return ok
}

// IsLocal reports whether a project runs locally and bypasses Nginx.
func IsLocal(kind string) bool {
	return IsValidKind(kind) && !HasCapability(kind, CAPABILITY_DOMAIN)
}

// UsesNginx reports whether a project is published through the local gateway.
func UsesNginx(kind string) bool {
	return HasCapability(kind, CAPABILITY_NGINX)
}

// IsProcess reports whether a project starts a development process.
func IsProcess(kind string) bool {
	return HasCapability(kind, CAPABILITY_COMMAND)
}

// RuntimeSoftwareKind maps a project type to the software catalog kind it uses.
// Rust intentionally returns an empty string because cargo is managed outside
// the software catalog and is resolved from the system PATH.
func RuntimeSoftwareKind(kind string) string {
	template, ok := ProjectTemplateForKind(kind)
	if !ok {
		return ""
	}
	return template.RuntimeKind
}

// ProjectKindsForSoftware returns project types that can use an installed runtime.
func ProjectKindsForSoftware(kind string) []string {
	result := make([]string, 0)
	for _, template := range ProjectTemplates() {
		if template.RuntimeKind == kind {
			result = append(result, template.Kind)
		}
	}
	return result
}

func ValidateNginxConfig(config string) error {
	depth := 0
	for _, line := range strings.Split(config, "\n") {
		trimmed := strings.TrimSpace(strings.ToLower(line))
		for _, block := range []string{"server", "http", "events"} {
			if strings.HasPrefix(trimmed, block+" ") || strings.HasPrefix(trimmed, block+"{") {
				return errors.New("nginx 高级配置不能包含 " + block + " 区块")
			}
		}
		depth += strings.Count(line, "{")
		depth -= strings.Count(line, "}")
		if depth < 0 {
			return errors.New("nginx 高级配置的花括号不匹配")
		}
	}
	if depth != 0 {
		return errors.New("nginx 高级配置的花括号不匹配")
	}
	return nil
}

func NormalizeRun(root, run string) (string, error) {
	trimmed := strings.TrimSpace(run)
	if trimmed == "" || trimmed == "/" || trimmed == `\` {
		return "", nil
	}
	trimmed = strings.TrimPrefix(trimmed, "/")
	trimmed = strings.TrimPrefix(trimmed, `\`)
	if trimmed == "" {
		return "", nil
	}
	if filepath.IsAbs(trimmed) || strings.HasPrefix(trimmed, "/") || strings.HasPrefix(trimmed, `\`) {
		return "", errors.New("运行目录不能是绝对路径")
	}
	cleaned := filepath.ToSlash(filepath.Clean(trimmed))
	if cleaned == "." {
		return "", nil
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.Contains(cleaned, "/../") {
		return "", errors.New("运行目录不能超出根目录")
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	absTarget, err := filepath.Abs(filepath.Join(absRoot, filepath.FromSlash(cleaned)))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(absRoot, absTarget)
	if err != nil {
		return "", err
	}
	relSlash := filepath.ToSlash(rel)
	if relSlash == ".." || strings.HasPrefix(relSlash, "../") {
		return "", errors.New("运行目录不能超出根目录")
	}
	return cleaned, nil
}

func ResolvePath(root, run string) string {
	if run == "" {
		return filepath.Clean(root)
	}
	return filepath.Join(filepath.Clean(root), filepath.FromSlash(run))
}

// FindEnvironmentFiles returns environment files directly under the project root.
func FindEnvironmentFiles(root string) []string {
	entries, err := os.ReadDir(filepath.Clean(root))
	if err != nil {
		return []string{}
	}
	files := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), ".env") {
			continue
		}
		if entry.Name() == ".env.example" || entry.Name() == ".env.sample" {
			continue
		}
		files = append(files, entry.Name())
	}
	sort.Strings(files)
	return files
}

// DefaultEnvironmentFile prefers development settings for the local gateway.
func DefaultEnvironmentFile(root string) string {
	files := FindEnvironmentFiles(root)
	for _, preferred := range []string{".env.dev", ".env.development", ".env.local", ".env"} {
		for _, file := range files {
			if file == preferred {
				return file
			}
		}
	}
	return ""
}

// NormalizeEnvironmentFile validates a selected environment file and returns a relative path.
func NormalizeEnvironmentFile(root, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", nil
	}
	if filepath.IsAbs(trimmed) || strings.HasPrefix(trimmed, "/") || strings.HasPrefix(trimmed, `\`) {
		return "", errors.New("环境变量文件必须位于项目根目录内")
	}
	cleaned := filepath.ToSlash(filepath.Clean(trimmed))
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.Contains(cleaned, "/../") {
		return "", errors.New("环境变量文件不能超出项目根目录")
	}
	if !strings.HasPrefix(cleaned, ".env") {
		return "", errors.New("环境变量文件必须以 .env 开头")
	}
	target := filepath.Join(filepath.Clean(root), filepath.FromSlash(cleaned))
	info, err := os.Stat(target)
	if err != nil {
		return "", errors.New("环境变量文件不存在: " + cleaned)
	}
	if info.IsDir() {
		return "", errors.New("环境变量文件不能是目录")
	}
	return cleaned, nil
}

// ResolveEnvironmentFile returns the absolute file used by the local gateway.
func ResolveEnvironmentFile(root, value string) (string, error) {
	selected := strings.TrimSpace(value)
	if selected == "" {
		selected = DefaultEnvironmentFile(root)
	}
	if selected == "" {
		return "", nil
	}
	cleaned, err := NormalizeEnvironmentFile(root, selected)
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Clean(root), filepath.FromSlash(cleaned)), nil
}

func FindRun(root, kind string) string {
	cleanRoot := filepath.Clean(strings.TrimSpace(root))
	if cleanRoot == "" {
		return ""
	}
	switch kind {
	case KIND_PHP:
		if fileExists(filepath.Join(cleanRoot, "public", "index.php")) {
			return "public"
		}
		if fileExists(filepath.Join(cleanRoot, "web", "index.php")) {
			return "web"
		}
		if fileExists(filepath.Join(cleanRoot, "index.php")) {
			return ""
		}
		return ""
	case KIND_STATIC:
		if fileExists(filepath.Join(cleanRoot, "index.html")) {
			return ""
		}
		if fileExists(filepath.Join(cleanRoot, "dist", "index.html")) {
			return "dist"
		}
		if fileExists(filepath.Join(cleanRoot, "public", "index.html")) {
			return "public"
		}
		return ""
	default:
		return ""
	}
}

func FindCommand(folder, kind string) string {
	cleanFolder := filepath.Clean(strings.TrimSpace(folder))
	if cleanFolder == "" {
		return ""
	}
	template, hasTemplate := ProjectTemplateForKind(kind)
	if hasTemplate && template.CommandMode == COMMAND_MODE_PACKAGE_SCRIPT {
		commands := FindCommands(cleanFolder, kind)
		if len(commands) > 0 {
			return commands[0]
		}
		return "pnpm dev:mp-weixin"
	}
	switch kind {
	case KIND_PYTHON:
		if fileExists(filepath.Join(cleanFolder, "manage.py")) {
			return "python manage.py runserver 127.0.0.1:%PORT%"
		}
		for _, entry := range []string{"app.py", "main.py", "server.py"} {
			if fileExists(filepath.Join(cleanFolder, entry)) {
				return "python " + entry
			}
		}
	case KIND_GO:
		if fileExists(filepath.Join(cleanFolder, "go.mod")) || fileExists(filepath.Join(cleanFolder, "main.go")) {
			return "go run ."
		}
	case KIND_RUST:
		if fileExists(filepath.Join(cleanFolder, "Cargo.toml")) {
			return "cargo run"
		}
	case KIND_JAVA:
		if fileExists(filepath.Join(cleanFolder, "pom.xml")) {
			return "mvn spring-boot:run"
		}
		if fileExists(filepath.Join(cleanFolder, "build.gradle")) || fileExists(filepath.Join(cleanFolder, "build.gradle.kts")) {
			if fileExists(filepath.Join(cleanFolder, "gradlew.bat")) {
				return "gradlew.bat bootRun"
			}
			return "gradle bootRun"
		}
		if jar := findJar(cleanFolder); jar != "" {
			return "java -jar " + jar
		}
		if fileExists(filepath.Join(cleanFolder, "Main.java")) {
			return "java Main.java"
		}
	}
	if kind != KIND_NODE {
		return ""
	}
	pkgPath := filepath.Join(cleanFolder, "package.json")
	if raw, err := os.ReadFile(pkgPath); err == nil {
		var pkg struct {
			Main    string            `json:"main"`
			Scripts map[string]string `json:"scripts"`
		}
		if json.Unmarshal(raw, &pkg) == nil {
			if strings.TrimSpace(pkg.Scripts["start"]) != "" {
				return "npm start"
			}
			if strings.TrimSpace(pkg.Main) != "" {
				return "node " + strings.TrimSpace(pkg.Main)
			}
		}
	}
	for _, entry := range []string{"server.js", "app.js", "index.js"} {
		if fileExists(filepath.Join(cleanFolder, entry)) {
			return "node " + entry
		}
	}
	return ""
}

// FindCommands returns the runnable package scripts for a mini-program project.
// The command keeps the package manager used by the project lockfile.
func FindCommands(folder, kind string) []string {
	template, ok := ProjectTemplateForKind(kind)
	if !ok || template.CommandMode != COMMAND_MODE_PACKAGE_SCRIPT {
		return []string{}
	}
	cleanFolder := filepath.Clean(strings.TrimSpace(folder))
	if cleanFolder == "" {
		return []string{}
	}
	raw, err := os.ReadFile(filepath.Join(cleanFolder, "package.json"))
	if err != nil {
		return []string{}
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if json.Unmarshal(raw, &pkg) != nil || len(pkg.Scripts) == 0 {
		return []string{}
	}

	ordered := []string{
		"dev:mp-weixin", "dev:mp-alipay", "dev:mp-baidu", "dev:mp-qq",
		"dev:mp-toutiao", "dev:mp-kuaishou", "dev:mp", "dev", "start", "serve", "preview",
	}
	seen := make(map[string]bool)
	scripts := make([]string, 0, len(pkg.Scripts))
	appendScript := func(name string) {
		if seen[name] || strings.TrimSpace(pkg.Scripts[name]) == "" {
			return
		}
		seen[name] = true
		scripts = append(scripts, packageScriptCommand(cleanFolder, name))
	}
	for _, name := range ordered {
		appendScript(name)
	}
	extra := make([]string, 0)
	for name := range pkg.Scripts {
		extra = append(extra, name)
	}
	sort.Strings(extra)
	for _, name := range extra {
		appendScript(name)
	}
	return scripts
}

func packageScriptCommand(folder, script string) string {
	runner := "npm run"
	if fileExists(filepath.Join(folder, "pnpm-lock.yaml")) {
		runner = "pnpm"
	} else if fileExists(filepath.Join(folder, "yarn.lock")) {
		runner = "yarn"
	}
	return runner + " " + script
}

func findJar(folder string) string {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return ""
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".jar") {
			return entry.Name()
		}
	}
	for _, directory := range []string{"target", filepath.Join("build", "libs")} {
		entries, err := os.ReadDir(filepath.Join(folder, directory))
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".jar") {
				return filepath.ToSlash(filepath.Join(directory, entry.Name()))
			}
		}
	}
	return ""
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
