package service

import (
	"time"

	"devenv/internal/domain/runtime"
)

type Status string

const (
	StatusStopped  Status = "stopped"
	StatusStarting Status = "starting"
	StatusRunning  Status = "running"
	StatusStopping Status = "stopping"
	StatusFailed   Status = "failed"
)

type Instance struct {
	ID               int64               `json:"id"`
	Name             string              `json:"name"`
	ServiceType      runtime.RuntimeType `json:"serviceType"`
	RuntimeVersionID int64               `json:"runtimeVersionId"`
	Port             *int                `json:"port"`
	ConfigPath       *string             `json:"configPath"`
	DataPath         *string             `json:"dataPath"`
	AutoStart        bool                `json:"autoStart"`
	Enabled          bool                `json:"enabled"`
	CreatedAt        time.Time           `json:"createdAt"`
	UpdatedAt        time.Time           `json:"updatedAt"`
}

type StatusInfo struct {
	ServiceID     int64      `json:"serviceId"`
	Status        Status     `json:"status"`
	PID           *int       `json:"pid"`
	StartedAt     *time.Time `json:"startedAt"`
	LastErrorCode *string    `json:"lastErrorCode"`
}
