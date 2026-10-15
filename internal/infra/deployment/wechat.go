package deployment

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	model "devenv/internal/domain/deployment"
	"devenv/internal/shared/paths"
)

//go:embed wechat.cjs
var wechatScript string

type WeChat struct{ base string }

func NewWeChat(base string) *WeChat { return &WeChat{base: base} }

func checkWeChat(config model.Config, root string) error {
	source, err := paths.DeploymentSource(root, config.Source)
	if err != nil {
		return err
	}
	if err := checkPrivateFile(config.UploadKey); err != nil {
		return err
	}
	content, err := os.ReadFile(filepath.Join(source, "project.config.json"))
	if err != nil {
		return fmt.Errorf("发布目录必须包含构建好的 project.config.json: %w", err)
	}
	var projectConfig map[string]json.RawMessage
	if err := json.Unmarshal(content, &projectConfig); err != nil {
		return err
	}
	var appID string
	if err := json.Unmarshal(projectConfig["appid"], &appID); err != nil || appID != config.AppID {
		return fmt.Errorf("project.config.json 的 appid 必须与发布配置一致")
	}
	for _, key := range []string{"miniprogramRoot", "srcMiniprogramRoot", "cloudfunctionRoot", "pluginRoot"} {
		if raw, ok := projectConfig[key]; ok {
			var value string
			if err := json.Unmarshal(raw, &value); err != nil {
				return err
			}
			if value != "" {
				if _, err := paths.DeploymentSource(source, value); err != nil {
					return fmt.Errorf("project.config.json 的 %s 不安全或不存在: %w", key, err)
				}
			}
		}
	}
	if config.ModulePath != "" && !filepath.IsAbs(config.ModulePath) {
		return fmt.Errorf("CI 工具目录必须是绝对路径（包含 node_modules 的目录）")
	}
	return nil
}

func (p *WeChat) Preview(ctx context.Context, config model.Config, root string) (model.Plan, error) {
	if err := checkWeChat(config, root); err != nil {
		return model.Plan{}, err
	}
	if err := p.Test(ctx, config, root, model.Credentials{}); err != nil {
		return model.Plan{}, err
	}
	plan, err := scan(ctx, config, root)
	plan.Target = "微信开发版本 · " + config.AppID + " · " + config.Version
	return plan, err
}
func (p *WeChat) Test(ctx context.Context, config model.Config, root string, _ model.Credentials) error {
	if err := checkWeChat(config, root); err != nil {
		return err
	}
	checkContext, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	return runWeChat(checkContext, config, root, "", true)
}
func (p *WeChat) Run(ctx context.Context, config model.Config, root string, plan model.Plan, _ model.Credentials, report func(int, string)) error {
	report(0, "正在校验并准备小程序文件快照")
	snapshot, cleanup, err := stage(ctx, p.base, config, root, plan)
	if err != nil {
		return err
	}
	defer cleanup()
	// Revalidate config paths against the filtered snapshot, not the original project.
	stagedConfig := config
	stagedConfig.Source = "."
	if err := checkWeChat(stagedConfig, snapshot); err != nil {
		return err
	}
	report(0, "微信工具正在编译并上传，请稍候；取消或网络中断后需在微信后台核对结果")
	if err := runWeChat(ctx, config, root, snapshot, false); err != nil {
		return err
	}
	report(len(plan.Files), "代码已上传到微信开发版本；请到微信后台手动提审、发布")
	return nil
}

func runWeChat(ctx context.Context, config model.Config, root, source string, check bool) error {
	binary := config.NodePath
	if binary == "" {
		var err error
		binary, err = exec.LookPath("node")
		if err != nil {
			return fmt.Errorf("未找到 Node.js，请安装或填写 node.exe 绝对路径")
		}
	} else if !filepath.IsAbs(binary) {
		return fmt.Errorf("Node.js 必须是可执行文件的绝对路径")
	}
	input := struct {
		Root   string `json:"root"`
		Source string `json:"source"`
		Check  bool   `json:"check"`
		model.Config
	}{Root: root, Source: source, Check: check, Config: config}
	raw, err := json.Marshal(input)
	if err != nil {
		return err
	}
	command := exec.CommandContext(ctx, binary, "-e", wechatScript)
	command.Dir = root
	command.Stdin = bytes.NewReader(raw)
	command.WaitDelay = 3 * time.Second
	prepareCommand(command)
	output := &limitedOutput{}
	command.Stdout = output
	command.Stderr = output
	if err := command.Run(); err != nil {
		return fmt.Errorf("微信上传工具失败（需预先安装 miniprogram-ci）: %w\n%s", err, output.String())
	}
	return nil
}

type limitedOutput struct {
	mu   sync.Mutex
	data []byte
}

func (w *limitedOutput) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.data = append(w.data, data...)
	if len(w.data) > 8192 {
		w.data = w.data[len(w.data)-8192:]
	}
	return len(data), nil
}
func (w *limitedOutput) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return strings.TrimSpace(string(w.data))
}
