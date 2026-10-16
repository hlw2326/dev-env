package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"devenv/internal/domain/project"
	"devenv/internal/domain/software"
	inframonitor "devenv/internal/infra/monitor"
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
	once    sync.Once
}

func (g *blockingProjectGateway) Start(port int) error {
	g.once.Do(func() { close(g.started) })
	<-g.release
	g.port = port
	return nil
}

type projectEvents struct{}

func (*projectEvents) Emit(string, interface{}) {}

func (g *projectGateway) Start(port int) error {
	g.port = port
	return nil
}

func (g *projectGateway) Stop() error {
	g.port = 0
	return nil
}

func (g *projectGateway) Restart() error {
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
	service := NewProjectService(home, projects, software, settings, &projectGateway{}, nil, nil, nil, nil, nil, inframonitor.NewCollector())
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

func TestCreateRejectsExistingProjectRoot(t *testing.T) {
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
	root := filepath.Join(home, "projects", "demo")
	if err := projects.Create(context.Background(), &project.Project{
		Name: "existing-project",
		Kind: project.KIND_PHP,
		Root: root,
	}); err != nil {
		t.Fatal(err)
	}

	service := NewProjectService(
		home,
		projects,
		sqlite.NewSoftwareRepository(database.DB),
		sqlite.NewSettingRepository(database.DB),
		&projectGateway{},
		nil,
		nil,
		nil,
		nil,
		nil,
		inframonitor.NewCollector(),
	)
	_, err = service.Create(context.Background(), project.CreateInput{
		Kind: project.KIND_PHP,
		Name: "new-project",
		Host: "new-project.localhost",
		Root: root,
	})
	if err == nil {
		t.Fatal("expected duplicate project root to be rejected")
	}
	var applicationError *appErrors.AppError
	if !errors.As(err, &applicationError) {
		t.Fatalf("expected application error, got %T", err)
	}
	if applicationError.Code != appErrors.ALREADY_EXISTS {
		t.Fatalf("expected %s, got %s", appErrors.ALREADY_EXISTS, applicationError.Code)
	}
	if _, err := os.Stat(root); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected duplicate project root to stay untouched, got %v", err)
	}
}

func TestCreateReturnsBeforeGatewayStarts(t *testing.T) {
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

	softwareRepository := sqlite.NewSoftwareRepository(database.DB)
	nginxPath := filepath.Join(home, "software", "nginx")
	if err := os.MkdirAll(nginxPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nginxPath, "nginx.exe"), []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}
	nginx := &software.Software{Kind: software.KIND_NGINX, Version: "test", Path: nginxPath, State: software.STATE_INSTALLED}
	if err := softwareRepository.Create(context.Background(), nginx); err != nil {
		t.Fatal(err)
	}
	if err := softwareRepository.SetMain(context.Background(), software.KIND_NGINX, nginx.ID); err != nil {
		t.Fatal(err)
	}

	gateway := &blockingProjectGateway{started: make(chan struct{}), release: make(chan struct{})}
	service := NewProjectService(
		home,
		sqlite.NewProjectRepository(database.DB),
		softwareRepository,
		sqlite.NewSettingRepository(database.DB),
		gateway,
		nil,
		nil,
		nil,
		&projectEvents{},
		nil,
		inframonitor.NewCollector(),
	)

	type createResult struct {
		view *ProjectView
		err  error
	}
	result := make(chan createResult, 1)
	go func() {
		view, createErr := service.Create(context.Background(), project.CreateInput{
			Kind: project.KIND_STATIC,
			Name: "demo",
			Host: "demo.localhost",
			Root: filepath.Join(home, "projects", "demo"),
		})
		result <- createResult{view: view, err: createErr}
	}()

	select {
	case created := <-result:
		if created.err != nil {
			t.Fatal(created.err)
		}
		if created.view.State != project.STATE_STARTING {
			t.Fatalf("expected %s state, got %s", project.STATE_STARTING, created.view.State)
		}
	case <-time.After(time.Second):
		t.Fatal("project creation waited for gateway startup")
	}

	select {
	case <-gateway.started:
	case <-time.After(time.Second):
		t.Fatal("background gateway startup did not begin")
	}
	close(gateway.release)
}

func TestRemoveNginxArtifacts(t *testing.T) {
	base := t.TempDir()
	item := project.Project{ID: 7, Name: "demo", Kind: project.KIND_PHP}
	hosts := filepath.Join(base, "data", "nginx", "vhosts")
	if err := os.MkdirAll(hosts, 0o755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(hosts, "website-7-demo.conf")
	environmentPath := filepath.Join(hosts, "website-7-demo.environment.conf")
	for _, target := range []string{configPath, environmentPath, configPath + ".devenv-backup"} {
		if err := os.WriteFile(target, []byte("generated"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if err := removeNginxArtifacts(base, item); err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{configPath, environmentPath, configPath + ".devenv-backup"} {
		if _, err := os.Stat(target); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("expected generated artifact to be removed: %s, got %v", target, err)
		}
	}
}
