package main

import (
	"context"
	"sync/atomic"

	"devenv/internal/app"
	trayinfra "devenv/internal/infra/tray"
	transport "devenv/internal/transport/wails"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx       context.Context
	container *app.Container
	bootError error
	Project   *transport.ProjectAPI
	Software  *transport.SoftwareAPI
	System    *transport.SystemAPI
	tray      trayinfra.Manager
	exiting   atomic.Bool
}

func NewApp() *App {
	trayManager := trayinfra.NewManager()
	container, err := app.Bootstrap()
	if err != nil {
		return &App{
			bootError: err,
			tray:      trayManager,
			Project:   transport.NewProjectAPI(nil, nil),
			Software:  transport.NewSoftwareAPI(nil),
			System:    transport.NewSystemAPI(nil, nil),
		}
	}

	return &App{
		container: container,
		tray:      trayManager,
		Project:   transport.NewProjectAPI(container.Project, container.Deployment),
		Software:  transport.NewSoftwareAPI(container.Software),
		System:    transport.NewSystemAPI(container.Setting, container.Monitor),
	}
}

func (a *App) startTray() {
	if a.bootError != nil {
		return
	}
	a.tray.Start(trayinfra.Callbacks{
		Show: a.showWindow,
		Hide: a.hideWindow,
		Quit: a.quit,
	})
}

func (a *App) showWindow() {
	if a.ctx != nil {
		if runtime.WindowIsMinimised(a.ctx) {
			runtime.WindowUnminimise(a.ctx)
		}
		runtime.WindowShow(a.ctx)
	}
}

func (a *App) hideWindow() {
	if a.ctx != nil {
		runtime.WindowHide(a.ctx)
	}
}

func (a *App) quit() {
	if a.ctx != nil {
		a.exiting.Store(true)
		runtime.Quit(a.ctx)
	}
}

func (a *App) beforeClose(ctx context.Context) bool {
	if a.bootError != nil || a.exiting.Load() {
		return false
	}
	runtime.WindowHide(ctx)
	return true
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if a.bootError != nil {
		if _, err := runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "DevEnv 启动失败",
			Message: a.bootError.Error() + "\n\n如需重置开发数据库，请设置环境变量 DEVENV_DB_RESET=1 后重新启动。",
		}); err != nil {
			runtime.Quit(ctx)
			return
		}
		runtime.Quit(ctx)
		return
	}
	a.Project.SetContext(ctx)
	a.Software.SetContext(ctx)
	a.System.SetContext(ctx)
	a.container.Bus.SetContext(ctx)
	a.container.Start(ctx)
}

func (a *App) shutdown(ctx context.Context) {
	a.tray.Stop()
	if a.bootError == nil {
		a.container.Shutdown()
	}
}
