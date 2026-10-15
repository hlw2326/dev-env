package git

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type versionLocation struct {
	path  string
	value string
	start int
	end   int
}

var versionPattern = regexp.MustCompile(`(?i)^v?([0-9]+)\.([0-9]+)\.([0-9]+)$`)

var jsonVersionValuePattern = regexp.MustCompile(`^"([^"\\]+)"$`)
var tomlVersionPattern = regexp.MustCompile(`(?m)^(\s*version\s*=\s*["'])([^"']+)(["']\s*(?:#.*)?)$`)
var propertyVersionPattern = regexp.MustCompile(`(?m)^(\s*version\s*=\s*)([^\s#]+)(\s*(?:#.*)?)$`)
var gradleVersionPattern = regexp.MustCompile(`(?m)^(\s*version\s*(?:=\s*)?['"])([^'"]+)(['"]\s*(?:\/\/.*)?)$`)
var plainVersionPattern = regexp.MustCompile(`(?m)^(\s*)(v?[0-9]+\.[0-9]+\.[0-9]+)(\s*)$`)

func versionCandidates(kind string) []string {
	switch kind {
	case "php":
		return []string{"VERSION", "composer.json"}
	case "node", "electron", "miniprogram":
		return []string{"VERSION", "package.json"}
	case "python":
		return []string{"VERSION", "pyproject.toml", "setup.cfg"}
	case "go":
		return []string{"VERSION"}
	case "rust":
		return []string{"VERSION", "Cargo.toml"}
	case "java":
		return []string{"VERSION", "pom.xml", "gradle.properties", "build.gradle", "build.gradle.kts"}
	default:
		return []string{"VERSION", "package.json", "composer.json"}
	}
}

func readVersion(root, kind string) (versionLocation, string, string, error) {
	var firstPresent string
	for _, name := range versionCandidates(kind) {
		path := filepath.Join(root, name)
		info, statError := os.Lstat(path)
		if statError == nil && info.Mode()&os.ModeSymlink != 0 {
			return versionLocation{}, "", "版本文件是符号链接，请手动维护版本号", nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return versionLocation{}, "", "", fmt.Errorf("读取 %s 失败: %w", name, err)
		}
		if firstPresent == "" {
			firstPresent = name
		}
		start, end, value, ok := findVersionValue(name, content)
		if !ok {
			continue
		}
		next, err := nextVersion(value)
		if err != nil {
			return versionLocation{path: name, value: value, start: start, end: end}, "", err.Error(), nil
		}
		return versionLocation{path: name, value: value, start: start, end: end}, next, "", nil
	}
	if firstPresent != "" {
		return versionLocation{path: firstPresent}, "", "未找到可识别的三段式版本号", nil
	}
	return versionLocation{}, "", "未找到可自动升级的版本文件（支持 VERSION、package.json、composer.json、Cargo.toml、pyproject.toml、pom.xml 和 Gradle 配置）", nil
}

func findVersionValue(name string, content []byte) (int, int, string, bool) {
	text := string(content)
	switch name {
	case "package.json", "composer.json":
		return jsonVersionValue(text)
	case "Cargo.toml":
		return sectionCapture(text, "package", tomlVersionPattern)
	case "pyproject.toml":
		return sectionCapture(text, "project", tomlVersionPattern)
	case "setup.cfg":
		return sectionCapture(text, "metadata", propertyVersionPattern)
	case "gradle.properties":
		return firstCapture(propertyVersionPattern, text, 2)
	case "build.gradle", "build.gradle.kts":
		return firstCapture(gradleVersionPattern, text, 2)
	case "pom.xml":
		return xmlCapture(text)
	case "VERSION":
		return firstCapture(plainVersionPattern, text, 2)
	default:
		return 0, 0, "", false
	}
}

func jsonVersionValue(content string) (int, int, string, bool) {
	decoder := json.NewDecoder(strings.NewReader(content))
	opening, err := decoder.Token()
	if err != nil || opening != json.Delim('{') {
		return 0, 0, "", false
	}
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return 0, 0, "", false
		}
		key, ok := keyToken.(string)
		if !ok {
			return 0, 0, "", false
		}
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			return 0, 0, "", false
		}
		if key != "version" {
			continue
		}
		valueEnd := int(decoder.InputOffset())
		valueStart := valueEnd - len(raw)
		segment := content[valueStart:valueEnd]
		trimmed := strings.TrimLeft(segment, " \r\n\t")
		leading := len(segment) - len(trimmed)
		match := jsonVersionValuePattern.FindStringSubmatchIndex(trimmed)
		if match == nil {
			return 0, 0, "", false
		}
		var value string
		if err := json.Unmarshal(raw, &value); err != nil || value != trimmed[match[2]:match[3]] {
			return 0, 0, "", false
		}
		return valueStart + leading + match[2], valueStart + leading + match[3], value, true
	}
	return 0, 0, "", false
}

