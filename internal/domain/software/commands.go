package software

// CommandDefinition describes a command exposed by an installed software
// version through the DevEnv user command directory.
type CommandDefinition struct {
	Name            string
	RelativePath    string
	RequiresPHP     bool
	RequiresJava    bool
	UseCommandShell bool
}

// GlobalCommandDefinitions returns the commands that can be used from any
// newly opened Windows command prompt. Missing optional executables are
// skipped during synchronization, so package layouts can differ by version.
func GlobalCommandDefinitions(kind string) []CommandDefinition {
	switch kind {
	case KIND_PHP:
		return []CommandDefinition{
			{Name: "php", RelativePath: "php.exe"},
			{Name: "php-cgi", RelativePath: "php-cgi.exe"},
		}
	case KIND_NODE:
		return []CommandDefinition{
			{Name: "node", RelativePath: "node.exe"},
			{Name: "npm", RelativePath: "npm.cmd", UseCommandShell: true},
			{Name: "npx", RelativePath: "npx.cmd", UseCommandShell: true},
			{Name: "corepack", RelativePath: "corepack.cmd", UseCommandShell: true},
		}
	case KIND_GO:
		return []CommandDefinition{
			{Name: "go", RelativePath: "bin/go.exe"},
			{Name: "gofmt", RelativePath: "bin/gofmt.exe"},
		}
	case KIND_PYTHON:
		return []CommandDefinition{
			{Name: "python", RelativePath: "python.exe"},
			{Name: "pythonw", RelativePath: "pythonw.exe"},
			{Name: "pip", RelativePath: "Scripts/pip.exe"},
			{Name: "pip3", RelativePath: "Scripts/pip3.exe"},
		}
	case KIND_COMPOSER:
		return []CommandDefinition{{Name: "composer", RelativePath: "composer.phar", RequiresPHP: true}}
	case KIND_MYSQL:
		return []CommandDefinition{
			{Name: "mysql", RelativePath: "bin/mysql.exe"},
			{Name: "mysqldump", RelativePath: "bin/mysqldump.exe"},
			{Name: "mysqladmin", RelativePath: "bin/mysqladmin.exe"},
		}
	case KIND_MARIADB:
		return []CommandDefinition{
			{Name: "mariadb", RelativePath: "bin/mariadb.exe"},
			{Name: "mariadb-dump", RelativePath: "bin/mariadb-dump.exe"},
			{Name: "mariadb-admin", RelativePath: "bin/mariadb-admin.exe"},
		}
	case KIND_POSTGRESQL:
		return []CommandDefinition{
			{Name: "psql", RelativePath: "bin/psql.exe"},
			{Name: "pg_dump", RelativePath: "bin/pg_dump.exe"},
			{Name: "pg_restore", RelativePath: "bin/pg_restore.exe"},
			{Name: "createdb", RelativePath: "bin/createdb.exe"},
			{Name: "createuser", RelativePath: "bin/createuser.exe"},
		}
	case KIND_MONGODB:
		return []CommandDefinition{
			{Name: "mongosh", RelativePath: "bin/mongosh.exe"},
			{Name: "mongodump", RelativePath: "bin/mongodump.exe"},
			{Name: "mongorestore", RelativePath: "bin/mongorestore.exe"},
		}
	case KIND_REDIS:
		return []CommandDefinition{
			{Name: "redis-cli", RelativePath: "redis-cli.exe"},
			{Name: "redis-server", RelativePath: "redis-server.exe"},
			{Name: "redis-benchmark", RelativePath: "redis-benchmark.exe"},
		}
	case KIND_MEMCACHED:
		return []CommandDefinition{{Name: "memcached", RelativePath: "memcached.exe"}}
	case KIND_JAVA:
		return []CommandDefinition{
			{Name: "java", RelativePath: "bin/java.exe"},
			{Name: "javac", RelativePath: "bin/javac.exe"},
			{Name: "jar", RelativePath: "bin/jar.exe"},
			{Name: "javadoc", RelativePath: "bin/javadoc.exe"},
			{Name: "jshell", RelativePath: "bin/jshell.exe"},
		}
	case KIND_DOTNET:
		return []CommandDefinition{{Name: "dotnet", RelativePath: "dotnet.exe"}}
	case KIND_DENO:
		return []CommandDefinition{{Name: "deno", RelativePath: "deno.exe"}}
	case KIND_BUN:
		return []CommandDefinition{
			{Name: "bun", RelativePath: "bun.exe"},
			{Name: "bunx", RelativePath: "bunx.exe"},
		}
	case KIND_NGINX:
		return []CommandDefinition{{Name: "nginx", RelativePath: "nginx.exe"}}
	case KIND_APACHE:
		return []CommandDefinition{
			{Name: "httpd", RelativePath: "bin/httpd.exe"},
			{Name: "ab", RelativePath: "bin/ab.exe"},
			{Name: "htpasswd", RelativePath: "bin/htpasswd.exe"},
		}
	case KIND_CADDY:
		return []CommandDefinition{{Name: "caddy", RelativePath: "caddy.exe"}}
	case KIND_SQLITE:
		return []CommandDefinition{{Name: "sqlite3", RelativePath: "sqlite3.exe"}}
	case KIND_GRADLE:
		return []CommandDefinition{{Name: "gradle", RelativePath: "bin/gradle.bat", RequiresJava: true, UseCommandShell: true}}
	case KIND_MAVEN:
		return []CommandDefinition{{Name: "mvn", RelativePath: "bin/mvn.cmd", RequiresJava: true, UseCommandShell: true}}
	default:
		return nil
	}
}

// HasGlobalCommands reports whether a software kind can expose command-line
// tools through the DevEnv user command directory.
func HasGlobalCommands(kind string) bool {
	return len(GlobalCommandDefinitions(kind)) > 0
}
