package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"devenv/internal/domain/settings"
	appErrors "devenv/internal/shared/errors"
)

type SettingsRepository struct {
	db *sql.DB
}

func NewSettingsRepository(db *sql.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

const settingsKey = "app.general"

func (r *SettingsRepository) Get(ctx context.Context) (*settings.AppSettings, error) {
	query := `SELECT value_json FROM system_app_settings WHERE key = ?`
	var valJSON string
	err := r.db.QueryRowContext(ctx, query, settingsKey).Scan(&valJSON)
	if errors.Is(err, sql.ErrNoRows) {
		def := settings.DefaultSettings()
		return &def, nil
	}
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to query settings")
	}

	var s settings.AppSettings
	if err := json.Unmarshal([]byte(valJSON), &s); err != nil {
		def := settings.DefaultSettings()
		return &def, nil
	}
	return &s, nil
}

func (r *SettingsRepository) Save(ctx context.Context, s settings.AppSettings) error {
	data, err := json.Marshal(s)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeInternalError, "failed to marshal settings")
	}

	query := `
		INSERT INTO system_app_settings (key, value_json, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value_json = excluded.value_json, updated_at = excluded.updated_at
	`
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = r.db.ExecContext(ctx, query, settingsKey, string(data), now)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to save settings")
	}
	return nil
}
