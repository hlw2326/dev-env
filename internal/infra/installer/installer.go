package installer

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"devenv/internal/domain/software"
	"devenv/internal/ports"
	appErrors "devenv/internal/shared/errors"
	"devenv/internal/shared/paths"
)

const (
	DOWNLOAD_TIMEOUT  = 30 * time.Minute
	PROGRESS_INTERVAL = 500 * time.Millisecond
	BUFFER_SIZE       = 64 * 1024
	BYTE_UNIT         = 1024
)

type Manager struct {
	http *http.Client
}

func NewManager() *Manager {
	return &Manager{
		http: &http.Client{Timeout: DOWNLOAD_TIMEOUT},
	}
}

func NewWithClient(client *http.Client) *Manager {
	return &Manager{http: client}
}

func (m *Manager) Install(ctx context.Context, request ports.Request, report func(progress int, message string)) error {
	partPath := request.Cache
	tempFolder := request.Target + ".tmp"

	cleanup := func() {
		if partPath != "" {
			if err := os.Remove(partPath); err != nil && !os.IsNotExist(err) {
				partPath = ""
			}
		}
		if tempFolder != "" {
			if err := os.RemoveAll(tempFolder); err != nil {
				tempFolder = ""
			}
		}
	}

	if err := os.RemoveAll(tempFolder); err != nil {
		return appErrors.Wrap(err, appErrors.INSTALL_FAILED, "清理临时解压目录失败")
	}
	if err := os.MkdirAll(filepath.Dir(partPath), 0o755); err != nil {
		return appErrors.Wrap(err, appErrors.INSTALL_FAILED, "创建下载缓存目录失败")
	}
	if err := os.MkdirAll(filepath.Dir(request.Target), 0o755); err != nil {
		return appErrors.Wrap(err, appErrors.INSTALL_FAILED, "创建软件安装父目录失败")
	}

	report(2, "正在下载安装包...")
	hash, err := m.downloadFile(ctx, request.URL, partPath, report)
	if err != nil {
		cleanup()
		return err
	}

	needHash := request.Kind == software.KIND_PHP || request.Kind == software.KIND_NODE || request.Kind == software.KIND_GO || request.Kind == software.KIND_JAVA || request.Kind == software.KIND_DOTNET
	if needHash && strings.TrimSpace(request.Checksum) == "" {
		cleanup()
		return appErrors.New(appErrors.DOWNLOAD_FAILED, "缺少安装包 SHA-256 校验值")
	}
	if strings.TrimSpace(request.Checksum) != "" {
		expected := strings.ToLower(strings.TrimSpace(request.Checksum))
		actual := hash
		if len(expected) == sha512.Size*2 {
			actual, err = fileSHA512(partPath)
			if err != nil {
				cleanup()
				return err
			}
		}
		if !strings.EqualFold(actual, expected) {
			cleanup()
			return appErrors.New(appErrors.DOWNLOAD_FAILED, fmt.Sprintf("安装包哈希校验失败: 期望 %s，实际 %s", expected, actual))
		}
	}

	if request.Kind == software.KIND_COMPOSER {
		report(80, "正在部署 Composer...")
		if err := placeComposer(partPath, tempFolder); err != nil {
			cleanup()
			return err
		}
	} else {
		report(72, "正在解压安装包...")
		if err := extractZip(ctx, partPath, tempFolder, report); err != nil {
			cleanup()
			return err
		}
	}

	if err := checkEntry(request.Kind, tempFolder); err != nil {
		cleanup()
		return err
	}

	report(92, "正在初始化配置...")
	if err := postProcess(request.Kind, tempFolder, request.Target); err != nil {
		cleanup()
		return err
	}

	if err := os.RemoveAll(request.Target); err != nil {
		cleanup()
		return appErrors.Wrap(err, appErrors.INSTALL_FAILED, "清理旧目标目录失败")
	}
	if err := os.Rename(tempFolder, request.Target); err != nil {
		cleanup()
		return appErrors.Wrap(err, appErrors.INSTALL_FAILED, "启用安装目录失败")
	}

	if err := os.Remove(partPath); err != nil && !os.IsNotExist(err) {
		return appErrors.Wrap(err, appErrors.INSTALL_FAILED, "清理下载缓存失败")
	}
	report(100, "安装完成")
	return nil
}

