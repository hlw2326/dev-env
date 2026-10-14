package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"devenv/internal/domain/runtime"
	appErrors "devenv/internal/shared/errors"
)

type RuntimeRepository struct {
	db *sql.DB
}

func NewRuntimeRepository(db *sql.DB) *RuntimeRepository {
	return &RuntimeRepository{db: db}
}

func (r *RuntimeRepository) Create(ctx context.Context, item runtime.RuntimeVersion) (int64, error) {
	query := `
		INSERT INTO system_runtime_versions (runtime_type, version, architecture, install_path, status, metadata_json, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	createdAt := item.CreatedAt.UTC().Format(time.RFC3339)
	updatedAt := item.UpdatedAt.UTC().Format(time.RFC3339)

	res, err := r.db.ExecContext(ctx, query,
		string(item.RuntimeType), item.Version, item.Architecture, item.InstallPath, string(item.Status), item.MetadataJSON, createdAt, updatedAt,
	)
	if err != nil {
		return 0, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to insert runtime version")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to get runtime version insert id")
	}
	return id, nil
}

func (r *RuntimeRepository) GetByID(ctx context.Context, id int64) (*runtime.RuntimeVersion, error) {
	query := `
		SELECT id, runtime_type, version, architecture, install_path, status, metadata_json, created_at, updated_at
		FROM system_runtime_versions WHERE id = ?
	`
	var v runtime.RuntimeVersion
	var typeStr, statusStr, createdAtStr, updatedAtStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&v.ID, &typeStr, &v.Version, &v.Architecture, &v.InstallPath, &statusStr, &v.MetadataJSON, &createdAtStr, &updatedAtStr,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, appErrors.New(appErrors.CodeNotFound, "runtime version not found")
	}
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to query runtime version")
	}

	v.RuntimeType = runtime.RuntimeType(typeStr)
	v.Status = runtime.RuntimeStatus(statusStr)
	v.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	v.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
	return &v, nil
}

func (r *RuntimeRepository) ListByType(ctx context.Context, runtimeType runtime.RuntimeType) ([]runtime.RuntimeVersion, error) {
	query := `
		SELECT id, runtime_type, version, architecture, install_path, status, metadata_json, created_at, updated_at
		FROM system_runtime_versions WHERE runtime_type = ? ORDER BY version DESC
	`
	rows, err := r.db.QueryContext(ctx, query, string(runtimeType))
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to list runtimes by type")
	}
	defer rows.Close()

	return r.scanList(rows)
}

func (r *RuntimeRepository) ListAll(ctx context.Context) ([]runtime.RuntimeVersion, error) {
	query := `
		SELECT id, runtime_type, version, architecture, install_path, status, metadata_json, created_at, updated_at
		FROM system_runtime_versions ORDER BY runtime_type, version DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to list all runtimes")
	}
	defer rows.Close()

	return r.scanList(rows)
}

func (r *RuntimeRepository) scanList(rows *sql.Rows) ([]runtime.RuntimeVersion, error) {
	var list []runtime.RuntimeVersion
	for rows.Next() {
		var v runtime.RuntimeVersion
		var typeStr, statusStr, createdAtStr, updatedAtStr string

		if err := rows.Scan(&v.ID, &typeStr, &v.Version, &v.Architecture, &v.InstallPath, &statusStr, &v.MetadataJSON, &createdAtStr, &updatedAtStr); err != nil {
			return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to scan runtime version")
		}
		v.RuntimeType = runtime.RuntimeType(typeStr)
		v.Status = runtime.RuntimeStatus(statusStr)
		v.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		v.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		list = append(list, v)
	}
	return list, nil
}

func (r *RuntimeRepository) Update(ctx context.Context, item runtime.RuntimeVersion) error {
	query := `
		UPDATE system_runtime_versions
		SET status = ?, metadata_json = ?, updated_at = ?
		WHERE id = ?
	`
	updatedAt := item.UpdatedAt.UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx, query, string(item.Status), item.MetadataJSON, updatedAt, item.ID)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to update runtime version")
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return appErrors.New(appErrors.CodeNotFound, "runtime version not found for update")
	}
	return nil
}

func (r *RuntimeRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM system_runtime_versions WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to delete runtime version")
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return appErrors.New(appErrors.CodeNotFound, "runtime version not found for delete")
	}
	return nil
}

func (r *RuntimeRepository) CountProjectsUsingRuntime(ctx context.Context, runtimeID int64) (int, error) {
	query := `SELECT COUNT(*) FROM system_projects WHERE php_version_id = ?`
	var count int
	err := r.db.QueryRowContext(ctx, query, runtimeID).Scan(&count)
	if err != nil {
		return 0, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to count projects using runtime")
	}
	return count, nil
}
