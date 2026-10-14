package site

import (
	"errors"
	"strings"
	"time"
)

type SiteType string

const (
	TypeStatic SiteType = "static"
	TypePHP    SiteType = "php"
	TypeNodeJS SiteType = "nodejs"
	TypeGo     SiteType = "go"
	TypePython SiteType = "python"
)

type Status string

const (
	StatusStopped Status = "stopped"
	StatusRunning Status = "running"
	StatusError   Status = "error"
)

type Metadata struct {
	EntryFile         string            `json:"entryFile,omitempty"`
	NodeCommand       string            `json:"nodeCommand,omitempty"`
	GoExecutable      string            `json:"goExecutable,omitempty"`
	PythonWsgi        string            `json:"pythonWsgi,omitempty"`
	AutoStart         bool              `json:"autoStart"`
	EnvVars           map[string]string `json:"envVars,omitempty"`
	NginxCustomConfig string            `json:"nginxCustomConfig,omitempty"`
	AccessLog         string            `json:"accessLog,omitempty"`
	ErrorLog          string            `json:"errorLog,omitempty"`
	ConfigHistory     []ConfigSnapshot  `json:"configHistory,omitempty"`
}

type ConfigSnapshot struct {
	Timestamp string `json:"timestamp"`
	Summary   string `json:"summary"`
	Content   string `json:"content"`
}

type Site struct {
	ID                int64     `json:"id"`
	Name              string    `json:"name"`
	SiteType          SiteType  `json:"siteType"`
	DomainList        []string  `json:"domainList"`
	RootPath          string    `json:"rootPath"`
	RuntimeVersionID  *int64    `json:"runtimeVersionId"`
	ServiceInstanceID *int64    `json:"serviceInstanceId"`
	GitEnabled        bool      `json:"gitEnabled"`
	GitRemoteURL      string    `json:"gitRemoteUrl"`
	GitBranch         string    `json:"gitBranch"`
	GitLastCommit     string    `json:"gitLastCommit"`
	GitLastPullAt     string    `json:"gitLastPullAt"`
	NginxConfigPath   string    `json:"nginxConfigPath"`
	ListenPort        int       `json:"listenPort"`
	SSLEnabled        bool      `json:"sslEnabled"`
	SSLCertPath       string    `json:"sslCertPath"`
	SSLKeyPath        string    `json:"sslKeyPath"`
	SSLProvider       string    `json:"sslProvider"`
	ProxyEnabled      bool      `json:"proxyEnabled"`
	ProxyTargetPort   *int      `json:"proxyTargetPort"`
	MetadataJSON      string    `json:"metadataJson"`
	Status            Status    `json:"status"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type CreateInput struct {
	Name              string            `json:"name"`
	SiteType          SiteType          `json:"siteType"`
	Domains           []string          `json:"domains"`
	RootPath          string            `json:"rootPath"`
	RuntimeVersionID  *int64            `json:"runtimeVersionId"`
	ServiceInstanceID *int64            `json:"serviceInstanceId"`
	AutoCreateService bool              `json:"autoCreateService"`
	ServicePort       *int              `json:"servicePort"`
	GitEnabled        bool              `json:"gitEnabled"`
	GitRemoteURL      string            `json:"gitRemoteUrl"`
	GitBranch         string            `json:"gitBranch"`
	ListenPort        int               `json:"listenPort"`
	SSLEnabled        bool              `json:"sslEnabled"`
	SSLProvider       string            `json:"sslProvider"`
	SSLCertPath       string            `json:"sslCertPath"`
	SSLKeyPath        string            `json:"sslKeyPath"`
	ProxyTargetPort   *int              `json:"proxyTargetPort"`
	EntryFile         string            `json:"entryFile"`
	StartCommand      string            `json:"startCommand"`
	AutoStart         bool              `json:"autoStart"`
	StartImmediately  bool              `json:"startImmediately"`
	NginxCustomConfig string            `json:"nginxCustomConfig"`
	EnvVars           map[string]string `json:"envVars"`
}

func (in *CreateInput) Validate() error {
	if strings.TrimSpace(in.Name) == "" {
		return errors.New("网站名称不能为空")
	}
	if in.SiteType == "" {
		in.SiteType = TypeStatic
	}
	cleanDomains := make([]string, 0, len(in.Domains))
	for _, d := range in.Domains {
		trimmed := strings.ToLower(strings.TrimSpace(d))
		if trimmed != "" {
			cleanDomains = append(cleanDomains, trimmed)
		}
	}
	if len(cleanDomains) == 0 {
		return errors.New("至少需要绑定一个域名")
	}
	in.Domains = cleanDomains
	if strings.TrimSpace(in.RootPath) == "" {
		return errors.New("网站根目录不能为空")
	}
	if in.ListenPort <= 0 || in.ListenPort > 65535 {
		in.ListenPort = 80
	}
	if in.GitBranch == "" {
		in.GitBranch = "main"
	}
	return nil
}

type UpdateInput struct {
	ID                int64    `json:"id"`
	Name              *string  `json:"name,omitempty"`
	Domains           []string `json:"domains,omitempty"`
	RootPath          *string  `json:"rootPath,omitempty"`
	RuntimeVersionID  *int64   `json:"runtimeVersionId,omitempty"`
	ServiceInstanceID *int64   `json:"serviceInstanceId,omitempty"`
	ListenPort        *int     `json:"listenPort,omitempty"`
	SSLEnabled        *bool    `json:"sslEnabled,omitempty"`
	SSLProvider       *string  `json:"sslProvider,omitempty"`
	SSLCertPath       *string  `json:"sslCertPath,omitempty"`
	SSLKeyPath        *string  `json:"sslKeyPath,omitempty"`
	ProxyTargetPort   *int     `json:"proxyTargetPort,omitempty"`
	EntryFile         *string  `json:"entryFile,omitempty"`
	NginxCustomConfig *string  `json:"nginxCustomConfig,omitempty"`
	AutoStart         *bool    `json:"autoStart,omitempty"`
}

type GitStatus struct {
	CurrentBranch  string   `json:"currentBranch"`
	CurrentCommit  string   `json:"currentCommit"`
	CommitMessage  string   `json:"commitMessage"`
	IsClean        bool     `json:"isClean"`
	AheadCount     int      `json:"aheadCount"`
	BehindCount    int      `json:"behindCount"`
	ModifiedFiles  []string `json:"modifiedFiles"`
	UntrackedFiles []string `json:"untrackedFiles"`
	LastCheckAt    string   `json:"lastCheckAt"`
}
