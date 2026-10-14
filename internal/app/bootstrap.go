package app

import (
	"context"
	"os"
	"path/filepath"

	"devenv/internal/domain/software"
	"devenv/internal/infra/catalog"
	"devenv/internal/infra/events"
	"devenv/internal/infra/hosts"
	"devenv/internal/infra/installer"
	"devenv/internal/infra/network"
	"devenv/internal/infra/nginx"
	"devenv/internal/infra/process"
	"devenv/internal/infra/sqlite"
)

type Container struct {
	BaseDir  string
	DB       *sqlite.DB
	Bus      *events.Publisher
	Gateway  *nginx.Gateway
	Runner   *process.Manager
	Site     *SiteService
	Software *SoftwareService
	Setting  *SettingService
}

func Bootstrap() (*Container, error) {
	localAppData := os.Getenv("LOCALAPPDATA")
	baseDir := filepath.Join(localAppData, "DevEnv")
	dbPath := filepath.Join(baseDir, "data", "devenv.db")

	resetOnDrift := os.Getenv("DEVENV_DB_RESET") == "1"
	db, err := sqlite.Open(dbPath, sqlite.Options{
		ResetOnDrift: resetOnDrift,
	})
	if err != nil {
		return nil, err
	}

	bus := events.NewPublisher()
	probe := network.NewChecker()
	runner := process.NewManager()
	gw := nginx.NewGateway(baseDir, runner, probe)
	cat := catalog.NewClient()
	inst := installer.NewManager()
	hostsWriter := hosts.NewWriter()

	softRepo := sqlite.NewSoftwareRepository(db.DB)
	siteRepo := sqlite.NewSiteRepository(db.DB)
	settingRepo := sqlite.NewSettingRepository(db.DB)

	softSvc := NewSoftwareService(baseDir, softRepo, siteRepo, cat, inst, runner, probe, bus)
	siteSvc := NewSiteService(baseDir, siteRepo, softRepo, settingRepo, gw, runner, hostsWriter, probe, bus)
	settingSvc := NewSettingService(baseDir, settingRepo, siteRepo, gw, hostsWriter)

	softSvc.SetOnInstalled(func(ctx context.Context, sw software.Software) {
		siteSvc.OnSoftwareInstalled(ctx, sw)
		if sw.Kind == software.KindNginx {
			settingSvc.StartGateway(ctx)
		}
	})

	return &Container{
		BaseDir:  baseDir,
		DB:       db,
		Bus:      bus,
		Gateway:  gw,
		Runner:   runner,
		Site:     siteSvc,
		Software: softSvc,
		Setting:  settingSvc,
	}, nil
}

func (c *Container) Start(ctx context.Context) {
	if c == nil {
		return
	}
	c.Setting.StartGateway(ctx)
	c.Software.ReconcileStartup(ctx)
	c.Site.ReconcileStartup(ctx)
}

func (c *Container) Shutdown() {
	if c == nil {
		return
	}
	if c.Site != nil {
		c.Site.StopAll()
	}
	if c.Software != nil {
		c.Software.StopAll()
	}
	if c.Gateway != nil {
		_ = c.Gateway.Stop()
	}
	if c.DB != nil {
		_ = c.DB.Close()
	}
}
