package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"devenv/internal/domain/site"
	appErrors "devenv/internal/shared/errors"
)

type SiteRepository struct {
	db *sql.DB
}

func NewSiteRepository(db *sql.DB) *SiteRepository {
	return &SiteRepository{db: db}
}

func (r *SiteRepository) List(ctx context.Context) ([]site.Site, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, kind, root, run, software, command, port, autostart, created, updated
		FROM system_sites
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.DATABASE_ERROR, "查询站点列表失败")
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			return
		}
	}()

	list, err := scanSites(rows)
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

func (r *SiteRepository) Get(ctx context.Context, id int64) (*site.Site, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, kind, root, run, software, command, port, autostart, created, updated
		FROM system_sites
		WHERE id = ?
	`, id)
	item, err := scanSite(row)
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

func (r *SiteRepository) GetByName(ctx context.Context, name string) (*site.Site, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, kind, root, run, software, command, port, autostart, created, updated
		FROM system_sites
		WHERE name = ?
	`, name)
	item, err := scanSite(row)
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

func (r *SiteRepository) ListBySoftware(ctx context.Context, softwareID int64) ([]site.Site, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, kind, root, run, software, command, port, autostart, created, updated
		FROM system_sites
		WHERE software = ?
		ORDER BY id ASC
	`, softwareID)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.DATABASE_ERROR, "查询绑定版本的站点失败")
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			return
		}
	}()

	list, err := scanSites(rows)
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

func (r *SiteRepository) ListMissingSoftware(ctx context.Context, kind string) ([]site.Site, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, kind, root, run, software, command, port, autostart, created, updated
		FROM system_sites
		WHERE kind = ? AND software IS NULL
		ORDER BY id ASC
	`, kind)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.DATABASE_ERROR, "查询缺少环境的站点失败")
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			return
		}
	}()

	list, err := scanSites(rows)
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

func (r *SiteRepository) Create(ctx context.Context, item *site.Site) error {
	now := time.Now().UTC().Format(time.RFC3339)
	item.Created = now
	item.Updated = now

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return appErrors.Wrap(err, appErrors.DATABASE_ERROR, "开启创建站点事务失败")
	}
	defer rollbackTx(tx)

	result, err := tx.ExecContext(ctx, `
		INSERT INTO system_sites (name, kind, root, run, software, command, port, autostart, created, updated)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		item.Name,
		item.Kind,
		item.Root,
		item.Run,
		item.Software,
		item.Command,
		item.Port,
		formatBool(item.Autostart),
		item.Created,
		item.Updated,
	)
	if err != nil {
		if checkUnique(err) {
			return appErrors.New(appErrors.ALREADY_EXISTS, "站点名称已存在")
		}
		return appErrors.Wrap(err, appErrors.DATABASE_ERROR, "写入站点记录失败")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return appErrors.Wrap(err, appErrors.DATABASE_ERROR, "读取站点主键失败")
	}
	item.ID = id

	for _, host := range item.Hosts {
		cleanHost := strings.ToLower(strings.TrimSpace(host))
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO system_site_domains (site, host, created)
			VALUES (?, ?, ?)
		`, item.ID, cleanHost, now); err != nil {
			if checkUnique(err) {
				return appErrors.New(appErrors.ALREADY_EXISTS, "域名已被占用: "+cleanHost)
			}
			return appErrors.Wrap(err, appErrors.DATABASE_ERROR, "写入站点域名失败")
		}
	}

	if err := tx.Commit(); err != nil {
		return appErrors.Wrap(err, appErrors.DATABASE_ERROR, "提交创建站点事务失败")
	}
	return nil
}

func (r *SiteRepository) Update(ctx context.Context, item *site.Site) error {
	now := time.Now().UTC().Format(time.RFC3339)
	item.Updated = now

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return appErrors.Wrap(err, appErrors.DATABASE_ERROR, "开启更新站点事务失败")
	}
	defer rollbackTx(tx)

	_, err = tx.ExecContext(ctx, `
		UPDATE system_sites
		SET root = ?, run = ?, software = ?, command = ?, autostart = ?, updated = ?
		WHERE id = ?
	`,
		item.Root,
		item.Run,
		item.Software,
		item.Command,
		formatBool(item.Autostart),
		item.Updated,
		item.ID,
	)
	if err != nil {
		return appErrors.Wrap(err, appErrors.DATABASE_ERROR, "更新站点记录失败")
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM system_site_domains WHERE site = ?`, item.ID); err != nil {
		return appErrors.Wrap(err, appErrors.DATABASE_ERROR, "清理旧域名失败")
	}

	for _, host := range item.Hosts {
		cleanHost := strings.ToLower(strings.TrimSpace(host))
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO system_site_domains (site, host, created)
			VALUES (?, ?, ?)
		`, item.ID, cleanHost, now); err != nil {
			if checkUnique(err) {
				return appErrors.New(appErrors.ALREADY_EXISTS, "域名已被占用: "+cleanHost)
			}
			return appErrors.Wrap(err, appErrors.DATABASE_ERROR, "更新站点域名失败")
		}
	}

	if err := tx.Commit(); err != nil {
		return appErrors.Wrap(err, appErrors.DATABASE_ERROR, "提交更新站点事务失败")
	}
	return nil
}

func (r *SiteRepository) BindSoftware(ctx context.Context, siteID, softwareID int64) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `
		UPDATE system_sites
		SET software = ?, updated = ?
		WHERE id = ?
	`, softwareID, now, siteID)
	if err != nil {
		return appErrors.Wrap(err, appErrors.DATABASE_ERROR, "绑定站点软件版本失败")
	}
	return nil
}

func (r *SiteRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM system_sites WHERE id = ?`, id)
	if err != nil {
		return appErrors.Wrap(err, appErrors.DATABASE_ERROR, "删除站点失败")
	}
	return nil
}

func (r *SiteRepository) listHosts(ctx context.Context, siteID int64) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT host FROM system_site_domains WHERE site = ? ORDER BY id ASC
	`, siteID)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.DATABASE_ERROR, "查询站点域名失败")
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
			return nil, appErrors.Wrap(err, appErrors.DATABASE_ERROR, "读取站点域名失败")
		}
		hosts = append(hosts, host)
	}
	return hosts, rows.Err()
}

func scanSites(rows *sql.Rows) ([]site.Site, error) {
	list := make([]site.Site, 0)
	for rows.Next() {
		var item site.Site
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
			&item.Port,
			&autostart,
			&item.Created,
			&item.Updated,
		); err != nil {
			return nil, appErrors.Wrap(err, appErrors.DATABASE_ERROR, "解析站点行失败")
		}
		if softwareID.Valid {
			val := softwareID.Int64
			item.Software = &val
		}
		item.Autostart = autostart == 1
		list = append(list, item)
	}
	if err := rows.Err(); err != nil {
		return nil, appErrors.Wrap(err, appErrors.DATABASE_ERROR, "遍历站点列表失败")
	}
	return list, nil
}

func scanSite(row *sql.Row) (*site.Site, error) {
	var item site.Site
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
		&item.Port,
		&autostart,
		&item.Created,
		&item.Updated,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.DATABASE_ERROR, "查询单条站点记录失败")
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
