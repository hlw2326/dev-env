package ports

import (
	"context"

	"devenv/internal/domain/setting"
	"devenv/internal/domain/site"
	"devenv/internal/domain/software"
)

type SoftwareRepository interface {
	List(ctx context.Context) ([]software.Software, error)
	ListByKind(ctx context.Context, kind string) ([]software.Software, error)
	Get(ctx context.Context, id int64) (*software.Software, error)
	GetByKindVersion(ctx context.Context, kind, version string) (*software.Software, error)
	GetMain(ctx context.Context, kind string) (*software.Software, error)
	Create(ctx context.Context, item *software.Software) error
	Update(ctx context.Context, item *software.Software) error
	SetMain(ctx context.Context, kind string, id int64) error
	Delete(ctx context.Context, id int64) error
}

type SiteRepository interface {
	List(ctx context.Context) ([]site.Site, error)
	Get(ctx context.Context, id int64) (*site.Site, error)
	GetByName(ctx context.Context, name string) (*site.Site, error)
	ListBySoftware(ctx context.Context, softwareID int64) ([]site.Site, error)
	ListMissingSoftware(ctx context.Context, kind string) ([]site.Site, error)
	Create(ctx context.Context, item *site.Site) error
	Update(ctx context.Context, item *site.Site) error
	BindSoftware(ctx context.Context, siteID, softwareID int64) error
	Delete(ctx context.Context, id int64) error
}

type SettingRepository interface {
	Load(ctx context.Context) (setting.Settings, error)
	Save(ctx context.Context, settings setting.Settings) error
}

type Route struct {
	ID    int64
	Hosts []string
	Kind  string
	Path  string
	PHP   string
	Port  int
	Log   string
}

type Gateway interface {
	Start(port int) error
	Stop() error
	SetRoutes(routes []Route)
	Port() int
}

type Options struct {
	Binary      string
	Args        []string
	Directory   string
	Environment []string
	Log         string
	OnExit      func(key string, message string)
}

type Status struct {
	Running bool
	PID     int
	Message string
}

type Runner interface {
	Start(key string, options Options) (int, error)
	Stop(key string) error
	Status(key string) Status
	Tail(key string, limit int) []string
}

type Catalog interface {
	Versions(ctx context.Context, kind string) ([]software.Version, error)
	Cached(kind string) []software.Version
}

type Request struct {
	Kind     string
	Version  string
	URL      string
	Checksum string
	Target   string
	Cache    string
}

type Installer interface {
	Install(ctx context.Context, request Request, report func(progress int, message string)) error
}

type HostsWriter interface {
	Sync(hosts []string) error
}

type PortChecker interface {
	Available(port int) bool
	Listening(port int) bool
}

type EventPublisher interface {
	Emit(event string, payload interface{})
}
