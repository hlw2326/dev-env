package app

import (
	"context"
	"log"
	"os"

	"devenv/internal/domain/software"
	"devenv/internal/infra/catalog"
	"devenv/internal/infra/events"
	"devenv/internal/infra/gateway"
	"devenv/internal/infra/hosts"
	"devenv/internal/infra/installer"
	"devenv/internal/infra/network"
	"devenv/internal/infra/process"
	"devenv/internal/infra/sqlite"
	"devenv/internal/ports"
	"devenv/internal/shared/paths"
)

type Container struct {
	Base     string
	DB       *sqlite.DB
	Bus      *events.Publisher
	Nginx    ports.Gateway
	Runner   *process.Manager
	Site     *SiteService
	Software *SoftwareService
	Setting  *SettingService
}

func Bootstrap() (*Container, error) {
	base := paths.Base()
	dbPath := paths.Database(base)

	resetDrift := os.Getenv("DEVENV_DB_RESET") == "1"
	db, err := sqlite.Open(dbPath, sqlite.Options{
		ResetOnDrift: resetDrift,
	})
	if err != nil {
		return nil, err
	}

	bus := events.NewPublisher()
	probe := network.NewChecker()
	runner := process.NewManager()
	catalogClient := catalog.NewClient()
	installerManager := installer.NewManager()
	hostsWriter := hosts.NewWriter()
	softwareRepo := sqlite.NewSoftwareRepository(db.DB)
	siteRepo := sqlite.NewSiteRepository(db.DB)
	settingRepo := sqlite.NewSettingRepository(db.DB)
	nginx := gateway.NewNginx(base, softwareRepo, runner, probe)

	siteService := NewSiteService(base, siteRepo, softwareRepo, settingRepo, nginx, runner, hostsWriter, probe, bus)
	softwareService := NewSoftwareService(base, softwareRepo, siteRepo, catalogClient, installerManager, runner, probe, settingRepo, nginx, bus, func(ctx context.Context, item software.Software) {
		siteService.BindInstalled(ctx, item)
	})
	settingService := NewSettingService(base, settingRepo, siteRepo, nginx, hostsWriter)

	return &Container{
		Base:     base,
		DB:       db,
		Bus:      bus,
		Nginx:    nginx,
		Runner:   runner,
		Site:     siteService,
		Software: softwareService,
		Setting:  settingService,
	}, nil
}

func (c *Container) Start(ctx context.Context) {
	c.Software.StartAuto(ctx)
	c.Setting.StartNginx(ctx)
	c.Site.StartAuto(ctx)
}

func (c *Container) Shutdown() {
	if err := c.Site.StopAll(); err != nil {
		log.Printf("停止站点进程失败: %v", err)
	}
	if err := c.Software.StopAll(); err != nil {
		log.Printf("停止软件进程失败: %v", err)
	}
	if err := c.Nginx.Stop(); err != nil {
		log.Printf("停止 Nginx 失败: %v", err)
	}
	if err := c.DB.Close(); err != nil {
		log.Printf("关闭数据库失败: %v", err)
	}
}
