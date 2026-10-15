package project

import "devenv/internal/domain/software"

// ProjectCapability describes a reusable section of the project detail form.
type ProjectCapability string

const (
	CAPABILITY_DOMAIN        ProjectCapability = "domain"
	CAPABILITY_RUN_DIRECTORY ProjectCapability = "run-directory"
	CAPABILITY_RUNTIME       ProjectCapability = "runtime"
	CAPABILITY_COMMAND       ProjectCapability = "command"
	CAPABILITY_AUTOSTART     ProjectCapability = "autostart"
	CAPABILITY_NGINX         ProjectCapability = "nginx"
	CAPABILITY_ACCESS_LOG    ProjectCapability = "access-log"
	CAPABILITY_PHP_RUNTIME   ProjectCapability = "php-runtime"
)

// CommandMode identifies how a project chooses its development command.
type CommandMode string

const (
	COMMAND_MODE_INPUT          CommandMode = "input"
	COMMAND_MODE_PACKAGE_SCRIPT CommandMode = "package-script"
)

// ProjectTemplate is the backend-owned description of a project type.
// Shared detail sections are selected by capabilities instead of project kind.
type ProjectTemplate struct {
	Kind               string              `json:"kind"`
	Label              string              `json:"label"`
	Icon               string              `json:"icon"`
	Description        string              `json:"description"`
	RuntimeKind        string              `json:"runtimeKind"`
	CommandMode        CommandMode         `json:"commandMode"`
	DefaultCommand     string              `json:"defaultCommand"`
	CommandLabel       string              `json:"commandLabel"`
	CommandPlaceholder string              `json:"commandPlaceholder"`
	Capabilities       []ProjectCapability `json:"capabilities"`
}

// ProjectTemplates returns the built-in templates. New project types should
// add one template here instead of extending several unrelated switch blocks.
func ProjectTemplates() []ProjectTemplate {
	return []ProjectTemplate{
		{
			Kind: "static", Label: "静态", Icon: "pi pi-file", Description: "静态文件项目",
			Capabilities: []ProjectCapability{CAPABILITY_DOMAIN, CAPABILITY_RUN_DIRECTORY, CAPABILITY_NGINX, CAPABILITY_ACCESS_LOG},
		},
		{
			Kind: "php", Label: "PHP", Icon: "pi pi-code", Description: "PHP Web 项目",
			RuntimeKind:  software.KIND_PHP,
			Capabilities: []ProjectCapability{CAPABILITY_DOMAIN, CAPABILITY_RUN_DIRECTORY, CAPABILITY_RUNTIME, CAPABILITY_PHP_RUNTIME, CAPABILITY_NGINX, CAPABILITY_ACCESS_LOG},
		},
		{
			Kind: "node", Label: "Node.js", Icon: "pi pi-bolt", Description: "Node.js Web 项目",
			RuntimeKind: software.KIND_NODE, CommandMode: COMMAND_MODE_INPUT,
			CommandLabel: "Node 启动命令", CommandPlaceholder: "如 npm start 或 node server.js",
			Capabilities: []ProjectCapability{CAPABILITY_DOMAIN, CAPABILITY_RUN_DIRECTORY, CAPABILITY_RUNTIME, CAPABILITY_COMMAND, CAPABILITY_AUTOSTART, CAPABILITY_NGINX, CAPABILITY_ACCESS_LOG},
		},
		{
			Kind: "python", Label: "Python", Icon: "pi pi-code", Description: "Python Web 项目",
			RuntimeKind: software.KIND_PYTHON, CommandMode: COMMAND_MODE_INPUT,
			CommandLabel: "Python 启动命令", CommandPlaceholder: "如 python app.py",
			Capabilities: []ProjectCapability{CAPABILITY_DOMAIN, CAPABILITY_RUN_DIRECTORY, CAPABILITY_RUNTIME, CAPABILITY_COMMAND, CAPABILITY_AUTOSTART, CAPABILITY_NGINX, CAPABILITY_ACCESS_LOG},
		},
		{
			Kind: "go", Label: "Go", Icon: "pi pi-code", Description: "Go Web 项目",
			RuntimeKind: software.KIND_GO, CommandMode: COMMAND_MODE_INPUT,
			CommandLabel: "Go 启动命令", CommandPlaceholder: "如 go run .",
			Capabilities: []ProjectCapability{CAPABILITY_DOMAIN, CAPABILITY_RUN_DIRECTORY, CAPABILITY_RUNTIME, CAPABILITY_COMMAND, CAPABILITY_AUTOSTART, CAPABILITY_NGINX, CAPABILITY_ACCESS_LOG},
		},
		{
			Kind: "rust", Label: "Rust", Icon: "pi pi-code", Description: "Rust Web 项目",
			CommandMode: COMMAND_MODE_INPUT, CommandLabel: "Rust 启动命令", CommandPlaceholder: "如 cargo run",
			Capabilities: []ProjectCapability{CAPABILITY_DOMAIN, CAPABILITY_RUN_DIRECTORY, CAPABILITY_COMMAND, CAPABILITY_AUTOSTART, CAPABILITY_NGINX, CAPABILITY_ACCESS_LOG},
		},
		{
			Kind: "java", Label: "Java", Icon: "pi pi-code", Description: "Java Web 项目",
			RuntimeKind: software.KIND_JAVA, CommandMode: COMMAND_MODE_INPUT,
			CommandLabel: "Java 启动命令", CommandPlaceholder: "如 java Main.java 或 mvn spring-boot:run",
			Capabilities: []ProjectCapability{CAPABILITY_DOMAIN, CAPABILITY_RUN_DIRECTORY, CAPABILITY_RUNTIME, CAPABILITY_COMMAND, CAPABILITY_AUTOSTART, CAPABILITY_NGINX, CAPABILITY_ACCESS_LOG},
		},
		{
			Kind: "electron", Label: "Electron", Icon: "pi pi-desktop", Description: "Electron 本地调试项目",
			RuntimeKind: software.KIND_NODE, CommandMode: COMMAND_MODE_INPUT, DefaultCommand: "pnpm dev",
			CommandLabel: "Electron 启动命令", CommandPlaceholder: "pnpm dev",
			Capabilities: []ProjectCapability{CAPABILITY_RUNTIME, CAPABILITY_COMMAND},
		},
		{
			Kind: "miniprogram", Label: "小程序", Icon: "pi pi-mobile", Description: "小程序本地调试项目",
			RuntimeKind: software.KIND_NODE, CommandMode: COMMAND_MODE_PACKAGE_SCRIPT,
			CommandLabel: "小程序启动命令", CommandPlaceholder: "选择 package.json 中的运行脚本",
			Capabilities: []ProjectCapability{CAPABILITY_RUNTIME, CAPABILITY_COMMAND},
		},
	}
}

// ProjectTemplateForKind finds the template that owns a project kind.
func ProjectTemplateForKind(kind string) (ProjectTemplate, bool) {
	for _, item := range ProjectTemplates() {
		if item.Kind == kind {
			return item, true
		}
	}
	return ProjectTemplate{}, false
}

// HasCapability reports whether a project kind exposes a detail section.
func HasCapability(kind string, capability ProjectCapability) bool {
	template, ok := ProjectTemplateForKind(kind)
	if !ok {
		return false
	}
	for _, item := range template.Capabilities {
		if item == capability {
			return true
		}
	}
	return false
}
