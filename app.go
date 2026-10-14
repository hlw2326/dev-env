package main

import (
	"context"

	"devenv/internal/app"
	transport "devenv/internal/transport/wails"
)

type App struct {
	ctx       context.Context
	container *app.Container
	Site      *transport.SiteAPI
	Git       *transport.GitAPI
	Runtime   *transport.RuntimeAPI
	Service   *transport.ServiceAPI
	Task      *transport.TaskAPI
	Setting   *transport.SettingAPI
}

func NewApp() *App {
	container, err := app.Bootstrap()
	if err != nil {
		panic(err)
	}

	return &App{
		container: container,
		Site:      transport.NewSiteAPI(container.Site),
		Git:       transport.NewGitAPI(container.Git),
		Runtime:   transport.NewRuntimeAPI(container.Runtime),
		Service:   transport.NewServiceAPI(container.Service),
		Task:      transport.NewTaskAPI(container.Task),
		Setting:   transport.NewSettingAPI(container.Setting),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.Site.SetContext(ctx)
	a.Git.SetContext(ctx)
	a.Runtime.SetContext(ctx)
	a.Service.SetContext(ctx)
	a.Task.SetContext(ctx)
	a.Setting.SetContext(ctx)
	a.container.Bus.SetContext(ctx)
}

func (a *App) shutdown(ctx context.Context) {
	if a.container != nil && a.container.DB != nil {
		if err := a.container.DB.Close(); err != nil {
			println("Failed to close database:", err.Error())
		}
	}
}
