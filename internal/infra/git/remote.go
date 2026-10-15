package git

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"devenv/internal/ports"
	appErrors "devenv/internal/shared/errors"
)

var remoteNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)
var sshAddressPattern = regexp.MustCompile(`^[^\s@:]+@[^\s:]+:[^\s]+$`)

func (c *Client) lock(root string) (func(), error) {
	path, err := validateRoot(root)
	if err != nil {
		return nil, err
	}
	key := strings.ToLower(path)
	if _, loaded := c.operations.LoadOrStore(key, true); loaded {
		return nil, appErrors.New(appErrors.GIT_OPERATION_FAILED, "此项目正在执行 Git 操作，请稍后重试")
	}
	return func() { c.operations.Delete(key) }, nil
}

func readRemotes(ctx context.Context, binary, root string, status *ports.GitStatus) error {
	output, err := run(ctx, binary, root, STATUS_TIMEOUT, "remote")
	if err != nil {
		return err
	}
	for _, name := range strings.Fields(string(output)) {
		fetchURL, err := run(ctx, binary, root, STATUS_TIMEOUT, "remote", "get-url", "--", name)
		if err != nil {
			return err
		}
		pushURL, err := run(ctx, binary, root, STATUS_TIMEOUT, "remote", "get-url", "--push", "--", name)
		if err != nil {
			return err
		}
		status.Remotes = append(status.Remotes, ports.GitRemote{Name: name, URL: strings.TrimSpace(string(fetchURL)), PushURL: strings.TrimSpace(string(pushURL))})
	}
	// Local refs only: opening the project must not trigger network access.
	upstream, err := runRaw(ctx, binary, root, STATUS_TIMEOUT, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}")
	if err != nil {
		return nil
	} // No upstream on new/local-only branches.
	status.Upstream = strings.TrimSpace(string(upstream))
	counts, err := run(ctx, binary, root, STATUS_TIMEOUT, "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
	if err != nil {
		return err
	}
	fields := strings.Fields(string(counts))
	if len(fields) != 2 {
		return appErrors.New(appErrors.GIT_OPERATION_FAILED, "无法解析远端同步状态")
	}
	status.Ahead, err = strconv.Atoi(fields[0])
	if err != nil {
		return err
	}
	status.Behind, err = strconv.Atoi(fields[1])
	return err
}

func validateRemote(remote ports.GitRemote) error {
	if !remoteNamePattern.MatchString(remote.Name) {
		return appErrors.New(appErrors.INVALID_ARGUMENT, "远端名称只能包含字母、数字、点、下划线和横线，且必须以字母或数字开头")
	}
	for _, address := range []string{remote.URL, remote.PushURL} {
		if strings.TrimSpace(address) == "" || strings.HasPrefix(address, "-") || strings.ContainsAny(address, "\x00\r\n") {
			return appErrors.New(appErrors.INVALID_ARGUMENT, "远端地址不能为空或包含控制字符")
		}
		if filepath.IsAbs(address) {
			continue
		}
		if strings.Contains(address, "://") {
			parsed, err := url.Parse(address)
			if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http" && parsed.Scheme != "ssh" && parsed.Scheme != "git") {
				return appErrors.New(appErrors.INVALID_ARGUMENT, "远端地址请使用 HTTPS、SSH、Git 协议或本地绝对路径")
			}
			if parsed.User != nil {
				if _, exists := parsed.User.Password(); exists {
					return appErrors.New(appErrors.INVALID_ARGUMENT, "请勿在远端地址中保存密码或令牌，请使用 Git 凭据管理器或 SSH 密钥")
				}
			}
		} else if !sshAddressPattern.MatchString(address) {
			return appErrors.New(appErrors.INVALID_ARGUMENT, "SSH 地址格式应为 git@host:group/repository.git")
		}
	}
	return nil
}

func (c *Client) repositoryOperation(ctx context.Context, root string, action func(string, string, ports.GitStatus) error) error {
	release, err := c.lock(root)
	if err != nil {
		return err
	}
	defer release()
	status, err := c.Status(ctx, root, "")
	if err != nil {
		return err
	}
	if !status.Repository {
		return appErrors.New(appErrors.GIT_NOT_INITIALIZED, "此项目尚未启用 Git")
	}
	if status.Nested {
		return appErrors.New(appErrors.GIT_OPERATION_FAILED, "请在项目的 Git 仓库根目录操作")
	}
	binary, err := exec.LookPath("git")
	if err != nil {
		return err
	}
	return action(binary, root, status)
}

