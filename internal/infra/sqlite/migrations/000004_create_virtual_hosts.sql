-- Migration: 000004_create_virtual_hosts
-- Description: Create system_virtual_hosts table with autoincrement ID

CREATE TABLE IF NOT EXISTS system_virtual_hosts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    domain TEXT NOT NULL UNIQUE,
    document_root TEXT NOT NULL,
    web_server_type TEXT NOT NULL DEFAULT 'nginx',
    port INTEGER NOT NULL DEFAULT 80,
    enabled INTEGER NOT NULL DEFAULT 1,
    config_path TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (project_id) REFERENCES system_projects(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_system_virtual_hosts_project ON system_virtual_hosts(project_id);
CREATE INDEX IF NOT EXISTS idx_system_virtual_hosts_domain ON system_virtual_hosts(domain);
