//go:build windows

package app

import (
	"os"
	"path/filepath"
	"testing"

	"devenv/internal/domain/software"
	"devenv/internal/infra/network"
	"devenv/internal/infra/windowsservice"
	"devenv/internal/shared/paths"
)

func TestServiceIdentityRequiresInstalledDirectory(t *testing.T) {
	for _, entry := range []struct{ kind, binary string }{
		{software.KIND_MYSQL, "bin/mysqld.exe"},
		{software.KIND_MARIADB, "bin/mariadbd.exe"},
		{software.KIND_POSTGRESQL, "bin/postgres.exe"},
		{software.KIND_MONGODB, "bin/mongod.exe"},
		{software.KIND_REDIS, "redis-server.exe"},
		{software.KIND_MEMCACHED, "memcached.exe"},
	} {
		t.Run(entry.kind, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "installed", "current")
			item := software.Software{Kind: entry.kind, Path: root}
			if !matchesInstalledServicePath(item, filepath.Join(root, filepath.FromSlash(entry.binary))) {
				t.Fatal("correct installation was rejected")
			}
			for _, other := range []string{
				filepath.Join(root+"-other", filepath.FromSlash(entry.binary)),
				filepath.Join(filepath.Dir(root), "other-version", filepath.FromSlash(entry.binary)),
				filepath.Join(root, "other.exe"),
				"",
			} {
				if matchesInstalledServicePath(item, other) {
					t.Fatalf("unrelated executable accepted: %s", other)
				}
			}
		})
	}
}

// This check only reads the installed MySQL service; it never stops a database.
func TestInstalledMySQLServiceIdentity(t *testing.T) {
	if os.Getenv("DEVENV_TEST_SERVICE_IDENTITY") != "1" {
		t.Skip("requires running installed MySQL service")
	}
	item := software.Software{Kind: software.KIND_MYSQL, Version: "8.4.11", Path: paths.Software(paths.Base(), software.KIND_MYSQL, "8.4.11"), Port: 3306}
	service := &SoftwareService{probe: network.NewChecker()}
	status := service.externalStatus(item)
	if !status.Running || status.PID == 0 {
		t.Fatal("installed MySQL was not recognized")
	}
	registeredPath, err := windowsservice.ExecutablePath(managedServiceName(item))
	if err != nil {
		t.Fatal(err)
	}
	if !matchesInstalledServicePath(item, registeredPath) || !windowsservice.OwnsProcess(managedServiceName(item), status.PID, registeredPath) {
		t.Fatal("SCM ownership did not identify the same installed MySQL process")
	}
	item.Path += "-different-installation"
	if service.externalStatus(item).Running {
		t.Fatal("same port was attributed to a different installation")
	}
}
