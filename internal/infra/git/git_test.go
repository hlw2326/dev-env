package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestInitializeCommitAndStatus(t *testing.T) {
	binary, err := exec.LookPath("git")
	if err != nil {
		t.Skip("Git executable is not installed")
	}
	root := t.TempDir()
	manifest := filepath.Join(root, "package.json")
	if err := os.WriteFile(manifest, []byte("{\n  \"name\": \"sample\",\n  \"version\": \"1.0.0\"\n}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	client := NewClient()
	ctx := context.Background()
	if err := client.Initialize(ctx, root); err != nil {
		t.Fatal(err)
	}
	configureGit(t, binary, root, "user.name", "DevEnv Test")
	configureGit(t, binary, root, "user.email", "devenv-test@example.invalid")
	if err := client.Commit(ctx, root, "node", "initial project", false); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("sample\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := client.Commit(ctx, root, "node", "release", true); err != nil {
		t.Fatal(err)
	}
	status, err := client.Status(ctx, root, "node")
	if err != nil {
		t.Fatal(err)
	}
	if !status.Available || !status.Repository || status.Changed != 0 {
		t.Fatalf("expected a clean Git repository, got %+v", status)
	}
	if status.Version != "1.0.01" || status.NextVersion != "1.0.02" || !status.VersionCanBump {
		t.Fatalf("unexpected bumped version status: %+v", status)
	}
	if status.LastCommit == "" || status.CommitMessage != "release" {
		t.Fatalf("expected latest commit details, got %+v", status)
	}
}

func configureGit(t *testing.T, binary, root, key, value string) {
	t.Helper()
	command := exec.Command(binary, "-C", root, "config", key, value)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git config %s failed: %v: %s", key, err, output)
	}
}
