package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"devenv/internal/domain/site"
)

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Clone(ctx context.Context, remote, root, branch string) error {
	parent := filepath.Dir(root)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return err
	}
	args := []string{"clone"}
	if branch != "" {
		args = append(args, "-b", branch)
	}
	args = append(args, remote, root)
	cmd := exec.CommandContext(ctx, "git", args...)
	return cmd.Run()
}

func (m *Manager) Pull(ctx context.Context, root string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", root, "pull")
	return cmd.Run()
}

func (m *Manager) Status(ctx context.Context, root string) (*site.GitState, error) {
	now := time.Now().Format(time.RFC3339)
	gitDir := filepath.Join(root, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return &site.GitState{
			Branch:    "",
			Commit:    "",
			Message:   "",
			Clean:     true,
			Modified:  make([]string, 0),
			Untracked: make([]string, 0),
			CheckedAt: now,
		}, nil
	}

	branchOut, err := exec.CommandContext(ctx, "git", "-C", root, "branch", "--show-current").Output()
	if err != nil {
		return nil, err
	}
	branch := strings.TrimSpace(string(branchOut))

	var commit, msg string
	if commitOut, err := exec.CommandContext(ctx, "git", "-C", root, "rev-parse", "--short", "HEAD").Output(); err == nil {
		commit = strings.TrimSpace(string(commitOut))
	}
	if msgOut, err := exec.CommandContext(ctx, "git", "-C", root, "log", "-1", "--pretty=%s").Output(); err == nil {
		msg = strings.TrimSpace(string(msgOut))
	}

	statusOut, err := exec.CommandContext(ctx, "git", "-C", root, "status", "--porcelain").Output()
	if err != nil {
		return nil, err
	}

	modified := make([]string, 0)
	untracked := make([]string, 0)
	lines := strings.Split(strings.ReplaceAll(string(statusOut), "\r\n", "\n"), "\n")
	for _, line := range lines {
		if len(line) < 3 {
			continue
		}
		st := line[0:2]
		file := strings.TrimSpace(line[3:])
		if st == "??" {
			untracked = append(untracked, file)
		} else if strings.TrimSpace(st) != "" {
			modified = append(modified, file)
		}
	}

	ahead := 0
	behind := 0
	if revOut, err := exec.CommandContext(ctx, "git", "-C", root, "rev-list", "--left-right", "--count", "HEAD...@{u}").Output(); err == nil {
		parts := strings.Fields(string(revOut))
		if len(parts) == 2 {
			if a, err := strconv.Atoi(parts[0]); err == nil {
				ahead = a
			}
			if b, err := strconv.Atoi(parts[1]); err == nil {
				behind = b
			}
		}
	}

	return &site.GitState{
		Branch:    branch,
		Commit:    commit,
		Message:   msg,
		Clean:     len(modified) == 0 && len(untracked) == 0,
		Ahead:     ahead,
		Behind:    behind,
		Modified:  modified,
		Untracked: untracked,
		CheckedAt: now,
	}, nil
}

func (m *Manager) Commit(ctx context.Context, root, message string, files []string) error {
	if len(files) == 0 {
		if err := exec.CommandContext(ctx, "git", "-C", root, "add", "-A").Run(); err != nil {
			return err
		}
	} else {
		for _, f := range files {
			if err := exec.CommandContext(ctx, "git", "-C", root, "add", f).Run(); err != nil {
				return err
			}
		}
	}
	return exec.CommandContext(ctx, "git", "-C", root, "commit", "-m", message).Run()
}

func (m *Manager) Push(ctx context.Context, root string) error {
	return exec.CommandContext(ctx, "git", "-C", root, "push").Run()
}
