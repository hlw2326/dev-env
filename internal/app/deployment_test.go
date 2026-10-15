package app

import (
	"context"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"devenv/internal/domain/deployment"
	"devenv/internal/domain/project"
	"devenv/internal/infra/sqlite"
	"devenv/internal/ports"
)

type testDeploymentProvider struct {
	runs    atomic.Int32
	started chan struct{}
}

func (p *testDeploymentProvider) Preview(_ context.Context, config deployment.Config, root string) (deployment.Plan, error) {
	return deployment.Plan{ProjectID: config.ProjectID, Provider: config.Provider, Source: root, Target: config.Remote, Files: []deployment.File{{Path: "index.php", Size: 1, Digest: "test"}}}, nil
}
func (p *testDeploymentProvider) Test(context.Context, deployment.Config, string, deployment.Credentials) error {
	return nil
}
func (p *testDeploymentProvider) Run(ctx context.Context, _ deployment.Config, _ string, _ deployment.Plan, _ deployment.Credentials, report func(int, string)) error {
	p.runs.Add(1)
	close(p.started)
	report(0, "upload started")
	<-ctx.Done()
	return ctx.Err()
}

func TestDeploymentOnlyStartsWithFreshExplicitPreview(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	database, err := sqlite.Open(filepath.Join(root, "test.db"), sqlite.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	projects := sqlite.NewProjectRepository(database.DB)
	repository := sqlite.NewDeploymentRepository(database.DB)
	item := project.Project{Name: "manual", Kind: project.KIND_PHP, Root: root}
	if err := projects.Create(ctx, &item); err != nil {
		t.Fatal(err)
	}
	provider := &testDeploymentProvider{started: make(chan struct{})}
	service := NewDeploymentService(projects, repository, map[string]ports.DeploymentProvider{deployment.PROVIDER_SFTP: provider})
	defer service.Close()
	config, err := service.Config(ctx, item.ID)
	if err != nil {
		t.Fatal(err)
	}
	config.Host = "test.invalid"
	config.Username = "deploy"
	config.Fingerprint = "SHA256:" + strings.Repeat("A", 43)
	config.Remote = "/deploy/project"
	if err := service.Save(ctx, config); err != nil {
		t.Fatal(err)
	}
	plan, err := service.Preview(ctx, item.ID)
	if err != nil {
		t.Fatal(err)
	}
	if provider.runs.Load() != 0 {
		t.Fatal("save or preview started deployment")
	}
	if err := service.Start(ctx, item.ID, "", deployment.Credentials{}); err == nil {
		t.Fatal("empty preview accepted")
	}
	config.Remote = "/deploy/other"
	if err := service.Save(ctx, config); err != nil {
		t.Fatal(err)
	}
	if err := service.Start(ctx, item.ID, plan.Token, deployment.Credentials{}); err == nil {
		t.Fatal("outdated preview accepted")
	}
	plan, err = service.Preview(ctx, item.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Start(ctx, item.ID, plan.Token, deployment.Credentials{}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-provider.started:
	case <-time.After(2 * time.Second):
		t.Fatal("async task did not start")
	}
	if err := service.Start(ctx, item.ID, plan.Token, deployment.Credentials{}); err == nil {
		t.Fatal("duplicate task accepted")
	}
	if err := service.Save(ctx, config); err == nil {
		t.Fatal("config changed during deployment")
	}
	if err := service.Cancel(item.ID); err != nil {
		t.Fatal(err)
	}
	service.Close()
	history, err := service.History(ctx, item.ID)
	if err != nil || len(history) != 1 || history[0].State != deployment.STATE_CANCELLED {
		t.Fatalf("wrong final state: %+v %v", history, err)
	}
	if provider.runs.Load() != 1 {
		t.Fatal("unexpected retries")
	}
}
