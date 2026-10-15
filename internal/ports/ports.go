package ports

import (
	"context"

	"devenv/internal/domain/project"
	"devenv/internal/domain/redis"
	"devenv/internal/domain/setting"
	"devenv/internal/domain/software"
)

const MAX_PORT = 65535

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

type ProjectRepository interface {
	List(ctx context.Context) ([]project.Project, error)
	Get(ctx context.Context, id int64) (*project.Project, error)
	GetByName(ctx context.Context, name string) (*project.Project, error)
	GetByRoot(ctx context.Context, root string) (*project.Project, error)
	ListBySoftware(ctx context.Context, softwareID int64) ([]project.Project, error)
	ListMissingSoftware(ctx context.Context, kind string) ([]project.Project, error)
	Create(ctx context.Context, item *project.Project) error
	Update(ctx context.Context, item *project.Project) error
	BindSoftware(ctx context.Context, projectID, softwareID int64) error
	Delete(ctx context.Context, id int64) error
}

type SettingRepository interface {
	Load(ctx context.Context) (setting.Settings, error)
	Save(ctx context.Context, settings setting.Settings) error
}

type Route struct {
	ID              int64
	Name            string
	Hosts           []string
	Kind            string
	Path            string
	PHP             string
	Nginx           string
	RewriteRule     string
	EnvironmentFile string
	Port            int
	Log             string
}

type Gateway interface {
	Start(port int) error
	Stop() error
	Restart() error
	SetRoutes(routes []Route) error
	Port() int
}

type Options struct {
	Binary      string
	Args        []string
	Directory   string
	Environment []string
	Log         string
	Detached    bool
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

// UserPathManager manages entries in the current Windows user's PATH.
// It never changes the machine-wide environment.
type UserPathManager interface {
	Ensure(path string) error
	Remove(path string) error
}

// CommandInspector only discovers commands; it never installs or imports them.
type CommandInspector interface {
	Inspect(ctx context.Context, kind, executable string) (software.EnvironmentScan, error)
}

type RedisBrowser interface {
	Databases(ctx context.Context, connection redis.Connection) ([]redis.Database, error)
	Keys(ctx context.Context, connection redis.Connection, query redis.KeyQuery) (redis.KeyPage, error)
	Value(ctx context.Context, connection redis.Connection, query redis.ValueQuery) (redis.Value, error)
}
