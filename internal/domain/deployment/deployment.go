package deployment

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"devenv/internal/domain/project"
)

const (
	PROVIDER_SFTP     = "sftp"
	PROVIDER_WECHAT   = "wechat"
	STATE_RUNNING     = "running"
	STATE_SUCCESS     = "success"
	STATE_FAILED      = "failed"
	STATE_CANCELLED   = "cancelled"
	STATE_INTERRUPTED = "interrupted"
)

// Config contains only non-secret settings. Passwords and key passphrases are never persisted.
type Config struct {
	ProjectID   int64  `json:"projectId"`
	Provider    string `json:"provider"`
	Source      string `json:"source"`
	Exclude     string `json:"exclude"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	PrivateKey  string `json:"privateKey"`
	Fingerprint string `json:"fingerprint"`
	Remote      string `json:"remote"`
	AppID       string `json:"appId"`
	UploadKey   string `json:"uploadKey"`
	NodePath    string `json:"nodePath"`
	ModulePath  string `json:"modulePath"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Robot       int    `json:"robot"`
}

type File struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	Digest string `json:"digest"`
}
type Plan struct {
	Token     string `json:"token"`
	ProjectID int64  `json:"projectId"`
	Provider  string `json:"provider"`
	Source    string `json:"source"`
	Target    string `json:"target"`
	Files     []File `json:"files"`
	Bytes     int64  `json:"bytes"`
	Excluded  int    `json:"excluded"`
}
type Task struct {
	ID        string `json:"id"`
	ProjectID int64  `json:"projectId"`
	Provider  string `json:"provider"`
	State     string `json:"state"`
	Message   string `json:"message"`
	Target    string `json:"target"`
	Completed int    `json:"completed"`
	Total     int    `json:"total"`
	Created   string `json:"created"`
	Finished  string `json:"finished"`
}
type Credentials struct {
	Password   string
	Passphrase string
}

func ProviderFor(kind string) string {
	if project.HasCapability(kind, project.CAPABILITY_SERVER_SYNC) {
		return PROVIDER_SFTP
	}
	if project.HasCapability(kind, project.CAPABILITY_WECHAT_UPLOAD) {
		return PROVIDER_WECHAT
	}
	return ""
}

func Default(projectID int64, kind string) Config {
	return Config{ProjectID: projectID, Provider: ProviderFor(kind), Source: ".", Port: 22, Robot: 1}
}

func Validate(config Config, kind string) error {
	if config.Provider == "" || config.Provider != ProviderFor(kind) {
		return fmt.Errorf("当前项目不支持此发布方式")
	}
	if strings.TrimSpace(config.Source) == "" {
		return fmt.Errorf("请填写相对项目根目录的发布目录，根目录填写 .")
	}
	if config.Provider == PROVIDER_SFTP {
		if strings.TrimSpace(config.Host) == "" || strings.ContainsAny(config.Host, "/\\\r\n\t ") || strings.TrimSpace(config.Username) == "" {
			return fmt.Errorf("请填写有效的 SSH 主机和用户名")
		}
		if config.Port < 1 || config.Port > 65535 {
			return fmt.Errorf("SSH 端口必须在 1–65535 之间")
		}
		if !regexp.MustCompile(`^SHA256:[A-Za-z0-9+/]{43}$`).MatchString(config.Fingerprint) {
			return fmt.Errorf("请填写并核验服务器 SSH SHA256 指纹")
		}
		if !strings.HasPrefix(config.Remote, "/") || strings.ContainsAny(config.Remote, "\\\x00\r\n") || path.Clean(config.Remote) != config.Remote || config.Remote == "/" || len(strings.Split(strings.Trim(config.Remote, "/"), "/")) < 2 {
			return fmt.Errorf("远端必须是明确的 Linux 项目绝对目录，如 /var/www/my-project，不能使用根目录或 ..")
		}
	} else {
		if !regexp.MustCompile(`^wx[0-9a-fA-F]{16}$`).MatchString(config.AppID) {
			return fmt.Errorf("请填写有效的微信小程序 AppID")
		}
		if config.UploadKey == "" {
			return fmt.Errorf("请选择微信代码上传私钥文件")
		}
		if strings.TrimSpace(config.Version) == "" || len(config.Version) > 64 {
			return fmt.Errorf("请填写上传版本号（最多 64 字节）")
		}
		if config.Robot < 1 || config.Robot > 30 {
			return fmt.Errorf("微信机器人编号必须在 1–30 之间")
		}
	}
	if len(config.Exclude) > 8192 || len(config.Description) > 1024 {
		return fmt.Errorf("排除规则或发布说明过长")
	}
	return nil
}
