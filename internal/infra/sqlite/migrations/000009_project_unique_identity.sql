CREATE UNIQUE INDEX idx_system_project_name_unique ON system_project (name COLLATE NOCASE);
CREATE UNIQUE INDEX idx_system_project_root_unique ON system_project (root COLLATE NOCASE);
