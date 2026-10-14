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

func (r *SiteRepository) Create(ctx context.Context, item *site.Site) error {
	query := `
		INSERT INTO system_sites (
			name, site_type, domain_list, root_path,
			runtime_version_id, service_instance_id,
			git_enabled, git_remote_url, git_branch, git_last_commit, git_last_pull_at,
			nginx_config_path, listen_port,
			ssl_enabled, ssl_cert_path, ssl_key_path, ssl_provider,
			proxy_enabled, proxy_target_port,
			metadata_json, status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now().UTC()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	if item.UpdatedAt.IsZero() {
		item.UpdatedAt = now
	}

	domainsJSON, err := json.Marshal(item.DomainList)
	if err != nil {
		domainsJSON = []byte("[]")
	}

	gitEnabled := 0
	if item.GitEnabled {
		gitEnabled = 1
	}
	sslEnabled := 0
	if item.SSLEnabled {
		sslEnabled = 1
	}
	proxyEnabled := 0
	if item.ProxyEnabled {
		proxyEnabled = 1
	}
	if item.MetadataJSON == "" {
		item.MetadataJSON = "{}"
	}

	res, err := r.db.ExecContext(ctx, query,
		item.Name,
		string(item.SiteType),
		string(domainsJSON),
		item.RootPath,
		item.RuntimeVersionID,
		item.ServiceInstanceID,
		gitEnabled,
		item.GitRemoteURL,
		item.GitBranch,
		item.GitLastCommit,
		item.GitLastPullAt,
		item.NginxConfigPath,
		item.ListenPort,
		sslEnabled,
		item.SSLCertPath,
		item.SSLKeyPath,
		item.SSLProvider,
		proxyEnabled,
		item.ProxyTargetPort,
		item.MetadataJSON,
		string(item.Status),
		item.CreatedAt.UTC().Format(time.RFC3339),
		item.UpdatedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to insert site")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to get site insert id")
	}
	item.ID = id
	return nil
}

func (r *SiteRepository) GetByID(ctx context.Context, id int64) (*site.Site, error) {
	query := `
		SELECT
			id, name, site_type, domain_list, root_path,
			runtime_version_id, service_instance_id,
			git_enabled, git_remote_url, git_branch, git_last_commit, git_last_pull_at,
			nginx_config_path, listen_port,
			ssl_enabled, ssl_cert_path, ssl_key_path, ssl_provider,
			proxy_enabled, proxy_target_port,
			metadata_json, status, created_at, updated_at
		FROM system_sites WHERE id = ?
	`
	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanRow(row)
}

func (r *SiteRepository) GetByDomain(ctx context.Context, domain string) (*site.Site, error) {
	domain = strings.ToLower(strings.TrimSpace(domain))
	if domain == "" {
		return nil, nil
	}
	list, err := r.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	for i := range list {
		for _, d := range list[i].DomainList {
			if strings.ToLower(strings.TrimSpace(d)) == domain {
				return &list[i], nil
			}
		}
	}
	return nil, nil
}

func (r *SiteRepository) ListAll(ctx context.Context) ([]site.Site, error) {
	query := `
		SELECT
			id, name, site_type, domain_list, root_path,
			runtime_version_id, service_instance_id,
			git_enabled, git_remote_url, git_branch, git_last_commit, git_last_pull_at,
			nginx_config_path, listen_port,
			ssl_enabled, ssl_cert_path, ssl_key_path, ssl_provider,
			proxy_enabled, proxy_target_port,
			metadata_json, status, created_at, updated_at
		FROM system_sites ORDER BY id DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to list sites")
	}
	defer rows.Close()

	var result []site.Site
	for rows.Next() {
		item, err := r.scanFromRows(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *item)
	}
	return result, nil
}

