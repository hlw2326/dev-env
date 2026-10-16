package uninstall

import (
	"os"
	"path/filepath"
	"testing"

	"devenv/internal/infra/sqlite"
	"devenv/internal/shared/paths"
)

func fixture(t *testing.T) string {
	t.Helper()
	base := filepath.Join(t.TempDir(), paths.APP_FOLDER)
	for _, relative := range []string{
		"software/mysql/8.4",
		"software/redis/8.0",
		"dbdata/mysql-8.4",
		"logs",
		"cache",
		"data/nginx",
		"data/backups",
		"data/services",
	} {
		folder := filepath.Join(base, filepath.FromSlash(relative))
		if err := os.MkdirAll(folder, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(folder, "keep.txt"), []byte("test"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	db, err := sqlite.Open(paths.Database(base), sqlite.Options{})
	if err != nil {
		t.Fatal(err)
	}
	for _, environment := range []Environment{{Kind: "mysql", Version: "8.4"}, {Kind: "redis", Version: "8.0"}} {
		_, err := db.Exec(`INSERT INTO system_software(kind,version,path,state,created,updated) VALUES (?,?,?,'installed','now','now')`, environment.Kind, environment.Version, paths.Software(base, environment.Kind, environment.Version))
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`INSERT INTO system_project(name,kind,root,software,created,updated) VALUES ('test','php','C:/project',1,'now','now')`); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	return base
}

func assertExists(t *testing.T, path string, expected bool) {
	t.Helper()
	_, err := os.Stat(path)
	if expected && err != nil {
		t.Fatalf("expected retained %s: %v", path, err)
	}
	if !expected && !os.IsNotExist(err) {
		t.Fatalf("expected deleted %s, got %v", path, err)
	}
}

func TestCleanSelectedEnvironment(t *testing.T) {
	base := fixture(t)
	if err := Clean(base, Options{Environments: []Environment{{Kind: "mysql", Version: "8.4"}}}); err != nil {
		t.Fatal(err)
	}
	assertExists(t, paths.Software(base, "mysql", "8.4"), false)
	assertExists(t, paths.Software(base, "redis", "8.0"), true)
	assertExists(t, filepath.Join(base, "dbdata", "mysql-8.4", "keep.txt"), true)
	assertExists(t, filepath.Join(base, "logs", "keep.txt"), true)

	db, err := sqlite.Open(paths.Database(base), sqlite.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var count, unbound int
	if err := db.QueryRow("SELECT COUNT(*) FROM system_software").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM system_project WHERE software IS NULL AND autostart = 0").Scan(&unbound); err != nil {
		t.Fatal(err)
	}
	if count != 1 || unbound != 1 {
		t.Fatalf("records=%d unbound=%d", count, unbound)
	}
}

func TestCleanUserDataRetainsInstallationsAndDatabaseData(t *testing.T) {
	base := fixture(t)
	if err := Clean(base, Options{UserData: true}); err != nil {
		t.Fatal(err)
	}
	for _, folder := range []string{"logs", "cache", "data/nginx", "data/backups", "data/services"} {
		assertExists(t, filepath.Join(base, filepath.FromSlash(folder)), false)
	}
	assertExists(t, paths.Software(base, "mysql", "8.4"), true)
	assertExists(t, filepath.Join(base, "dbdata", "mysql-8.4", "keep.txt"), true)

	db, err := sqlite.Open(paths.Database(base), sqlite.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var projects, software, settings int
	for query, destination := range map[string]*int{
		"SELECT COUNT(*) FROM system_project":     &projects,
		"SELECT COUNT(*) FROM system_software":    &software,
		"SELECT COUNT(*) FROM system_app_setting": &settings,
	} {
		if err := db.QueryRow(query).Scan(destination); err != nil {
			t.Fatal(err)
		}
	}
	if projects != 0 || software != 2 || settings != 4 {
		t.Fatalf("projects=%d software=%d settings=%d", projects, software, settings)
	}
}

func TestCleanDatabaseDataAndDefaultPreservation(t *testing.T) {
	base := fixture(t)
	if err := Clean(base, Options{}); err != nil {
		t.Fatal(err)
	}
	assertExists(t, filepath.Join(base, "logs", "keep.txt"), true)
	if err := Clean(base, Options{DatabaseData: true}); err != nil {
		t.Fatal(err)
	}
	assertExists(t, filepath.Join(base, paths.DBDATA_FOLDER), false)
	assertExists(t, paths.Software(base, "mysql", "8.4"), true)
}

func TestCleanAllAndRejectUnsafePaths(t *testing.T) {
	base := fixture(t)
	for _, invalid := range []Environment{{Kind: "../outside", Version: "8"}, {Kind: "mysql", Version: ".."}, {Kind: "mysql", Version: "C:/data"}} {
		if _, _, err := Plan(base, Options{Environments: []Environment{invalid}}); err == nil {
			t.Fatalf("accepted %#v", invalid)
		}
	}
	if _, _, err := Plan(filepath.Dir(base), Options{UserData: true}); err == nil {
		t.Fatal("accepted workspace root")
	}
	if err := Clean(base, Options{AllEnvironments: true, DatabaseData: true, UserData: true}); err != nil {
		t.Fatal(err)
	}
	assertExists(t, filepath.Join(base, paths.SOFTWARE_FOLDER), false)
	assertExists(t, filepath.Join(base, paths.DBDATA_FOLDER), false)
}

func TestRejectLinkWithoutDeletingTarget(t *testing.T) {
	base := fixture(t)
	outside := t.TempDir()
	link := filepath.Join(base, paths.LOG_FOLDER, "external")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if err := Clean(base, Options{UserData: true}); err == nil {
		t.Fatal("accepted linked target")
	}
	assertExists(t, outside, true)
	assertExists(t, filepath.Join(base, paths.LOG_FOLDER, "keep.txt"), true)
}
