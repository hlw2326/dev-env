package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
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

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func (r *SiteRepository) Create(ctx context.Context, item *site.Site) error {
	query := `
		INSERT INTO system_sites (
			name, kind, domains, root,
			runtime_id, service_id,
			git, remote, branch, commit_hash, pulled_at,
			conf, port,
			tls, cert, key, issuer,
			proxy, upstream,
			meta, status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now().UTC()
	item.CreatedAt = now
	item.UpdatedAt = now

	domainsRaw, err := json.Marshal(item.Domains)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeInternalError, "failed to marshal domains")
	}
	metaRaw, err := json.Marshal(item.Meta)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeInternalError, "failed to marshal meta")
	}

	res, err := r.db.ExecContext(ctx, query,
		item.Name,
		string(item.Kind),
		string(domainsRaw),
		item.Root,
		item.RuntimeID,
		item.ServiceID,
		boolToInt(item.Git),
		item.Remote,
		item.Branch,
		item.Commit,
		item.PulledAt,
		item.Conf,
		item.Port,
		boolToInt(item.TLS),
		item.Cert,
		item.Key,
		item.Issuer,
		boolToInt(item.Proxy),
		item.Upstream,
		string(metaRaw),
		string(item.Status),
		item.CreatedAt.Format(time.RFC3339),
		item.UpdatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to insert site")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to get site id")
	}
	item.ID = id
	return nil
}

func (r *SiteRepository) Get(ctx context.Context, id int64) (*site.Site, error) {
	query := `
		SELECT
			id, name, kind, domains, root,
			runtime_id, service_id,
			git, remote, branch, commit_hash, pulled_at,
			conf, port,
			tls, cert, key, issuer,
			proxy, upstream,
			meta, status, created_at, updated_at
		FROM system_sites WHERE id = ?
	`
	row := r.db.QueryRowContext(ctx, query, id)
	item, err := r.scan(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to query site")
	}
	return item, nil
}

func (r *SiteRepository) GetByDomain(ctx context.Context, domain string) (*site.Site, error) {
	domain = strings.ToLower(strings.TrimSpace(domain))
	list, err := r.List(ctx)
	if err != nil {
		return nil, err
	}
	for i := range list {
		for _, d := range list[i].Domains {
			if strings.ToLower(strings.TrimSpace(d)) == domain {
				return &list[i], nil
			}
		}
	}
	return nil, nil
}

func (r *SiteRepository) List(ctx context.Context) ([]site.Site, error) {
	query := `
		SELECT
			id, name, kind, domains, root,
			runtime_id, service_id,
			git, remote, branch, commit_hash, pulled_at,
			conf, port,
			tls, cert, key, issuer,
			proxy, upstream,
			meta, status, created_at, updated_at
		FROM system_sites ORDER BY id DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to list sites")
	}
	defer rows.Close()

	list := make([]site.Site, 0)
	for rows.Next() {
		item, err := r.scan(rows)
		if err != nil {
			return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to scan site")
		}
		list = append(list, *item)
	}
	return list, nil
}

func (r *SiteRepository) Update(ctx context.Context, item *site.Site) error {
	query := `
		UPDATE system_sites SET
			name = ?, kind = ?, domains = ?, root = ?,
			runtime_id = ?, service_id = ?,
			git = ?, remote = ?, branch = ?, commit_hash = ?, pulled_at = ?,
			conf = ?, port = ?,
			tls = ?, cert = ?, key = ?, issuer = ?,
			proxy = ?, upstream = ?,
			meta = ?, status = ?, updated_at = ?
		WHERE id = ?
	`
	item.UpdatedAt = time.Now().UTC()
	domainsRaw, err := json.Marshal(item.Domains)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeInternalError, "failed to marshal domains")
	}
	metaRaw, err := json.Marshal(item.Meta)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeInternalError, "failed to marshal meta")
	}

	res, err := r.db.ExecContext(ctx, query,
		item.Name,
		string(item.Kind),
		string(domainsRaw),
		item.Root,
		item.RuntimeID,
		item.ServiceID,
		boolToInt(item.Git),
		item.Remote,
		item.Branch,
		item.Commit,
		item.PulledAt,
		item.Conf,
		item.Port,
		boolToInt(item.TLS),
		item.Cert,
		item.Key,
		item.Issuer,
		boolToInt(item.Proxy),
		item.Upstream,
		string(metaRaw),
		string(item.Status),
		item.UpdatedAt.Format(time.RFC3339),
		item.ID,
	)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to update site")
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to get rows affected")
	}
	if affected == 0 {
		return appErrors.New(appErrors.CodeNotFound, "site not found")
	}
	return nil
}

func (r *SiteRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM system_sites WHERE id = ?`, id)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to delete site")
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to get rows affected")
	}
	if affected == 0 {
		return appErrors.New(appErrors.CodeNotFound, "site not found")
	}
	return nil
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func (r *SiteRepository) scan(sc scanner) (*site.Site, error) {
	var s site.Site
	var kindStr, domainsStr, metaStr, statusStr, createdStr, updatedStr string
	var runtimeID, serviceID, upstream sql.NullInt64
	var gitInt, tlsInt, proxyInt int

	err := sc.Scan(
		&s.ID,
		&s.Name,
		&kindStr,
		&domainsStr,
		&s.Root,
		&runtimeID,
		&serviceID,
		&gitInt,
		&s.Remote,
		&s.Branch,
		&s.Commit,
		&s.PulledAt,
		&s.Conf,
		&s.Port,
		&tlsInt,
		&s.Cert,
		&s.Key,
		&s.Issuer,
		&proxyInt,
		&upstream,
		&metaStr,
		&statusStr,
		&createdStr,
		&updatedStr,
	)
	if err != nil {
		return nil, err
	}

	s.Kind = site.Kind(kindStr)
	s.Domains = make([]string, 0)
	if err := json.Unmarshal([]byte(domainsStr), &s.Domains); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(metaStr), &s.Meta); err != nil {
		return nil, err
	}
	if s.Meta.History == nil {
		s.Meta.History = make([]site.Snapshot, 0)
	}
	if s.Meta.Env == nil {
		s.Meta.Env = make(map[string]string)
	}

	if runtimeID.Valid {
		v := runtimeID.Int64
		s.RuntimeID = &v
	}
	if serviceID.Valid {
		v := serviceID.Int64
		s.ServiceID = &v
	}
	if upstream.Valid {
		p := int(upstream.Int64)
		s.Upstream = &p
	}

	s.Git = gitInt == 1
	s.TLS = tlsInt == 1
	s.Proxy = proxyInt == 1
	s.Status = site.Status(statusStr)

	createdAt, err := time.Parse(time.RFC3339, createdStr)
	if err != nil {
		return nil, err
	}
	updatedAt, err := time.Parse(time.RFC3339, updatedStr)
	if err != nil {
		return nil, err
	}
	s.CreatedAt = createdAt
	s.UpdatedAt = updatedAt

	return &s, nil
}
