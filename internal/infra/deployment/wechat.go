package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	model "devenv/internal/domain/deployment"
	"devenv/internal/shared/paths"
)

const weChatCLIHelp = "--help"

type WeChat struct{ base string }

func NewWeChat(base string) *WeChat { return &WeChat{base: base} }

func checkWeChat(config model.Config, root string) error {
	source, err := paths.DeploymentSource(root, config.Source)
	if err != nil {
		return fmt.Errorf("小程序编译目录 %q 不存在或不可用，请先在项目根目录完成 uni 编译: %w", config.Source, err)
	}
	if err := checkDeveloperToolsCLI(config.DeveloperToolsCLI); err != nil {
		return err
	}
	content, err := os.ReadFile(filepath.Join(source, "project.config.json"))
	if err != nil {
		return fmt.Errorf("发布目录必须包含构建好的 project.config.json: %w", err)
	}
	var projectConfig map[string]json.RawMessage
	if err := json.Unmarshal(content, &projectConfig); err != nil {
		return fmt.Errorf("project.config.json 格式无效: %w", err)
	}
	var appID string
	if err := json.Unmarshal(projectConfig["appid"], &appID); err != nil {
		return fmt.Errorf("project.config.json 的 appid 格式无效: %w", err)
	}
	if strings.TrimSpace(appID) == "" {
		return fmt.Errorf("project.config.json 的 appid 不能为空")
	}
	for _, key := range []string{"miniprogramRoot", "srcMiniprogramRoot", "cloudfunctionRoot", "pluginRoot"} {
		if raw, ok := projectConfig[key]; ok {
			var value string
			if err := json.Unmarshal(raw, &value); err != nil {
				return fmt.Errorf("project.config.json 的 %s 格式无效: %w", key, err)
			}
			if value != "" {
				if _, err := paths.DeploymentSource(source, value); err != nil {
					return fmt.Errorf("project.config.json 的 %s 不安全或不存在: %w", key, err)
				}
			}
		}
	}
	return nil
}

func checkDeveloperToolsCLI(binary string) error {
	if strings.TrimSpace(binary) == "" {
		return fmt.Errorf("请选择微信开发者工具 CLI；通常是安装目录下的 cli.bat")
	}
	if !filepath.IsAbs(binary) {
		return fmt.Errorf("微信开发者工具 CLI 必须是绝对路径")
	}
	info, err := os.Stat(binary)
	if err != nil {
		return fmt.Errorf("微信开发者工具 CLI 不存在: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("微信开发者工具 CLI 不能是目录")
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
	plan.Target = "微信开发版本 · " + strings.Join(config.AppIDValues(), ", ") + " · " + config.Version
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
	// Revalidate the project configuration against the filtered snapshot.
	stagedConfig := config
	stagedConfig.Source = "."
	if err := checkWeChat(stagedConfig, snapshot); err != nil {
		return err
	}
	report(0, "微信开发者工具正在上传，请稍候；取消或网络中断后需在微信后台核对结果")
	if err := runWeChat(ctx, config, root, snapshot, false); err != nil {
		return err
	}
	report(len(plan.Files)*len(config.AppIDValues()), "代码已上传到微信开发版本；请到微信后台手动提审、发布")
	return nil
}

func runWeChat(ctx context.Context, config model.Config, root, source string, check bool) error {
	if err := checkDeveloperToolsCLI(config.DeveloperToolsCLI); err != nil {
		return err
	}
	if check {
		return executeWeChatCLI(ctx, config, root, []string{weChatCLIHelp}, true, "")
	}
	projectPath := source
	var err error
	if projectPath == "" {
		projectPath, err = paths.DeploymentSource(root, config.Source)
		if err != nil {
			return err
		}
	} else {
		if !filepath.IsAbs(projectPath) {
			return fmt.Errorf("微信上传快照路径必须是绝对路径")
		}
		info, err := os.Stat(projectPath)
		if err != nil {
			return fmt.Errorf("微信上传快照不存在: %w", err)
		}
		if !info.IsDir() {
			return fmt.Errorf("微信上传快照必须是目录")
		}
	}
	projectConfigPath := filepath.Join(projectPath, "project.config.json")
	original, err := os.ReadFile(projectConfigPath)
	if err != nil {
		return fmt.Errorf("读取小程序 project.config.json 失败: %w", err)
	}
	defer func() { _ = os.WriteFile(projectConfigPath, original, 0o600) }()
	for _, appID := range config.AppIDValues() {
		if err := writeWeChatAppID(projectConfigPath, original, appID); err != nil {
			return err
		}
		args := []string{
			"upload",
			"--project", projectPath,
			"--version", config.Version,
			"--desc", config.Description,
			"--robot", strconv.Itoa(config.Robot),
		}
		if err := executeWeChatCLI(ctx, config, root, args, false, appID); err != nil {
			return err
		}
	}
	return nil
}

func writeWeChatAppID(path string, original []byte, appID string) error {
	var projectConfig map[string]json.RawMessage
	if err := json.Unmarshal(original, &projectConfig); err != nil {
		return fmt.Errorf("project.config.json 格式无效: %w", err)
	}
	value, err := json.Marshal(appID)
	if err != nil {
		return err
	}
	projectConfig["appid"] = value
	updated, err := json.MarshalIndent(projectConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("生成 AppID 配置失败: %w", err)
	}
	if err := os.WriteFile(path, updated, 0o600); err != nil {
		return fmt.Errorf("写入 AppID 配置失败: %w", err)
	}
	return nil
}

func executeWeChatCLI(ctx context.Context, config model.Config, root string, args []string, check bool, appID string) error {
	command := newWeChatCLICommand(ctx, config.DeveloperToolsCLI, args...)
	command.Dir = root
	command.WaitDelay = 3 * time.Second
	prepareCommand(command)
	output := &limitedOutput{}
	command.Stdout = output
	command.Stderr = output
	if err := command.Run(); err != nil {
		message := "微信开发者工具 CLI 执行失败"
		if check {
			message = "微信开发者工具 CLI 检查失败"
		}
		if appID != "" {
			message += "（AppID " + appID + "）"
		}
		return fmt.Errorf("%s，请确认已在微信开发者工具中开启服务端口并登录有上传权限的账号: %w\n%s", message, err, output.String())
	}
	return nil
}

func newWeChatCLICommand(ctx context.Context, binary string, args ...string) *exec.Cmd {
	if runtime.GOOS == "windows" && (strings.EqualFold(filepath.Ext(binary), ".bat") || strings.EqualFold(filepath.Ext(binary), ".cmd")) {
		commandLine := "call " + quoteWindowsArgument(binary)
		for _, arg := range args {
			commandLine += " " + quoteWindowsArgument(arg)
		}
		return exec.CommandContext(ctx, "cmd.exe", "/d", "/c", commandLine)
	}
	return exec.CommandContext(ctx, binary, args...)
}

func quoteWindowsArgument(value string) string {
	if value != "" && !strings.ContainsAny(value, " \t\"&|<>^%!") {
		return value
	}
	return `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
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
