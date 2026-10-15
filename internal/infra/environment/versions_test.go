package environment

import (
	"reflect"
	"regexp"
	"testing"

	"devenv/internal/domain/software"
)

func TestEveryMarketKindHasSafeInspection(t *testing.T) {
	for _, kind := range software.AllKinds() {
		probes := software.InspectionProbes(kind)
		if len(probes) == 0 {
			t.Fatalf("missing inspection for %s", kind)
		}
		for _, probe := range probes {
			if probe.Name == "" || probe.Role == "" || len(probe.Arguments) == 0 {
				t.Fatalf("incomplete probe: %+v", probe)
			}
			if _, err := regexp.Compile(probe.Pattern); err != nil {
				t.Fatal(err)
			}
			safeExit := false
			for _, argument := range probe.Arguments {
				switch argument {
				case "--version", "-v", "-version", "version", "-h", "--list-sdks", "--list-runtimes":
					safeExit = true
				case "-I", "-S", "-n", "--no-defaults", "--no-plugins", "--no-scripts", "--no-interaction", "--no-ansi":
				default:
					t.Fatalf("unapproved probe argument: %s %s", kind, argument)
				}
			}
			if !safeExit {
				t.Fatalf("probe may start application: %+v", probe)
			}
		}
	}
}

func TestParseSoftwareVersions(t *testing.T) {
	fixtures := []struct {
		kind     string
		index    int
		output   string
		versions []string
	}{
		{software.KIND_NODE, 0, "v24.22.0\r\n", []string{"24.22.0"}},
		{software.KIND_PHP, 0, "PHP 8.3.35 (cli)\r\nZend Engine v4.3.35", []string{"8.3.35"}},
		{software.KIND_PYTHON, 0, "Python 3.13.15", []string{"3.13.15"}},
		{software.KIND_GO, 0, "go version go1.27.0 windows/amd64", []string{"1.27.0"}},
		{software.KIND_COMPOSER, 0, "Composer version 2.10.3 2026-01-01\r\nPHP version 8.3.35", []string{"2.10.3"}},
		{software.KIND_MYSQL, 0, "mysql  Ver 14.14 Distrib 5.7.44, for Win64", []string{"5.7.44"}},
		{software.KIND_MYSQL, 1, "C:\\mysql\\mysqld.exe  Ver 8.4.11 for Win64 (MySQL Community Server)", []string{"8.4.11"}},
		{software.KIND_MARIADB, 0, "mariadb  Ver 15.1 Distrib 11.8.2-MariaDB", []string{"11.8.2"}},
		{software.KIND_POSTGRESQL, 0, "psql (PostgreSQL) 18.6", []string{"18.6"}},
		{software.KIND_POSTGRESQL, 2, "pg_ctl (PostgreSQL) 18.6", []string{"18.6"}},
		{software.KIND_MONGODB, 0, "db version v8.0.10\nBuild Info: {}", []string{"8.0.10"}},
		{software.KIND_REDIS, 0, "Redis server v=8.10.2 sha=abc bits=64", []string{"8.10.2"}},
		{software.KIND_REDIS, 1, "redis-cli 8.10.2", []string{"8.10.2"}},
		{software.KIND_MEMCACHED, 0, "memcached 1.6.39\n-p <num> TCP port", []string{"1.6.39"}},
		{software.KIND_JAVA, 0, "openjdk version \"21.0.8\" 2025-07-15\nOpenJDK Runtime Environment", []string{"21.0.8"}},
		{software.KIND_JAVA, 1, "javac 21.0.8", []string{"21.0.8"}},
		{software.KIND_DOTNET, 0, "8.0.404 [C:\\dotnet\\sdk]\r\n9.0.102 [C:\\dotnet\\sdk]\r\n", []string{"8.0.404", "9.0.102"}},
		{software.KIND_DOTNET, 1, "Microsoft.NETCore.App 8.0.15 [C:\\dotnet]\r\nMicrosoft.AspNetCore.App 8.0.15 [C:\\dotnet]", []string{"8.0.15"}},
		{software.KIND_DENO, 0, "deno 2.3.2 (stable)\nv8 13.4.1\ntypescript 5.8.3", []string{"2.3.2"}},
		{software.KIND_BUN, 0, "1.2.14", []string{"1.2.14"}},
		{software.KIND_NGINX, 0, "nginx version: nginx/1.30.6", []string{"1.30.6"}},
		{software.KIND_APACHE, 0, "Server version: Apache/2.4.63 (Win64)\nServer built: abc", []string{"2.4.63"}},
		{software.KIND_CADDY, 0, "v2.10.0 h1:hash=", []string{"2.10.0"}},
		{software.KIND_SQLITE, 0, "3.54.0 2026-01-01 abc (64-bit)", []string{"3.54.0"}},
		{software.KIND_GRADLE, 0, "Welcome!\n-----\nGradle 8.14.2\n-----\nJVM: 21.0.8", []string{"8.14.2"}},
		{software.KIND_MAVEN, 0, "\x1b[1mApache Maven 3.9.10\x1b[0m (abc)\nJava version: 21.0.8", []string{"3.9.10"}},
	}
	for _, fixture := range fixtures {
		t.Run(fixture.kind+fixture.output[:2], func(t *testing.T) {
			probe := software.InspectionProbes(fixture.kind)[fixture.index]
			versions, err := parseVersions(probe, fixture.output)
			if err != nil || !reflect.DeepEqual(versions, fixture.versions) {
				t.Fatalf("%s: %v, %v", fixture.kind, versions, err)
			}
		})
	}
	if _, err := parseVersions(software.InspectionProbes(software.KIND_MYSQL)[0], "mysql Ver 15.1 Distrib 11.8.2-MariaDB"); err == nil {
		t.Fatal("MariaDB misidentified as MySQL")
	}
	if _, err := parseVersions(software.InspectionProbes(software.KIND_DOTNET)[0], ""); err == nil {
		t.Fatal("runtime-only .NET reported as SDK")
	}
}