func (m *Manager) downloadFile(ctx context.Context, url, partPath string, report func(progress int, message string)) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", appErrors.Wrap(err, appErrors.DOWNLOAD_FAILED, "创建下载请求失败")
	}
	request.Header.Set("User-Agent", "curl/8.13.0")

	response, err := m.http.Do(request)
	if err != nil {
		return "", appErrors.Wrap(err, appErrors.DOWNLOAD_FAILED, "下载请求失败")
	}
	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			return
		}
	}()

	if response.StatusCode != http.StatusOK {
		return "", appErrors.New(appErrors.DOWNLOAD_FAILED, fmt.Sprintf("下载失败，HTTP 状态码 %d", response.StatusCode))
	}

	file, err := os.Create(partPath)
	if err != nil {
		return "", appErrors.Wrap(err, appErrors.DOWNLOAD_FAILED, "创建缓存文件失败")
	}

	hasher := sha256.New()
	total := response.ContentLength
	var downloaded int64
	lastTime := time.Now()
	buffer := make([]byte, BUFFER_SIZE)

	for {
		if err := ctx.Err(); err != nil {
			if closeErr := file.Close(); closeErr != nil {
				return "", appErrors.Wrap(err, appErrors.DOWNLOAD_FAILED, "下载已取消")
			}
			return "", appErrors.Wrap(err, appErrors.DOWNLOAD_FAILED, "下载已取消")
		}
		count, readErr := response.Body.Read(buffer)
		if count > 0 {
			if _, writeErr := file.Write(buffer[:count]); writeErr != nil {
				if closeErr := file.Close(); closeErr != nil {
					return "", appErrors.Wrap(writeErr, appErrors.DOWNLOAD_FAILED, "写入缓存文件失败")
				}
				return "", appErrors.Wrap(writeErr, appErrors.DOWNLOAD_FAILED, "写入缓存文件失败")
			}
			if _, hashErr := hasher.Write(buffer[:count]); hashErr != nil {
				if closeErr := file.Close(); closeErr != nil {
					return "", appErrors.Wrap(hashErr, appErrors.DOWNLOAD_FAILED, "计算哈希失败")
				}
				return "", appErrors.Wrap(hashErr, appErrors.DOWNLOAD_FAILED, "计算哈希失败")
			}
			downloaded += int64(count)

			if time.Since(lastTime) >= PROGRESS_INTERVAL {
				lastTime = time.Now()
				progress := 35
				var message string
				if total > 0 {
					progress = int(float64(downloaded) / float64(total) * 70)
					if progress > 70 {
						progress = 70
					}
					message = fmt.Sprintf("正在下载安装包 (%s / %s)...", formatBytes(downloaded), formatBytes(total))
				} else {
					message = fmt.Sprintf("正在下载安装包 (%s)...", formatBytes(downloaded))
				}
				report(progress, message)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			if closeErr := file.Close(); closeErr != nil {
				return "", appErrors.Wrap(readErr, appErrors.DOWNLOAD_FAILED, "读取下载流失败")
			}
			return "", appErrors.Wrap(readErr, appErrors.DOWNLOAD_FAILED, "读取下载流失败")
		}
	}

	if err := file.Close(); err != nil {
		return "", appErrors.Wrap(err, appErrors.DOWNLOAD_FAILED, "关闭缓存文件失败")
	}
	report(70, "下载完成，正在准备解压...")
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func fileSHA512(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", appErrors.Wrap(err, appErrors.DOWNLOAD_FAILED, "读取缓存文件失败")
	}
	defer file.Close()
	hasher := sha512.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", appErrors.Wrap(err, appErrors.DOWNLOAD_FAILED, "计算 SHA-512 失败")
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func formatBytes(size int64) string {
	if size < BYTE_UNIT {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(BYTE_UNIT), 0
	for n := size / BYTE_UNIT; n >= BYTE_UNIT; n /= BYTE_UNIT {
		div *= BYTE_UNIT
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func extractZip(ctx context.Context, zipPath, targetFolder string, report func(progress int, message string)) error {
	archive, err := zip.OpenReader(zipPath)
	if err != nil {
		return appErrors.Wrap(err, appErrors.INSTALL_FAILED, "打开压缩包失败")
	}
	defer func() {
		if closeErr := archive.Close(); closeErr != nil {
			return
		}
	}()

	if len(archive.File) == 0 {
		return appErrors.New(appErrors.INSTALL_FAILED, "压缩包为空")
	}

	prefix := findPrefix(archive.File)
	if err := os.MkdirAll(targetFolder, 0o755); err != nil {
		return appErrors.Wrap(err, appErrors.INSTALL_FAILED, "创建解压目录失败")
	}
	absTarget, err := filepath.Abs(targetFolder)
	if err != nil {
		return appErrors.Wrap(err, appErrors.INSTALL_FAILED, "解析解压目录路径失败")
	}

	total := len(archive.File)
	for i, item := range archive.File {
		if err := ctx.Err(); err != nil {
			return appErrors.Wrap(err, appErrors.INSTALL_FAILED, "解压已取消")
		}
		relName := strings.ReplaceAll(item.Name, `\`, "/")
		if prefix != "" {
			relName = strings.TrimPrefix(relName, prefix)
		}
		relName = strings.TrimPrefix(relName, "/")
		if relName == "" {
			continue
		}

		cleaned := filepath.ToSlash(filepath.Clean(relName))
		if cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.Contains(cleaned, "/../") || filepath.IsAbs(relName) {
			return appErrors.New(appErrors.INSTALL_FAILED, "检测到非法压缩包路径: "+item.Name)
		}

		destPath := filepath.Join(absTarget, filepath.FromSlash(cleaned))
		if err := paths.CheckWithin(absTarget, destPath); err != nil {
			return appErrors.New(appErrors.INSTALL_FAILED, "检测到越界压缩包路径: "+item.Name)
		}

		if item.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, 0o755); err != nil {
				return appErrors.Wrap(err, appErrors.INSTALL_FAILED, "创建目录失败")
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return appErrors.Wrap(err, appErrors.INSTALL_FAILED, "创建父目录失败")
		}

		reader, err := item.Open()
		if err != nil {
			return appErrors.Wrap(err, appErrors.INSTALL_FAILED, "读取压缩包条目失败")
		}
		writer, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
		if err != nil {
			if closeErr := reader.Close(); closeErr != nil {
				return appErrors.Wrap(err, appErrors.INSTALL_FAILED, "创建目标文件失败")
			}
			return appErrors.Wrap(err, appErrors.INSTALL_FAILED, "创建目标文件失败")
		}
		_, copyErr := io.Copy(writer, reader)
		closeWriteErr := writer.Close()
		closeReadErr := reader.Close()
		if copyErr != nil {
			return appErrors.Wrap(copyErr, appErrors.INSTALL_FAILED, "解压写入文件失败")
		}
		if closeWriteErr != nil {
			return appErrors.Wrap(closeWriteErr, appErrors.INSTALL_FAILED, "关闭目标文件失败")
		}
		if closeReadErr != nil {
			return appErrors.Wrap(closeReadErr, appErrors.INSTALL_FAILED, "关闭压缩条目失败")
		}

		if total > 0 && i%25 == 0 {
			progress := 70 + int(float64(i+1)/float64(total)*20)
			if progress > 90 {
				progress = 90
			}
			report(progress, "正在解压安装包...")
		}
	}
	return nil
}

func findPrefix(files []*zip.File) string {
	var root string
	for _, item := range files {
		name := strings.TrimPrefix(strings.ReplaceAll(item.Name, `\`, "/"), "/")
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

func placeComposer(partPath, targetFolder string) error {
	if err := os.MkdirAll(targetFolder, 0o755); err != nil {
		return appErrors.Wrap(err, appErrors.INSTALL_FAILED, "创建 Composer 目录失败")
	}
	raw, err := os.ReadFile(partPath)
	if err != nil {
		return appErrors.Wrap(err, appErrors.INSTALL_FAILED, "读取 Composer 下载包失败")
	}
	pharPath := filepath.Join(targetFolder, "composer.phar")
	if err := os.WriteFile(pharPath, raw, 0o755); err != nil {
		return appErrors.Wrap(err, appErrors.INSTALL_FAILED, "写入 composer.phar 失败")
	}
	batPath := filepath.Join(targetFolder, "composer.bat")
	batContent := "@php \"%~dp0composer.phar\" %*\r\n"
	if err := os.WriteFile(batPath, []byte(batContent), 0o755); err != nil {
		return appErrors.Wrap(err, appErrors.INSTALL_FAILED, "写入 composer.bat 失败")
	}
	return nil
}

func checkEntry(kind, folder string) error {
	switch kind {
	case software.KIND_PHP:
		if !fileExists(filepath.Join(folder, "php-cgi.exe")) {
			return appErrors.New(appErrors.INSTALL_FAILED, "安装包中未找到 php-cgi.exe")
		}
	case software.KIND_NODE:
		if !fileExists(filepath.Join(folder, "node.exe")) {
			return appErrors.New(appErrors.INSTALL_FAILED, "安装包中未找到 node.exe")
		}
	case software.KIND_GO:
		if !fileExists(filepath.Join(folder, "bin", "go.exe")) {
			return appErrors.New(appErrors.INSTALL_FAILED, "安装包中未找到 bin\\go.exe")
		}
	case software.KIND_PYTHON:
		if !fileExists(filepath.Join(folder, "python.exe")) {
			return appErrors.New(appErrors.INSTALL_FAILED, "安装包中未找到 python.exe")
		}
	case software.KIND_COMPOSER:
		if !fileExists(filepath.Join(folder, "composer.phar")) {
			return appErrors.New(appErrors.INSTALL_FAILED, "安装包中未找到 composer.phar")
		}
	case software.KIND_MYSQL:
		if !fileExists(filepath.Join(folder, "bin", "mysqld.exe")) {
			return appErrors.New(appErrors.INSTALL_FAILED, "安装包中未找到 bin\\mysqld.exe")
		}
	case software.KIND_MARIADB:
		if !fileExists(filepath.Join(folder, "bin", "mariadbd.exe")) && !fileExists(filepath.Join(folder, "bin", "mysqld.exe")) {
			return appErrors.New(appErrors.INSTALL_FAILED, "安装包中未找到 bin\\mariadbd.exe")
		}
	case software.KIND_POSTGRESQL:
		if !fileExists(filepath.Join(folder, "bin", "postgres.exe")) {
			return appErrors.New(appErrors.INSTALL_FAILED, "安装包中未找到 bin\\postgres.exe")
		}
	case software.KIND_MONGODB:
		if !fileExists(filepath.Join(folder, "bin", "mongod.exe")) {
			return appErrors.New(appErrors.INSTALL_FAILED, "安装包中未找到 bin\\mongod.exe")
		}
	case software.KIND_REDIS:
		if FindRedis(folder) == "" {
			return appErrors.New(appErrors.INSTALL_FAILED, "安装包中未找到 redis-server.exe")
		}
	case software.KIND_MEMCACHED:
		if FindBinary(folder, "memcached.exe") == "" {
			return appErrors.New(appErrors.INSTALL_FAILED, "安装包中未找到 memcached.exe")
		}
	case software.KIND_JAVA:
		return requireEntry(folder, filepath.Join("bin", "java.exe"))
	case software.KIND_DOTNET:
		return requireEntry(folder, "dotnet.exe")
	case software.KIND_DENO:
		return requireEntry(folder, "deno.exe")
	case software.KIND_BUN:
		return requireEntry(folder, "bun.exe")
	case software.KIND_NGINX:
		return requireEntry(folder, "nginx.exe")
	case software.KIND_APACHE:
		return requireEntry(folder, filepath.Join("bin", "httpd.exe"))
	case software.KIND_CADDY:
		return requireEntry(folder, "caddy.exe")
	case software.KIND_SQLITE:
		return requireEntry(folder, "sqlite3.exe")
	case software.KIND_GRADLE:
		return requireEntry(folder, filepath.Join("bin", "gradle.bat"))
	case software.KIND_MAVEN:
		return requireEntry(folder, filepath.Join("bin", "mvn.cmd"))
	}
	return nil
}

func requireEntry(folder, relative string) error {
	if fileExists(filepath.Join(folder, relative)) {
		return nil
	}
	return appErrors.New(appErrors.INSTALL_FAILED, "安装包中未找到 "+relative)
}

func FindRedis(folder string) string {
	return FindBinary(folder, "redis-server.exe")
}

func FindBinary(folder, name string) string {
	direct := filepath.Join(folder, name)
	if fileExists(direct) {
		return direct
	}
	if strings.EqualFold(name, "memcached.exe") {
		preferred := filepath.Join(folder, "libevent-2.1", "x64", "memcached.exe")
		if fileExists(preferred) {
			return preferred
		}
		fallback := filepath.Join(folder, "libevent-2.0", "x64", "memcached.exe")
		if fileExists(fallback) {
			return fallback
		}
	}
	var found string
	if err := filepath.WalkDir(folder, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		if strings.EqualFold(entry.Name(), name) {
			found = path
			return filepath.SkipAll
		}
		return nil
	}); err != nil {
		return ""
	}
	return found
}

func postProcess(kind, tempFolder, targetFolder string) error {
	if kind != software.KIND_PHP {
		return nil
	}
	sourceIni := filepath.Join(tempFolder, "php.ini-development")
	raw, err := os.ReadFile(sourceIni)
	if err != nil {
		return appErrors.Wrap(err, appErrors.INSTALL_FAILED, "读取 php.ini-development 失败")
	}

	text := string(raw)
	extensionFolder := filepath.Join(targetFolder, "ext")
	text = strings.ReplaceAll(text, `;extension_dir = "ext"`, fmt.Sprintf(`extension_dir = "%s"`, extensionFolder))

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
	for _, extension := range extensions {
		text = strings.ReplaceAll(text, ";extension="+extension, "extension="+extension)
	}
	text = strings.ReplaceAll(text, ";date.timezone =", `date.timezone = "Asia/Shanghai"`)

	targetIni := filepath.Join(tempFolder, "php.ini")
	if err := os.WriteFile(targetIni, []byte(text), 0o644); err != nil {
		return appErrors.Wrap(err, appErrors.INSTALL_FAILED, "生成 php.ini 失败")
	}
	return nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
