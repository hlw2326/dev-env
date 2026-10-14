package ports

import (
	"context"

	"devenv/internal/domain/task"
)

type TaskRepository interface {
	Create(ctx context.Context, item task.Task) (int64, error)
	GetByID(ctx context.Context, id int64) (*task.Task, error)
	ListAll(ctx context.Context) ([]task.Task, error)
	Update(ctx context.Context, item task.Task) error
}
