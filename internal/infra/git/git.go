package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"devenv/internal/ports"
	appErrors "devenv/internal/shared/errors"
)

const (
	STATUS_TIMEOUT = 8 * time.Second
	WRITE_TIMEOUT  = 2 * time.Minute
)

type Client struct{ operations sync.Map }

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Status(ctx context.Context, root, kind string) (ports.GitStatus, error) {
	status := ports.GitStatus{Files: make([]ports.GitFileStatus, 0), Remotes: make([]ports.GitRemote, 0)}
	cleanRoot, err := validateRoot(root)
	if err != nil {
		return status, err
	}
	binary, err := exec.LookPath("git")
	if err != nil {
		return status, appErrors.New(appErrors.GIT_NOT_INSTALLED, "未检测到 Git，请先安装 Git 并添加到 PATH")
	}
	status.Available = true
	rootOutput, rootErr := runRaw(ctx, binary, cleanRoot, STATUS_TIMEOUT, "rev-parse", "--show-toplevel")
	if rootErr != nil {
		if strings.Contains(strings.ToLower(string(rootOutput)), "not a git repository") {
			c.readVersion(cleanRoot, kind, &status)
			return status, nil
		}
		return status, gitError("读取 Git 仓库根目录失败", rootOutput, rootErr)
	}
	status.Repository = true
	repositoryRoot, err := filepath.Abs(strings.TrimSpace(string(rootOutput)))
	if err != nil {
		return status, appErrors.Wrap(err, appErrors.GIT_OPERATION_FAILED, "解析 Git 仓库目录失败")
	}
	if !samePath(cleanRoot, repositoryRoot) {
		status.Nested = true
		c.readVersion(cleanRoot, kind, &status)
		return status, nil
	}

	output, err := run(ctx, binary, cleanRoot, STATUS_TIMEOUT, "status", "--porcelain=v1", "--branch", "--untracked-files=all", "-z")
	if err != nil {
		return status, err
	}
	c.parseStatus(string(output), &status)
	if err := readRemotes(ctx, binary, cleanRoot, &status); err != nil {
		return status, err
	}
	if commit, err := run(ctx, binary, cleanRoot, STATUS_TIMEOUT, "log", "-1", "--format=%h%x00%s"); err == nil {
		parts := strings.SplitN(strings.TrimSpace(string(commit)), "\x00", 2)
		if len(parts) > 0 {
			status.LastCommit = parts[0]
		}
		if len(parts) > 1 {
			status.CommitMessage = parts[1]
		}
	}
	c.readVersion(cleanRoot, kind, &status)
	return status, nil
}

func (c *Client) Initialize(ctx context.Context, root string) error {
	release, err := c.lock(root)
	if err != nil {
		return err
	}
	defer release()
	cleanRoot, err := validateRoot(root)
	if err != nil {
		return err
	}
	binary, err := exec.LookPath("git")
	if err != nil {
		return appErrors.New(appErrors.GIT_NOT_INSTALLED, "未检测到 Git，请先安装 Git 并添加到 PATH")
	}
	output, probeErr := runRaw(ctx, binary, cleanRoot, STATUS_TIMEOUT, "rev-parse", "--show-toplevel")
	if probeErr == nil {
		repositoryRoot, absErr := filepath.Abs(strings.TrimSpace(string(output)))
		if absErr != nil {
			return appErrors.Wrap(absErr, appErrors.GIT_OPERATION_FAILED, "解析 Git 仓库目录失败")
		}
		if !samePath(cleanRoot, repositoryRoot) {
			return appErrors.New(appErrors.GIT_OPERATION_FAILED, "项目目录位于另一个 Git 仓库内，请将仓库根目录添加为项目")
		}
		return nil
	}
	if !strings.Contains(strings.ToLower(string(output)), "not a git repository") {
		return gitError("检查 Git 仓库失败", output, probeErr)
	}
	if _, err := run(ctx, binary, cleanRoot, WRITE_TIMEOUT, "init"); err != nil {
		return err
	}
	return nil
}

func (c *Client) Commit(ctx context.Context, root, kind, message string, bumpVersion bool) (result error) {
	release, err := c.lock(root)
	if err != nil {
		return err
	}
	defer release()
	cleanRoot, err := validateRoot(root)
	if err != nil {
		return err
	}
	message = strings.TrimSpace(message)
	if message == "" {
		return appErrors.New(appErrors.INVALID_ARGUMENT, "提交说明不能为空")
	}
	status, err := c.Status(ctx, cleanRoot, kind)
	if err != nil {
		return err
	}
	if !status.Repository {
		return appErrors.New(appErrors.GIT_NOT_INITIALIZED, "项目尚未初始化 Git 仓库")
	}
	if status.Nested {
		return appErrors.New(appErrors.GIT_OPERATION_FAILED, "项目目录位于另一个 Git 仓库内，不能在此提交")
	}
	binary, err := exec.LookPath("git")
	if err != nil {
		return appErrors.New(appErrors.GIT_NOT_INSTALLED, "未检测到 Git，请先安装 Git 并添加到 PATH")
	}
	var originalVersion []byte
	var versionPath string
	var versionRelative string
	var updatedVersion []byte
	index, err := captureIndex(ctx, binary, cleanRoot)
	if err != nil {
		return err
	}
	defer func() {
		if result == nil {
			return
		}
		if rollbackError := restoreCommitState(binary, cleanRoot, index, versionPath, originalVersion, updatedVersion); rollbackError != nil {
			result = errors.Join(result, fmt.Errorf("恢复提交前状态失败，请检查工作区与暂存区: %w", rollbackError))
		}
	}()
	if bumpVersion {
		location, next, versionError, err := readVersion(cleanRoot, kind)
		if err != nil {
			return appErrors.Wrap(err, appErrors.GIT_OPERATION_FAILED, "读取版本号失败")
		}
		if next == "" {
			return appErrors.New(appErrors.INVALID_ARGUMENT, versionError)
		}
		originalVersion, err = writeNextVersion(cleanRoot, location, next)
		if err != nil {
			return appErrors.Wrap(err, appErrors.GIT_OPERATION_FAILED, "自动升级版本号失败")
		}
		versionPath = filepath.Join(cleanRoot, location.path)
		versionRelative = filepath.ToSlash(location.path)
		updatedVersion, err = os.ReadFile(versionPath)
		if err != nil {
			return err
		}
	}

	if _, err := run(ctx, binary, cleanRoot, WRITE_TIMEOUT, "add", "--all"); err != nil {
		return err
	}
	if bumpVersion {
		stagedVersion, err := runRaw(ctx, binary, cleanRoot, STATUS_TIMEOUT, "show", ":"+versionRelative)
		if err != nil || !bytes.Equal(normalizeLineEndings(stagedVersion), normalizeLineEndings(updatedVersion)) {
			return appErrors.New(appErrors.GIT_OPERATION_FAILED, "版本文件未正确暂存，请检查 .gitignore、Git 过滤器或 skip-worktree 设置")
		}
	}
	status, err = c.Status(ctx, cleanRoot, kind)
	if err != nil {
		return err
	}
	if status.Changed == 0 {
		return appErrors.New(appErrors.GIT_NO_CHANGES, "没有可提交的变更")
	}
	if _, err := run(ctx, binary, cleanRoot, WRITE_TIMEOUT, "commit", "-m", message); err != nil {
		return err
	}
	return nil
}

