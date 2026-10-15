package app

import (
	"context"
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
	"devenv/internal/shared/paths"
)

type Container struct {
	Base     string
	DB       *sqlite.DB
	Bus      *events.Publisher
	Gateway  *gateway.Server
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
	gw := gateway.New()
	catalogClient := catalog.NewClient()
	installerManager := installer.NewManager()
	hostsWriter := hosts.NewWriter()

	softwareRepo := sqlite.NewSoftwareRepository(db.DB)
	siteRepo := sqlite.NewSiteRepository(db.DB)
	settingRepo := sqlite.NewSettingRepository(db.DB)

	siteService := NewSiteService(base, siteRepo, softwareRepo, settingRepo, gw, runner, hostsWriter, probe, bus)
	softwareService := NewSoftwareService(base, softwareRepo, siteRepo, catalogClient, installerManager, runner, probe, bus, func(ctx context.Context, item software.Software) {
		siteService.BindInstalled(ctx, item)
	})
	settingService := NewSettingService(base, settingRepo, siteRepo, gw, hostsWriter)

	return &Container{
		Base:     base,
		DB:       db,
		Bus:      bus,
		Gateway:  gw,
		Runner:   runner,
		Site:     siteService,
		Software: softwareService,
		Setting:  settingService,
	}, nil
}

func (c *Container) Start(ctx context.Context) {
	c.Setting.StartGateway(ctx)
	c.Software.StartAuto(ctx)
	c.Site.StartAuto(ctx)
}

func (c *Container) Shutdown() {
	c.Site.StopAll()
	c.Software.StopAll()
	if err := c.Gateway.Stop(); err != nil {
		return
	}
	if err := c.DB.Close(); err != nil {
		return
	}
}
