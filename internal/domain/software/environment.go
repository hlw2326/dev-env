package software

// CommandLocation describes a command discovered in the persisted Windows PATH.
type CommandLocation struct {
	Path         string
	Command      string
	Role         string
	Source       string
	Versions     []string
	Error        string
	Managed      bool
	Active       bool
	ServiceName  string
	ServiceState string
}

type CommandEnvironment struct {
	Kind           string
	Name           string
	Error          string
	Command        string
	Entries        []CommandLocation
	DefaultVersion string
	Conflict       bool
	Warnings       []string
}

type EnvironmentScan struct {
	Entries  []CommandLocation
	Warnings []string
}
