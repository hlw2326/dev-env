package ports

import (
	"context"

	"devenv/internal/domain/runtime"
)

type RuntimeRepository interface {
	Create(ctx context.Context, item runtime.RuntimeVersion) (int64, error)
	GetByID(ctx context.Context, id int64) (*runtime.RuntimeVersion, error)
	ListByType(ctx context.Context, runtimeType runtime.RuntimeType) ([]runtime.RuntimeVersion, error)
	ListAll(ctx context.Context) ([]runtime.RuntimeVersion, error)
	Update(ctx context.Context, item runtime.RuntimeVersion) error
	Delete(ctx context.Context, id int64) error
	CountProjectsUsingRuntime(ctx context.Context, runtimeID int64) (int, error)
}
