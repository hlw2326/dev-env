package wails

import (
	"context"

	"devenv/internal/app"
	"devenv/internal/domain/setting"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type SystemAPI struct {
	ctx     context.Context
	service *app.SettingService
}

func NewSystemAPI(service *app.SettingService) *SystemAPI {
	return &SystemAPI{
		ctx:     context.Background(),
		service: service,
	}
}

func (a *SystemAPI) SetContext(ctx context.Context) {
	a.ctx = ctx
}

func (a *SystemAPI) Settings() (SettingsDTO, error) {
	view, err := a.service.Settings(a.ctx)
	if err != nil {
		return SettingsDTO{}, err
	}
	return SettingsDTO{
		Theme:     view.Theme,
		Root:      view.Root,
		Port:      view.Port,
		Hosts:     view.Hosts,
		Listening: view.Listening,
		Error:     view.Error,
	}, nil
}

func (a *SystemAPI) SaveSettings(in SettingsInput) Result {
	err := a.service.SaveSettings(a.ctx, setting.Settings{
		Theme: in.Theme,
		Root:  in.Root,
		Port:  in.Port,
		Hosts: in.Hosts,
	})
	if err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "设置已保存"}
}

func (a *SystemAPI) PickFolder(title string) (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: title,
	})
}

func (a *SystemAPI) OpenURL(url string) Result {
	runtime.BrowserOpenURL(a.ctx, url)
	return Result{OK: true, Message: ""}
}

func (a *SystemAPI) OpenFolder(path string) Result {
	if err := a.service.OpenFolder(path); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: ""}
}

func (a *SystemAPI) DataFolder() string {
	return a.service.DataFolder()
}
