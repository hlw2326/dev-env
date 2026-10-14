package app

import (
	"context"

	"devenv/internal/domain/settings"
	"devenv/internal/ports"
)

type SettingsService struct {
	repo ports.SettingsRepository
}

func NewSettingsService(repo ports.SettingsRepository) *SettingsService {
	return &SettingsService{repo: repo}
}

func (s *SettingsService) GetSettings(ctx context.Context) (*settings.AppSettings, error) {
	return s.repo.Get(ctx)
}

func (s *SettingsService) UpdateSettings(ctx context.Context, in settings.AppSettings) (*settings.AppSettings, error) {
	if err := s.repo.Save(ctx, in); err != nil {
		return nil, err
	}
	return &in, nil
}

func (s *SettingsService) ResetSettings(ctx context.Context) (*settings.AppSettings, error) {
	def := settings.DefaultSettings()
	if err := s.repo.Save(ctx, def); err != nil {
		return nil, err
	}
	return &def, nil
}
