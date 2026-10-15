package app

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"devenv/internal/domain/software"
	"devenv/internal/infra/sqlite"
	"devenv/internal/shared/paths"
)

type environmentInspectorFixture struct {
	entries []software.CommandLocation
	failure error
}

func (fixture environmentInspectorFixture) Inspect(context.Context, string, string) (software.EnvironmentScan, error) {
	return software.EnvironmentScan{Entries: fixture.entries, Warnings: []string{}}, fixture.failure
}

func TestInspectCommandsSeparatesExternalInstallations(t *testing.T) {
	base := t.TempDir()
	ctx := context.Background()
	db, err := sqlite.Open(paths.Database(base), sqlite.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := sqlite.NewSoftwareRepository(db.DB)
	item := software.Software{Kind: software.KIND_NODE, Version: "24.22.0", Path: paths.Software(base, software.KIND_NODE, "24.22.0"), State: software.STATE_INSTALLED, Main: true}
	if err := repo.Create(ctx, &item); err != nil {
		t.Fatal(err)
	}
	external := filepath.Join(t.TempDir(), "node.exe")
	managed := filepath.Join(paths.Commands(base), "node.bat")
	for _, test := range []struct {
		name     string
		entries  []software.CommandLocation
		conflict bool
	}{
		{"external wins", []software.CommandLocation{{Command: "node", Role: software.ROLE_RUNTIME, Path: external, Versions: []string{"20.13.1"}, Active: true}, {Command: "node", Role: software.ROLE_RUNTIME, Path: managed, Versions: []string{item.Version}}}, true},
		{"same version external", []software.CommandLocation{{Command: "node", Role: software.ROLE_RUNTIME, Path: external, Versions: []string{item.Version}, Active: true}}, true},
		{"managed wins", []software.CommandLocation{{Command: "node", Role: software.ROLE_RUNTIME, Path: managed, Versions: []string{item.Version}, Active: true}, {Command: "node", Role: software.ROLE_RUNTIME, Path: external, Versions: []string{"20.13.1"}}}, false},
		{"managed executable", []software.CommandLocation{{Command: "node", Role: software.ROLE_RUNTIME, Path: filepath.Join(item.Path, "node.exe"), Versions: []string{item.Version}, Active: true}}, false},
		{"broken managed wrapper", []software.CommandLocation{{Command: "node", Role: software.ROLE_RUNTIME, Path: managed, Versions: []string{}, Active: true, Error: "not found"}}, true},
		{"missing PATH", []software.CommandLocation{}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			service := &SoftwareService{base: base, repo: repo, commands: environmentInspectorFixture{entries: test.entries}}
			result, err := service.InspectEnvironment(ctx, software.KIND_NODE, "")
			if err != nil {
				t.Fatal(err)
			}
			if result.Conflict != test.conflict || result.DefaultVersion != item.Version || result.Entries == nil {
				t.Fatalf("unexpected environment: %+v", result)
			}
			for _, entry := range result.Entries {
				if entry.Path == external && entry.Managed {
					t.Fatalf("external command treated as managed: %+v", entry)
				}
			}
			list, err := repo.List(ctx)
			if err != nil || len(list) != 1 {
				t.Fatalf("inspection changed installation records: %+v, %v", list, err)
			}
		})
	}
	if err := repo.Delete(ctx, item.ID); err != nil {
		t.Fatal(err)
	}
	service := &SoftwareService{base: base, repo: repo, commands: environmentInspectorFixture{entries: []software.CommandLocation{{Command: "node", Role: software.ROLE_RUNTIME, Path: external, Versions: []string{"20.13.1"}, Active: true}}}}
	result, err := service.InspectEnvironment(ctx, software.KIND_NODE, "")
	if err != nil || result.DefaultVersion != "" || result.Conflict {
		t.Fatalf("external-only install: %+v, %v", result, err)
	}
	service.commands = environmentInspectorFixture{failure: errors.New("denied")}
	if _, err := service.InspectEnvironment(ctx, software.KIND_NODE, ""); err == nil {
		t.Fatal("inspection error swallowed")
	}
	if _, err := service.InspectEnvironment(ctx, "invalid", ""); err == nil {
		t.Fatal("invalid kind accepted")
	}
}