func (r *SiteRepository) Update(ctx context.Context, item *site.Site) error {
	query := `
		UPDATE system_sites SET
			name = ?, site_type = ?, domain_list = ?, root_path = ?,
			runtime_version_id = ?, service_instance_id = ?,
			git_enabled = ?, git_remote_url = ?, git_branch = ?, git_last_commit = ?, git_last_pull_at = ?,
			nginx_config_path = ?, listen_port = ?,
			ssl_enabled = ?, ssl_cert_path = ?, ssl_key_path = ?, ssl_provider = ?,
			proxy_enabled = ?, proxy_target_port = ?,
			metadata_json = ?, status = ?, updated_at = ?
		WHERE id = ?
	`
	item.UpdatedAt = time.Now().UTC()
	domainsJSON, err := json.Marshal(item.DomainList)
	if err != nil {
		domainsJSON = []byte("[]")
	}
	gitEnabled := 0
	if item.GitEnabled {
		gitEnabled = 1
	}
	sslEnabled := 0
	if item.SSLEnabled {
		sslEnabled = 1
	}
	proxyEnabled := 0
	if item.ProxyEnabled {
		proxyEnabled = 1
	}
	if item.MetadataJSON == "" {
		item.MetadataJSON = "{}"
	}

	res, err := r.db.ExecContext(ctx, query,
		item.Name,
		string(item.SiteType),
		string(domainsJSON),
		item.RootPath,
		item.RuntimeVersionID,
		item.ServiceInstanceID,
		gitEnabled,
		item.GitRemoteURL,
		item.GitBranch,
		item.GitLastCommit,
		item.GitLastPullAt,
		item.NginxConfigPath,
		item.ListenPort,
		sslEnabled,
		item.SSLCertPath,
		item.SSLKeyPath,
		item.SSLProvider,
		proxyEnabled,
		item.ProxyTargetPort,
		item.MetadataJSON,
		string(item.Status),
		item.UpdatedAt.UTC().Format(time.RFC3339),
		item.ID,
	)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to update site")
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return appErrors.New(appErrors.CodeNotFound, "site not found for update")
	}
	return nil
}

func (r *SiteRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM system_sites WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to delete site")
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return appErrors.New(appErrors.CodeNotFound, "site not found for delete")
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func (r *SiteRepository) scanRow(row *sql.Row) (*site.Site, error) {
	item, err := r.scanCommon(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to scan site")
	}
	return item, nil
}

func (r *SiteRepository) scanFromRows(rows *sql.Rows) (*site.Site, error) {
	item, err := r.scanCommon(rows)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to scan site row")
	}
	return item, nil
}

func (r *SiteRepository) scanCommon(scanner rowScanner) (*site.Site, error) {
	var s site.Site
	var siteTypeStr, domainListStr, statusStr, createdAtStr, updatedAtStr string
	var runtimeID, serviceID, proxyPort sql.NullInt64
	var gitRemote, gitBranch, gitCommit, gitPullAt, nginxPath, sslCert, sslKey, sslProvider, metaJSON sql.NullString
	var gitEnabled, sslEnabled, proxyEnabled int

	err := scanner.Scan(
		&s.ID,
		&s.Name,
		&siteTypeStr,
		&domainListStr,
		&s.RootPath,
		&runtimeID,
		&serviceID,
		&gitEnabled,
		&gitRemote,
		&gitBranch,
		&gitCommit,
		&gitPullAt,
		&nginxPath,
		&s.ListenPort,
		&sslEnabled,
		&sslCert,
		&sslKey,
		&sslProvider,
		&proxyEnabled,
		&proxyPort,
		&metaJSON,
		&statusStr,
		&createdAtStr,
		&updatedAtStr,
	)
	if err != nil {
		return nil, err
	}

	s.SiteType = site.SiteType(siteTypeStr)
	var domains []string
	if err := json.Unmarshal([]byte(domainListStr), &domains); err != nil {
		domains = []string{}
	}
	s.DomainList = domains

	if runtimeID.Valid {
		v := runtimeID.Int64
		s.RuntimeVersionID = &v
	}
	if serviceID.Valid {
		v := serviceID.Int64
		s.ServiceInstanceID = &v
	}
	s.GitEnabled = gitEnabled == 1
	s.GitRemoteURL = gitRemote.String
	s.GitBranch = gitBranch.String
	if s.GitBranch == "" {
		s.GitBranch = "main"
	}
	s.GitLastCommit = gitCommit.String
	s.GitLastPullAt = gitPullAt.String
	s.NginxConfigPath = nginxPath.String
	s.SSLEnabled = sslEnabled == 1
	s.SSLCertPath = sslCert.String
	s.SSLKeyPath = sslKey.String
	s.SSLProvider = sslProvider.String
	s.ProxyEnabled = proxyEnabled == 1
	if proxyPort.Valid {
		p := int(proxyPort.Int64)
		s.ProxyTargetPort = &p
	}
	s.MetadataJSON = metaJSON.String
	if s.MetadataJSON == "" {
		s.MetadataJSON = "{}"
	}
	s.Status = site.Status(statusStr)
	s.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	s.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

	return &s, nil
}
