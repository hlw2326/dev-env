package ports

import (
	"context"

	"devenv/internal/domain/service"
)

type ProcessSpec struct {
	Executable  string
	Args        []string
	WorkingDir  string
	Environment []string
}

type ProcessInfo struct {
	PID     int
	Running bool
}

type ProcessManager interface {
	Start(ctx context.Context, serviceID int64, spec ProcessSpec) (ProcessInfo, error)
	Stop(ctx context.Context, serviceID int64) error
	GetStatus(ctx context.Context, serviceID int64) service.StatusInfo
	GetLogs(ctx context.Context, serviceID int64, limit int) []string
}

type PortChecker interface {
	IsPortAvailable(port int) bool
	FindAvailablePort(startPort int) (int, error)
}

type EventPublisher interface {
	Emit(eventName string, data interface{})
}
