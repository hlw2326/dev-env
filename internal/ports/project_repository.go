package ports

import (
	"context"

	"devenv/internal/domain/project"
)

type ProjectRepository interface {
	Create(ctx context.Context, item project.Project) (int64, error)
	GetByID(ctx context.Context, id int64) (*project.Project, error)
	GetByPathKey(ctx context.Context, pathKey string) (*project.Project, error)
	List(ctx context.Context) ([]project.Project, error)
	Update(ctx context.Context, item project.Project) error
	Delete(ctx context.Context, id int64) error
}
