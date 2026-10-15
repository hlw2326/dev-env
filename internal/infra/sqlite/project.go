package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"
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

func (r *ProjectRepository) List(ctx context.Context) ([]project.Project, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, kind, root, run, software, command, nginx, environment_file, port, autostart, created, updated
		FROM system_project
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.DATABASE_ERROR, "查询项目列表失败")
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			return
		}
	}()

	list, err := scanProjects(rows)
	if err != nil {
		return nil, err
	}
	for i := range list {
		hosts, err := r.listHosts(ctx, list[i].ID)
		if err != nil {
			return nil, err
		}
		list[i].Hosts = hosts
	}
	return list, nil
}

func (r *ProjectRepository) Get(ctx context.Context, id int64) (*project.Project, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, kind, root, run, software, command, nginx, environment_file, port, autostart, created, updated
		FROM system_project
		WHERE id = ?
	`, id)
	item, err := scanProject(row)
	if err != nil || item == nil {
		return item, err
	}
	hosts, err := r.listHosts(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	item.Hosts = hosts
	return item, nil
}

func (r *ProjectRepository) GetByName(ctx context.Context, name string) (*project.Project, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, kind, root, run, software, command, nginx, environment_file, port, autostart, created, updated
		FROM system_project
		WHERE name = ?
	`, name)
	item, err := scanProject(row)
	if err != nil || item == nil {
		return item, err
	}
	hosts, err := r.listHosts(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	item.Hosts = hosts
	return item, nil
}

func (r *ProjectRepository) ListBySoftware(ctx context.Context, softwareID int64) ([]project.Project, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, kind, root, run, software, command, nginx, environment_file, port, autostart, created, updated
		FROM system_project
		WHERE software = ?
		ORDER BY id ASC
	`, softwareID)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.DATABASE_ERROR, "查询绑定版本的项目失败")
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			return
		}
	}()

	list, err := scanProjects(rows)
	if err != nil {
		return nil, err
	}
	for i := range list {
		hosts, err := r.listHosts(ctx, list[i].ID)
		if err != nil {
			return nil, err
		}
		list[i].Hosts = hosts
	}
	return list, nil
}

func (r *ProjectRepository) ListMissingSoftware(ctx context.Context, kind string) ([]project.Project, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, kind, root, run, software, command, nginx, environment_file, port, autostart, created, updated
		FROM system_project
		WHERE kind = ? AND software IS NULL
		ORDER BY id ASC
	`, kind)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.DATABASE_ERROR, "查询缺少环境的项目失败")
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			return
		}
	}()

	list, err := scanProjects(rows)
	if err != nil {
		return nil, err
	}
	for i := range list {
		hosts, err := r.listHosts(ctx, list[i].ID)
		if err != nil {
			return nil, err
		}
		list[i].Hosts = hosts
	}
	return list, nil
}

func (r *ProjectRepository) Create(ctx context.Context, item *project.Project) error {
	now := time.Now().UTC().Format(time.RFC3339)
	item.Created = now
	item.Updated = now

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return appErrors.Wrap(err, appErrors.DATABASE_ERROR, "开启创建项目事务失败")
	}
	defer rollbackTx(tx)

	result, err := tx.ExecContext(ctx, `
		INSERT INTO system_project (name, kind, root, run, software, command, nginx, environment_file, port, autostart, created, updated)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		item.Name,
		item.Kind,
		item.Root,
		item.Run,
		item.Software,
		item.Command,
		item.Nginx,
		item.EnvironmentFile,
		item.Port,
		formatBool(item.Autostart),
		item.Created,
		item.Updated,
	)
	if err != nil {
		if checkUnique(err) {
			return appErrors.New(appErrors.ALREADY_EXISTS, "项目名称已存在")
		}
		return appErrors.Wrap(err, appErrors.DATABASE_ERROR, "写入项目记录失败")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return appErrors.Wrap(err, appErrors.DATABASE_ERROR, "读取项目主键失败")
	}
	item.ID = id

	for _, host := range item.Hosts {
		cleanHost := strings.ToLower(strings.TrimSpace(host))
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO system_project_domain (project, host, created)
			VALUES (?, ?, ?)
		`, item.ID, cleanHost, now); err != nil {
			if checkUnique(err) {
				return appErrors.New(appErrors.ALREADY_EXISTS, "域名已被占用: "+cleanHost)
			}
			return appErrors.Wrap(err, appErrors.DATABASE_ERROR, "写入项目域名失败")
		}
	}

	if err := tx.Commit(); err != nil {
		return appErrors.Wrap(err, appErrors.DATABASE_ERROR, "提交创建项目事务失败")
	}
	return nil
}

