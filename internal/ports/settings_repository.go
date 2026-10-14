package ports

import (
	"context"

	"devenv/internal/domain/settings"
)

type SettingsRepository interface {
	Get(ctx context.Context) (*settings.AppSettings, error)
	Save(ctx context.Context, s settings.AppSettings) error
}
