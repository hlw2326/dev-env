package software

// 检测清单只允许版本查询，不能复用服务启动参数或运行项目脚本。
type CommandProbe struct {
	Name      string
	Role      string
	Arguments []string
	Pattern   string
	Reject    string
	Home      string
	Relative  string
}

const (
	ROLE_RUNTIME   = "运行环境"
	ROLE_CLIENT    = "客户端"
	ROLE_SERVER    = "服务端"
	ROLE_TOOL      = "工具"
	ROLE_SDK       = "SDK"
	SOURCE_PATH    = "PATH"
	SOURCE_HOME    = "环境变量"
	SOURCE_SERVICE = "Windows 服务"
)

func InspectionProbes(kind string) []CommandProbe {
	switch kind {
	case KIND_NODE:
		return []CommandProbe{{Name: "node", Role: ROLE_RUNTIME, Arguments: []string{"--version"}, Pattern: `^v(\d+\.\d+\.\d+(?:-[\w.-]+)?)$`}}
	case KIND_PHP:
		return []CommandProbe{{Name: "php", Role: ROLE_RUNTIME, Arguments: []string{"-n", "-v"}, Pattern: `(?m)^PHP (\d+\.\d+\.\d+(?:[\w.-]+)?)`, Home: "PHP_HOME", Relative: "php.exe"}}
	case KIND_PYTHON:
		return []CommandProbe{
			{Name: "python", Role: ROLE_RUNTIME, Arguments: []string{"-I", "-S", "--version"}, Pattern: `(?m)^Python (\d+\.\d+\.\d+[\w.-]*)`, Home: "PYTHONHOME", Relative: "python.exe"},
			{Name: "python3", Role: ROLE_RUNTIME, Arguments: []string{"-I", "-S", "--version"}, Pattern: `(?m)^Python (\d+\.\d+\.\d+[\w.-]*)`},
		}
	case KIND_GO:
		return []CommandProbe{{Name: "go", Role: ROLE_RUNTIME, Arguments: []string{"version"}, Pattern: `(?m)^go version go(\d+\.\d+(?:\.\d+)?[\w.-]*)`, Home: "GOROOT", Relative: "bin/go.exe"}}
	case KIND_COMPOSER:
		return []CommandProbe{{Name: "composer", Role: ROLE_TOOL, Arguments: []string{"--no-plugins", "--no-scripts", "--no-interaction", "--no-ansi", "--version"}, Pattern: `(?m)^Composer version (\d+\.\d+\.\d+[\w.-]*)`}}
	case KIND_MYSQL:
		return []CommandProbe{
			{Name: "mysql", Role: ROLE_CLIENT, Arguments: []string{"--no-defaults", "--version"}, Pattern: `(?i)\b(?:Distrib|Ver)\s+(\d+\.\d+\.\d+)`, Reject: `(?i)MariaDB`},
			{Name: "mysqld", Role: ROLE_SERVER, Arguments: []string{"--no-defaults", "--version"}, Pattern: `(?i)\bVer\s+(\d+\.\d+\.\d+)`, Reject: `(?i)MariaDB`},
		}
	case KIND_MARIADB:
		return []CommandProbe{
			{Name: "mariadb", Role: ROLE_CLIENT, Arguments: []string{"--no-defaults", "--version"}, Pattern: `(\d+\.\d+\.\d+)-MariaDB`},
			{Name: "mariadbd", Role: ROLE_SERVER, Arguments: []string{"--no-defaults", "--version"}, Pattern: `(\d+\.\d+\.\d+)-MariaDB`},
			{Name: "mysqld", Role: ROLE_SERVER, Arguments: []string{"--no-defaults", "--version"}, Pattern: `(\d+\.\d+\.\d+)-MariaDB`},
		}
	case KIND_POSTGRESQL:
		return []CommandProbe{
			{Name: "psql", Role: ROLE_CLIENT, Arguments: []string{"--version"}, Pattern: `(?m)^psql \(PostgreSQL\) (\d+\.\d+(?:\.\d+)?)`},
			{Name: "postgres", Role: ROLE_SERVER, Arguments: []string{"--version"}, Pattern: `(?m)^postgres \(PostgreSQL\) (\d+\.\d+(?:\.\d+)?)`},
			{Name: "pg_ctl", Role: ROLE_SERVER, Arguments: []string{"--version"}, Pattern: `(?m)^pg_ctl \(PostgreSQL\) (\d+\.\d+(?:\.\d+)?)`},
		}
	case KIND_MONGODB:
		return []CommandProbe{{Name: "mongod", Role: ROLE_SERVER, Arguments: []string{"--version"}, Pattern: `(?m)^db version v(\d+\.\d+\.\d+[\w.-]*)`}}
	case KIND_REDIS:
		return []CommandProbe{
			{Name: "redis-server", Role: ROLE_SERVER, Arguments: []string{"--version"}, Pattern: `Redis server v=(\d+\.\d+\.\d+[\w.-]*)`},
			{Name: "redis-cli", Role: ROLE_CLIENT, Arguments: []string{"--version"}, Pattern: `(?m)^redis-cli (\d+\.\d+\.\d+[\w.-]*)`},
		}
	case KIND_MEMCACHED:
		return []CommandProbe{{Name: "memcached", Role: ROLE_SERVER, Arguments: []string{"-h"}, Pattern: `(?m)^memcached (\d+\.\d+\.\d+)`}}
	case KIND_JAVA:
		return []CommandProbe{
			{Name: "java", Role: ROLE_RUNTIME, Arguments: []string{"-version"}, Pattern: `(?:openjdk|java) version "([\d][\w.+_-]*)"`, Home: "JAVA_HOME", Relative: "bin/java.exe"},
			{Name: "javac", Role: "JDK 编译器", Arguments: []string{"-version"}, Pattern: `(?m)^javac ([\d][\w.+_-]*)`, Home: "JAVA_HOME", Relative: "bin/javac.exe"},
		}
	case KIND_DOTNET:
		return []CommandProbe{
			{Name: "dotnet", Role: ROLE_SDK, Arguments: []string{"--list-sdks"}, Pattern: `(?m)^(\d+\.\d+\.\d+[\w.-]*)\s+\[`, Home: "DOTNET_ROOT", Relative: "dotnet.exe"},
			{Name: "dotnet", Role: ".NET 运行时", Arguments: []string{"--list-runtimes"}, Pattern: `(?m)^Microsoft\.[\w.]+ (\d+\.\d+\.\d+[\w.-]*)\s+\[`, Home: "DOTNET_ROOT", Relative: "dotnet.exe"},
		}
	case KIND_DENO:
		return []CommandProbe{{Name: "deno", Role: ROLE_RUNTIME, Arguments: []string{"--version"}, Pattern: `(?m)^deno (\d+\.\d+\.\d+[\w.-]*)`, Home: "DENO_INSTALL", Relative: "bin/deno.exe"}}
	case KIND_BUN:
		return []CommandProbe{{Name: "bun", Role: ROLE_RUNTIME, Arguments: []string{"--version"}, Pattern: `^(\d+\.\d+\.\d+[\w.-]*)$`, Home: "BUN_INSTALL", Relative: "bin/bun.exe"}}
	case KIND_NGINX:
		return []CommandProbe{{Name: "nginx", Role: ROLE_SERVER, Arguments: []string{"-v"}, Pattern: `nginx version: nginx/(\d+\.\d+\.\d+)`}}
	case KIND_APACHE:
		return []CommandProbe{{Name: "httpd", Role: ROLE_SERVER, Arguments: []string{"-v"}, Pattern: `Server version: Apache/(\d+\.\d+\.\d+)`}}
	case KIND_CADDY:
		return []CommandProbe{{Name: "caddy", Role: ROLE_SERVER, Arguments: []string{"version"}, Pattern: `(?m)^v(\d+\.\d+\.\d+[\w.-]*)`}}
	case KIND_SQLITE:
		return []CommandProbe{{Name: "sqlite3", Role: ROLE_TOOL, Arguments: []string{"--version"}, Pattern: `^(\d+\.\d+\.\d+)\s`}}
	case KIND_GRADLE:
		return []CommandProbe{{Name: "gradle", Role: ROLE_TOOL, Arguments: []string{"--version"}, Pattern: `(?m)^Gradle (\d+\.\d+(?:\.\d+)?[\w.-]*)`, Home: "GRADLE_HOME", Relative: "bin/gradle.bat"}}
	case KIND_MAVEN:
		return []CommandProbe{{Name: "mvn", Role: ROLE_TOOL, Arguments: []string{"--version"}, Pattern: `(?m)^Apache Maven (\d+\.\d+\.\d+[\w.-]*)`, Home: "MAVEN_HOME", Relative: "bin/mvn.cmd"}}
	default:
		return []CommandProbe{}
	}
}
