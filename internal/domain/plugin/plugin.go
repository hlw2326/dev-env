package plugin

const (
	ID_COMPOSER       = "composer"
	ID_NODE_PACKAGE   = "node-package"
	ID_PYTHON_PACKAGE = "python-package"
	ID_GO_MODULE      = "go-module"
	ID_RUST_PACKAGE   = "rust-package"
	ID_JAVA_BUILD     = "java-build"

	ACTION_INSTALL       = "install"
	ACTION_UPDATE        = "update"
	ACTION_DUMP_AUTOLOAD = "dump-autoload"
)

type Action struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ProjectPlugin struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Available   bool     `json:"available"`
	Reason      string   `json:"reason"`
	Version     string   `json:"version"`
	Actions     []Action `json:"actions"`
}

type ActionInput struct {
	ProjectID int64  `json:"projectId"`
	PluginID  string `json:"pluginId"`
	ActionID  string `json:"actionId"`
}
