ALTER TABLE system_site RENAME TO system_project;
ALTER TABLE system_site_domain RENAME TO system_project_domain;
ALTER TABLE system_project_domain RENAME COLUMN site TO project;
DROP INDEX idx_system_site_domain_site;
CREATE INDEX idx_system_project_domain_project ON system_project_domain (project);
