package wails

import (
	"context"

	"devenv/internal/app"
	"devenv/internal/domain/plugin"
	"devenv/internal/domain/project"
	"devenv/internal/ports"
)

type ProjectAPI struct {
	ctx        context.Context
	service    *app.ProjectService
	deployment *app.DeploymentService
}

func NewProjectAPI(service *app.ProjectService, deployment *app.DeploymentService) *ProjectAPI {
	return &ProjectAPI{
		ctx:        context.Background(),
		service:    service,
		deployment: deployment,
	}
}

func (a *ProjectAPI) SetContext(ctx context.Context) {
	a.ctx = ctx
}

func (a *ProjectAPI) List() ([]ProjectDTO, error) {
	list, err := a.service.List(a.ctx)
	if err != nil {
		return nil, err
	}
	result := make([]ProjectDTO, 0, len(list))
	for _, item := range list {
		result = append(result, convertProject(item))
	}
	return result, nil
}

func (a *ProjectAPI) Get(id int64) (*ProjectDTO, error) {
	item, err := a.service.Get(a.ctx, id)
	if err != nil {
		return nil, err
	}
	dto := convertProject(*item)
	return &dto, nil
}

func (a *ProjectAPI) Create(in ProjectCreateInput) (*ProjectDTO, error) {
	item, err := a.service.Create(a.ctx, project.CreateInput{
		Kind:    in.Kind,
		Name:    in.Name,
		Host:    in.Host,
		Root:    in.Root,
		Run:     in.Run,
		Command: in.Command,
	})
	if err != nil {
		return nil, err
	}
	dto := convertProject(*item)
	return &dto, nil
}

func (a *ProjectAPI) Update(in ProjectUpdateInput) (*ProjectDTO, error) {
	item, err := a.service.Update(a.ctx, project.UpdateInput{
		ID:        in.ID,
		Hosts:     in.Hosts,
		Root:      in.Root,
		Run:       in.Run,
		Software:  in.Software,
		Command:   in.Command,
		Nginx:     in.Nginx,
		Autostart: in.Autostart,
	})
	if err != nil {
		return nil, err
	}
	dto := convertProject(*item)
	return &dto, nil
}

func (a *ProjectAPI) FindRun(root, kind string) string {
	return a.service.FindRun(root, kind)
}

func (a *ProjectAPI) FindCommands(root, kind string) []string {
	return a.service.FindCommands(root, kind)
}

func (a *ProjectAPI) Templates() []ProjectTemplateDTO {
	items := a.service.Templates()
	result := make([]ProjectTemplateDTO, 0, len(items))
	for _, item := range items {
		capabilities := make([]string, 0, len(item.Capabilities))
		for _, capability := range item.Capabilities {
			capabilities = append(capabilities, string(capability))
		}
		result = append(result, ProjectTemplateDTO{
			Kind:               item.Kind,
			Label:              item.Label,
			Icon:               item.Icon,
			Description:        item.Description,
			RuntimeKind:        item.RuntimeKind,
			CommandMode:        string(item.CommandMode),
			DefaultCommand:     item.DefaultCommand,
			CommandLabel:       item.CommandLabel,
			CommandPlaceholder: item.CommandPlaceholder,
			Capabilities:       capabilities,
		})
	}
	return result
}

func (a *ProjectAPI) Start(id int64) Result {
	if err := a.service.Start(a.ctx, id); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "项目已启动"}
}

func (a *ProjectAPI) Stop(id int64) Result {
	if err := a.service.Stop(a.ctx, id); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "项目已停止"}
}

func (a *ProjectAPI) Restart(id int64) Result {
	if err := a.service.Restart(a.ctx, id); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "项目已重启"}
}

func (a *ProjectAPI) Delete(id int64) Result {
	if err := a.service.Delete(a.ctx, id); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "项目记录已删除"}
}

func (a *ProjectAPI) Logs(id int64, kind string, limit int) ([]string, error) {
	return a.service.Logs(a.ctx, id, kind, limit)
}

func (a *ProjectAPI) Plugins(id int64) ([]ProjectPluginDTO, error) {
	list, err := a.service.Plugins(a.ctx, id)
	if err != nil {
		return nil, err
	}
	result := make([]ProjectPluginDTO, 0, len(list))
	for _, item := range list {
		actions := make([]PluginActionDTO, 0, len(item.Actions))
		for _, action := range item.Actions {
			actions = append(actions, PluginActionDTO{ID: action.ID, Name: action.Name})
		}
		result = append(result, ProjectPluginDTO{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			Available:   item.Available,
			Reason:      item.Reason,
			Version:     item.Version,
			Actions:     actions,
		})
	}
	return result, nil
}

