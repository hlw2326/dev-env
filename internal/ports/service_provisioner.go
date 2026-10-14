package ports

import (
	"context"

	"devenv/internal/domain/runtime"
	"devenv/internal/domain/service"
)

type ServiceProvisioner interface {
	// EnsureForRuntime 保证该运行环境类型下至少有一个服务实例，已有则原样返回
	EnsureForRuntime(ctx context.Context, rt runtime.RuntimeVersion) (*service.Instance, bool, error)
	// IsRuntimeRunning 判断关联该运行环境的服务实例是否正在运行
	IsRuntimeRunning(ctx context.Context, runtimeID int64) (bool, error)
	// OnRuntimeRemoved 运行环境被卸载时清理孤立的服务实例
	OnRuntimeRemoved(ctx context.Context, rt runtime.RuntimeVersion) error
}
