package deployment

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	model "devenv/internal/domain/deployment"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func writeFixture(t *testing.T, root, name, content string) {
	t.Helper()
	target := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestManifestAndStalePreview(t *testing.T) {
	root := t.TempDir()
	config := model.Config{Provider: model.PROVIDER_SFTP, Source: ".", Exclude: "tests/\n*.map"}
	for _, name := range []string{"index.php", "public/app.js", ".env", ".git/config", "node_modules/test.js", "storage/session", "secret.pem", "id_ed25519", "tests/test.php", "public/app.js.map", ".npmrc", "backup.sql", "database.sqlite", ".pgpass"} {
		writeFixture(t, root, name, name)
	}
	plan, err := scan(context.Background(), config, root)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Files) != 2 {
		t.Fatalf("unexpected manifest: %+v", plan.Files)
	}
	snapshot, cleanup, err := stage(context.Background(), t.TempDir(), config, root, plan)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	if _, err := os.Stat(filepath.Join(snapshot, ".env")); !os.IsNotExist(err) {
		t.Fatal("snapshot leaked secret")
	}
	writeFixture(t, root, "index.php", "new content")
	if _, _, err := stage(context.Background(), t.TempDir(), config, root, plan); err == nil {
		t.Fatal("stale preview must fail")
	}
	config.Source = "../outside"
	if _, err := scan(context.Background(), config, root); err == nil {
		t.Fatal("traversal accepted")
	}
}

func TestWeChatManifestIgnoresServerExcludeRules(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "project.config.json", `{"appid":"wx1234567890abcdef"}`)
	writeFixture(t, root, "app.json", `{"pages":[]}`)
	config := model.Config{Provider: model.PROVIDER_WECHAT, Source: ".", Exclude: "app.json"}
	plan, err := scan(context.Background(), config, root)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range plan.Files {
		if file.Path == "app.json" {
			return
		}
	}
	t.Fatal("wechat extra exclude should not remove compiled files")
}

func testSFTPServer(t *testing.T) (model.Config, model.Credentials) {
	t.Helper()
	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(key)
	if err != nil {
		t.Fatal(err)
	}
	serverConfig := &ssh.ServerConfig{PasswordCallback: func(meta ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
		if meta.User() != "deploy" || string(password) != "test-password" {
			return nil, fmt.Errorf("bad credentials")
		}
		return nil, nil
	}}
	serverConfig.AddHostKey(signer)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	handlers := sftp.InMemHandler()
	go func() {
		for {
			connection, err := listener.Accept()
			if err != nil {
				return
			}
			go func() {
				defer connection.Close()
				transport, channels, requests, err := ssh.NewServerConn(connection, serverConfig)
				if err != nil {
					return
				}
				defer transport.Close()
				go ssh.DiscardRequests(requests)
				for incoming := range channels {
					if incoming.ChannelType() != "session" {
						_ = incoming.Reject(ssh.UnknownChannelType, "unsupported")
						continue
					}
					channel, requests, err := incoming.Accept()
					if err != nil {
						return
					}
					go func() {
						defer channel.Close()
						for request := range requests {
							var payload struct{ Name string }
							if request.Type == "subsystem" && ssh.Unmarshal(request.Payload, &payload) == nil && payload.Name == "sftp" {
								_ = request.Reply(true, nil)
								server := sftp.NewRequestServer(channel, handlers)
								_ = server.Serve()
								_ = server.Close()
								return
							}
							_ = request.Reply(false, nil)
						}
					}()
				}
			}()
		}
	}()
	_, portText, _ := net.SplitHostPort(listener.Addr().String())
	port, _ := strconv.Atoi(portText)
	return model.Config{Provider: model.PROVIDER_SFTP, Source: ".", Host: "127.0.0.1", Port: port, Username: "deploy", Fingerprint: ssh.FingerprintSHA256(signer.PublicKey()), Remote: "/deploy/project"}, model.Credentials{Password: "test-password"}
}