func (a *ProjectAPI) RunPlugin(in ProjectPluginActionInput) Result {
	err := a.service.RunPlugin(a.ctx, plugin.ActionInput{
		ProjectID: in.ProjectID,
		PluginID:  in.PluginID,
		ActionID:  in.ActionID,
	})
	if err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "插件任务已启动"}
}

func (a *ProjectAPI) OpenNginxConfig(id int64) Result {
	if err := a.service.OpenNginxConfig(a.ctx, id); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: ""}
}

func (a *ProjectAPI) ReadNginxConfig(id int64) (string, error) {
	return a.service.ReadNginxConfig(a.ctx, id)
}

func (a *ProjectAPI) WriteNginxConfig(id int64, content string) Result {
	if err := a.service.WriteNginxConfig(a.ctx, id, content); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "Nginx 配置已保存"}
}

func (a *ProjectAPI) GitStatuses() ([]ProjectGitStatusDTO, error) {
	items, err := a.service.GitStatuses(a.ctx)
	if err != nil {
		return nil, err
	}
	result := make([]ProjectGitStatusDTO, 0, len(items))
	for _, item := range items {
		result = append(result, convertProjectGitStatus(item))
	}
	return result, nil
}

func (a *ProjectAPI) GitStatus(id int64) (ProjectGitStatusDTO, error) {
	item, err := a.service.GitStatus(a.ctx, id)
	if err != nil {
		return ProjectGitStatusDTO{}, err
	}
	return convertProjectGitStatus(item), nil
}

func (a *ProjectAPI) InitializeGit(id int64) Result {
	if err := a.service.InitializeGit(a.ctx, id); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "Git 仓库已初始化"}
}

func (a *ProjectAPI) CommitGit(input ProjectGitCommitInput) Result {
	if err := a.service.CommitGit(a.ctx, input.ProjectID, input.Message, input.BumpVersion); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "Git 提交成功"}
}

func convertProject(view app.ProjectView) ProjectDTO {
	hosts := view.Hosts
	if hosts == nil {
		hosts = make([]string, 0)
	}
	return ProjectDTO{
		ID:        view.ID,
		Name:      view.Name,
		Kind:      view.Kind,
		Hosts:     hosts,
		Root:      view.Root,
		Run:       view.Run,
		Path:      view.Path,
		Software:  view.Software,
		Version:   view.Version,
		Command:   view.Command,
		Nginx:     view.Nginx,
		Port:      view.Port,
		Autostart: view.Autostart,
		State:     view.State,
		URL:       view.URL,
		Need:      view.Need,
		Created:   view.Created,
		Updated:   view.Updated,
	}
}

func convertProjectGitStatus(item ports.ProjectGitStatus) ProjectGitStatusDTO {
	remotes := make([]ProjectGitRemoteDTO, 0, len(item.Remotes))
	for _, remote := range item.Remotes {
		remotes = append(remotes, ProjectGitRemoteDTO{Name: remote.Name, URL: remote.URL, PushURL: remote.PushURL})
	}
	files := make([]ProjectGitFileDTO, 0, len(item.Files))
	for _, file := range item.Files {
		files = append(files, ProjectGitFileDTO{
			Path: file.Path, Status: file.Status, Staged: file.Staged,
			Unstaged: file.Unstaged, Untracked: file.Untracked,
		})
	}
	return ProjectGitStatusDTO{
		Remotes: remotes, Upstream: item.Upstream, Ahead: item.Ahead, Behind: item.Behind,
		ProjectID: item.ProjectID, ProjectName: item.ProjectName,
		Available: item.Available, Repository: item.Repository, Nested: item.Nested,
		Branch: item.Branch, LastCommit: item.LastCommit,
		CommitMessage: item.CommitMessage, Changed: item.Changed, Staged: item.Staged,
		Unstaged: item.Unstaged, Untracked: item.Untracked, Files: files,
		Version: item.Version, NextVersion: item.NextVersion, VersionFile: item.VersionFile,
		VersionCanBump: item.VersionCanBump, VersionError: item.VersionError, Error: item.Error,
	}
}

func (a *ProjectAPI) SaveGitRemote(input ProjectGitRemoteInput) Result {
	if err := a.service.SaveGitRemote(a.ctx, input.ProjectID, ports.GitRemote{Name: input.Name, URL: input.URL, PushURL: input.PushURL}); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "远端地址已保存"}
}

func (a *ProjectAPI) RemoveGitRemote(id int64, name string) Result {
	if err := a.service.RemoveGitRemote(a.ctx, id, name); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "远端配置已移除，不影响远端仓库本身"}
}

func (a *ProjectAPI) FetchGit(id int64, name string) Result {
	if err := a.service.FetchGit(a.ctx, id, name); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "远端状态已更新，未合并或修改工作区"}
}

func (a *ProjectAPI) PushGit(id int64, name string) Result {
	if err := a.service.PushGit(a.ctx, id, name); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "当前分支已推送"}
}
