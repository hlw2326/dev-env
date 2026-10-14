package ports

import (
	"context"

	"devenv/internal/domain/site"
)

type SiteRepository interface {
	Create(ctx context.Context, item *site.Site) error
	GetByID(ctx context.Context, id int64) (*site.Site, error)
	GetByDomain(ctx context.Context, domain string) (*site.Site, error)
	ListAll(ctx context.Context) ([]site.Site, error)
	Update(ctx context.Context, item *site.Site) error
	Delete(ctx context.Context, id int64) error
}

type GitManager interface {
	Clone(ctx context.Context, remoteURL, localPath, branch string) error
	Pull(ctx context.Context, repoPath string) error
	GetStatus(ctx context.Context, repoPath string) (*site.GitStatus, error)
	Commit(ctx context.Context, repoPath, message string, files []string) error
	Push(ctx context.Context, repoPath string) error
}

type NginxManager interface {
	GenerateConfig(item *site.Site, meta site.Metadata) string
	WriteConfig(ctx context.Context, siteID int64, domain string, content string) (string, error)
	ReadConfig(ctx context.Context, configPath string) (string, error)
	DeleteConfig(ctx context.Context, configPath string) error
	TestConfig(ctx context.Context, configContent string) (bool, string, error)
	Reload(ctx context.Context) error
}
