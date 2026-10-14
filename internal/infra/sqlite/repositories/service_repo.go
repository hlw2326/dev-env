package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"devenv/internal/domain/runtime"
	"devenv/internal/domain/service"
	appErrors "devenv/internal/shared/errors"
)

type ServiceRepository struct {
	db *sql.DB
}

func NewServiceRepository(db *sql.DB) *ServiceRepository {
	return &ServiceRepository{db: db}
}

func (r *ServiceRepository) Create(ctx context.Context, item service.Instance) (int64, error) {
	query := `
		INSERT INTO system_service_instances (name, runtime_version_id, service_type, port, config_path, data_path, auto_start, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	createdAt := item.CreatedAt.UTC().Format(time.RFC3339)
	updatedAt := item.UpdatedAt.UTC().Format(time.RFC3339)
	autoStart := 0
	if item.AutoStart {
		autoStart = 1
	}
	enabled := 0
	if item.Enabled {
		enabled = 1
	}

	res, err := r.db.ExecContext(ctx, query,
		item.Name, item.RuntimeVersionID, string(item.ServiceType), item.Port, item.ConfigPath, item.DataPath, autoStart, enabled, createdAt, updatedAt,
	)
	if err != nil {
		return 0, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to insert service instance")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to get service instance insert id")
	}
	return id, nil
}

func (r *ServiceRepository) GetByID(ctx context.Context, id int64) (*service.Instance, error) {
	query := `
		SELECT id, name, runtime_version_id, service_type, port, config_path, data_path, auto_start, enabled, created_at, updated_at
		FROM system_service_instances WHERE id = ?
	`
	var s service.Instance
	var sType string
	var port sql.NullInt64
	var cfgPath, dataPath sql.NullString
	var autoStart, enabled int
	var createdAtStr, updatedAtStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.Name, &s.RuntimeVersionID, &sType, &port, &cfgPath, &dataPath, &autoStart, &enabled, &createdAtStr, &updatedAtStr,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, appErrors.New(appErrors.CodeNotFound, "service instance not found")
	}
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to query service instance")
	}

	s.ServiceType = runtime.RuntimeType(sType)
	if port.Valid {
		p := int(port.Int64)
		s.Port = &p
	}
	if cfgPath.Valid {
		s.ConfigPath = &cfgPath.String
	}
	if dataPath.Valid {
		s.DataPath = &dataPath.String
	}
	s.AutoStart = autoStart == 1
	s.Enabled = enabled == 1
	s.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	s.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
	return &s, nil
}

func (r *ServiceRepository) ListAll(ctx context.Context) ([]service.Instance, error) {
	query := `
		SELECT id, name, runtime_version_id, service_type, port, config_path, data_path, auto_start, enabled, created_at, updated_at
		FROM system_service_instances ORDER BY name ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to list services")
	}
	defer rows.Close()

	return r.scanList(rows)
}

func (r *ServiceRepository) ListByType(ctx context.Context, sType runtime.RuntimeType) ([]service.Instance, error) {
	query := `
		SELECT id, name, runtime_version_id, service_type, port, config_path, data_path, auto_start, enabled, created_at, updated_at
		FROM system_service_instances WHERE service_type = ? ORDER BY name ASC
	`
	rows, err := r.db.QueryContext(ctx, query, string(sType))
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to list services by type")
	}
	defer rows.Close()

	return r.scanList(rows)
}

func (r *ServiceRepository) scanList(rows *sql.Rows) ([]service.Instance, error) {
	var list []service.Instance
	for rows.Next() {
		var s service.Instance
		var sType string
		var port sql.NullInt64
		var cfgPath, dataPath sql.NullString
		var autoStart, enabled int
		var createdAtStr, updatedAtStr string

		if err := rows.Scan(
			&s.ID, &s.Name, &s.RuntimeVersionID, &sType, &port, &cfgPath, &dataPath, &autoStart, &enabled, &createdAtStr, &updatedAtStr,
		); err != nil {
			return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to scan service instance")
		}

		s.ServiceType = runtime.RuntimeType(sType)
		if port.Valid {
			p := int(port.Int64)
			s.Port = &p
		}
		if cfgPath.Valid {
			s.ConfigPath = &cfgPath.String
		}
		if dataPath.Valid {
			s.DataPath = &dataPath.String
		}
		s.AutoStart = autoStart == 1
		s.Enabled = enabled == 1
		s.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		s.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		list = append(list, s)
	}
	return list, nil
}

func (r *ServiceRepository) Update(ctx context.Context, item service.Instance) error {
	query := `
		UPDATE system_service_instances
		SET name = ?, port = ?, config_path = ?, data_path = ?, auto_start = ?, enabled = ?, updated_at = ?
		WHERE id = ?
	`
	updatedAt := item.UpdatedAt.UTC().Format(time.RFC3339)
	autoStart := 0
	if item.AutoStart {
		autoStart = 1
	}
	enabled := 0
	if item.Enabled {
		enabled = 1
	}

	res, err := r.db.ExecContext(ctx, query, item.Name, item.Port, item.ConfigPath, item.DataPath, autoStart, enabled, updatedAt, item.ID)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to update service instance")
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return appErrors.New(appErrors.CodeNotFound, "service instance not found for update")
	}
	return nil
}

func (r *ServiceRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM system_service_instances WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to delete service instance")
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return appErrors.New(appErrors.CodeNotFound, "service instance not found for delete")
	}
	return nil
}

func (r *ServiceRepository) CountServicesUsingRuntime(ctx context.Context, runtimeID int64) (int, error) {
	query := `SELECT COUNT(*) FROM system_service_instances WHERE runtime_version_id = ?`
	var count int
	err := r.db.QueryRowContext(ctx, query, runtimeID).Scan(&count)
	if err != nil {
		return 0, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to count services using runtime")
	}
	return count, nil
}
