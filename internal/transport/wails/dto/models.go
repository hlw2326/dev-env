package dto

import "devenv/internal/domain/site"

type OperationResultDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type RuntimeDTO struct {
	ID           int64  `json:"id"`
	RuntimeType  string `json:"runtimeType"`
	Version      string `json:"version"`
	Variant      string `json:"variant"`
	Architecture string `json:"architecture"`
	InstallPath  string `json:"installPath"`
	Status       string `json:"status"`
	IsDefault    bool   `json:"isDefault"`
	SiteCount    int    `json:"siteCount"`
	ServiceCount int    `json:"serviceCount"`
}

type RegisterRuntimeInput struct {
	RuntimeType  string `json:"runtimeType"`
	Version      string `json:"version"`
	Variant      string `json:"variant"`
	Architecture string `json:"architecture"`
	InstallPath  string `json:"installPath"`
}

type InstallRuntimeInput struct {
	RuntimeType  string `json:"runtimeType"`
	Version      string `json:"version"`
	Variant      string `json:"variant"`
	Architecture string `json:"architecture"`
	DownloadURL  string `json:"downloadUrl"`
	InstallPath  string `json:"installPath"`
	Mirror       string `json:"mirror"`
}

type InstallResultDTO struct {
	RuntimeID      int64  `json:"runtimeId"`
	ServiceID      *int64 `json:"serviceId,omitempty"`
	ServiceCreated bool   `json:"serviceCreated"`
}

type PHPExtensionDTO struct {
	Name      string `json:"name"`
	Enabled   bool   `json:"enabled"`
	Available bool   `json:"available"`
}

type ServiceInstanceDTO struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	ServiceType    string  `json:"serviceType"`
	RuntimeID      int64   `json:"runtimeId"`
	RuntimeVersion string  `json:"runtimeVersion"`
	RuntimePath    string  `json:"runtimePath"`
	Port           *int    `json:"port"`
	ConfigPath     *string `json:"configPath"`
	DataPath       *string `json:"dataPath"`
	Enabled        bool    `json:"enabled"`
	AutoStart      bool    `json:"autoStart"`
	Status         string  `json:"status"`
	PID            *int    `json:"pid"`
}

type SwitchRuntimeInput struct {
	ServiceID int64 `json:"serviceId"`
	RuntimeID int64 `json:"runtimeId"`
}

type RuntimeOptionDTO struct {
	ID           int64  `json:"id"`
	Version      string `json:"version"`
	Variant      string `json:"variant"`
	Architecture string `json:"architecture"`
	InstallPath  string `json:"installPath"`
}

type ServiceActionInput struct {
	ServiceID int64 `json:"serviceId"`
}

type ServiceStatusInfoDTO struct {
	ServiceID int64   `json:"serviceId"`
	Status    string  `json:"status"`
	PID       *int    `json:"pid"`
	StartedAt *string `json:"startedAt"`
	ErrorCode *string `json:"errorCode"`
}

type ServiceLogInput struct {
	ServiceID int64  `json:"serviceId"`
	Level     string `json:"level,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

type LogResultDTO struct {
	Lines []string `json:"lines"`
}

type TaskDTO struct {
	ID           int64   `json:"id"`
	TaskType     string  `json:"taskType"`
	Status       string  `json:"status"`
	Progress     int     `json:"progress"`
	Message      string  `json:"message"`
	ErrorCode    *string `json:"errorCode"`
	ErrorMessage *string `json:"errorMessage"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
	FinishedAt   *string `json:"finishedAt"`
}

type AppSettingsDTO struct {
	Language           string `json:"language"`
	Theme              string `json:"theme"`
	RuntimeRootPath    string `json:"runtimeRootPath"`
	DefaultProjectPath string `json:"defaultProjectPath"`
	AutoStartServices  bool   `json:"autoStartServices"`
}

