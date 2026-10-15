package plugin

const (
	HANDLER_COMPOSER = "composer"
	HANDLER_NODE     = "node"
	HANDLER_RUNTIME  = "runtime"
)

// Definition describes when a project plugin appears and which service
// handler executes its actions. Definitions are the extension registry for
// project integrations; the project table does not need plugin-specific fields.
type Definition struct {
	ID               string
	Name             string
	Description      string
	ProjectKinds     []string
	RequiredFiles    []string
	RequiredAnyFiles []string
	Handler          string
	Actions          []Action
}

// Definitions returns all built-in project integrations.
func Definitions() []Definition {
	return []Definition{
		{
			ID: ID_COMPOSER, Name: "Composer", Description: "管理 PHP 项目的依赖包和自动加载文件",
			ProjectKinds: []string{"php"}, RequiredFiles: []string{"composer.json"}, Handler: HANDLER_COMPOSER,
			Actions: []Action{{ID: ACTION_INSTALL, Name: "安装依赖"}, {ID: ACTION_UPDATE, Name: "更新依赖"}, {ID: ACTION_DUMP_AUTOLOAD, Name: "生成自动加载"}},
		},
		{
			ID: ID_NODE_PACKAGE, Name: "Node 包管理", Description: "根据项目锁文件选择包管理器并安装依赖",
			ProjectKinds: []string{"node", "electron", "miniprogram"}, RequiredFiles: []string{"package.json"}, Handler: HANDLER_NODE,
			Actions: []Action{{ID: ACTION_INSTALL, Name: "安装依赖"}},
		},
		{
			ID: ID_PYTHON_PACKAGE, Name: "Python 依赖", Description: "根据 requirements.txt 或 pyproject.toml 安装 Python 项目依赖",
			ProjectKinds: []string{"python"}, RequiredAnyFiles: []string{"requirements.txt", "pyproject.toml"}, Handler: HANDLER_RUNTIME,
			Actions: []Action{{ID: ACTION_INSTALL, Name: "安装依赖"}},
		},
		{
			ID: ID_GO_MODULE, Name: "Go 模块", Description: "整理 Go 模块依赖并生成校验信息",
			ProjectKinds: []string{"go"}, RequiredFiles: []string{"go.mod"}, Handler: HANDLER_RUNTIME,
			Actions: []Action{{ID: ACTION_UPDATE, Name: "整理依赖"}},
		},
		{
			ID: ID_RUST_PACKAGE, Name: "Cargo", Description: "获取 Rust 项目的 Cargo 依赖",
			ProjectKinds: []string{"rust"}, RequiredFiles: []string{"Cargo.toml"}, Handler: HANDLER_RUNTIME,
			Actions: []Action{{ID: ACTION_INSTALL, Name: "获取依赖"}},
		},
		{
			ID: ID_JAVA_BUILD, Name: "Java 构建", Description: "解析 Maven 或 Gradle 项目的依赖",
			ProjectKinds: []string{"java"}, RequiredAnyFiles: []string{"pom.xml", "build.gradle", "build.gradle.kts"}, Handler: HANDLER_RUNTIME,
			Actions: []Action{{ID: ACTION_INSTALL, Name: "解析依赖"}},
		},
	}
}

// DefinitionFor finds a registered project plugin by identifier.
func DefinitionFor(pluginID string) (Definition, bool) {
	for _, definition := range Definitions() {
		if definition.ID == pluginID {
			return definition, true
		}
	}
	return Definition{}, false
}
