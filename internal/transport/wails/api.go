package wails

import (
	"context"
	"time"

	"devenv/internal/app"
	"devenv/internal/domain/host"
	"devenv/internal/domain/project"
	"devenv/internal/domain/settings"
	"devenv/internal/transport/wails/dto"
)

type AppAPI struct {
	ctx          context.Context
	projectSvc   *app.ProjectService
	phpSvc       *app.PHPService
	serviceSvc   *app.ServiceAppService
	hostSvc      *app.HostService
	taskSvc      *app.TaskService
	settingsSvc  *app.SettingsService
}

func NewAppAPI(
	projectSvc *app.ProjectService,
	phpSvc *app.PHPService,
	serviceSvc *app.ServiceAppService,
	hostSvc *app.HostService,
	taskSvc *app.TaskService,
	settingsSvc *app.SettingsService,
) *AppAPI {
	return &AppAPI{
		projectSvc:  projectSvc,
		phpSvc:      phpSvc,
		serviceSvc:  serviceSvc,
		hostSvc:     hostSvc,
		taskSvc:     taskSvc,
		settingsSvc: settingsSvc,
	}
}

func (a *AppAPI) SetContext(ctx context.Context) {
	a.ctx = ctx
}

func (a *AppAPI) getCtx() context.Context {
	if a.ctx != nil {
		return a.ctx
	}
	return context.Background()
}

// ----------------- Projects API -----------------

func (a *AppAPI) ListProjects() ([]dto.ProjectDTO, error) {
	list, err := a.projectSvc.ListProjects(a.getCtx())
	if err != nil {
		return nil, err
	}
	result := make([]dto.ProjectDTO, len(list))
	for i, p := range list {
		result[i] = dto.ProjectDTO{
			ID:           p.ID,
			Name:         p.Name,
			Path:         p.Path,
			PHPVersionID: p.PHPVersionID,
			Status:       string(p.Status),
			CreatedAt:    p.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    p.UpdatedAt.Format(time.RFC3339),
		}
	}
	return result, nil
}

