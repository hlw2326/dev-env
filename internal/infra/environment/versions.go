package environment

import (
	"fmt"
	"regexp"
	"strings"

	"devenv/internal/domain/software"
)

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func parseVersions(probe software.CommandProbe, output string) ([]string, error) {
	output = strings.TrimSpace(ansiPattern.ReplaceAllString(output, ""))
	if probe.Reject != "" && regexp.MustCompile(probe.Reject).MatchString(output) {
		return nil, fmt.Errorf("%s 入口指向其他软件，不能视为当前软件已安装", probe.Name)
	}
	matches := regexp.MustCompile(probe.Pattern).FindAllStringSubmatch(output, -1)
	versions := make([]string, 0, len(matches))
	seen := make(map[string]bool)
	for _, match := range matches {
		if !seen[match[1]] {
			seen[match[1]] = true
			versions = append(versions, match[1])
		}
	}
	if len(versions) == 0 {
		return versions, fmt.Errorf("%s 未返回可识别的版本（可能缺少运行依赖、SDK 或入口不是对应软件）", probe.Name)
	}
	return versions, nil
}
