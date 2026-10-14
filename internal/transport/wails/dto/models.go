package dto

type OperationResultDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type ProjectDTO struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Path         string  `json:"path"`
	PHPVersionID *int64  `json:"phpVersionId"`
	Status       string  `json:"status"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
}

type CreateProjectInput struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	PHPVersionID *int64 `json:"phpVersionId,omitempty"`
}

type UpdateProjectInput struct {
	ID           int64   `json:"id"`
	Name         *string `json:"name,omitempty"`
	PHPVersionID *int64  `json:"phpVersionId,omitempty"`
}

type ProjectPathValidationDTO struct {
	Valid          bool     `json:"valid"`
	NormalizedPath string   `json:"normalizedPath"`
	Exists         bool     `json:"exists"`
	Writable       bool     `json:"writable"`
	Errors         []string `json:"errors"`
}

type PHPVersionDTO struct {
	ID           int64  `json:"id"`
	Version      string `json:"version"`
	Architecture string `json:"architecture"`
	InstallPath  string `json:"installPath"`
	Status       string `json:"status"`
	IsDefault    bool   `json:"isDefault"`
}

type RegisterPHPVersionInput struct {
	Version      string `json:"version"`
	Architecture string `json:"architecture"`
	InstallPath  string `json:"installPath"`
}

type PHPExtensionDTO struct {
	Name      string `json:"name"`
	Enabled   bool   `json:"enabled"`
	Available bool   `json:"available"`
}

type ServiceInstanceDTO struct {
	ID               int64   `json:"id"`
	Name             string  `json:"name"`
	ServiceType      string  `json:"serviceType"`
	RuntimeVersionID int64   `json:"runtimeVersionId"`
	Port             *int    `json:"port"`
	ConfigPath       *string `json:"configPath"`
	DataPath         *string `json:"dataPath"`
	Enabled          bool    `json:"enabled"`
	AutoStart        bool    `json:"autoStart"`
	Status           string  `json:"status"`
	PID              *int    `json:"pid"`
}

type ServiceActionInput struct {
	ServiceID int64 `json:"serviceId"`
}

type ServiceStatusInfoDTO struct {
	ServiceID     int64   `json:"serviceId"`
	Status        string  `json:"status"`
	PID           *int    `json:"pid"`
	StartedAt     *string `json:"startedAt"`
	LastErrorCode *string `json:"lastErrorCode"`
}

type ServiceLogInput struct {
	ServiceID int64  `json:"serviceId"`
	Level     string `json:"level,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

type ServiceLogResultDTO struct {
	Lines []string `json:"lines"`
}

type VirtualHostDTO struct {
	ID            int64   `json:"id"`
	ProjectID     int64   `json:"projectId"`
	Domain        string  `json:"domain"`
	DocumentRoot  string  `json:"documentRoot"`
	WebServerType string  `json:"webServerType"`
	Port          int     `json:"port"`
	Enabled       bool    `json:"enabled"`
	ConfigPath    *string `json:"configPath"`
}

type CreateVirtualHostInput struct {
	ProjectID     int64  `json:"projectId"`
	Domain        string `json:"domain"`
	DocumentRoot  string `json:"documentRoot"`
	WebServerType string `json:"webServerType"`
	Port          int    `json:"port"`
}

type ValidationResultDTO struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
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