type SiteDTO struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Kind      string    `json:"kind"`
	Domains   []string  `json:"domains"`
	Root      string    `json:"root"`
	RuntimeID *int64    `json:"runtimeId"`
	ServiceID *int64    `json:"serviceId"`
	Git       bool      `json:"git"`
	Remote    string    `json:"remote"`
	Branch    string    `json:"branch"`
	Commit    string    `json:"commit"`
	PulledAt  string    `json:"pulledAt"`
	Conf      string    `json:"conf"`
	Port      int       `json:"port"`
	TLS       bool      `json:"tls"`
	Cert      string    `json:"cert"`
	Key       string    `json:"key"`
	Issuer    string    `json:"issuer"`
	Proxy     bool      `json:"proxy"`
	Upstream  *int      `json:"upstream"`
	Meta      site.Meta `json:"meta"`
	Status    string    `json:"status"`
	CreatedAt string    `json:"createdAt"`
	UpdatedAt string    `json:"updatedAt"`
}

type CreateSiteInput struct {
	Name        string            `json:"name"`
	Kind        string            `json:"kind"`
	Domains     []string          `json:"domains"`
	Root        string            `json:"root"`
	RuntimeID   *int64            `json:"runtimeId,omitempty"`
	ServiceID   *int64            `json:"serviceId,omitempty"`
	AutoService bool              `json:"autoService"`
	Git         bool              `json:"git"`
	Remote      string            `json:"remote"`
	Branch      string            `json:"branch"`
	Port        int               `json:"port"`
	TLS         bool              `json:"tls"`
	Issuer      string            `json:"issuer"`
	Cert        string            `json:"cert"`
	Key         string            `json:"key"`
	Upstream    *int              `json:"upstream,omitempty"`
	Entry       string            `json:"entry"`
	Cmd         string            `json:"cmd"`
	AutoStart   bool              `json:"autoStart"`
	StartNow    bool              `json:"startNow"`
	Custom      string            `json:"custom"`
	Env         map[string]string `json:"env,omitempty"`
}

type UpdateSiteInput struct {
	ID        int64    `json:"id"`
	Name      *string  `json:"name,omitempty"`
	Domains   []string `json:"domains,omitempty"`
	Root      *string  `json:"root,omitempty"`
	RuntimeID *int64   `json:"runtimeId,omitempty"`
	ServiceID *int64   `json:"serviceId,omitempty"`
	Port      *int     `json:"port,omitempty"`
	TLS       *bool    `json:"tls,omitempty"`
	Issuer    *string  `json:"issuer,omitempty"`
	Cert      *string  `json:"cert,omitempty"`
	Key       *string  `json:"key,omitempty"`
	Upstream  *int     `json:"upstream,omitempty"`
	Entry     *string  `json:"entry,omitempty"`
	Cmd       *string  `json:"cmd,omitempty"`
	Custom    *string  `json:"custom,omitempty"`
	AutoStart *bool    `json:"autoStart,omitempty"`
}

type DeleteSiteInput struct {
	ID    int64 `json:"id"`
	Purge bool  `json:"purge"`
}

type GitStatusDTO struct {
	Branch    string   `json:"branch"`
	Commit    string   `json:"commit"`
	Message   string   `json:"message"`
	Clean     bool     `json:"clean"`
	Ahead     int      `json:"ahead"`
	Behind    int      `json:"behind"`
	Modified  []string `json:"modified"`
	Untracked []string `json:"untracked"`
	CheckedAt string   `json:"checkedAt"`
}

type GitCommitInput struct {
	SiteID  int64    `json:"siteId"`
	Message string   `json:"message"`
	Files   []string `json:"files"`
}

type SaveConfInput struct {
	SiteID  int64  `json:"siteId"`
	Content string `json:"content"`
	Summary string `json:"summary"`
}

type ConfCheckDTO struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}

type SiteLogInput struct {
	SiteID int64  `json:"siteId"`
	Kind   string `json:"kind"`
	Limit  int    `json:"limit"`
}
