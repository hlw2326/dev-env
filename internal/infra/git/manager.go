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

func (m *Manager) Clone(ctx context.Context, remoteURL, localPath, branch string) error {
	if branch == "" {
		branch = "main"
	}
	parentDir := filepath.Dir(localPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "git", "clone", "-b", branch, remoteURL, localPath)
	if err := cmd.Run(); err != nil {
		// Fallback to default branch if specified branch fails
		cmdFallback := exec.CommandContext(ctx, "git", "clone", remoteURL, localPath)
		return cmdFallback.Run()
	}
	return nil
}

func (m *Manager) Pull(ctx context.Context, repoPath string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "pull")
	return cmd.Run()
}

func (m *Manager) GetStatus(ctx context.Context, repoPath string) (*site.GitStatus, error) {
	now := time.Now().Format(time.RFC3339)
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return &site.GitStatus{
			CurrentBranch:  "main",
			CurrentCommit:  "-",
			CommitMessage:  "尚未初始化 Git 仓库",
			IsClean:        true,
			ModifiedFiles:  []string{},
			UntrackedFiles: []string{},
			LastCheckAt:    now,
		}, nil
	}

	branchCmd := exec.CommandContext(ctx, "git", "-C", repoPath, "branch", "--show-current")
	branchOut, _ := branchCmd.Output()
	branch := strings.TrimSpace(string(branchOut))
	if branch == "" {
		branch = "main"
	}

	commitCmd := exec.CommandContext(ctx, "git", "-C", repoPath, "rev-parse", "--short", "HEAD")
	commitOut, _ := commitCmd.Output()
	commit := strings.TrimSpace(string(commitOut))

	msgCmd := exec.CommandContext(ctx, "git", "-C", repoPath, "log", "-1", "--pretty=%s")
	msgOut, _ := msgCmd.Output()
	commitMsg := strings.TrimSpace(string(msgOut))

	statusCmd := exec.CommandContext(ctx, "git", "-C", repoPath, "status", "--porcelain")
	statusOut, _ := statusCmd.Output()

	modifiedFiles := make([]string, 0)
	untrackedFiles := make([]string, 0)
	lines := strings.Split(strings.ReplaceAll(string(statusOut), "\r\n", "\n"), "\n")
	for _, line := range lines {
		if len(line) < 3 {
			continue
		}
		st := line[0:2]
		file := strings.TrimSpace(line[3:])
		if st == "??" {
			untrackedFiles = append(untrackedFiles, file)
		} else if strings.TrimSpace(st) != "" {
			modifiedFiles = append(modifiedFiles, file)
		}
	}

	aheadCount := 0
	behindCount := 0
	revCmd := exec.CommandContext(ctx, "git", "-C", repoPath, "rev-list", "--left-right", "--count", "HEAD...@{u}")
	if revOut, err := revCmd.Output(); err == nil {
		parts := strings.Fields(string(revOut))
		if len(parts) == 2 {
			aheadCount, _ = strconv.Atoi(parts[0])
			behindCount, _ = strconv.Atoi(parts[1])
		}
	}

	return &site.GitStatus{
		CurrentBranch:  branch,
		CurrentCommit:  commit,
		CommitMessage:  commitMsg,
		IsClean:        len(modifiedFiles) == 0 && len(untrackedFiles) == 0,
		AheadCount:     aheadCount,
		BehindCount:    behindCount,
		ModifiedFiles:  modifiedFiles,
		UntrackedFiles: untrackedFiles,
		LastCheckAt:    now,
	}, nil
}

func (m *Manager) Commit(ctx context.Context, repoPath, message string, files []string) error {
	if len(files) == 0 {
		addAll := exec.CommandContext(ctx, "git", "-C", repoPath, "add", "-A")
		if err := addAll.Run(); err != nil {
			return err
		}
	} else {
		for _, f := range files {
			addCmd := exec.CommandContext(ctx, "git", "-C", repoPath, "add", f)
			if err := addCmd.Run(); err != nil {
				return err
			}
		}
	}

	commitCmd := exec.CommandContext(ctx, "git", "-C", repoPath, "commit", "-m", message)
	return commitCmd.Run()
}

func (m *Manager) Push(ctx context.Context, repoPath string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "push")
	return cmd.Run()
}
