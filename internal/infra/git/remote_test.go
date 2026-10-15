package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"devenv/internal/ports"
)

func newTestRepository(t *testing.T) (*Client, string, string) {
	t.Helper()
	binary, err := exec.LookPath("git")
	if err != nil {
		t.Skip("Git is not installed")
	}
	root := t.TempDir()
	client := NewClient()
	if err := client.Initialize(context.Background(), root); err != nil {
		t.Fatal(err)
	}
	configureGit(t, binary, root, "user.name", "DevEnv Test")
	configureGit(t, binary, root, "user.email", "test@example.invalid")
	configureGit(t, binary, root, "commit.gpgsign", "false")
	configureGit(t, binary, root, "core.hooksPath", filepath.Join(root, ".git", "hooks"))
	return client, root, binary
}

func TestRemoteLifecycle(t *testing.T) {
	client, root, binary := newTestRepository(t)
	ctx := context.Background()
	bare := t.TempDir()
	if _, err := run(ctx, binary, bare, WRITE_TIMEOUT, "init", "--bare"); err != nil {
		t.Fatal(err)
	}
	remote := ports.GitRemote{Name: "origin", URL: filepath.ToSlash(bare)}
	if err := client.SaveRemote(ctx, root, remote); err != nil {
		t.Fatal(err)
	}
	status, err := client.Status(ctx, root, "go")
	if err != nil {
		t.Fatal(err)
	}
	if len(status.Remotes) != 1 || status.Remotes[0].URL != remote.URL || status.Remotes[0].PushURL != remote.URL {
		t.Fatalf("unexpected remotes: %+v", status.Remotes)
	}
	if err := client.Push(ctx, root, "origin"); err == nil {
		t.Fatal("push must reject unborn branch")
	}
	if err := os.WriteFile(filepath.Join(root, "VERSION"), []byte("1.0.0\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := client.Commit(ctx, root, "go", "initial", false); err != nil {
		t.Fatal(err)
	}
	if err := client.Push(ctx, root, "origin"); err != nil {
		t.Fatal(err)
	}
	if err := client.Commit(ctx, root, "go", "release", true); err != nil {
		t.Fatal(err)
	}
	status, err = client.Status(ctx, root, "go")
	if err != nil {
		t.Fatal(err)
	}
	if status.Upstream == "" || status.Ahead != 1 || status.Behind != 0 {
		t.Fatalf("unexpected tracking status: %+v", status)
	}
	if err := client.Fetch(ctx, root, "origin"); err != nil {
		t.Fatal(err)
	}
	if err := client.Push(ctx, root, "origin"); err != nil {
		t.Fatal(err)
	}
	status, err = client.Status(ctx, root, "go")
	if err != nil || status.Ahead != 0 {
		t.Fatalf("push did not update tracking refs: %+v %v", status, err)
	}
	remote.URL = "https://example.invalid/changed.git"
	if err := client.SaveRemote(ctx, root, remote); err != nil {
		t.Fatal(err)
	}
	status, err = client.Status(ctx, root, "go")
	if err != nil || status.Remotes[0].URL != remote.URL {
		t.Fatalf("remote edit failed: %+v %v", status.Remotes, err)
	}
	if err := client.RemoveRemote(ctx, root, "origin"); err != nil {
		t.Fatal(err)
	}
	status, err = client.Status(ctx, root, "go")
	if err != nil || len(status.Remotes) != 0 {
		t.Fatalf("remote removal failed: %+v %v", status.Remotes, err)
	}
}

func TestGitDisabledAndNestedRepository(t *testing.T) {
	client, root, _ := newTestRepository(t)
	ctx := context.Background()
	plain := t.TempDir()
	status, err := client.Status(ctx, plain, "go")
	if err != nil || !status.Available || status.Repository {
		t.Fatalf("expected disabled Git: %+v %v", status, err)
	}
	if err := client.SaveRemote(ctx, plain, ports.GitRemote{Name: "origin", URL: "https://example.invalid/repo.git"}); err == nil {
		t.Fatal("must explicitly enable Git")
	}
	nested := filepath.Join(root, "child")
	if err := os.Mkdir(nested, 0o700); err != nil {
		t.Fatal(err)
	}
	status, err = client.Status(ctx, nested, "go")
	if err != nil || !status.Nested {
		t.Fatalf("expected nested repository: %+v %v", status, err)
	}
	if err := client.Initialize(ctx, nested); err == nil {
		t.Fatal("must refuse parent repository mutation")
	}
	if err := client.Commit(ctx, nested, "go", "unsafe", false); err == nil {
		t.Fatal("must refuse nested commit")
	}
}

func TestRemoteValidation(t *testing.T) {
	for _, address := range []string{"ext::command", "-bad", "https://user:secret@example.invalid/repo.git", "https://host/a\nb", "unknown://host/repo"} {
		if err := validateRemote(ports.GitRemote{Name: "origin", URL: address, PushURL: address}); err == nil {
			t.Errorf("accepted unsafe address %q", address)
		}
	}
}

func TestCommitFailureRestoresVersionAndIndex(t *testing.T) {
	client, root, binary := newTestRepository(t)
	ctx := context.Background()
	path := filepath.Join(root, "VERSION")
	if err := os.WriteFile(path, []byte("1.0.0\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := client.Commit(ctx, root, "go", "initial", false); err != nil {
		t.Fatal(err)
	}
	staged := filepath.Join(root, "staged.txt")
	if err := os.WriteFile(staged, []byte("staged\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := run(ctx, binary, root, STATUS_TIMEOUT, "add", "--", "staged.txt"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(staged, []byte("unstaged edit\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	before, err := captureIndex(ctx, binary, root)
	if err != nil {
		t.Fatal(err)
	}
	hook := filepath.Join(root, ".git", "hooks", "pre-commit")
	if err := os.WriteFile(hook, []byte("#!/bin/sh\nexit 1\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := client.Commit(ctx, root, "go", "must fail", true); err == nil {
		t.Fatal("expected hook failure")
	}
	version, err := os.ReadFile(path)
	if err != nil || string(version) != "1.0.0\n" {
		t.Fatalf("version was not restored: %s %v", version, err)
	}
	after, err := os.ReadFile(before.path)
	if err != nil || string(after) != string(before.content) {
		t.Fatalf("index was not restored: %v", err)
	}
	worktree, err := os.ReadFile(staged)
	if err != nil || string(worktree) != "unstaged edit\n" {
		t.Fatal("worktree changed")
	}
}
