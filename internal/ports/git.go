package ports

import "context"

type GitFileStatus struct {
	Path      string
	Status    string
	Staged    bool
	Unstaged  bool
	Untracked bool
}

type GitStatus struct {
	Available      bool
	Repository     bool
	Nested         bool
	Branch         string
	Remotes        []GitRemote
	Upstream       string
	Ahead          int
	Behind         int
	LastCommit     string
	CommitMessage  string
	Changed        int
	Staged         int
	Unstaged       int
	Untracked      int
	Files          []GitFileStatus
	Version        string
	NextVersion    string
	VersionFile    string
	VersionCanBump bool
	VersionError   string
	Error          string
}

type GitRemote struct {
	Name    string
	URL     string
	PushURL string
}

type ProjectGitStatus struct {
	ProjectID   int64
	ProjectName string
	GitStatus
}

type GitManager interface {
	Status(ctx context.Context, root, kind string) (GitStatus, error)
	Initialize(ctx context.Context, root string) error
	Commit(ctx context.Context, root, kind, message string, bumpVersion bool) error
	SaveRemote(ctx context.Context, root string, remote GitRemote) error
	RemoveRemote(ctx context.Context, root, name string) error
	Fetch(ctx context.Context, root, name string) error
	Push(ctx context.Context, root, name string) error
}
