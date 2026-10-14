package sqlite

import (
	"context"
	"database/sql"
	"os"
	"strconv"
	"strings"
	"time"

	"devenv/internal/domain/setting"
	appErrors "devenv/internal/shared/errors"
)

type SettingRepository struct {
	db *sql.DB
}

func NewSettingRepository(db *sql.DB) *SettingRepository {
	return &SettingRepository{db: db}
}

func (r *SettingRepository) Load(ctx context.Context) (setting.Settings, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT key, value FROM system_app_settings`)
	if err != nil {
		return setting.Settings{}, appErrors.Wrap(err, appErrors.CodeDatabaseError, "读取系统设置失败")
	}
	defer rows.Close()

	var result setting.Settings
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return setting.Settings{}, appErrors.Wrap(err, appErrors.CodeDatabaseError, "解析系统设置失败")
		}
		switch key {
		case "theme":
			result.Theme = value
		case "root":
			result.Root = expandUserProfile(value)
		case "port":
			port, convErr := strconv.Atoi(value)
			if convErr != nil {
				return setting.Settings{}, appErrors.Wrap(convErr, appErrors.CodeDatabaseError, "解析网关端口设置失败")
			}
			result.Port = port
		case "hosts":
			result.Hosts = value == "true"
		case "autostart":
			result.Autostart = value == "true"
		}
	}
	if err := rows.Err(); err != nil {
		return setting.Settings{}, appErrors.Wrap(err, appErrors.CodeDatabaseError, "遍历系统设置失败")
	}
	return result, nil
}

func (r *SettingRepository) Save(ctx context.Context, settings setting.Settings) error {
	now := time.Now().UTC().Format(time.RFC3339)
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "开启保存设置事务失败")
	}
	defer func() { _ = tx.Rollback() }()

	pairs := map[string]string{
		"theme":     settings.Theme,
		"root":      settings.Root,
		"port":      strconv.Itoa(settings.Port),
		"hosts":     strconv.FormatBool(settings.Hosts),
		"autostart": strconv.FormatBool(settings.Autostart),
	}
	for k, v := range pairs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO system_app_settings (key, value, updated)
			VALUES (?, ?, ?)
			ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated = excluded.updated
		`, k, v, now); err != nil {
			return appErrors.Wrap(err, appErrors.CodeDatabaseError, "写入系统设置失败")
		}
	}

	if err := tx.Commit(); err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "提交系统设置事务失败")
	}
	return nil
}

func expandUserProfile(value string) string {
	if strings.Contains(value, "%USERPROFILE%") {
		profile := os.Getenv("USERPROFILE")
		return strings.ReplaceAll(value, "%USERPROFILE%", profile)
	}
	return value
}
