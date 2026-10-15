package sqlite

import (
	"context"
	"devenv/internal/domain/deployment"
	"devenv/internal/domain/project"
	"path/filepath"
	"testing"
)

func TestDeploymentConfigHistoryAndInterruptedRecovery(t *testing.T) {
	ctx := context.Background()
	db, err := Open(filepath.Join(t.TempDir(), "test.db"), Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	projects := NewProjectRepository(db.DB)
	item := project.Project{Name: "publication", Kind: "php", Root: t.TempDir()}
	if err := projects.Create(ctx, &item); err != nil {
		t.Fatal(err)
	}
	repository := NewDeploymentRepository(db.DB)
	if value, err := repository.Load(ctx, item.ID); err != nil || value != nil {
		t.Fatalf("expected no stored config: %+v %v", value, err)
	}
	config := deployment.Default(item.ID, item.Kind)
	config.PrivateKey = filepath.Join(t.TempDir(), "deploy-key")
	if err := repository.Save(ctx, config); err != nil {
		t.Fatal(err)
	}
	loaded, err := repository.Load(ctx, item.ID)
	if err != nil || *loaded != config {
		t.Fatalf("config mismatch: %+v %v", loaded, err)
	}
	task := deployment.Task{ID: "test-task", ProjectID: item.ID, Provider: deployment.PROVIDER_SFTP, State: deployment.STATE_RUNNING, Total: 5, Created: "2026-10-15T00:00:00Z"}
	if err := repository.Record(ctx, task); err != nil {
		t.Fatal(err)
	}
	if err := repository.Interrupt(ctx); err != nil {
		t.Fatal(err)
	}
	history, err := repository.History(ctx, item.ID)
	if err != nil || len(history) != 1 || history[0].State != deployment.STATE_INTERRUPTED || history[0].Finished == "" {
		t.Fatalf("invalid recovered history: %+v %v", history, err)
	}
	if err := projects.Delete(ctx, item.ID); err != nil {
		t.Fatal(err)
	}
	history, err = repository.History(ctx, item.ID)
	if err != nil || len(history) != 0 || history == nil {
		t.Fatalf("history not removed with project: %+v %v", history, err)
	}
}
