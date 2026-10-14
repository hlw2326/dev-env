package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"devenv/internal/domain/software"
	appErrors "devenv/internal/shared/errors"
)

type SoftwareRepository struct {
	db *sql.DB
}

func NewSoftwareRepository(db *sql.DB) *SoftwareRepository {
	return &SoftwareRepository{db: db}
}

func (r *SoftwareRepository) List(ctx context.Context) ([]software.Software, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, kind, version, path, state, progress, error, main, port, autostart, created, updated
		FROM system_software
		ORDER BY kind ASC, main DESC, version DESC, id DESC
	`)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "查询软件列表失败")
	}
	defer rows.Close()
	return scanSoftwareRows(rows)
}

func (r *SoftwareRepository) ListByKind(ctx context.Context, kind string) ([]software.Software, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, kind, version, path, state, progress, error, main, port, autostart, created, updated
		FROM system_software
		WHERE kind = ?
		ORDER BY main DESC, version DESC, id DESC
	`, kind)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "按类型查询软件失败")
	}
	defer rows.Close()
	return scanSoftwareRows(rows)
}

func (r *SoftwareRepository) Get(ctx context.Context, id int64) (*software.Software, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, kind, version, path, state, progress, error, main, port, autostart, created, updated
		FROM system_software
		WHERE id = ?
	`, id)
	return scanSoftwareRow(row)
}

func (r *SoftwareRepository) GetByKindVersion(ctx context.Context, kind, version string) (*software.Software, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, kind, version, path, state, progress, error, main, port, autostart, created, updated
		FROM system_software
		WHERE kind = ? AND version = ?
	`, kind, version)
	return scanSoftwareRow(row)
}

func (r *SoftwareRepository) GetMain(ctx context.Context, kind string) (*software.Software, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, kind, version, path, state, progress, error, main, port, autostart, created, updated
		FROM system_software
		WHERE kind = ? AND main = 1 AND state = ?
		LIMIT 1
	`, kind, software.StateInstalled)
	return scanSoftwareRow(row)
}

func (r *SoftwareRepository) Create(ctx context.Context, item *software.Software) error {
	now := time.Now().UTC().Format(time.RFC3339)
	item.Created = now
	item.Updated = now

	res, err := r.db.ExecContext(ctx, `
		INSERT INTO system_software (kind, version, path, state, progress, error, main, port, autostart, created, updated)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		item.Kind,
		item.Version,
		item.Path,
		item.State,
		item.Progress,
		item.Error,
		boolToInt(item.Main),
		item.Port,
		boolToInt(item.Autostart),
		item.Created,
		item.Updated,
	)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "创建软件记录失败")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "读取软件记录主键失败")
	}
	item.ID = id
	return nil
}

func (r *SoftwareRepository) Update(ctx context.Context, item *software.Software) error {
	item.Updated = time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `
		UPDATE system_software
		SET path = ?, state = ?, progress = ?, error = ?, main = ?, port = ?, autostart = ?, updated = ?
		WHERE id = ?
	`,
		item.Path,
		item.State,
		item.Progress,
		item.Error,
		boolToInt(item.Main),
		item.Port,
		boolToInt(item.Autostart),
		item.Updated,
		item.ID,
	)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "更新软件记录失败")
	}
	return nil
}

func (r *SoftwareRepository) SetMain(ctx context.Context, kind string, id int64) error {
	now := time.Now().UTC().Format(time.RFC3339)
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "开启事务失败")
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `UPDATE system_software SET main = 0, updated = ? WHERE kind = ?`, now, kind); err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "清除旧默认版本失败")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE system_software SET main = 1, updated = ? WHERE id = ? AND kind = ?`, now, id, kind); err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "设置默认版本失败")
	}
	if err := tx.Commit(); err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "提交默认版本失败")
	}
	return nil
}

func (r *SoftwareRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM system_software WHERE id = ?`, id)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "删除软件记录失败")
	}
	return nil
}

func scanSoftwareRows(rows *sql.Rows) ([]software.Software, error) {
	list := make([]software.Software, 0)
	for rows.Next() {
		var item software.Software
		var mainInt, autostartInt int
		if err := rows.Scan(
			&item.ID,
			&item.Kind,
			&item.Version,
			&item.Path,
			&item.State,
			&item.Progress,
			&item.Error,
			&mainInt,
			&item.Port,
			&autostartInt,
			&item.Created,
			&item.Updated,
		); err != nil {
			return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "解析软件行失败")
		}
		item.Main = mainInt == 1
		item.Autostart = autostartInt == 1
		list = append(list, item)
	}
	if err := rows.Err(); err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "遍历软件列表失败")
	}
	return list, nil
}

func scanSoftwareRow(row *sql.Row) (*software.Software, error) {
	var item software.Software
	var mainInt, autostartInt int
	err := row.Scan(
		&item.ID,
		&item.Kind,
		&item.Version,
		&item.Path,
		&item.State,
		&item.Progress,
		&item.Error,
		&mainInt,
		&item.Port,
		&autostartInt,
		&item.Created,
		&item.Updated,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "查询单条软件记录失败")
	}
	item.Main = mainInt == 1
	item.Autostart = autostartInt == 1
	return &item, nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
