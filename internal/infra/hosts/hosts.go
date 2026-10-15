package hosts

import (
	"fmt"
	"os"
	"strings"

	appErrors "devenv/internal/shared/errors"
	"devenv/internal/shared/paths"
)

const (
	BEGIN_MARKER = "# BEGIN DevEnv"
	END_MARKER   = "# END DevEnv"
)

type Writer struct {
	path string
}

func NewWriter() *Writer {
	return &Writer{
		path: paths.Hosts(),
	}
}

func (w *Writer) Sync(hosts []string) error {
	var custom []string
	seen := make(map[string]bool)
	for _, host := range hosts {
		clean := strings.ToLower(strings.TrimSpace(host))
		if clean == "" || clean == "localhost" || clean == paths.LOOPBACK || strings.HasSuffix(clean, ".localhost") {
			continue
		}
		if !seen[clean] {
			seen[clean] = true
			custom = append(custom, clean)
		}
	}

	raw, err := os.ReadFile(w.path)
	if err != nil {
		return appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "读取 hosts 文件失败")
	}

	lines := strings.Split(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\n")
	var kept []string
	inside := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == BEGIN_MARKER {
			inside = true
			continue
		}
		if trimmed == END_MARKER {
			inside = false
			continue
		}
		if !inside {
			kept = append(kept, line)
		}
	}

	for len(kept) > 0 && strings.TrimSpace(kept[len(kept)-1]) == "" {
		kept = kept[:len(kept)-1]
	}

	if len(custom) > 0 {
		kept = append(kept, BEGIN_MARKER)
		for _, host := range custom {
			kept = append(kept, fmt.Sprintf("%s %s", paths.LOOPBACK, host))
		}
		kept = append(kept, END_MARKER)
	}

	output := strings.Join(kept, "\r\n") + "\r\n"
	if err := os.WriteFile(w.path, []byte(output), 0o644); err != nil {
		manual := ""
		if len(custom) > 0 {
			manual = fmt.Sprintf("%s %s", paths.LOOPBACK, strings.Join(custom, " "))
		}
		return appErrors.Wrap(
			err,
			appErrors.PERMISSION_DENIED,
			fmt.Sprintf("无权限写入 hosts 文件，请以管理员身份运行或手动添加: %s", manual),
		)
	}
	return nil
}
