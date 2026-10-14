package installer

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"devenv/internal/domain/software"
	"devenv/internal/ports"
	appErrors "devenv/internal/shared/errors"
)

type Manager struct {
	http *http.Client
}

func NewManager() *Manager {
	return &Manager{
		http: &http.Client{Timeout: 30 * time.Minute},
	}
}

func NewManagerWithClient(client *http.Client) *Manager {
	return &Manager{http: client}
}

func (m *Manager) Install(ctx context.Context, request ports.Request, report func(progress int, message string)) error {
	partPath := request.Cache
	tmpDir := request.Target + ".tmp"

	cleanup := func() {
		if partPath != "" {
			_ = os.Remove(partPath)
		}
		if tmpDir != "" {
			_ = os.RemoveAll(tmpDir)
		}
	}

	if err := os.RemoveAll(tmpDir); err != nil {
		return appErrors.Wrap(err, appErrors.CodeInstallFailed, "清理临时解压目录失败")
	}
	if err := os.MkdirAll(filepath.Dir(partPath), 0755); err != nil {
		return appErrors.Wrap(err, appErrors.CodeInstallFailed, "创建下载缓存目录失败")
	}
	if err := os.MkdirAll(filepath.Dir(request.Target), 0755); err != nil {
		return appErrors.Wrap(err, appErrors.CodeInstallFailed, "创建软件安装父目录失败")
	}

	report(2, "正在下载安装包...")
	computedHash, err := m.downloadFile(ctx, request.URL, partPath, report)
	if err != nil {
		cleanup()
		return err
	}

	requiresChecksum := request.Kind == software.KindPHP || request.Kind == software.KindNode
	if requiresChecksum && strings.TrimSpace(request.Checksum) == "" {
		cleanup()
		return appErrors.New(appErrors.CodeDownloadFailed, "缺少安装包 SHA-256 校验值")
	}
	if strings.TrimSpace(request.Checksum) != "" {
		expected := strings.ToLower(strings.TrimSpace(request.Checksum))
		if !strings.EqualFold(computedHash, expected) {
			cleanup()
			return appErrors.New(appErrors.CodeDownloadFailed, fmt.Sprintf("安装包 SHA-256 校验失败: 期望 %s，实际 %s", expected, computedHash))
		}
	}

	report(72, "正在解压安装包...")
	if err := extractZipSafe(ctx, partPath, tmpDir, report); err != nil {
		cleanup()
		return err
	}

	if err := verifyEntryBinary(request.Kind, tmpDir); err != nil {
		cleanup()
		return err
	}

	report(92, "正在初始化配置...")
	if err := postProcess(request.Kind, tmpDir, request.Target); err != nil {
		cleanup()
		return err
	}

	if err := os.RemoveAll(request.Target); err != nil {
		cleanup()
		return appErrors.Wrap(err, appErrors.CodeInstallFailed, "清理旧目标目录失败")
	}
	if err := os.Rename(tmpDir, request.Target); err != nil {
		cleanup()
		return appErrors.Wrap(err, appErrors.CodeInstallFailed, "启用安装目录失败")
	}

	_ = os.Remove(partPath)
	report(100, "安装完成")
	return nil
}

