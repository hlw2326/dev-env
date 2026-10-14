package ports

import (
	"context"

	"devenv/internal/domain/runtime"
	"devenv/internal/domain/service"
)

type ServiceRepository interface {
	Create(ctx context.Context, item service.Instance) (int64, error)
	GetByID(ctx context.Context, id int64) (*service.Instance, error)
	ListAll(ctx context.Context) ([]service.Instance, error)
	ListByType(ctx context.Context, sType runtime.RuntimeType) ([]service.Instance, error)
	Update(ctx context.Context, item service.Instance) error
	Delete(ctx context.Context, id int64) error
	CountServicesUsingRuntime(ctx context.Context, runtimeID int64) (int, error)
}
