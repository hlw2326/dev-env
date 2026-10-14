package wails

import (
	"context"
	"time"

	"devenv/internal/app"
	"devenv/internal/domain/runtime"
	"devenv/internal/domain/settings"
	"devenv/internal/domain/site"
	"devenv/internal/transport/wails/dto"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type baseAPI struct {
	ctx context.Context
}

func (b *baseAPI) SetContext(ctx context.Context) {
	b.ctx = ctx
}

func (b *baseAPI) getCtx() context.Context {
	if b.ctx != nil {
		return b.ctx
	}
	return context.Background()
}

// ----------------- SiteAPI -----------------

type SiteAPI struct {
	baseAPI
	svc *app.SiteService
}

func NewSiteAPI(svc *app.SiteService) *SiteAPI {
	return &SiteAPI{svc: svc}
}

func toSiteDTO(cur site.Site) dto.SiteDTO {
	domains := cur.Domains
	if domains == nil {
		domains = make([]string, 0)
	}
	meta := cur.Meta
	if meta.History == nil {
		meta.History = make([]site.Snapshot, 0)
	}
	if meta.Env == nil {
		meta.Env = make(map[string]string)
	}
	return dto.SiteDTO{
		ID:        cur.ID,
		Name:      cur.Name,
		Kind:      string(cur.Kind),
		Domains:   domains,
		Root:      cur.Root,
		RuntimeID: cur.RuntimeID,
		ServiceID: cur.ServiceID,
		Git:       cur.Git,
		Remote:    cur.Remote,
		Branch:    cur.Branch,
		Commit:    cur.Commit,
		PulledAt:  cur.PulledAt,
		Conf:      cur.Conf,
		Port:      cur.Port,
		TLS:       cur.TLS,
		Cert:      cur.Cert,
		Key:       cur.Key,
		Issuer:    cur.Issuer,
		Proxy:     cur.Proxy,
		Upstream:  cur.Upstream,
		Meta:      meta,
		Status:    string(cur.Status),
		CreatedAt: cur.CreatedAt.Format(time.RFC3339),
		UpdatedAt: cur.UpdatedAt.Format(time.RFC3339),
	}
}

func (a *SiteAPI) List() ([]dto.SiteDTO, error) {
	list, err := a.svc.List(a.getCtx())
	if err != nil {
		return nil, err
	}
	res := make([]dto.SiteDTO, 0, len(list))
	for _, cur := range list {
		res = append(res, toSiteDTO(cur))
	}
	return res, nil
}

func (a *SiteAPI) Get(id int64) (*dto.SiteDTO, error) {
	cur, err := a.svc.Get(a.getCtx(), id)
	if err != nil {
		return nil, err
	}
	res := toSiteDTO(*cur)
	return &res, nil
}

func (a *SiteAPI) Create(in dto.CreateSiteInput) (*dto.SiteDTO, error) {
	cur, err := a.svc.Create(a.getCtx(), site.CreateInput{
		Name:        in.Name,
		Kind:        site.Kind(in.Kind),
		Domains:     in.Domains,
		Root:        in.Root,
		RuntimeID:   in.RuntimeID,
		ServiceID:   in.ServiceID,
		AutoService: in.AutoService,
		Git:         in.Git,
		Remote:      in.Remote,
		Branch:      in.Branch,
		Port:        in.Port,
		TLS:         in.TLS,
		Issuer:      in.Issuer,
		Cert:        in.Cert,
		Key:         in.Key,
		Upstream:    in.Upstream,
		Entry:       in.Entry,
		Cmd:         in.Cmd,
		AutoStart:   in.AutoStart,
		StartNow:    in.StartNow,
		Custom:      in.Custom,
		Env:         in.Env,
	})
	if err != nil {
		return nil, err
	}
	res := toSiteDTO(*cur)
	return &res, nil
}

func (a *SiteAPI) Update(in dto.UpdateSiteInput) (*dto.SiteDTO, error) {
	cur, err := a.svc.Update(a.getCtx(), site.UpdateInput{
		ID:        in.ID,
		Name:      in.Name,
		Domains:   in.Domains,
		Root:      in.Root,
		RuntimeID: in.RuntimeID,
		ServiceID: in.ServiceID,
		Port:      in.Port,
		TLS:       in.TLS,
		Issuer:    in.Issuer,
		Cert:      in.Cert,
		Key:       in.Key,
		Upstream:  in.Upstream,
		Entry:     in.Entry,
		Cmd:       in.Cmd,
		Custom:    in.Custom,
		AutoStart: in.AutoStart,
	})
	if err != nil {
		return nil, err
	}
	res := toSiteDTO(*cur)
	return &res, nil
}

func (a *SiteAPI) Start(id int64) (dto.OperationResultDTO, error) {
	if err := a.svc.Start(a.getCtx(), id); err != nil {
		return dto.OperationResultDTO{Success: false, Message: err.Error()}, err
	}
	return dto.OperationResultDTO{Success: true}, nil
}

func (a *SiteAPI) Stop(id int64) (dto.OperationResultDTO, error) {
	if err := a.svc.Stop(a.getCtx(), id); err != nil {
		return dto.OperationResultDTO{Success: false, Message: err.Error()}, err
	}
	return dto.OperationResultDTO{Success: true}, nil
}

func (a *SiteAPI) Restart(id int64) (dto.OperationResultDTO, error) {
	if err := a.svc.Restart(a.getCtx(), id); err != nil {
		return dto.OperationResultDTO{Success: false, Message: err.Error()}, err
	}
	return dto.OperationResultDTO{Success: true}, nil
}

func (a *SiteAPI) Delete(in dto.DeleteSiteInput) (dto.OperationResultDTO, error) {
	if err := a.svc.Delete(a.getCtx(), in.ID, in.Purge); err != nil {
		return dto.OperationResultDTO{Success: false, Message: err.Error()}, err
	}
	return dto.OperationResultDTO{Success: true}, nil
}

func (a *SiteAPI) Conf(id int64) (string, error) {
	return a.svc.ReadConf(a.getCtx(), id)
}

func (a *SiteAPI) RenderConf(id int64) (string, error) {
	return a.svc.RenderConf(a.getCtx(), id)
}

func (a *SiteAPI) SaveConf(in dto.SaveConfInput) (dto.OperationResultDTO, error) {
	if err := a.svc.WriteConf(a.getCtx(), in.SiteID, in.Content, in.Summary); err != nil {
		return dto.OperationResultDTO{Success: false, Message: err.Error()}, err
	}
	return dto.OperationResultDTO{Success: true}, nil
}

func (a *SiteAPI) CheckConf(content string) dto.ConfCheckDTO {
	ok, msg := a.svc.CheckConf(a.getCtx(), content)
	return dto.ConfCheckDTO{
		Valid:   ok,
		Message: msg,
	}
}

func (a *SiteAPI) Logs(in dto.SiteLogInput) (dto.LogResultDTO, error) {
	lines, err := a.svc.Logs(a.getCtx(), in.SiteID, in.Kind, in.Limit)
	if err != nil {
		return dto.LogResultDTO{Lines: make([]string, 0)}, err
	}
	return dto.LogResultDTO{Lines: lines}, nil
}

// ----------------- GitAPI -----------------

type GitAPI struct {
	baseAPI
	svc *app.GitService
}

func NewGitAPI(svc *app.GitService) *GitAPI {
	return &GitAPI{svc: svc}
}

func (a *GitAPI) Status(siteID int64) (*dto.GitStatusDTO, error) {
	st, err := a.svc.Status(a.getCtx(), siteID)
	if err != nil {
		return nil, err
	}
	modified := st.Modified
	if modified == nil {
		modified = make([]string, 0)
	}
	untracked := st.Untracked
	if untracked == nil {
		untracked = make([]string, 0)
	}
	return &dto.GitStatusDTO{
		Branch:    st.Branch,
		Commit:    st.Commit,
		Message:   st.Message,
		Clean:     st.Clean,
		Ahead:     st.Ahead,
		Behind:    st.Behind,
		Modified:  modified,
		Untracked: untracked,
		CheckedAt: st.CheckedAt,
	}, nil
}

func (a *GitAPI) Pull(siteID int64) (dto.OperationResultDTO, error) {
	if err := a.svc.Pull(a.getCtx(), siteID); err != nil {
		return dto.OperationResultDTO{Success: false, Message: err.Error()}, err
	}
	return dto.OperationResultDTO{Success: true}, nil
}

func (a *GitAPI) Commit(in dto.GitCommitInput) (dto.OperationResultDTO, error) {
	if err := a.svc.Commit(a.getCtx(), in.SiteID, in.Message, in.Files); err != nil {
		return dto.OperationResultDTO{Success: false, Message: err.Error()}, err
	}
	return dto.OperationResultDTO{Success: true}, nil
}

func (a *GitAPI) Push(siteID int64) (dto.OperationResultDTO, error) {
	if err := a.svc.Push(a.getCtx(), siteID); err != nil {
		return dto.OperationResultDTO{Success: false, Message: err.Error()}, err
	}
	return dto.OperationResultDTO{Success: true}, nil
}

// ----------------- RuntimeAPI -----------------

type RuntimeAPI struct {
	baseAPI
	svc *app.RuntimeService
}

func NewRuntimeAPI(svc *app.RuntimeService) *RuntimeAPI {
	return &RuntimeAPI{svc: svc}
}

func (a *RuntimeAPI) List() ([]dto.RuntimeDTO, error) {
	list, err := a.svc.List(a.getCtx())
	if err != nil {
		return nil, err
	}
	res := make([]dto.RuntimeDTO, 0, len(list))
	for _, v := range list {
		res = append(res, dto.RuntimeDTO{
			ID:           v.ID,
			RuntimeType:  string(v.RuntimeType),
			Version:      v.Version,
			Variant:      v.Variant,
			Architecture: v.Architecture,
			InstallPath:  v.InstallPath,
			Status:       string(v.Status),
			IsDefault:    v.IsDefault,
			SiteCount:    v.SiteCount,
			ServiceCount: v.ServiceCount,
		})
	}
	return res, nil
}

func (a *RuntimeAPI) Register(in dto.RegisterRuntimeInput) (*dto.RuntimeDTO, error) {
	v, err := a.svc.Register(a.getCtx(), in.RuntimeType, in.Version, in.Variant, in.Architecture, in.InstallPath)
	if err != nil {
		return nil, err
	}
	return &dto.RuntimeDTO{
		ID:           v.ID,
		RuntimeType:  string(v.RuntimeType),
		Version:      v.Version,
		Variant:      v.Variant,
		Architecture: v.Architecture,
		InstallPath:  v.InstallPath,
		Status:       string(v.Status),
		IsDefault:    v.IsDefault,
		SiteCount:    v.SiteCount,
		ServiceCount: v.ServiceCount,
	}, nil
}

func (a *RuntimeAPI) Install(in dto.InstallRuntimeInput) (dto.InstallResultDTO, error) {
	res, err := a.svc.Install(a.getCtx(), runtime.InstallInput{
		RuntimeType:  runtime.RuntimeType(in.RuntimeType),
		Version:      in.Version,
		Variant:      in.Variant,
		Architecture: in.Architecture,
		DownloadURL:  in.DownloadURL,
		InstallPath:  in.InstallPath,
		Mirror:       in.Mirror,
	})
	if err != nil {
		return dto.InstallResultDTO{}, err
	}
	return dto.InstallResultDTO{
		RuntimeID:      res.RuntimeID,
		ServiceID:      res.ServiceID,
		ServiceCreated: res.ServiceCreated,
	}, nil
}

func (a *RuntimeAPI) SetDefault(id int64) (dto.OperationResultDTO, error) {
	if err := a.svc.SetDefault(a.getCtx(), id); err != nil {
		return dto.OperationResultDTO{Success: false, Message: err.Error()}, err
	}
	return dto.OperationResultDTO{Success: true}, nil
}

func (a *RuntimeAPI) Extensions(id int64) ([]dto.PHPExtensionDTO, error) {
	list, err := a.svc.Extensions(a.getCtx(), id)
	if err != nil {
		return nil, err
	}
	res := make([]dto.PHPExtensionDTO, 0, len(list))
	for _, e := range list {
		res = append(res, dto.PHPExtensionDTO{
			Name:      e.Name,
			Enabled:   e.Enabled,
			Available: e.Available,
		})
	}
	return res, nil
}

func (a *RuntimeAPI) Delete(id int64, purge bool) (dto.OperationResultDTO, error) {
	if err := a.svc.Delete(a.getCtx(), id, purge); err != nil {
		return dto.OperationResultDTO{Success: false, Message: err.Error()}, err
	}
	return dto.OperationResultDTO{Success: true, Message: "运行环境已成功卸载"}, nil
}

// ----------------- ServiceAPI -----------------

type ServiceAPI struct {
	baseAPI
	svc *app.ServiceService
}

func NewServiceAPI(svc *app.ServiceService) *ServiceAPI {
	return &ServiceAPI{svc: svc}
}

func (a *ServiceAPI) List() ([]dto.ServiceInstanceDTO, error) {
	list, err := a.svc.List(a.getCtx())
	if err != nil {
		return nil, err
	}
	res := make([]dto.ServiceInstanceDTO, 0, len(list))
	for _, cur := range list {
		st, err := a.svc.Status(a.getCtx(), cur.ID)
		if err != nil {
			return nil, err
		}
		var rtVer, rtPath string
		if rt, err := a.svc.GetRuntime(a.getCtx(), cur.RuntimeID); err == nil && rt != nil {
			rtVer = rt.Version
			rtPath = rt.InstallPath
		}
		res = append(res, dto.ServiceInstanceDTO{
			ID:             cur.ID,
			Name:           cur.Name,
			ServiceType:    string(cur.ServiceType),
			RuntimeID:      cur.RuntimeID,
			RuntimeVersion: rtVer,
			RuntimePath:    rtPath,
			Port:           cur.Port,
			ConfigPath:     cur.ConfigPath,
			DataPath:       cur.DataPath,
			Enabled:        cur.Enabled,
			AutoStart:      cur.AutoStart,
			Status:         string(st.Status),
			PID:            st.PID,
		})
	}
	return res, nil
}

func (a *ServiceAPI) Get(id int64) (*dto.ServiceInstanceDTO, error) {
	cur, err := a.svc.Get(a.getCtx(), id)
	if err != nil {
		return nil, err
	}
	st, err := a.svc.Status(a.getCtx(), cur.ID)
	if err != nil {
		return nil, err
	}
	var rtVer, rtPath string
	if rt, err := a.svc.GetRuntime(a.getCtx(), cur.RuntimeID); err == nil && rt != nil {
		rtVer = rt.Version
		rtPath = rt.InstallPath
	}
	return &dto.ServiceInstanceDTO{
		ID:             cur.ID,
		Name:           cur.Name,
		ServiceType:    string(cur.ServiceType),
		RuntimeID:      cur.RuntimeID,
		RuntimeVersion: rtVer,
		RuntimePath:    rtPath,
		Port:           cur.Port,
		ConfigPath:     cur.ConfigPath,
		DataPath:       cur.DataPath,
		Enabled:        cur.Enabled,
		AutoStart:      cur.AutoStart,
		Status:         string(st.Status),
		PID:            st.PID,
	}, nil
}

func (a *ServiceAPI) Status(id int64) (dto.ServiceStatusInfoDTO, error) {
	st, err := a.svc.Status(a.getCtx(), id)
	if err != nil {
		return dto.ServiceStatusInfoDTO{}, err
	}
	var startedStr *string
	if st.StartedAt != nil {
		s := st.StartedAt.Format(time.RFC3339)
		startedStr = &s
	}
	return dto.ServiceStatusInfoDTO{
		ServiceID: st.ServiceID,
		Status:    string(st.Status),
		PID:       st.PID,
		StartedAt: startedStr,
		ErrorCode: st.ErrorCode,
	}, nil
}

func (a *ServiceAPI) Start(in dto.ServiceActionInput) (dto.OperationResultDTO, error) {
	if err := a.svc.Start(a.getCtx(), in.ServiceID); err != nil {
		return dto.OperationResultDTO{Success: false, Message: err.Error()}, err
	}
	return dto.OperationResultDTO{Success: true}, nil
}

func (a *ServiceAPI) Stop(in dto.ServiceActionInput) (dto.OperationResultDTO, error) {
	if err := a.svc.Stop(a.getCtx(), in.ServiceID); err != nil {
		return dto.OperationResultDTO{Success: false, Message: err.Error()}, err
	}
	return dto.OperationResultDTO{Success: true}, nil
}

func (a *ServiceAPI) Restart(in dto.ServiceActionInput) (dto.OperationResultDTO, error) {
	if err := a.svc.Restart(a.getCtx(), in.ServiceID); err != nil {
		return dto.OperationResultDTO{Success: false, Message: err.Error()}, err
	}
	return dto.OperationResultDTO{Success: true}, nil
}

func (a *ServiceAPI) SwitchRuntime(in dto.SwitchRuntimeInput) (dto.OperationResultDTO, error) {
	if err := a.svc.SwitchRuntime(a.getCtx(), in.ServiceID, in.RuntimeID); err != nil {
		return dto.OperationResultDTO{Success: false, Message: err.Error()}, err
	}
	return dto.OperationResultDTO{Success: true, Message: "服务运行版本已切换"}, nil
}

func (a *ServiceAPI) RuntimeOptions(id int64) ([]dto.RuntimeOptionDTO, error) {
	list, err := a.svc.RuntimeOptions(a.getCtx(), id)
	if err != nil {
		return nil, err
	}
	res := make([]dto.RuntimeOptionDTO, 0, len(list))
	for _, r := range list {
		res = append(res, dto.RuntimeOptionDTO{
			ID:           r.ID,
			Version:      r.Version,
			Architecture: r.Architecture,
			InstallPath:  r.InstallPath,
		})
	}
	return res, nil
}

func (a *ServiceAPI) Logs(in dto.ServiceLogInput) (dto.LogResultDTO, error) {
	lines := a.svc.Logs(a.getCtx(), in.ServiceID, in.Limit)
	if lines == nil {
		lines = make([]string, 0)
	}
	return dto.LogResultDTO{Lines: lines}, nil
}

func (a *ServiceAPI) Delete(id int64) (dto.OperationResultDTO, error) {
	if err := a.svc.Delete(a.getCtx(), id); err != nil {
		return dto.OperationResultDTO{Success: false, Message: err.Error()}, err
	}
	return dto.OperationResultDTO{Success: true, Message: "服务实例已删除"}, nil
}

// ----------------- TaskAPI -----------------

type TaskAPI struct {
	baseAPI
	svc *app.TaskService
}

func NewTaskAPI(svc *app.TaskService) *TaskAPI {
	return &TaskAPI{svc: svc}
}

func (a *TaskAPI) List() ([]dto.TaskDTO, error) {
	list, err := a.svc.List(a.getCtx())
	if err != nil {
		return nil, err
	}
	res := make([]dto.TaskDTO, 0, len(list))
	for _, t := range list {
		var finStr *string
		if t.FinishedAt != nil {
			s := t.FinishedAt.Format(time.RFC3339)
			finStr = &s
		}
		res = append(res, dto.TaskDTO{
			ID:           t.ID,
			TaskType:     t.TaskType,
			Status:       string(t.Status),
			Progress:     t.Progress,
			Message:      t.Message,
			ErrorCode:    t.ErrorCode,
			ErrorMessage: t.ErrorMessage,
			CreatedAt:    t.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    t.UpdatedAt.Format(time.RFC3339),
			FinishedAt:   finStr,
		})
	}
	return res, nil
}

func (a *TaskAPI) Get(id int64) (*dto.TaskDTO, error) {
	t, err := a.svc.Get(a.getCtx(), id)
	if err != nil {
		return nil, err
	}
	var finStr *string
	if t.FinishedAt != nil {
		s := t.FinishedAt.Format(time.RFC3339)
		finStr = &s
	}
	return &dto.TaskDTO{
		ID:           t.ID,
		TaskType:     t.TaskType,
		Status:       string(t.Status),
		Progress:     t.Progress,
		Message:      t.Message,
		ErrorCode:    t.ErrorCode,
		ErrorMessage: t.ErrorMessage,
		CreatedAt:    t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    t.UpdatedAt.Format(time.RFC3339),
		FinishedAt:   finStr,
	}, nil
}

func (a *TaskAPI) Cancel(id int64) (dto.OperationResultDTO, error) {
	if err := a.svc.Cancel(a.getCtx(), id); err != nil {
		return dto.OperationResultDTO{Success: false, Message: err.Error()}, err
	}
	return dto.OperationResultDTO{Success: true}, nil
}

// ----------------- SettingAPI -----------------

type SettingAPI struct {
	baseAPI
	svc *app.SettingService
}

func NewSettingAPI(svc *app.SettingService) *SettingAPI {
	return &SettingAPI{svc: svc}
}

func (a *SettingAPI) Get() (dto.AppSettingsDTO, error) {
	s, err := a.svc.Get(a.getCtx())
	if err != nil {
		return dto.AppSettingsDTO{}, err
	}
	return dto.AppSettingsDTO{
		Language:           s.Language,
		Theme:              s.Theme,
		RuntimeRootPath:    s.RuntimeRootPath,
		DefaultProjectPath: s.DefaultProjectPath,
		AutoStartServices:  s.AutoStartServices,
	}, nil
}

func (a *SettingAPI) Update(in dto.AppSettingsDTO) (dto.AppSettingsDTO, error) {
	s, err := a.svc.Update(a.getCtx(), settings.AppSettings{
		Language:           in.Language,
		Theme:              in.Theme,
		RuntimeRootPath:    in.RuntimeRootPath,
		DefaultProjectPath: in.DefaultProjectPath,
		AutoStartServices:  in.AutoStartServices,
	})
	if err != nil {
		return dto.AppSettingsDTO{}, err
	}
	return dto.AppSettingsDTO{
		Language:           s.Language,
		Theme:              s.Theme,
		RuntimeRootPath:    s.RuntimeRootPath,
		DefaultProjectPath: s.DefaultProjectPath,
		AutoStartServices:  s.AutoStartServices,
	}, nil
}

func (a *SettingAPI) Reset() (dto.AppSettingsDTO, error) {
	s, err := a.svc.Reset(a.getCtx())
	if err != nil {
		return dto.AppSettingsDTO{}, err
	}
	return dto.AppSettingsDTO{
		Language:           s.Language,
		Theme:              s.Theme,
		RuntimeRootPath:    s.RuntimeRootPath,
		DefaultProjectPath: s.DefaultProjectPath,
		AutoStartServices:  s.AutoStartServices,
	}, nil
}

func (a *SettingAPI) SelectDir(title string) (string, error) {
	if title == "" {
		title = "选择目录"
	}
	return wailsRuntime.OpenDirectoryDialog(a.getCtx(), wailsRuntime.OpenDialogOptions{
		Title: title,
	})
}

func (a *SettingAPI) SelectFile(title string) (string, error) {
	if title == "" {
		title = "选择文件"
	}
	return wailsRuntime.OpenFileDialog(a.getCtx(), wailsRuntime.OpenDialogOptions{
		Title: title,
	})
}
