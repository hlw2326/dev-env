package deployment

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	model "devenv/internal/domain/deployment"
	"github.com/google/uuid"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

const CONNECT_TIMEOUT = 15 * time.Second
const TASK_TIMEOUT = 30 * time.Minute

type SFTP struct{ base string }

func NewSFTP(base string) *SFTP { return &SFTP{base: base} }

func (p *SFTP) Preview(ctx context.Context, config model.Config, root string) (model.Plan, error) {
	plan, err := scan(ctx, config, root)
	plan.Target = fmt.Sprintf("%s@%s:%d%s", config.Username, config.Host, config.Port, config.Remote)
	if err != nil {
		return plan, err
	}
	if config.PrivateKey != "" {
		if err := checkPrivateFile(config.PrivateKey); err != nil {
			return plan, err
		}
	}
	return plan, nil
}

func (p *SFTP) Test(ctx context.Context, config model.Config, _ string, credentials model.Credentials) error {
	client, closeClient, err := connect(ctx, config, credentials)
	if err != nil {
		return err
	}
	defer closeClient()
	return checkRemote(client, config.Remote)
}

func (p *SFTP) Run(ctx context.Context, config model.Config, root string, plan model.Plan, credentials model.Credentials, report func(int, string)) error {
	report(0, "正在校验并准备已确认的文件快照")
	snapshot, cleanup, err := stage(ctx, p.base, config, root, plan)
	if err != nil {
		return err
	}
	defer cleanup()
	client, closeClient, err := connect(ctx, config, credentials)
	if err != nil {
		return err
	}
	defer closeClient()
	if err := checkRemote(client, config.Remote); err != nil {
		return err
	}
	for index, item := range plan.Files {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := uploadFile(client, snapshot, config.Remote, item); err != nil {
			return fmt.Errorf("同步 %s 失败（此前已完成 %d/%d 个文件，不自动回滚）: %w", item.Path, index, len(plan.Files), err)
		}
		report(index+1, fmt.Sprintf("已同步 %d/%d · %s", index+1, len(plan.Files), item.Path))
	}
	return nil
}

func connect(ctx context.Context, config model.Config, credentials model.Credentials) (*sftp.Client, func(), error) {
	var auth ssh.AuthMethod
	if config.PrivateKey != "" {
		if err := checkPrivateFile(config.PrivateKey); err != nil {
			return nil, nil, err
		}
		key, err := os.ReadFile(config.PrivateKey)
		if err != nil {
			return nil, nil, err
		}
		var signer ssh.Signer
		if credentials.Passphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(credentials.Passphrase))
		} else {
			signer, err = ssh.ParsePrivateKey(key)
		}
		if err != nil {
			return nil, nil, fmt.Errorf("读取 SSH 私钥失败（加密私钥需要口令）: %w", err)
		}
		auth = ssh.PublicKeys(signer)
	} else {
		if credentials.Password == "" {
			return nil, nil, fmt.Errorf("请选择 SSH 私钥或输入本次连接密码，密码不会保存")
		}
		auth = ssh.Password(credentials.Password)
	}
	address := net.JoinHostPort(config.Host, strconv.Itoa(config.Port))
	connection, err := (&net.Dialer{Timeout: CONNECT_TIMEOUT}).DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, nil, err
	}
	stop := context.AfterFunc(ctx, func() { _ = connection.Close() })
	closeConnection := func() { stop(); _ = connection.Close() }
	if err := connection.SetDeadline(time.Now().Add(CONNECT_TIMEOUT)); err != nil {
		closeConnection()
		return nil, nil, err
	}
	configuration := &ssh.ClientConfig{User: config.Username, Auth: []ssh.AuthMethod{auth}, Timeout: CONNECT_TIMEOUT, HostKeyCallback: func(_ string, _ net.Addr, key ssh.PublicKey) error {
		actual := ssh.FingerprintSHA256(key)
		if actual != config.Fingerprint {
			return fmt.Errorf("服务器 SSH 指纹不匹配，实际为 %s；请通过可信渠道核验，禁止直接忽略", actual)
		}
		return nil
	}}
	transport, channels, requests, err := ssh.NewClientConn(connection, address, configuration)
	if err != nil {
		closeConnection()
		return nil, nil, err
	}
	sshClient := ssh.NewClient(transport, channels, requests)
	if err := connection.SetDeadline(time.Now().Add(TASK_TIMEOUT)); err != nil {
		_ = sshClient.Close()
		closeConnection()
		return nil, nil, err
	}
	client, err := sftp.NewClient(sshClient)
	if err != nil {
		_ = sshClient.Close()
		closeConnection()
		return nil, nil, err
	}
	return client, func() { _ = client.Close(); _ = sshClient.Close(); closeConnection() }, nil
}

func checkPrivateFile(file string) error {
	if !filepath.IsAbs(file) {
		return fmt.Errorf("私钥必须使用本机绝对路径")
	}
	info, err := os.Stat(file)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Size() > 1024*1024 {
		return fmt.Errorf("私钥不是有效的小型普通文件")
	}
	return nil
}

func checkRemote(client *sftp.Client, root string) error {
	if _, ok := client.HasExtension("posix-rename@openssh.com"); !ok {
		return fmt.Errorf("服务器不支持原子覆盖文件所需的 SFTP POSIX rename 扩展，未上传任何文件")
	}
	return remoteDirectory(client, root, false)
}

func remoteDirectory(client *sftp.Client, directory string, create bool) error {
	current := "/"
	for _, part := range strings.Split(strings.Trim(directory, "/"), "/") {
		if part == "" || part == "." || part == ".." {
			return fmt.Errorf("远端目录不合法")
		}
		current = path.Join(current, part)
		info, err := client.Lstat(current)
		if os.IsNotExist(err) && create {
			if err := client.Mkdir(current); err != nil {
				return err
			}
			info, err = client.Lstat(current)
		}
		if err != nil {
			return fmt.Errorf("远端目录 %s 不可访问（目标根目录需提前创建）: %w", current, err)
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return fmt.Errorf("拒绝写入符号链接或非目录路径: %s", current)
		}
	}
	return nil
}

func uploadFile(client *sftp.Client, snapshot, root string, item model.File) error {
	destination := path.Join(root, item.Path)
	if err := remoteDirectory(client, path.Dir(destination), true); err != nil {
		return err
	}
	permission := os.FileMode(0o644)
	if info, err := client.Lstat(destination); err == nil {
		if !info.Mode().IsRegular() {
			return fmt.Errorf("拒绝覆盖符号链接或非普通文件")
		}
		permission = info.Mode().Perm()
	} else if !os.IsNotExist(err) {
		return err
	}
	input, err := os.Open(filepath.Join(snapshot, filepath.FromSlash(item.Path)))
	if err != nil {
		return err
	}
	defer input.Close()
	temporary := path.Join(path.Dir(destination), ".devenv-upload-"+uuid.NewString())
	output, err := client.OpenFile(temporary, os.O_CREATE|os.O_EXCL|os.O_WRONLY)
	if err != nil {
		return err
	}
	defer client.Remove(temporary)
	_, copyError := io.Copy(output, input)
	closeError := output.Close()
	if copyError != nil {
		return copyError
	}
	if closeError != nil {
		return closeError
	}
	if err := client.Chmod(temporary, permission); err != nil {
		return err
	}
	return client.PosixRename(temporary, destination)
}