func (r *ProjectRepository) Update(ctx context.Context, item *project.Project) error {
	now := time.Now().UTC().Format(time.RFC3339)
	item.Updated = now

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return appErrors.Wrap(err, appErrors.DATABASE_ERROR, "开启更新项目事务失败")
	}
	defer rollbackTx(tx)

	_, err = tx.ExecContext(ctx, `
		UPDATE system_project
		SET root = ?, run = ?, software = ?, command = ?, nginx = ?, environment_file = ?, autostart = ?, updated = ?
		WHERE id = ?
	`,
		item.Root,
		item.Run,
		item.Software,
		item.Command,
		item.Nginx,
		item.EnvironmentFile,
		formatBool(item.Autostart),
		item.Updated,
		item.ID,
	)
	if err != nil {
		return appErrors.Wrap(err, appErrors.DATABASE_ERROR, "更新项目记录失败")
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM system_project_domain WHERE project = ?`, item.ID); err != nil {
		return appErrors.Wrap(err, appErrors.DATABASE_ERROR, "清理旧域名失败")
	}

	for _, host := range item.Hosts {
		cleanHost := strings.ToLower(strings.TrimSpace(host))
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO system_project_domain (project, host, created)
			VALUES (?, ?, ?)
		`, item.ID, cleanHost, now); err != nil {
			if checkUnique(err) {
				return appErrors.New(appErrors.ALREADY_EXISTS, "域名已被占用: "+cleanHost)
			}
			return appErrors.Wrap(err, appErrors.DATABASE_ERROR, "更新项目域名失败")
		}
	}

	if err := tx.Commit(); err != nil {
		return appErrors.Wrap(err, appErrors.DATABASE_ERROR, "提交更新项目事务失败")
	}
	return nil
}

func (r *ProjectRepository) BindSoftware(ctx context.Context, projectID, softwareID int64) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `
		UPDATE system_project
		SET software = ?, updated = ?
		WHERE id = ?
	`, softwareID, now, projectID)
	if err != nil {
		return appErrors.Wrap(err, appErrors.DATABASE_ERROR, "绑定项目软件版本失败")
	}
	return nil
}

func (r *ProjectRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM system_project WHERE id = ?`, id)
	if err != nil {
		return appErrors.Wrap(err, appErrors.DATABASE_ERROR, "删除项目失败")
	}
	return nil
}

func (r *ProjectRepository) listHosts(ctx context.Context, projectID int64) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT host FROM system_project_domain WHERE project = ? ORDER BY id ASC
	`, projectID)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.DATABASE_ERROR, "查询项目域名失败")
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			return
		}
	}()

	hosts := make([]string, 0)
	for rows.Next() {
		var host string
		if err := rows.Scan(&host); err != nil {
			return nil, appErrors.Wrap(err, appErrors.DATABASE_ERROR, "读取项目域名失败")
		}
		hosts = append(hosts, host)
	}
	return hosts, rows.Err()
}

func scanProjects(rows *sql.Rows) ([]project.Project, error) {
	list := make([]project.Project, 0)
	for rows.Next() {
		var item project.Project
		var softwareID sql.NullInt64
		var autostart int
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Kind,
			&item.Root,
			&item.Run,
			&softwareID,
			&item.Command,
			&item.Nginx,
			&item.EnvironmentFile,
			&item.Port,
			&autostart,
			&item.Created,
			&item.Updated,
		); err != nil {
			return nil, appErrors.Wrap(err, appErrors.DATABASE_ERROR, "解析项目行失败")
		}
		if softwareID.Valid {
			val := softwareID.Int64
			item.Software = &val
		}
		item.Autostart = autostart == 1
		list = append(list, item)
	}
	if err := rows.Err(); err != nil {
		return nil, appErrors.Wrap(err, appErrors.DATABASE_ERROR, "遍历项目列表失败")
	}
	return list, nil
}

func scanProject(row *sql.Row) (*project.Project, error) {
	var item project.Project
	var softwareID sql.NullInt64
	var autostart int
	err := row.Scan(
		&item.ID,
		&item.Name,
		&item.Kind,
		&item.Root,
		&item.Run,
		&softwareID,
		&item.Command,
		&item.Nginx,
		&item.EnvironmentFile,
		&item.Port,
		&autostart,
		&item.Created,
		&item.Updated,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.DATABASE_ERROR, "查询单条项目记录失败")
	}
	if softwareID.Valid {
		val := softwareID.Int64
		item.Software = &val
	}
	item.Autostart = autostart == 1
	return &item, nil
}

func checkUnique(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unique constraint failed")
}
