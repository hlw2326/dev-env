package ports

import (
	"context"

	"devenv/internal/domain/host"
)

type HostRepository interface {
	Create(ctx context.Context, item host.VirtualHost) (int64, error)
	GetByID(ctx context.Context, id int64) (*host.VirtualHost, error)
	GetByDomain(ctx context.Context, domain string) (*host.VirtualHost, error)
	ListAll(ctx context.Context) ([]host.VirtualHost, error)
	ListByProjectID(ctx context.Context, projectID int64) ([]host.VirtualHost, error)
	Update(ctx context.Context, item host.VirtualHost) error
	Delete(ctx context.Context, id int64) error
}
