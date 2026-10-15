package site

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"devenv/internal/domain/software"
)

const (
	KIND_STATIC   = "static"
	KIND_PHP      = "php"
	KIND_NODE     = "node"
	KIND_PYTHON   = "python"
	KIND_GO       = "go"
	KIND_RUST     = "rust"
	KIND_JAVA     = "java"
	KIND_ELECTRON = "electron"
	KIND_MINIPROGRAM = "miniprogram"
)

const (
	STATE_RUNNING = "running"
	STATE_STOPPED = "stopped"
	STATE_ERROR   = "error"
)

type Site struct {
	ID        int64    `json:"id"`
	Name      string   `json:"name"`
	Kind      string   `json:"kind"`
	Hosts     []string `json:"hosts"`
	Root      string   `json:"root"`
	Run       string   `json:"run"`
	Software  *int64   `json:"software"`
	Command   string   `json:"command"`
	Nginx     string   `json:"nginx"`
	Port      int      `json:"port"`
	Autostart bool     `json:"autostart"`
	Created   string   `json:"created"`
	Updated   string   `json:"updated"`
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
	ID        int64    `json:"id"`
	Hosts     []string `json:"hosts"`
	Root      string   `json:"root"`
	Run       string   `json:"run"`
	Software  *int64   `json:"software"`
	Command   string   `json:"command"`
	Nginx     string   `json:"nginx"`
	Autostart bool     `json:"autostart"`
}

func IsValidKind(kind string) bool {
	return kind == KIND_STATIC || kind == KIND_PHP || kind == KIND_NODE ||
		kind == KIND_PYTHON || kind == KIND_GO || kind == KIND_RUST ||
		kind == KIND_JAVA || kind == KIND_ELECTRON || kind == KIND_MINIPROGRAM
}

// IsLocal reports whether a project runs locally and bypasses Nginx.
func IsLocal(kind string) bool {
	return kind == KIND_ELECTRON || kind == KIND_MINIPROGRAM
}

// UsesNginx reports whether a project is published through the local gateway.
func UsesNginx(kind string) bool {
	return IsValidKind(kind) && !IsLocal(kind)
}

// IsProcess reports whether a project starts a development process.
func IsProcess(kind string) bool {
	return kind == KIND_NODE || kind == KIND_PYTHON || kind == KIND_GO ||
		kind == KIND_RUST || kind == KIND_JAVA || kind == KIND_ELECTRON || kind == KIND_MINIPROGRAM
}

// RuntimeSoftwareKind maps a project type to the software catalog kind it uses.
// Rust intentionally returns an empty string because cargo is managed outside
// the software catalog and is resolved from the system PATH.
func RuntimeSoftwareKind(kind string) string {
	switch kind {
	case KIND_PHP:
		return software.KIND_PHP
	case KIND_NODE, KIND_ELECTRON, KIND_MINIPROGRAM:
		return software.KIND_NODE
	case KIND_PYTHON:
		return software.KIND_PYTHON
	case KIND_GO:
		return software.KIND_GO
	case KIND_JAVA:
		return software.KIND_JAVA
	default:
		return ""
	}
}

// ProjectKindsForSoftware returns project types that can use an installed runtime.
func ProjectKindsForSoftware(kind string) []string {
	switch kind {
	case software.KIND_PHP:
		return []string{KIND_PHP}
	case software.KIND_NODE:
		return []string{KIND_NODE, KIND_ELECTRON, KIND_MINIPROGRAM}
	case software.KIND_PYTHON:
		return []string{KIND_PYTHON}
	case software.KIND_GO:
		return []string{KIND_GO}
	case software.KIND_JAVA:
		return []string{KIND_JAVA}
	default:
		return make([]string, 0)
	}
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
	if kind != KIND_NODE && kind != KIND_MINIPROGRAM {
		return ""
	}
	pkgPath := filepath.Join(cleanFolder, "package.json")
	if raw, err := os.ReadFile(pkgPath); err == nil {
		var pkg struct {
			Main    string            `json:"main"`
			Scripts map[string]string `json:"scripts"`
		}
		if json.Unmarshal(raw, &pkg) == nil {
			if kind == KIND_MINIPROGRAM {
				for _, script := range []string{
					"dev:mp-weixin", "dev:mp-alipay", "dev:mp-baidu", "dev:mp-qq",
					"dev:mp-toutiao", "dev:mp-kuaishou", "dev:mp", "dev", "start",
				} {
					if strings.TrimSpace(pkg.Scripts[script]) != "" {
						return packageScriptCommand(cleanFolder, script)
					}
				}
			}
			if strings.TrimSpace(pkg.Scripts["start"]) != "" {
				return "npm start"
			}
			if strings.TrimSpace(pkg.Main) != "" {
				return "node " + strings.TrimSpace(pkg.Main)
			}
		}
	}
	if kind == KIND_MINIPROGRAM {
		return "pnpm dev:mp-weixin"
	}
	for _, entry := range []string{"server.js", "app.js", "index.js"} {
		if fileExists(filepath.Join(cleanFolder, entry)) {
			return "node " + entry
		}
	}
	return ""
}

func packageScriptCommand(folder, script string) string {
	runner := "npm run"
	if fileExists(filepath.Join(folder, "pnpm-lock.yaml")) {
		runner = "pnpm"
	} else if fileExists(filepath.Join(folder, "yarn.lock")) {
		runner = "yarn"
	}
	if runner == "npm run" {
		return runner + " " + script
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
