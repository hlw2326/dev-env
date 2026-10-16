package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"devenv/internal/domain/project"
	"devenv/internal/domain/software"
	"devenv/internal/infra/sqlite"
	"devenv/internal/ports"
	appErrors "devenv/internal/shared/errors"
)

type projectGateway struct {
	port int
}

type blockingProjectGateway struct {
	projectGateway
	started chan struct{}
	release chan struct{}
}

func (g *blockingProjectGateway) Start(port int) error {
	close(g.started)
	<-g.release
	g.port = port
	return nil
}

type projectEvents struct {
	events chan string
}

func (e *projectEvents) Emit(event string, _ interface{}) {
	e.events <- event
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

func TestCreateReturnsBeforeBackgroundGatewayStartCompletes(t *testing.T) {
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
	softwareRepository := sqlite.NewSoftwareRepository(database.DB)
	settings := sqlite.NewSettingRepository(database.DB)
	nginxDirectory := filepath.Join(home, "software", "nginx")
	if err := os.MkdirAll(nginxDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nginxDirectory, "nginx.exe"), []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}
	nginx := software.Software{
		Kind:    software.KIND_NGINX,
		Version: "test",
		Path:    nginxDirectory,
		State:   software.STATE_INSTALLED,
		Main:    true,
	}
	if err := softwareRepository.Create(context.Background(), &nginx); err != nil {
		t.Fatal(err)
	}

	gateway := &blockingProjectGateway{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	events := &projectEvents{events: make(chan string, 1)}
	service := NewProjectService(home, projects, softwareRepository, settings, gateway, nil, nil, nil, events)
	result := make(chan error, 1)
	go func() {
		_, createError := service.Create(context.Background(), project.CreateInput{
			Kind: project.KIND_STATIC,
			Name: "demo",
			Host: "demo.localhost",
			Root: filepath.Join(home, "projects", "demo"),
		})
		result <- createError
	}()

	select {
	case err := <-result:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("project creation waited for the background gateway start")
	}
	select {
	case <-gateway.started:
	case <-time.After(time.Second):
		t.Fatal("background gateway start did not begin")
	}
	close(gateway.release)
	select {
	case event := <-events.events:
		if event != "project:status-changed" {
			t.Fatalf("unexpected event: %s", event)
		}
	case <-time.After(time.Second):
		t.Fatal("background project start did not publish status")
	}
}