func sectionCapture(content, section string, pattern *regexp.Regexp) (int, int, string, bool) {
	current := ""
	for _, line := range lineRanges(content) {
		trimmed := strings.TrimSpace(content[line.start:line.end])
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			current = strings.TrimSpace(trimmed[1 : len(trimmed)-1])
			continue
		}
		if current != section {
			continue
		}
		if index := pattern.FindStringSubmatchIndex(content[line.start:line.end]); index != nil {
			start := line.start + index[4]
			end := line.start + index[5]
			return start, end, content[start:end], true
		}
	}
	return 0, 0, "", false
}

type sourceLine struct {
	start int
	end   int
}

func lineRanges(content string) []sourceLine {
	lines := make([]sourceLine, 0)
	start := 0
	for index, char := range content {
		if char == '\n' {
			end := index
			if end > start && content[end-1] == '\r' {
				end--
			}
			lines = append(lines, sourceLine{start: start, end: end})
			start = index + 1
		}
	}
	if start <= len(content) {
		lines = append(lines, sourceLine{start: start, end: len(content)})
	}
	return lines
}

func xmlCapture(content string) (int, int, string, bool) {
	decoder := xml.NewDecoder(strings.NewReader(content))
	stack := make([]string, 0)
	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}
		switch item := token.(type) {
		case xml.StartElement:
			if len(stack) == 1 && stack[0] == "project" && item.Name.Local == "version" {
				start := int(decoder.InputOffset())
				var value string
				if decoder.DecodeElement(&value, &item) != nil {
					return 0, 0, "", false
				}
				value = strings.TrimSpace(value)
				segment := content[start:]
				start += len(segment) - len(strings.TrimLeft(segment, " \t\r\n"))
				if strings.HasPrefix(content[start:], value+"</version>") || (value != "" && strings.HasPrefix(content[start:], value)) {
					return start, start + len(value), value, true
				}
				return 0, 0, "", false
			}
			stack = append(stack, item.Name.Local)
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		}
	}
	return 0, 0, "", false
}

func firstCapture(pattern *regexp.Regexp, content string, capture int) (int, int, string, bool) {
	index := pattern.FindStringSubmatchIndex(content)
	if index == nil {
		return 0, 0, "", false
	}
	start := index[capture*2]
	end := index[capture*2+1]
	return start, end, content[start:end], true
}

func nextVersion(current string) (string, error) {
	match := versionPattern.FindStringSubmatch(current)
	if match == nil {
		return "", fmt.Errorf("版本号 %q 不是 major.minor.patch 三段式格式", current)
	}
	major, err := strconv.Atoi(match[1])
	if err != nil {
		return "", fmt.Errorf("解析主版本号失败: %w", err)
	}
	minor, err := strconv.Atoi(match[2])
	if err != nil {
		return "", fmt.Errorf("解析次版本号失败: %w", err)
	}
	patch, err := strconv.Atoi(match[3])
	if err != nil {
		return "", fmt.Errorf("解析补丁版本号失败: %w", err)
	}
	if patch > 99 || minor > 99 {
		return "", fmt.Errorf("版本号 %q 超出补丁号和次版本号 0–99 的递增范围", current)
	}
	if patch < 99 {
		patch++
	} else {
		patch = 0
		if minor < 99 {
			minor++
		} else {
			major++
			minor = 0
		}
	}
	if patch == 0 {
		return versionPrefix(current) + fmt.Sprintf("%d.%d.0", major, minor), nil
	}
	return versionPrefix(current) + fmt.Sprintf("%d.%d.%02d", major, minor, patch), nil
}

func versionPrefix(value string) string {
	if strings.HasPrefix(value, "v") || strings.HasPrefix(value, "V") {
		return value[:1]
	}
	return ""
}

func writeNextVersion(root string, location versionLocation, next string) ([]byte, error) {
	path := filepath.Join(root, location.path)
	linkInfo, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if linkInfo.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("不能自动修改符号链接版本文件")
	}
	original, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取版本文件失败: %w", err)
	}
	if location.start < 0 || location.end > len(original) || location.start > location.end || string(original[location.start:location.end]) != location.value {
		return nil, fmt.Errorf("版本文件位置已变化，请刷新 Git 状态后重试")
	}
	updated := make([]byte, 0, len(original)-location.end+location.start+len(next))
	updated = append(updated, original[:location.start]...)
	updated = append(updated, next...)
	updated = append(updated, original[location.end:]...)
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("读取版本文件权限失败: %w", err)
	}
	if err := os.WriteFile(path, updated, info.Mode().Perm()); err != nil {
		return nil, fmt.Errorf("写入新版本号失败: %w", err)
	}
	return original, nil
}