func (c *Client) readVersion(root, kind string, status *ports.GitStatus) {
	location, next, versionError, err := readVersion(root, kind)
	if err != nil {
		status.VersionError = err.Error()
		return
	}
	status.Version = location.value
	status.NextVersion = next
	status.VersionFile = filepath.ToSlash(location.path)
	status.VersionCanBump = next != ""
	status.VersionError = versionError
}

func (c *Client) parseStatus(output string, status *ports.GitStatus) {
	parts := strings.Split(output, "\x00")
	for index := 0; index < len(parts); index++ {
		entry := parts[index]
		if entry == "" {
			continue
		}
		if strings.HasPrefix(entry, "## ") {
			branch := strings.TrimPrefix(entry, "## ")
			branch = strings.TrimPrefix(branch, "No commits yet on ")
			branch = strings.TrimPrefix(branch, "Initial commit on ")
			if separator := strings.Index(branch, "..."); separator >= 0 {
				branch = branch[:separator]
			}
			status.Branch = branch
			continue
		}
		if len(entry) < 3 {
			continue
		}
		indexStatus := entry[0]
		worktreeStatus := entry[1]
		code := strings.TrimSpace(entry[:2])
		path := entry[3:]
		untracked := code == "??"
		staged := !untracked && indexStatus != ' '
		unstaged := !untracked && worktreeStatus != ' '
		status.Changed++
		if staged {
			status.Staged++
		}
		if unstaged {
			status.Unstaged++
		}
		if untracked {
			status.Untracked++
		}
		status.Files = append(status.Files, ports.GitFileStatus{
			Path: path, Status: code, Staged: staged, Unstaged: unstaged, Untracked: untracked,
		})
		if strings.ContainsRune(code, 'R') || strings.ContainsRune(code, 'C') {
			index++
		}
	}
}

func validateRoot(root string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", appErrors.New(appErrors.PATH_INVALID, "项目目录不能为空")
	}
	cleanRoot, err := filepath.Abs(strings.TrimSpace(root))
	if err != nil {
		return "", appErrors.Wrap(err, appErrors.PATH_INVALID, "解析项目目录失败")
	}
	info, err := os.Stat(cleanRoot)
	if err != nil {
		return "", appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "读取项目目录失败")
	}
	if !info.IsDir() {
		return "", appErrors.New(appErrors.PATH_INVALID, "项目路径不是目录")
	}
	return cleanRoot, nil
}

func samePath(left, right string) bool {
	return strings.EqualFold(filepath.Clean(left), filepath.Clean(right))
}

func run(ctx context.Context, binary, root string, timeout time.Duration, args ...string) ([]byte, error) {
	output, err := runRaw(ctx, binary, root, timeout, args...)
	if err != nil {
		return output, gitError("Git 命令执行失败", output, err)
	}
	return output, nil
}

func runRaw(ctx context.Context, binary, root string, timeout time.Duration, args ...string) ([]byte, error) {
	commandContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	commandArgs := make([]string, 0, len(args)+2)
	commandArgs = append(commandArgs, "-C", root)
	commandArgs = append(commandArgs, args...)
	command := exec.CommandContext(commandContext, binary, commandArgs...)
	for _, entry := range os.Environ() {
		key := strings.ToUpper(strings.SplitN(entry, "=", 2)[0])
		if !strings.HasPrefix(key, "GIT_") && key != "LC_ALL" {
			command.Env = append(command.Env, entry)
		}
	}
	command.Env = append(command.Env, "GIT_TERMINAL_PROMPT=0", "GCM_INTERACTIVE=never", "LC_ALL=C")
	command.WaitDelay = 2 * time.Second
	prepareCommand(command)
	output, err := command.CombinedOutput()
	return output, err
}

func gitError(message string, output []byte, err error) error {
	detail := strings.TrimSpace(string(output))
	if detail == "" {
		return appErrors.Wrap(err, appErrors.GIT_OPERATION_FAILED, message)
	}
	return appErrors.Wrap(fmt.Errorf("%s: %w", detail, err), appErrors.GIT_OPERATION_FAILED, message)
}

var _ ports.GitManager = (*Client)(nil)
