package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"devenv/internal/shared/errors"
)

const (
	APP_FOLDER      = "DevEnv"
	DATA_FOLDER     = "data"
	BACKUP_FOLDER   = "backups"
	SOFTWARE_FOLDER = "software"
	DBDATA_FOLDER   = "dbdata"
	CACHE_FOLDER    = "cache"
	LOG_FOLDER      = "logs"
	SITE_LOG_FOLDER = "sites"
	NGINX_FOLDER    = "nginx"
	VHOST_FOLDER    = "vhosts"
	TEMP_FOLDER     = "temp"
	DB_FILE         = "devenv.db"
	NGINX_CONFIG    = "nginx.conf"
	NGINX_PID       = "nginx.pid"
	LOOPBACK        = "127.0.0.1"
)

func Base() string {
	return filepath.Join(os.Getenv("LOCALAPPDATA"), APP_FOLDER)
}

func Database(base string) string {
	return filepath.Join(base, DATA_FOLDER, DB_FILE)
}

func Backups(databasePath string) string {
	return filepath.Join(filepath.Dir(databasePath), BACKUP_FOLDER)
}

func Software(base, kind, version string) string {
	return filepath.Join(base, SOFTWARE_FOLDER, kind, version)
}

func MySQLData(base, version string) string {
	return DatabaseData(base, "mysql", version)
}

func DatabaseData(base, kind, version string) string {
	return filepath.Join(base, DBDATA_FOLDER, fmt.Sprintf("%s-%s", kind, version))
}

func Cache(base, kind, version string) string {
	return filepath.Join(base, CACHE_FOLDER, fmt.Sprintf("%s-%s.zip.part", kind, version))
}

func SiteLog(base, name, file string) string {
	return filepath.Join(base, LOG_FOLDER, SITE_LOG_FOLDER, name, file)
}

func SoftwareLog(base, kind, version string) string {
	return filepath.Join(base, LOG_FOLDER, SOFTWARE_FOLDER, fmt.Sprintf("%s-%s.log", kind, version))
}

func Nginx(base string) string {
	return filepath.Join(base, DATA_FOLDER, NGINX_FOLDER)
}

func NginxHosts(base string) string {
	return filepath.Join(Nginx(base), VHOST_FOLDER)
}

func NginxTemp(base, name string) string {
	return filepath.Join(Nginx(base), TEMP_FOLDER, name)
}

func NginxLog(base string) string {
	return filepath.Join(Nginx(base), LOG_FOLDER)
}

func NginxConfig(base string) string {
	return filepath.Join(Nginx(base), NGINX_CONFIG)
}

func NginxPID(base string) string {
	return filepath.Join(Nginx(base), NGINX_PID)
}

func GatewayLog(base, name string) string {
	return filepath.Join(base, LOG_FOLDER, NGINX_FOLDER, name)
}

func Hosts() string {
	return filepath.Join(os.Getenv("SystemRoot"), "System32", "drivers", "etc", "hosts")
}

func Shell() string {
	return filepath.Join(os.Getenv("SystemRoot"), "System32", "cmd.exe")
}

func ExpandProfile(value string) string {
	return strings.ReplaceAll(value, "%USERPROFILE%", os.Getenv("USERPROFILE"))
}

func CheckWithin(root, target string) error {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return errors.Wrap(err, errors.PATH_INVALID, "解析根目录失败")
	}
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return errors.Wrap(err, errors.PATH_INVALID, "解析目标路径失败")
	}
	rel, err := filepath.Rel(absRoot, absTarget)
	if err != nil {
		return errors.Wrap(err, errors.PATH_INVALID, "计算相对路径失败")
	}
	slash := filepath.ToSlash(rel)
	if slash == ".." || strings.HasPrefix(slash, "../") {
		return errors.New(errors.PATH_INVALID, "路径超出根目录范围")
	}
	return nil
}
