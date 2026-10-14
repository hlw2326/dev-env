package main

import (
	"context"

	"devenv/internal/app"
	transport "devenv/internal/transport/wails"
)

// App struct manages desktop lifecycle
type App struct {
	ctx       context.Context
	container *app.Container
	api       *transport.AppAPI
}

// NewApp creates a new App application struct
func NewApp() *App {
	container, err := app.Bootstrap()
	if err != nil {
		panic(err)
	}

	api := transport.NewAppAPI(
		container.ProjectSvc,
		container.PHPSvc,
		container.ServiceSvc,
		container.HostSvc,
		container.TaskSvc,
		container.SettingsSvc,
	)

	return &App{
		container: container,
		api:       api,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.api.SetContext(ctx)
	a.container.Publisher.SetContext(ctx)
}

// shutdown is called when the app terminates
func (a *App) shutdown(ctx context.Context) {
	if a.container != nil && a.container.DB != nil {
		_ = a.container.DB.Close()
	}
}
