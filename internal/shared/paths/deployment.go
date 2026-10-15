package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func DeploymentCache(base string) string { return filepath.Join(base, TEMP_FOLDER, "deployment") }

func DeploymentSource(root, source string) (string, error) {
	if !filepath.IsAbs(root) || filepath.IsAbs(source) || strings.Contains(source, ":") {
		return "", fmt.Errorf("发布源目录必须是项目内相对路径")
	}
	for _, part := range strings.FieldsFunc(source, func(r rune) bool { return r == '/' || r == '\\' }) {
		if part == ".." {
			return "", fmt.Errorf("发布目录不能包含 ..")
		}
	}
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(filepath.Join(resolvedRoot, source))
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(resolvedRoot, resolved)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		return "", fmt.Errorf("发布源目录不能越出项目根目录")
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("发布源必须是目录")
	}
	return resolved, nil
}
