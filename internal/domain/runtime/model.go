package runtime

import "time"

type RuntimeType string

const (
	TypePHP        RuntimeType = "php"
	TypeNginx      RuntimeType = "nginx"
	TypeMySQL      RuntimeType = "mysql"
	TypePostgreSQL RuntimeType = "postgresql"
	TypeRedis      RuntimeType = "redis"
	TypeApache     RuntimeType = "apache"
	TypeNodeJS     RuntimeType = "nodejs"
)

type RuntimeStatus string

const (
	StatusAvailable  RuntimeStatus = "available"
	StatusInstalling RuntimeStatus = "installing"
	StatusInstalled  RuntimeStatus = "installed"
	StatusFailed     RuntimeStatus = "failed"
	StatusRemoving   RuntimeStatus = "removing"
)

type RuntimeVersion struct {
	ID           int64         `json:"id"`
	RuntimeType  RuntimeType   `json:"runtimeType"`
	Version      string        `json:"version"`
	Architecture string        `json:"architecture"`
	InstallPath  string        `json:"installPath"`
	Status       RuntimeStatus `json:"status"`
	MetadataJSON string        `json:"metadataJson"`
	CreatedAt    time.Time     `json:"createdAt"`
	UpdatedAt    time.Time     `json:"updatedAt"`
}

type PHPExtension struct {
	Name      string `json:"name"`
	Enabled   bool   `json:"enabled"`
	Available bool   `json:"available"`
}

type RuntimeView struct {
	ID           int64         `json:"id"`
	RuntimeType  RuntimeType   `json:"runtimeType"`
	Version      string        `json:"version"`
	Variant      string        `json:"variant"`
	Architecture string        `json:"architecture"`
	InstallPath  string        `json:"installPath"`
	Status       RuntimeStatus `json:"status"`
	IsDefault    bool          `json:"isDefault"`
	SiteCount    int           `json:"siteCount"`
	ServiceCount int           `json:"serviceCount"`
}

type InstallInput struct {
	RuntimeType  RuntimeType `json:"runtimeType"`
	Version      string      `json:"version"`
	Variant      string      `json:"variant"`
	Architecture string      `json:"architecture"`
	DownloadURL  string      `json:"downloadUrl"`
	InstallPath  string      `json:"installPath"`
	Mirror       string      `json:"mirror"`
}

type InstallResult struct {
	RuntimeID      int64  `json:"runtimeId"`
	ServiceID      *int64 `json:"serviceId"`
	ServiceCreated bool   `json:"serviceCreated"`
}
