package app

import (
	"context"

	"devenv/internal/domain/settings"
	"devenv/internal/ports"
)

type SettingService struct {
	settings ports.SettingsRepository
}

func NewSettingService(settings ports.SettingsRepository) *SettingService {
	return &SettingService{settings: settings}
}

func (s *SettingService) Get(ctx context.Context) (settings.AppSettings, error) {
	cur, err := s.settings.Get(ctx)
	if err != nil {
		return settings.AppSettings{}, err
	}
	return *cur, nil
}

func (s *SettingService) Update(ctx context.Context, in settings.AppSettings) (settings.AppSettings, error) {
	if err := s.settings.Save(ctx, in); err != nil {
		return settings.AppSettings{}, err
	}
	return in, nil
}

func (s *SettingService) Reset(ctx context.Context) (settings.AppSettings, error) {
	def := settings.DefaultSettings()
	if err := s.settings.Save(ctx, def); err != nil {
		return settings.AppSettings{}, err
	}
	return def, nil
}
