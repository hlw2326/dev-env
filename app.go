package main

import (
	"context"

	"devenv/internal/app"
	transport "devenv/internal/transport/wails"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx       context.Context
	container *app.Container
	bootError error
	Site      *transport.SiteAPI
	Software  *transport.SoftwareAPI
	System    *transport.SystemAPI
}

func NewApp() *App {
	container, err := app.Bootstrap()
	if err != nil {
		return &App{
			bootError: err,
			Site:      transport.NewSiteAPI(nil),
			Software:  transport.NewSoftwareAPI(nil),
			System:    transport.NewSystemAPI(nil),
		}
	}

	return &App{
		container: container,
		Site:      transport.NewSiteAPI(container.Site),
		Software:  transport.NewSoftwareAPI(container.Software),
		System:    transport.NewSystemAPI(container.Setting),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if a.bootError != nil {
		_, _ = runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "DevEnv 启动失败",
			Message: a.bootError.Error() + "\n\n如需重置开发数据库，请设置环境变量 DEVENV_DB_RESET=1 后重新启动。",
		})
		runtime.Quit(ctx)
		return
	}
	a.Site.SetContext(ctx)
	a.Software.SetContext(ctx)
	a.System.SetContext(ctx)
	a.container.Bus.SetContext(ctx)
	a.container.Start(ctx)
}

func (a *App) shutdown(ctx context.Context) {
	if a.container != nil {
		a.container.Shutdown()
	}
}
