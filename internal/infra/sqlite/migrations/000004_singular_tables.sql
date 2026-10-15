ALTER TABLE system_sites RENAME TO system_site;
ALTER TABLE system_site_domains RENAME TO system_site_domain;
ALTER TABLE system_app_settings RENAME TO system_app_setting;
DROP INDEX idx_system_site_domains_site;
CREATE INDEX idx_system_site_domain_site ON system_site_domain (site);
