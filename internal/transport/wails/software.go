package wails

import (
	"context"

	"devenv/internal/app"
	"devenv/internal/domain/software"
)

type SoftwareAPI struct {
	ctx     context.Context
	service *app.SoftwareService
}

func NewSoftwareAPI(service *app.SoftwareService) *SoftwareAPI {
	return &SoftwareAPI{
		ctx:     context.Background(),
		service: service,
	}
}

func (a *SoftwareAPI) SetContext(ctx context.Context) {
	a.ctx = ctx
}

func (a *SoftwareAPI) Market() ([]ItemDTO, error) {
	items, err := a.service.Market(a.ctx)
	if err != nil {
		return nil, err
	}
	res := make([]ItemDTO, 0, len(items))
	for _, it := range items {
		res = append(res, ItemDTO{
			Kind:        it.Kind,
			Name:        it.Name,
			Description: it.Description,
			Category:    it.Category,
			Count:       it.Count,
			Recommended: it.Recommended,
		})
	}
	return res, nil
}

func (a *SoftwareAPI) List() ([]SoftwareDTO, error) {
	list, err := a.service.List(a.ctx)
	if err != nil {
		return nil, err
	}
	res := make([]SoftwareDTO, 0, len(list))
	for _, item := range list {
		res = append(res, SoftwareDTO{
			ID:        item.ID,
			Kind:      item.Kind,
			Name:      software.Name(item.Kind),
			Version:   item.Version,
			Path:      item.Path,
			State:     item.State,
			Progress:  item.Progress,
			Error:     item.Error,
			Main:      item.Main,
			Port:      item.Port,
			Autostart: item.Autostart,
			Running:   item.Running,
			PID:       item.PID,
			Daemon:    software.IsDaemon(item.Kind),
		})
	}
	return res, nil
}

func (a *SoftwareAPI) Versions(kind string) ([]VersionDTO, error) {
	list, err := a.service.Versions(a.ctx, kind)
	if err != nil {
		return nil, err
	}
	res := make([]VersionDTO, 0, len(list))
	for _, v := range list {
		res = append(res, VersionDTO{
			Version:     v.Version,
			Recommended: v.Recommended,
			Installed:   v.Installed,
			Note:        v.Note,
		})
	}
	return res, nil
}

func (a *SoftwareAPI) Config(id int64) (ConfigDTO, error) {
	cfg, err := a.service.Config(a.ctx, id)
	if err != nil {
		return ConfigDTO{}, err
	}
	items := make([]ConfigItemDTO, 0, len(cfg.Items))
	for _, it := range cfg.Items {
		items = append(items, ConfigItemDTO{
			Key:         it.Key,
			Label:       it.Label,
			Value:       it.Value,
			Description: it.Description,
		})
	}
	return ConfigDTO{
		Items: items,
		File:  cfg.File,
	}, nil
}

func (a *SoftwareAPI) SaveConfig(in SaveConfigInput) SaveConfigResult {
	restart, err := a.service.SaveConfig(a.ctx, in.ID, in.Items)
	if err != nil {
		return SaveConfigResult{OK: false, Message: err.Error(), Restart: false}
	}
	msg := "配置已保存并立即生效"
	if restart {
		msg = "配置已保存，重启服务后生效"
	}
	return SaveConfigResult{OK: true, Message: msg, Restart: restart}
}

func (a *SoftwareAPI) ReadConfig(id int64) (string, error) {
	return a.service.ReadConfig(a.ctx, id)
}

func (a *SoftwareAPI) WriteConfig(id int64, text string) Result {
	if err := a.service.WriteConfig(a.ctx, id, text); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "配置文件已保存"}
}

func (a *SoftwareAPI) Extensions(id int64) ([]ExtensionDTO, error) {
	exts, err := a.service.Extensions(a.ctx, id)
	if err != nil {
		return nil, err
	}
	res := make([]ExtensionDTO, 0, len(exts))
	for _, e := range exts {
		res = append(res, ExtensionDTO{
			Name:    e.Name,
			Enabled: e.Enabled,
		})
	}
	return res, nil
}

func (a *SoftwareAPI) SetExtension(id int64, name string, on bool) Result {
	if err := a.service.SetExtension(a.ctx, id, name, on); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "扩展状态已更新"}
}

func (a *SoftwareAPI) OpenFolder(id int64) Result {
	if err := a.service.OpenFolder(a.ctx, id); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "已打开目录"}
}

func (a *SoftwareAPI) Install(in InstallInput) Result {
	if err := a.service.Install(a.ctx, in.Kind, in.Version); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "已开始安装"}
}

func (a *SoftwareAPI) Cancel(kind, version string) Result {
	if err := a.service.Cancel(a.ctx, kind, version); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "已取消安装"}
}

func (a *SoftwareAPI) Remove(id int64) Result {
	if err := a.service.Remove(a.ctx, id); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "软件已卸载"}
}

func (a *SoftwareAPI) SetMain(id int64) Result {
	if err := a.service.SetMain(a.ctx, id); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "已设为默认版本"}
}

func (a *SoftwareAPI) Start(id int64) Result {
	if err := a.service.Start(a.ctx, id); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "服务已启动"}
}

func (a *SoftwareAPI) Stop(id int64) Result {
	if err := a.service.Stop(a.ctx, id); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "服务已停止"}
}

func (a *SoftwareAPI) SetAutostart(id int64, on bool) Result {
	if err := a.service.SetAutostart(a.ctx, id, on); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "自启动设置已更新"}
}

func (a *SoftwareAPI) Logs(id int64, limit int) ([]string, error) {
	return a.service.Logs(a.ctx, id, limit)
}
