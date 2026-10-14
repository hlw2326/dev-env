package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"devenv/internal/domain/project"
	appErrors "devenv/internal/shared/errors"
)

type ProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(ctx context.Context, item project.Project) (int64, error) {
	query := `
		INSERT INTO system_projects (name, path, path_key, php_version_id, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	createdAt := item.CreatedAt.UTC().Format(time.RFC3339)
	updatedAt := item.UpdatedAt.UTC().Format(time.RFC3339)

	res, err := r.db.ExecContext(ctx, query,
		item.Name, item.Path, item.PathKey, item.PHPVersionID, string(item.Status), createdAt, updatedAt,
	)
	if err != nil {
		return 0, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to insert project")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to get project insert id")
	}
	return id, nil
}

func (r *ProjectRepository) GetByID(ctx context.Context, id int64) (*project.Project, error) {
	query := `
		SELECT id, name, path, path_key, php_version_id, status, created_at, updated_at
		FROM system_projects WHERE id = ?
	`
	var p project.Project
	var phpVersionID sql.NullInt64
	var statusStr, createdAtStr, updatedAtStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Path, &p.PathKey, &phpVersionID, &statusStr, &createdAtStr, &updatedAtStr,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, appErrors.New(appErrors.CodeNotFound, "project not found")
	}
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to query project")
	}

	if phpVersionID.Valid {
		val := phpVersionID.Int64
		p.PHPVersionID = &val
	}
	p.Status = project.Status(statusStr)
	p.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	p.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
	return &p, nil
}

func (r *ProjectRepository) GetByPathKey(ctx context.Context, pathKey string) (*project.Project, error) {
	query := `
		SELECT id, name, path, path_key, php_version_id, status, created_at, updated_at
		FROM system_projects WHERE path_key = ?
	`
	var p project.Project
	var phpVersionID sql.NullInt64
	var statusStr, createdAtStr, updatedAtStr string

	err := r.db.QueryRowContext(ctx, query, pathKey).Scan(
		&p.ID, &p.Name, &p.Path, &p.PathKey, &phpVersionID, &statusStr, &createdAtStr, &updatedAtStr,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to query project by path key")
	}

	if phpVersionID.Valid {
		val := phpVersionID.Int64
		p.PHPVersionID = &val
	}
	p.Status = project.Status(statusStr)
	p.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	p.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
	return &p, nil
}

func (r *ProjectRepository) List(ctx context.Context) ([]project.Project, error) {
	query := `
		SELECT id, name, path, path_key, php_version_id, status, created_at, updated_at
		FROM system_projects ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to list projects")
	}
	defer rows.Close()

	var list []project.Project
	for rows.Next() {
		var p project.Project
		var phpVersionID sql.NullInt64
		var statusStr, createdAtStr, updatedAtStr string

		if err := rows.Scan(&p.ID, &p.Name, &p.Path, &p.PathKey, &phpVersionID, &statusStr, &createdAtStr, &updatedAtStr); err != nil {
			return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to scan project")
		}
		if phpVersionID.Valid {
			val := phpVersionID.Int64
			p.PHPVersionID = &val
		}
		p.Status = project.Status(statusStr)
		p.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		p.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		list = append(list, p)
	}
	return list, nil
}

func (r *ProjectRepository) Update(ctx context.Context, item project.Project) error {
	query := `
		UPDATE system_projects
		SET name = ?, php_version_id = ?, status = ?, updated_at = ?
		WHERE id = ?
	`
	updatedAt := item.UpdatedAt.UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx, query, item.Name, item.PHPVersionID, string(item.Status), updatedAt, item.ID)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to update project")
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return appErrors.New(appErrors.CodeNotFound, "project not found for update")
	}
	return nil
}

func (r *ProjectRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM system_projects WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to delete project")
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return appErrors.New(appErrors.CodeNotFound, "project not found for delete")
	}
	return nil
}
