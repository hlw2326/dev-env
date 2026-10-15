package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"devenv/internal/domain/project"
	"devenv/internal/infra/sqlite"
	"devenv/internal/ports"
	appErrors "devenv/internal/shared/errors"
)

type projectGateway struct {
	port int
}

func (g *projectGateway) Start(port int) error {
	g.port = port
	return nil
}

func (g *projectGateway) Stop() error {
	g.port = 0
	return nil
}

func (g *projectGateway) SetRoutes([]ports.Route) error {
	return nil
}

func (g *projectGateway) Port() int {
	return g.port
}

func TestCreateRequiresInstalledNginx(t *testing.T) {
	home := t.TempDir()
	database, err := sqlite.Open(filepath.Join(home, "devenv.db"), sqlite.Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("close test database: %v", err)
		}
	})

	projects := sqlite.NewProjectRepository(database.DB)
	software := sqlite.NewSoftwareRepository(database.DB)
	settings := sqlite.NewSettingRepository(database.DB)
	service := NewProjectService(home, projects, software, settings, &projectGateway{}, nil, nil, nil, nil)
	root := filepath.Join(home, "projects", "demo")

	_, err = service.Create(context.Background(), project.CreateInput{
		Kind: project.KIND_STATIC,
		Name: "demo",
		Host: "demo.localhost",
		Root: root,
		Run:  "",
	})
	if err == nil {
		t.Fatal("expected missing Nginx to reject project creation")
	}
	var applicationError *appErrors.AppError
	if !errors.As(err, &applicationError) {
		t.Fatalf("expected application error, got %T", err)
	}
	if applicationError.Code != appErrors.RUNTIME_NOT_INSTALLED {
		t.Fatalf("expected %s, got %s", appErrors.RUNTIME_NOT_INSTALLED, applicationError.Code)
	}

	list, err := projects.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("expected no project records, got %d", len(list))
	}
	if _, err := os.Stat(root); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected no project directory, got %v", err)
	}
}