func (m *Manager) downloadFile(ctx context.Context, url, partPath string, report func(progress int, message string)) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", appErrors.Wrap(err, appErrors.CodeDownloadFailed, "创建下载请求失败")
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) DevEnv/2.0")

	resp, err := m.http.Do(req)
	if err != nil {
		return "", appErrors.Wrap(err, appErrors.CodeDownloadFailed, "下载请求失败")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", appErrors.New(appErrors.CodeDownloadFailed, fmt.Sprintf("下载失败，HTTP 状态码 %d", resp.StatusCode))
	}

	out, err := os.Create(partPath)
	if err != nil {
		return "", appErrors.Wrap(err, appErrors.CodeDownloadFailed, "创建缓存文件失败")
	}
	defer out.Close()

	hasher := sha256.New()
	total := resp.ContentLength
	var downloaded int64
	lastReport := time.Now()
	buf := make([]byte, 64*1024)

	for {
		if err := ctx.Err(); err != nil {
			_ = out.Close()
			return "", appErrors.Wrap(err, appErrors.CodeDownloadFailed, "下载已取消")
		}
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				_ = out.Close()
				return "", appErrors.Wrap(writeErr, appErrors.CodeDownloadFailed, "写入缓存文件失败")
			}
			_, _ = hasher.Write(buf[:n])
			downloaded += int64(n)

			if time.Since(lastReport) >= 200*time.Millisecond {
				lastReport = time.Now()
				pct := 35
				var msg string
				if total > 0 {
					pct = int(float64(downloaded) / float64(total) * 70)
					if pct > 70 {
						pct = 70
					}
					msg = fmt.Sprintf("正在下载安装包 (%s / %s)...", formatBytes(downloaded), formatBytes(total))
				} else {
					msg = fmt.Sprintf("正在下载安装包 (%s)...", formatBytes(downloaded))
				}
				report(pct, msg)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			_ = out.Close()
			return "", appErrors.Wrap(readErr, appErrors.CodeDownloadFailed, "读取下载流失败")
		}
	}

	if err := out.Close(); err != nil {
		return "", appErrors.Wrap(err, appErrors.CodeDownloadFailed, "关闭缓存文件失败")
	}
	report(70, "下载完成，正在准备解压...")
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func extractZipSafe(ctx context.Context, zipPath, destDir string, report func(progress int, message string)) error {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeInstallFailed, "打开压缩包失败")
	}
	defer zr.Close()

	if len(zr.File) == 0 {
		return appErrors.New(appErrors.CodeInstallFailed, "压缩包为空")
	}

	commonPrefix := detectSingleRootPrefix(zr.File)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return appErrors.Wrap(err, appErrors.CodeInstallFailed, "创建解压目录失败")
	}
	absDest, err := filepath.Abs(destDir)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeInstallFailed, "解析解压目录路径失败")
	}

	total := len(zr.File)
	for i, f := range zr.File {
		if err := ctx.Err(); err != nil {
			return appErrors.Wrap(err, appErrors.CodeInstallFailed, "解压已取消")
		}
		relName := strings.ReplaceAll(f.Name, `\`, "/")
		if commonPrefix != "" {
			relName = strings.TrimPrefix(relName, commonPrefix)
		}
		relName = strings.TrimPrefix(relName, "/")
		if relName == "" {
			continue
		}

		// Check for zip-slip before joining
		cleanedSlash := filepath.ToSlash(filepath.Clean(relName))
		if cleanedSlash == ".." || strings.HasPrefix(cleanedSlash, "../") || strings.Contains(cleanedSlash, "/../") || filepath.IsAbs(relName) {
			return appErrors.New(appErrors.CodeInstallFailed, "检测到非法压缩包路径: "+f.Name)
		}

		targetPath := filepath.Join(absDest, filepath.FromSlash(cleanedSlash))
		relCheck, relErr := filepath.Rel(absDest, targetPath)
		if relErr != nil || strings.HasPrefix(filepath.ToSlash(relCheck), "../") || relCheck == ".." {
			return appErrors.New(appErrors.CodeInstallFailed, "检测到越界压缩包路径: "+f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return appErrors.Wrap(err, appErrors.CodeInstallFailed, "创建目录失败")
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return appErrors.Wrap(err, appErrors.CodeInstallFailed, "创建父目录失败")
		}

		rc, err := f.Open()
		if err != nil {
			return appErrors.Wrap(err, appErrors.CodeInstallFailed, "读取压缩包条目失败")
		}
		outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			_ = rc.Close()
			return appErrors.Wrap(err, appErrors.CodeInstallFailed, "创建目标文件失败")
		}
		_, copyErr := io.Copy(outFile, rc)
		_ = outFile.Close()
		_ = rc.Close()
		if copyErr != nil {
			return appErrors.Wrap(copyErr, appErrors.CodeInstallFailed, "解压写入文件失败")
		}

		if total > 0 && i%25 == 0 {
			pct := 70 + int(float64(i+1)/float64(total)*20)
			if pct > 90 {
				pct = 90
			}
			report(pct, "正在解压安装包...")
		}
	}
	return nil
}

func detectSingleRootPrefix(files []*zip.File) string {
	var root string
	for _, f := range files {
		name := strings.TrimPrefix(strings.ReplaceAll(f.Name, `\`, "/"), "/")
		if name == "" {
			continue
		}
		idx := strings.IndexByte(name, '/')
		if idx <= 0 {
			return ""
		}
		top := name[:idx+1]
		if root == "" {
			root = top
		} else if root != top {
			return ""
		}
	}
	return root
}

func verifyEntryBinary(kind, dir string) error {
	switch kind {
	case software.KindPHP:
		if !fileExists(filepath.Join(dir, "php-cgi.exe")) {
			return appErrors.New(appErrors.CodeInstallFailed, "安装包中未找到 php-cgi.exe")
		}
	case software.KindNode:
		if !fileExists(filepath.Join(dir, "node.exe")) {
			return appErrors.New(appErrors.CodeInstallFailed, "安装包中未找到 node.exe")
		}
	case software.KindMySQL:
		if !fileExists(filepath.Join(dir, "bin", "mysqld.exe")) {
			return appErrors.New(appErrors.CodeInstallFailed, "安装包中未找到 bin\\mysqld.exe")
		}
	case software.KindRedis:
		if FindRedisBinary(dir) == "" {
			return appErrors.New(appErrors.CodeInstallFailed, "安装包中未找到 redis-server.exe")
		}
	case software.KindNginx:
		if !fileExists(filepath.Join(dir, "nginx.exe")) {
			return appErrors.New(appErrors.CodeInstallFailed, "安装包中未找到 nginx.exe")
		}
	case software.KindApache:
		if FindApacheBinary(dir) == "" {
			return appErrors.New(appErrors.CodeInstallFailed, "安装包中未找到 httpd.exe")
		}
	}
	return nil
}

func FindApacheBinary(dir string) string {
	direct := filepath.Join(dir, "bin", "httpd.exe")
	if fileExists(direct) {
		return direct
	}
	nested := filepath.Join(dir, "Apache24", "bin", "httpd.exe")
	if fileExists(nested) {
		return nested
	}
	var found string
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.EqualFold(d.Name(), "httpd.exe") {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

func FindRedisBinary(dir string) string {
	direct := filepath.Join(dir, "redis-server.exe")
	if fileExists(direct) {
		return direct
	}
	var found string
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.EqualFold(d.Name(), "redis-server.exe") {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

func postProcess(kind, tmpDir, finalTarget string) error {
	if kind == software.KindApache {
		apacheSubdir := filepath.Join(tmpDir, "Apache24")
		if info, err := os.Stat(apacheSubdir); err == nil && info.IsDir() {
			entries, readErr := os.ReadDir(apacheSubdir)
			if readErr == nil {
				for _, ent := range entries {
					src := filepath.Join(apacheSubdir, ent.Name())
					dst := filepath.Join(tmpDir, ent.Name())
					_ = os.RemoveAll(dst)
					_ = os.Rename(src, dst)
				}
				_ = os.Remove(apacheSubdir)
			}
		}

		confPath := filepath.Join(tmpDir, "conf", "httpd.conf")
		if raw, err := os.ReadFile(confPath); err == nil {
			text := string(raw)
			targetSlash := filepath.ToSlash(finalTarget)
			reSrvRoot := regexp.MustCompile(`(?mi)^\s*Define\s+SRVROOT\s+.*$`)
			if reSrvRoot.MatchString(text) {
				text = reSrvRoot.ReplaceAllString(text, fmt.Sprintf(`Define SRVROOT "%s"`, targetSlash))
			} else {
				text = fmt.Sprintf("Define SRVROOT \"%s\"\r\n%s", targetSlash, text)
			}
			_ = os.WriteFile(confPath, []byte(text), 0644)
		}
		return nil
	}
	if kind != software.KindPHP {
		return nil
	}
	devIni := filepath.Join(tmpDir, "php.ini-development")
	raw, err := os.ReadFile(devIni)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeInstallFailed, "读取 php.ini-development 失败")
	}

	text := string(raw)
	extDir := filepath.Join(finalTarget, "ext")
	text = strings.ReplaceAll(text, `;extension_dir = "ext"`, fmt.Sprintf(`extension_dir = "%s"`, extDir))

	extensions := []string{
		"curl",
		"fileinfo",
		"gd",
		"mbstring",
		"mysqli",
		"openssl",
		"pdo_mysql",
		"pdo_sqlite",
		"sqlite3",
		"zip",
	}
	for _, ext := range extensions {
		text = strings.ReplaceAll(text, ";extension="+ext, "extension="+ext)
	}
	text = strings.ReplaceAll(text, ";date.timezone =", `date.timezone = "Asia/Shanghai"`)

	targetIni := filepath.Join(tmpDir, "php.ini")
	if err := os.WriteFile(targetIni, []byte(text), 0644); err != nil {
		return appErrors.Wrap(err, appErrors.CodeInstallFailed, "生成 php.ini 失败")
	}
	return nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
