package wails

type Result struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

type ConfigResult struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
	Restart bool   `json:"restart"`
}

type ProjectDTO struct {
	ID        int64    `json:"id"`
	Name      string   `json:"name"`
	Kind      string   `json:"kind"`
	Hosts     []string `json:"hosts"`
	Root      string   `json:"root"`
	Run       string   `json:"run"`
	Path      string   `json:"path"`
	Software  *int64   `json:"software"`
	Version   string   `json:"version"`
	Command   string   `json:"command"`
	Nginx     string   `json:"nginx"`
	Port      int      `json:"port"`
	Autostart bool     `json:"autostart"`
	State     string   `json:"state"`
	URL       string   `json:"url"`
	Need      string   `json:"need"`
	Created   string   `json:"created"`
	Updated   string   `json:"updated"`
}

type ProjectTemplateDTO struct {
	Kind               string   `json:"kind"`
	Label              string   `json:"label"`
	Icon               string   `json:"icon"`
	Description        string   `json:"description"`
	RuntimeKind        string   `json:"runtimeKind"`
	CommandMode        string   `json:"commandMode"`
	DefaultCommand     string   `json:"defaultCommand"`
	CommandLabel       string   `json:"commandLabel"`
	CommandPlaceholder string   `json:"commandPlaceholder"`
	Capabilities       []string `json:"capabilities"`
}

type ProjectCreateInput struct {
	Kind    string `json:"kind"`
	Name    string `json:"name"`
	Host    string `json:"host"`
	Root    string `json:"root"`
	Run     string `json:"run"`
	Command string `json:"command"`
}

type ProjectUpdateInput struct {
	ID        int64    `json:"id"`
	Hosts     []string `json:"hosts"`
	Root      string   `json:"root"`
	Run       string   `json:"run"`
	Software  *int64   `json:"software"`
	Command   string   `json:"command"`
	Nginx     string   `json:"nginx"`
	Autostart bool     `json:"autostart"`
}

type ProjectPluginActionInput struct {
	ProjectID int64  `json:"projectId"`
	PluginID  string `json:"pluginId"`
	ActionID  string `json:"actionId"`
}

type ProjectGitFileDTO struct {
	Path      string `json:"path"`
	Status    string `json:"status"`
	Staged    bool   `json:"staged"`
	Unstaged  bool   `json:"unstaged"`
	Untracked bool   `json:"untracked"`
}

type ProjectGitStatusDTO struct {
	Remotes        []ProjectGitRemoteDTO `json:"remotes"`
	Upstream       string                `json:"upstream"`
	Ahead          int                   `json:"ahead"`
	Behind         int                   `json:"behind"`
	ProjectID      int64                 `json:"projectId"`
	ProjectName    string                `json:"projectName"`
	Available      bool                  `json:"available"`
	Repository     bool                  `json:"repository"`
	Nested         bool                  `json:"nested"`
	Branch         string                `json:"branch"`
	LastCommit     string                `json:"lastCommit"`
	CommitMessage  string                `json:"commitMessage"`
	Changed        int                   `json:"changed"`
	Staged         int                   `json:"staged"`
	Unstaged       int                   `json:"unstaged"`
	Untracked      int                   `json:"untracked"`
	Files          []ProjectGitFileDTO   `json:"files"`
	Version        string                `json:"version"`
	NextVersion    string                `json:"nextVersion"`
	VersionFile    string                `json:"versionFile"`
	VersionCanBump bool                  `json:"versionCanBump"`
	VersionError   string                `json:"versionError"`
	Error          string                `json:"error"`
}

type ProjectGitCommitInput struct {
	ProjectID   int64  `json:"projectId"`
	Message     string `json:"message"`
	BumpVersion bool   `json:"bumpVersion"`
}

type ProjectGitRemoteDTO struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	PushURL string `json:"pushUrl"`
}

type ProjectGitRemoteInput struct {
	ProjectID int64  `json:"projectId"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	PushURL   string `json:"pushUrl"`
}

type PluginActionDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ProjectPluginDTO struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Available   bool              `json:"available"`
	Reason      string            `json:"reason"`
	Version     string            `json:"version"`
	Actions     []PluginActionDTO `json:"actions"`
}

type ItemDTO struct {
	Kind        string `json:"kind"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Count       int    `json:"count"`
	Recommended string `json:"recommended"`
}

type SoftwareDTO struct {
	ID        int64  `json:"id"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Version   string `json:"version"`
	Path      string `json:"path"`
	State     string `json:"state"`
	Progress  int    `json:"progress"`
	Error     string `json:"error"`
	Main      bool   `json:"main"`
	Port      int    `json:"port"`
	Autostart bool   `json:"autostart"`
	Running   bool   `json:"running"`
	PID       int    `json:"pid"`
	Daemon    bool   `json:"daemon"`
	Service   bool   `json:"service"`
}

type VersionDTO struct {
	Version     string `json:"version"`
	Recommended bool   `json:"recommended"`
	Installed   bool   `json:"installed"`
	Note        string `json:"note"`
}

type ConfigItem struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

type ConfigDTO struct {
	Items []ConfigItem `json:"items"`
	File  string       `json:"file"`
}

type ConfigInput struct {
	ID    int64             `json:"id"`
	Items map[string]string `json:"items"`
}

type ExtensionDTO struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

type InstallInput struct {
	Kind    string `json:"kind"`
	Version string `json:"version"`
}

type SettingsDTO struct {
	Theme             string `json:"theme"`
	Root              string `json:"root"`
	Port              int    `json:"port"`
	Hosts             bool   `json:"hosts"`
	DeveloperToolsCLI string `json:"developerToolsCLI"`
	Listening         bool   `json:"listening"`
	Error             string `json:"error"`
}

type SettingsInput struct {
	Theme             string `json:"theme"`
	Root              string `json:"root"`
	Port              int    `json:"port"`
	Hosts             bool   `json:"hosts"`
	DeveloperToolsCLI string `json:"developerToolsCLI"`
}
