package site

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	KindStatic = "static"
	KindPHP    = "php"
	KindNode   = "node"
)

const (
	StateRunning = "running"
	StateStopped = "stopped"
	StateError   = "error"
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
	Autostart bool     `json:"autostart"`
}

func IsValidKind(kind string) bool {
	return kind == KindStatic || kind == KindPHP || kind == KindNode
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
	case KindPHP:
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
	case KindStatic:
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

func FindCommand(dir string) string {
	cleanDir := filepath.Clean(strings.TrimSpace(dir))
	if cleanDir == "" {
		return ""
	}
	pkgPath := filepath.Join(cleanDir, "package.json")
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
		if fileExists(filepath.Join(cleanDir, entry)) {
			return "node " + entry
		}
	}
	return ""
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