func (a *AppAPI) GetProject(id int64) (*dto.ProjectDTO, error) {
	p, err := a.projectSvc.GetProject(a.getCtx(), id)
	if err != nil {
		return nil, err
	}
	return &dto.ProjectDTO{
		ID:           p.ID,
		Name:         p.Name,
		Path:         p.Path,
		PHPVersionID: p.PHPVersionID,
		Status:       string(p.Status),
		CreatedAt:    p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    p.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (a *AppAPI) CreateProject(in dto.CreateProjectInput) (*dto.ProjectDTO, error) {
	p, err := a.projectSvc.CreateProject(a.getCtx(), project.CreateInput{
		Name:         in.Name,
		Path:         in.Path,
		PHPVersionID: in.PHPVersionID,
	})
	if err != nil {
		return nil, err
	}
	return &dto.ProjectDTO{
		ID:           p.ID,
		Name:         p.Name,
		Path:         p.Path,
		PHPVersionID: p.PHPVersionID,
		Status:       string(p.Status),
		CreatedAt:    p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    p.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (a *AppAPI) ImportProject(path string) (*dto.ProjectDTO, error) {
	p, err := a.projectSvc.ImportProject(a.getCtx(), path)
	if err != nil {
		return nil, err
	}
	return &dto.ProjectDTO{
		ID:           p.ID,
		Name:         p.Name,
		Path:         p.Path,
		PHPVersionID: p.PHPVersionID,
		Status:       string(p.Status),
		CreatedAt:    p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    p.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (a *AppAPI) UpdateProject(in dto.UpdateProjectInput) (*dto.ProjectDTO, error) {
	p, err := a.projectSvc.UpdateProject(a.getCtx(), project.UpdateInput{
		ID:           in.ID,
		Name:         in.Name,
		PHPVersionID: in.PHPVersionID,
	})
	if err != nil {
		return nil, err
	}
	return &dto.ProjectDTO{
		ID:           p.ID,
		Name:         p.Name,
		Path:         p.Path,
		PHPVersionID: p.PHPVersionID,
		Status:       string(p.Status),
		CreatedAt:    p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    p.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (a *AppAPI) DeleteProject(id int64) (dto.OperationResultDTO, error) {
	err := a.projectSvc.DeleteProject(a.getCtx(), id)
	if err != nil {
		return dto.OperationResultDTO{Success: false, Message: err.Error()}, err
	}
	return dto.OperationResultDTO{Success: true}, nil
}

func (a *AppAPI) ValidateProjectPath(path string) dto.ProjectPathValidationDTO {
	val := a.projectSvc.ValidateProjectPath(a.getCtx(), path)
	return dto.ProjectPathValidationDTO{
		Valid:          val.Valid,
		NormalizedPath: val.NormalizedPath,
		Exists:         val.Exists,
		Writable:       val.Writable,
		Errors:         val.Errors,
	}
}

// ----------------- PHP API -----------------

func (a *AppAPI) ListVersions() ([]dto.PHPVersionDTO, error) {
	list, err := a.phpSvc.ListVersions(a.getCtx())
	if err != nil {
		return nil, err
	}
	result := make([]dto.PHPVersionDTO, len(list))
	for i, v := range list {
		result[i] = dto.PHPVersionDTO{
			ID:           v.ID,
			Version:      v.Version,
			Architecture: v.Architecture,
			InstallPath:  v.InstallPath,
			Status:       string(v.Status),
			IsDefault:    v.IsDefault,
		}
	}
	return result, nil
}

func (a *AppAPI) RegisterVersion(in dto.RegisterPHPVersionInput) (*dto.PHPVersionDTO, error) {
	v, err := a.phpSvc.RegisterVersion(a.getCtx(), in.Version, in.Architecture, in.InstallPath)
	if err != nil {
		return nil, err
	}
	return &dto.PHPVersionDTO{
		ID:           v.ID,
		Version:      v.Version,
		Architecture: v.Architecture,
		InstallPath:  v.InstallPath,
		Status:       string(v.Status),
		IsDefault:    false,
	}, nil
}

func (a *AppAPI) SetDefaultVersion(versionID int64) (dto.OperationResultDTO, error) {
	err := a.phpSvc.SetDefaultVersion(a.getCtx(), versionID)
	if err != nil {
		return dto.OperationResultDTO{Success: false, Message: err.Error()}, err
	}
	return dto.OperationResultDTO{Success: true}, nil
}

func (a *AppAPI) GetExtensions(versionID int64) ([]dto.PHPExtensionDTO, error) {
	list, err := a.phpSvc.GetExtensions(a.getCtx(), versionID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.PHPExtensionDTO, len(list))
	for i, e := range list {
		result[i] = dto.PHPExtensionDTO{
			Name:      e.Name,
			Enabled:   e.Enabled,
			Available: e.Available,
		}
	}
	return result, nil
}

// ----------------- Services API -----------------

func (a *AppAPI) ListServices() ([]dto.ServiceInstanceDTO, error) {
	list, err := a.serviceSvc.ListServices(a.getCtx())
	if err != nil {
		return nil, err
	}
	result := make([]dto.ServiceInstanceDTO, len(list))
	for i, s := range list {
		status := a.serviceSvc.GetServiceLogs(a.getCtx(), s.ID, 0)
		_ = status
		stInfo, _ := a.serviceSvc.GetServiceStatus(a.getCtx(), s.ID)
		result[i] = dto.ServiceInstanceDTO{
			ID:               s.ID,
			Name:             s.Name,
			ServiceType:      string(s.ServiceType),
			RuntimeVersionID: s.RuntimeVersionID,
			Port:             s.Port,
			ConfigPath:       s.ConfigPath,
			DataPath:         s.DataPath,
			Enabled:          s.Enabled,
			AutoStart:        s.AutoStart,
			Status:           string(stInfo.Status),
			PID:              stInfo.PID,
		}
	}
	return result, nil
}

func (a *AppAPI) GetService(serviceID int64) (*dto.ServiceInstanceDTO, error) {
	s, err := a.serviceSvc.GetService(a.getCtx(), serviceID)
	if err != nil {
		return nil, err
	}
	stInfo, _ := a.serviceSvc.GetServiceStatus(a.getCtx(), s.ID)
	return &dto.ServiceInstanceDTO{
		ID:               s.ID,
		Name:             s.Name,
		ServiceType:      string(s.ServiceType),
		RuntimeVersionID: s.RuntimeVersionID,
		Port:             s.Port,
		ConfigPath:       s.ConfigPath,
		DataPath:         s.DataPath,
		Enabled:          s.Enabled,
		AutoStart:        s.AutoStart,
		Status:           string(stInfo.Status),
		PID:              stInfo.PID,
	}, nil
}

func (a *AppAPI) GetServiceStatus(serviceID int64) (dto.ServiceStatusInfoDTO, error) {
	st, err := a.serviceSvc.GetServiceStatus(a.getCtx(), serviceID)
	if err != nil {
		return dto.ServiceStatusInfoDTO{}, err
	}
	var startedStr *string
	if st.StartedAt != nil {
		s := st.StartedAt.Format(time.RFC3339)
		startedStr = &s
	}
	return dto.ServiceStatusInfoDTO{
		ServiceID:     st.ServiceID,
		Status:        string(st.Status),
		PID:           st.PID,
		StartedAt:     startedStr,
		LastErrorCode: st.LastErrorCode,
	}, nil
}

func (a *AppAPI) StartService(in dto.ServiceActionInput) (dto.OperationResultDTO, error) {
	err := a.serviceSvc.StartService(a.getCtx(), in.ServiceID)
	if err != nil {
		return dto.OperationResultDTO{Success: false, Message: err.Error()}, err
	}
	return dto.OperationResultDTO{Success: true}, nil
}

func (a *AppAPI) StopService(in dto.ServiceActionInput) (dto.OperationResultDTO, error) {
	err := a.serviceSvc.StopService(a.getCtx(), in.ServiceID)
	if err != nil {
		return dto.OperationResultDTO{Success: false, Message: err.Error()}, err
	}
	return dto.OperationResultDTO{Success: true}, nil
}

func (a *AppAPI) RestartService(in dto.ServiceActionInput) (dto.OperationResultDTO, error) {
	err := a.serviceSvc.RestartService(a.getCtx(), in.ServiceID)
	if err != nil {
		return dto.OperationResultDTO{Success: false, Message: err.Error()}, err
	}
	return dto.OperationResultDTO{Success: true}, nil
}

func (a *AppAPI) GetServiceLogs(in dto.ServiceLogInput) (dto.ServiceLogResultDTO, error) {
	lines := a.serviceSvc.GetServiceLogs(a.getCtx(), in.ServiceID, in.Limit)
	return dto.ServiceLogResultDTO{Lines: lines}, nil
}

// ----------------- Virtual Hosts API -----------------

func (a *AppAPI) ListVirtualHosts() ([]dto.VirtualHostDTO, error) {
	list, err := a.hostSvc.ListVirtualHosts(a.getCtx())
	if err != nil {
		return nil, err
	}
	result := make([]dto.VirtualHostDTO, len(list))
	for i, h := range list {
		result[i] = dto.VirtualHostDTO{
			ID:            h.ID,
			ProjectID:     h.ProjectID,
			Domain:        h.Domain,
			DocumentRoot:  h.DocumentRoot,
			WebServerType: h.WebServerType,
			Port:          h.Port,
			Enabled:       h.Enabled,
			ConfigPath:    h.ConfigPath,
		}
	}
	return result, nil
}

func (a *AppAPI) GetVirtualHost(id int64) (*dto.VirtualHostDTO, error) {
	h, err := a.hostSvc.GetVirtualHost(a.getCtx(), id)
	if err != nil {
		return nil, err
	}
	return &dto.VirtualHostDTO{
		ID:            h.ID,
		ProjectID:     h.ProjectID,
		Domain:        h.Domain,
		DocumentRoot:  h.DocumentRoot,
		WebServerType: h.WebServerType,
		Port:          h.Port,
		Enabled:       h.Enabled,
		ConfigPath:    h.ConfigPath,
	}, nil
}

func (a *AppAPI) CreateVirtualHost(in dto.CreateVirtualHostInput) (*dto.VirtualHostDTO, error) {
	h, err := a.hostSvc.CreateVirtualHost(a.getCtx(), host.CreateInput{
		ProjectID:     in.ProjectID,
		Domain:        in.Domain,
		DocumentRoot:  in.DocumentRoot,
		WebServerType: in.WebServerType,
		Port:          in.Port,
	})
	if err != nil {
		return nil, err
	}
	return &dto.VirtualHostDTO{
		ID:            h.ID,
		ProjectID:     h.ProjectID,
		Domain:        h.Domain,
		DocumentRoot:  h.DocumentRoot,
		WebServerType: h.WebServerType,
		Port:          h.Port,
		Enabled:       h.Enabled,
		ConfigPath:    h.ConfigPath,
	}, nil
}

func (a *AppAPI) UpdateVirtualHost(in dto.VirtualHostDTO) (*dto.VirtualHostDTO, error) {
	h, err := a.hostSvc.UpdateVirtualHost(a.getCtx(), host.VirtualHost{
		ID:            in.ID,
		ProjectID:     in.ProjectID,
		Domain:        in.Domain,
		DocumentRoot:  in.DocumentRoot,
		WebServerType: in.WebServerType,
		Port:          in.Port,
		Enabled:       in.Enabled,
		ConfigPath:    in.ConfigPath,
	})
	if err != nil {
		return nil, err
	}
	return &dto.VirtualHostDTO{
		ID:            h.ID,
		ProjectID:     h.ProjectID,
		Domain:        h.Domain,
		DocumentRoot:  h.DocumentRoot,
		WebServerType: h.WebServerType,
		Port:          h.Port,
		Enabled:       h.Enabled,
		ConfigPath:    h.ConfigPath,
	}, nil
}

func (a *AppAPI) DeleteVirtualHost(id int64) (dto.OperationResultDTO, error) {
	err := a.hostSvc.DeleteVirtualHost(a.getCtx(), id)
	if err != nil {
		return dto.OperationResultDTO{Success: false, Message: err.Error()}, err
	}
	return dto.OperationResultDTO{Success: true}, nil
}

func (a *AppAPI) ValidateVirtualHost(in dto.CreateVirtualHostInput) dto.ValidationResultDTO {
	res := a.hostSvc.ValidateVirtualHost(a.getCtx(), host.CreateInput{
		ProjectID:     in.ProjectID,
		Domain:        in.Domain,
		DocumentRoot:  in.DocumentRoot,
		WebServerType: in.WebServerType,
		Port:          in.Port,
	})
	return dto.ValidationResultDTO{
		Valid:    res.Valid,
		Errors:   res.Errors,
		Warnings: res.Warnings,
	}
}

func (a *AppAPI) EnableVirtualHost(id int64) (dto.OperationResultDTO, error) {
	err := a.hostSvc.EnableVirtualHost(a.getCtx(), id)
	if err != nil {
		return dto.OperationResultDTO{Success: false, Message: err.Error()}, err
	}
	return dto.OperationResultDTO{Success: true}, nil
}

func (a *AppAPI) DisableVirtualHost(id int64) (dto.OperationResultDTO, error) {
	err := a.hostSvc.DisableVirtualHost(a.getCtx(), id)
	if err != nil {
		return dto.OperationResultDTO{Success: false, Message: err.Error()}, err
	}
	return dto.OperationResultDTO{Success: true}, nil
}

// ----------------- Tasks API -----------------

func (a *AppAPI) ListTasks() ([]dto.TaskDTO, error) {
	list, err := a.taskSvc.ListTasks(a.getCtx())
	if err != nil {
		return nil, err
	}
	result := make([]dto.TaskDTO, len(list))
	for i, t := range list {
		var finStr *string
		if t.FinishedAt != nil {
			s := t.FinishedAt.Format(time.RFC3339)
			finStr = &s
		}
		result[i] = dto.TaskDTO{
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
		}
	}
	return result, nil
}

func (a *AppAPI) GetTask(taskId int64) (*dto.TaskDTO, error) {
	t, err := a.taskSvc.GetTask(a.getCtx(), taskId)
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

func (a *AppAPI) CancelTask(taskId int64) (dto.OperationResultDTO, error) {
	err := a.taskSvc.CancelTask(a.getCtx(), taskId)
	if err != nil {
		return dto.OperationResultDTO{Success: false, Message: err.Error()}, err
	}
	return dto.OperationResultDTO{Success: true}, nil
}

// ----------------- Settings API -----------------

func (a *AppAPI) GetSettings() (dto.AppSettingsDTO, error) {
	s, err := a.settingsSvc.GetSettings(a.getCtx())
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

func (a *AppAPI) UpdateSettings(in dto.AppSettingsDTO) (dto.AppSettingsDTO, error) {
	s, err := a.settingsSvc.UpdateSettings(a.getCtx(), settings.AppSettings{
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

func (a *AppAPI) ResetSettings() (dto.AppSettingsDTO, error) {
	s, err := a.settingsSvc.ResetSettings(a.getCtx())
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
