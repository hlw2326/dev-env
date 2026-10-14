package hosts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	appErrors "devenv/internal/shared/errors"
)

const (
	beginMarker = "# BEGIN DevEnv"
	endMarker   = "# END DevEnv"
)

type Writer struct {
	path string
}

func NewWriter() *Writer {
	sysRoot := os.Getenv("SystemRoot")
	if sysRoot == "" {
		sysRoot = `C:\Windows`
	}
	return &Writer{
		path: filepath.Join(sysRoot, "System32", "drivers", "etc", "hosts"),
	}
}

func NewWriterWithPath(path string) *Writer {
	return &Writer{path: path}
}

func (w *Writer) Sync(hosts []string) error {
	var custom []string
	seen := make(map[string]bool)
	for _, h := range hosts {
		clean := strings.ToLower(strings.TrimSpace(h))
		if clean == "" || clean == "localhost" || clean == "127.0.0.1" || strings.HasSuffix(clean, ".localhost") {
			continue
		}
		if !seen[clean] {
			seen[clean] = true
			custom = append(custom, clean)
		}
	}

	raw, err := os.ReadFile(w.path)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodePathNotAccessible, "读取 hosts 文件失败")
	}

	lines := strings.Split(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\n")
	var kept []string
	inBlock := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == beginMarker {
			inBlock = true
			continue
		}
		if trimmed == endMarker {
			inBlock = false
			continue
		}
		if !inBlock {
			kept = append(kept, line)
		}
	}

	for len(kept) > 0 && strings.TrimSpace(kept[len(kept)-1]) == "" {
		kept = kept[:len(kept)-1]
	}

	if len(custom) > 0 {
		kept = append(kept, beginMarker)
		for _, h := range custom {
			kept = append(kept, fmt.Sprintf("127.0.0.1 %s", h))
		}
		kept = append(kept, endMarker)
	}

	out := strings.Join(kept, "\r\n") + "\r\n"
	if err := os.WriteFile(w.path, []byte(out), 0644); err != nil {
		manual := ""
		if len(custom) > 0 {
			manual = fmt.Sprintf("127.0.0.1 %s", strings.Join(custom, " "))
		}
		return appErrors.Wrap(
			err,
			appErrors.CodePermissionDenied,
			fmt.Sprintf("无权限写入 hosts 文件，请以管理员身份运行或手动添加: %s", manual),
		)
	}
	return nil
}