func TestSFTPSyncAndFingerprint(t *testing.T) {
	ctx := context.Background()
	config, credentials := testSFTPServer(t)
	client, closeClient, err := connect(ctx, config, credentials)
	if err != nil {
		t.Fatal(err)
	}
	defer closeClient()
	if err := client.MkdirAll(config.Remote); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"index.php", ".env", "extra.txt"} {
		file, err := client.Create(config.Remote + "/" + name)
		if err != nil {
			t.Fatal(err)
		}
		_, err = file.Write([]byte("server value"))
		if err != nil {
			t.Fatal(err)
		}
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
	}
	root := t.TempDir()
	writeFixture(t, root, "index.php", "new php")
	writeFixture(t, root, "public/app.js", "new js")
	writeFixture(t, root, ".env", "local secret")
	provider := NewSFTP(t.TempDir())
	plan, err := provider.Preview(ctx, config, root)
	if err != nil {
		t.Fatal(err)
	}
	if err := provider.Test(ctx, config, root, credentials); err != nil {
		t.Fatal(err)
	}
	completed := 0
	if err := provider.Run(ctx, config, root, plan, credentials, func(count int, _ string) { completed = count }); err != nil {
		t.Fatal(err)
	}
	if completed != 2 {
		t.Fatalf("completed %d", completed)
	}
	for name, expected := range map[string]string{"index.php": "new php", "public/app.js": "new js", ".env": "server value", "extra.txt": "server value"} {
		file, err := client.Open(config.Remote + "/" + name)
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(file)
		_ = file.Close()
		if err != nil || string(data) != expected {
			t.Fatalf("%s: %s %v", name, data, err)
		}
	}
	if err := client.Symlink("/deploy/project", "/deploy/link"); err != nil {
		t.Fatal(err)
	}
	linked := config
	linked.Remote = "/deploy/link"
	if err := provider.Test(ctx, linked, root, credentials); err == nil {
		t.Fatal("symlink target must be rejected")
	}
	config.Fingerprint = "SHA256:" + strings.Repeat("A", 43)
	if err := provider.Test(ctx, config, root, credentials); err == nil {
		t.Fatal("host key mismatch accepted")
	}
}

func TestWeChatAdapterUsesDeveloperToolsCLIAndSnapshot(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	tools := t.TempDir()
	output := filepath.Join(tools, "uploaded.txt")
	writeFixture(t, root, "project.config.json", `{"appid":"wx1234567890abcdef","miniprogramRoot":"./"}`)
	writeFixture(t, root, "app.json", `{"pages":[]}`)
	writeFixture(t, root, ".env", "secret")
	cli := filepath.Join(tools, "wechat-cli")
	if runtime.GOOS == "windows" {
		cli += ".cmd"
		writeFixture(t, tools, filepath.Base(cli), "@echo off\r\nif \"%1\"==\"--help\" exit /b 0\r\nif not \"%1\"==\"upload\" exit /b 1\r\necho %* >> \""+output+"\"\r\n")
	} else {
		writeFixture(t, tools, filepath.Base(cli), "#!/bin/sh\nif [ \"$1\" = \"--help\" ]; then exit 0; fi\nif [ \"$1\" != \"upload\" ]; then exit 1; fi\nprintf '%s\\n' \"$*\" > \""+output+"\"\n")
		if err := os.Chmod(cli, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	config := model.Config{Provider: model.PROVIDER_WECHAT, Source: ".", AppIDs: []string{"wx1111222233334444", "wxabcdef1234567890"}, DeveloperToolsCLI: cli, Version: "1.0.01", Description: "manual upload", Robot: 3}
	provider := NewWeChat(t.TempDir())
	plan, err := provider.Preview(ctx, config, root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatal("preview triggered upload")
	}
	if err := provider.Run(ctx, config, root, plan, model.Credentials{}, func(int, string) {}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	commandLine := string(raw)
	if strings.Count(commandLine, "upload --project") != len(config.AppIDs) || !strings.Contains(commandLine, "--version "+config.Version) || !strings.Contains(commandLine, "--robot 3") || !strings.Contains(commandLine, config.Description) {
		t.Fatalf("invalid CLI arguments: %s", raw)
	}
	if strings.Contains(commandLine, root) {
		t.Fatal("CLI received the original project path instead of the snapshot")
	}
}
