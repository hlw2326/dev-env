package ports

import (
	"context"

	"devenv/internal/domain/site"
)

type SiteRepository interface {
	Create(ctx context.Context, item *site.Site) error
	Get(ctx context.Context, id int64) (*site.Site, error)
	GetByDomain(ctx context.Context, domain string) (*site.Site, error)
	List(ctx context.Context) ([]site.Site, error)
	Update(ctx context.Context, item *site.Site) error
	Delete(ctx context.Context, id int64) error
}

type GitManager interface {
	Clone(ctx context.Context, remote, root, branch string) error
	Pull(ctx context.Context, root string) error
	Status(ctx context.Context, root string) (*site.GitState, error)
	Commit(ctx context.Context, root, message string, files []string) error
	Push(ctx context.Context, root string) error
}

type NginxManager interface {
	Generate(item *site.Site) string
	Write(ctx context.Context, id int64, domain, content string) (string, error)
	Read(ctx context.Context, path string) (string, error)
	Delete(ctx context.Context, path string) error
	Check(ctx context.Context, content string) (bool, string)
	Reload(ctx context.Context) error
}
