package software

const (
	KIND_PHP        = "php"
	KIND_NODE       = "nodejs"
	KIND_GO         = "go"
	KIND_PYTHON     = "python"
	KIND_COMPOSER   = "composer"
	KIND_MYSQL      = "mysql"
	KIND_MARIADB    = "mariadb"
	KIND_POSTGRESQL = "postgresql"
	KIND_MONGODB    = "mongodb"
	KIND_REDIS      = "redis"
	KIND_MEMCACHED  = "memcached"
	KIND_JAVA       = "java"
	KIND_DOTNET     = "dotnet"
	KIND_DENO       = "deno"
	KIND_BUN        = "bun"
	KIND_NGINX      = "nginx"
	KIND_APACHE     = "apache"
	KIND_CADDY      = "caddy"
	KIND_SQLITE     = "sqlite"
	KIND_GRADLE     = "gradle"
	KIND_MAVEN      = "maven"
)

const (
	STATE_INSTALLING = "installing"
	STATE_INSTALLED  = "installed"
	STATE_FAILED     = "failed"
)

type Software struct {
	ID        int64  `json:"id"`
	Kind      string `json:"kind"`
	Version   string `json:"version"`
	Path      string `json:"path"`
	State     string `json:"state"`
	Progress  int    `json:"progress"`
	Error     string `json:"error"`
	Main      bool   `json:"main"`
	Port      int    `json:"port"`
	Autostart bool   `json:"autostart"`
	Created   string `json:"created"`
	Updated   string `json:"updated"`
}

type Item struct {
	Kind        string `json:"kind"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Count       int    `json:"count"`
	Recommended string `json:"recommended"`
}

type Version struct {
	Version     string `json:"version"`
	Recommended bool   `json:"recommended"`
	Installed   bool   `json:"installed"`
	Note        string `json:"note"`
	URL         string `json:"-"`
	Checksum    string `json:"-"`
}

type ConfigItem struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

type Config struct {
	Items []ConfigItem `json:"items"`
	File  string       `json:"file"`
}

type Extension struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

func AllKinds() []string {
	return []string{
		KIND_PHP,
		KIND_NODE,
		KIND_GO,
		KIND_PYTHON,
		KIND_COMPOSER,
		KIND_MYSQL,
		KIND_MARIADB,
		KIND_POSTGRESQL,
		KIND_MONGODB,
		KIND_REDIS,
		KIND_MEMCACHED,
		KIND_JAVA,
		KIND_DOTNET,
		KIND_DENO,
		KIND_BUN,
		KIND_NGINX,
		KIND_APACHE,
		KIND_CADDY,
		KIND_SQLITE,
		KIND_GRADLE,
		KIND_MAVEN,
	}
}

func IsDaemon(kind string) bool {
	return kind == KIND_MYSQL ||
		kind == KIND_MARIADB ||
		kind == KIND_POSTGRESQL ||
		kind == KIND_MONGODB ||
		kind == KIND_REDIS ||
		kind == KIND_MEMCACHED
}

// IsService reports whether an installed software item participates in the
// service management screen and can expose a running state.
// Keep this list as the single extension point when new managed services are added.
func IsService(kind string) bool {
	return IsDaemon(kind) || kind == KIND_NGINX
}

func HasData(kind string) bool {
	return kind == KIND_MYSQL ||
		kind == KIND_MARIADB ||
		kind == KIND_POSTGRESQL ||
		kind == KIND_MONGODB
}

func IsValidKind(kind string) bool {
	for _, item := range AllKinds() {
		if item == kind {
			return true
		}
	}
	return false
}

func Name(kind string) string {
	switch kind {
	case KIND_PHP:
		return "PHP"
	case KIND_NODE:
		return "Node.js"
	case KIND_GO:
		return "Go"
	case KIND_PYTHON:
		return "Python"
	case KIND_COMPOSER:
		return "Composer"
	case KIND_MYSQL:
		return "MySQL"
	case KIND_MARIADB:
		return "MariaDB"
	case KIND_POSTGRESQL:
		return "PostgreSQL"
	case KIND_MONGODB:
		return "MongoDB"
	case KIND_REDIS:
		return "Redis"
	case KIND_MEMCACHED:
		return "Memcached"
	case KIND_JAVA:
		return "OpenJDK"
	case KIND_DOTNET:
		return ".NET SDK"
	case KIND_DENO:
		return "Deno"
	case KIND_BUN:
		return "Bun"
	case KIND_NGINX:
		return "Nginx"
	case KIND_APACHE:
		return "Apache"
	case KIND_CADDY:
		return "Caddy"
	case KIND_SQLITE:
		return "SQLite"
	case KIND_GRADLE:
		return "Gradle"
	case KIND_MAVEN:
		return "Maven"
	default:
		return kind
	}
}