func (c *Client) SaveRemote(ctx context.Context, root string, remote ports.GitRemote) error {
	remote.Name = strings.TrimSpace(remote.Name)
	remote.URL = strings.TrimSpace(remote.URL)
	remote.PushURL = strings.TrimSpace(remote.PushURL)
	if remote.PushURL == "" {
		remote.PushURL = remote.URL
	}
	if err := validateRemote(remote); err != nil {
		return err
	}
	return c.repositoryOperation(ctx, root, func(binary, root string, status ports.GitStatus) error {
		var previous *ports.GitRemote
		for _, item := range status.Remotes {
			if item.Name == remote.Name {
				copy := item
				previous = &copy
				break
			}
		}
		if previous != nil {
			for _, key := range []string{"url", "pushurl"} {
				output, readError := runRaw(ctx, binary, root, STATUS_TIMEOUT, "config", "--get-all", "remote."+remote.Name+"."+key)
				if readError != nil {
					exitError, ok := readError.(*exec.ExitError)
					if !ok || exitError.ExitCode() != 1 {
						return gitError("读取远端配置失败", output, readError)
					}
				}
				if len(strings.Split(strings.TrimSpace(string(output)), "\n")) > 1 {
					return appErrors.New(appErrors.GIT_OPERATION_FAILED, "此远端配置了多个地址，请在 Git 配置中管理，避免覆盖已有设置")
				}
			}
			if _, err := run(ctx, binary, root, WRITE_TIMEOUT, "remote", "set-url", "--", remote.Name, remote.URL); err != nil {
				return err
			}
		} else {
			if _, err := run(ctx, binary, root, WRITE_TIMEOUT, "remote", "add", "--", remote.Name, remote.URL); err != nil {
				return err
			}
		}
		if _, err := run(ctx, binary, root, WRITE_TIMEOUT, "remote", "set-url", "--push", "--", remote.Name, remote.PushURL); err != nil {
			return fmt.Errorf("拉取地址已保存，但推送地址保存失败，请检查远端设置: %w", err)
		}
		return nil
	})
}

func (c *Client) remoteAction(ctx context.Context, root, name string, action func(string, string, ports.GitStatus) error) error {
	if !remoteNamePattern.MatchString(name) {
		return appErrors.New(appErrors.INVALID_ARGUMENT, "无效的远端名称")
	}
	return c.repositoryOperation(ctx, root, func(binary, root string, status ports.GitStatus) error {
		for _, remote := range status.Remotes {
			if remote.Name == name {
				return action(binary, root, status)
			}
		}
		return appErrors.New(appErrors.NOT_FOUND, "远端不存在，请刷新后重试")
	})
}

func (c *Client) RemoveRemote(ctx context.Context, root, name string) error {
	return c.remoteAction(ctx, root, name, func(binary, root string, _ ports.GitStatus) error {
		_, err := run(ctx, binary, root, WRITE_TIMEOUT, "remote", "remove", name)
		return err
	})
}

func (c *Client) Fetch(ctx context.Context, root, name string) error {
	return c.remoteAction(ctx, root, name, func(binary, root string, _ ports.GitStatus) error {
		_, err := run(ctx, binary, root, WRITE_TIMEOUT, "-c", "protocol.ext.allow=never", "fetch", "--", name)
		return err
	})
}

func (c *Client) Push(ctx context.Context, root, name string) error {
	return c.remoteAction(ctx, root, name, func(binary, root string, status ports.GitStatus) error {
		if status.LastCommit == "" {
			return appErrors.New(appErrors.INVALID_ARGUMENT, "请先创建本地提交再推送")
		}
		branch, err := run(ctx, binary, root, STATUS_TIMEOUT, "symbolic-ref", "--quiet", "--short", "HEAD")
		if err != nil {
			return appErrors.New(appErrors.GIT_OPERATION_FAILED, "当前处于分离 HEAD 状态，请先切换到分支")
		}
		_, err = run(ctx, binary, root, WRITE_TIMEOUT, "-c", "protocol.ext.allow=never", "push", "--no-force", "--no-mirror", "--no-follow-tags", "--set-upstream", "--", name, "HEAD:refs/heads/"+strings.TrimSpace(string(branch)))
		return err
	})
}
