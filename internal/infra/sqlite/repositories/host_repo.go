package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"devenv/internal/domain/host"
	appErrors "devenv/internal/shared/errors"
)

type HostRepository struct {
	db *sql.DB
}

func NewHostRepository(db *sql.DB) *HostRepository {
	return &HostRepository{db: db}
}

func (r *HostRepository) Create(ctx context.Context, item host.VirtualHost) (int64, error) {
	query := `
		INSERT INTO system_virtual_hosts (project_id, domain, document_root, web_server_type, port, enabled, config_path, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	createdAt := item.CreatedAt.UTC().Format(time.RFC3339)
	updatedAt := item.UpdatedAt.UTC().Format(time.RFC3339)
	enabled := 0
	if item.Enabled {
		enabled = 1
	}

	res, err := r.db.ExecContext(ctx, query,
		item.ProjectID, item.Domain, item.DocumentRoot, item.WebServerType, item.Port, enabled, item.ConfigPath, createdAt, updatedAt,
	)
	if err != nil {
		return 0, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to insert virtual host")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to get virtual host insert id")
	}
	return id, nil
}

func (r *HostRepository) GetByID(ctx context.Context, id int64) (*host.VirtualHost, error) {
	query := `
		SELECT id, project_id, domain, document_root, web_server_type, port, enabled, config_path, created_at, updated_at
		FROM system_virtual_hosts WHERE id = ?
	`
	return r.queryOne(ctx, query, id)
}

func (r *HostRepository) GetByDomain(ctx context.Context, domain string) (*host.VirtualHost, error) {
	query := `
		SELECT id, project_id, domain, document_root, web_server_type, port, enabled, config_path, created_at, updated_at
		FROM system_virtual_hosts WHERE domain = ?
	`
	return r.queryOne(ctx, query, domain)
}

func (r *HostRepository) queryOne(ctx context.Context, query string, arg interface{}) (*host.VirtualHost, error) {
	var h host.VirtualHost
	var cfgPath sql.NullString
	var enabled int
	var createdAtStr, updatedAtStr string

	err := r.db.QueryRowContext(ctx, query, arg).Scan(
		&h.ID, &h.ProjectID, &h.Domain, &h.DocumentRoot, &h.WebServerType, &h.Port, &enabled, &cfgPath, &createdAtStr, &updatedAtStr,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to query virtual host")
	}

	if cfgPath.Valid {
		h.ConfigPath = &cfgPath.String
	}
	h.Enabled = enabled == 1
	h.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	h.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
	return &h, nil
}

func (r *HostRepository) ListAll(ctx context.Context) ([]host.VirtualHost, error) {
	query := `
		SELECT id, project_id, domain, document_root, web_server_type, port, enabled, config_path, created_at, updated_at
		FROM system_virtual_hosts ORDER BY domain ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to list virtual hosts")
	}
	defer rows.Close()

	return r.scanList(rows)
}

func (r *HostRepository) ListByProjectID(ctx context.Context, projectID int64) ([]host.VirtualHost, error) {
	query := `
		SELECT id, project_id, domain, document_root, web_server_type, port, enabled, config_path, created_at, updated_at
		FROM system_virtual_hosts WHERE project_id = ? ORDER BY domain ASC
	`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to list virtual hosts by project")
	}
	defer rows.Close()

	return r.scanList(rows)
}

func (r *HostRepository) scanList(rows *sql.Rows) ([]host.VirtualHost, error) {
	var list []host.VirtualHost
	for rows.Next() {
		var h host.VirtualHost
		var cfgPath sql.NullString
		var enabled int
		var createdAtStr, updatedAtStr string

		if err := rows.Scan(
			&h.ID, &h.ProjectID, &h.Domain, &h.DocumentRoot, &h.WebServerType, &h.Port, &enabled, &cfgPath, &createdAtStr, &updatedAtStr,
		); err != nil {
			return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to scan virtual host")
		}

		if cfgPath.Valid {
			h.ConfigPath = &cfgPath.String
		}
		h.Enabled = enabled == 1
		h.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		h.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		list = append(list, h)
	}
	return list, nil
}

func (r *HostRepository) Update(ctx context.Context, item host.VirtualHost) error {
	query := `
		UPDATE system_virtual_hosts
		SET domain = ?, document_root = ?, web_server_type = ?, port = ?, enabled = ?, config_path = ?, updated_at = ?
		WHERE id = ?
	`
	updatedAt := item.UpdatedAt.UTC().Format(time.RFC3339)
	enabled := 0
	if item.Enabled {
		enabled = 1
	}

	res, err := r.db.ExecContext(ctx, query, item.Domain, item.DocumentRoot, item.WebServerType, item.Port, enabled, item.ConfigPath, updatedAt, item.ID)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to update virtual host")
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return appErrors.New(appErrors.CodeNotFound, "virtual host not found for update")
	}
	return nil
}

func (r *HostRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM system_virtual_hosts WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to delete virtual host")
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return appErrors.New(appErrors.CodeNotFound, "virtual host not found for delete")
	}
	return nil
}
