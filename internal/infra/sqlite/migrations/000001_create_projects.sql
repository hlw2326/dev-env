-- Migration: 000001_create_projects
-- Description: Create system_projects table with autoincrement ID

CREATE TABLE IF NOT EXISTS system_projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    path TEXT NOT NULL,
    path_key TEXT NOT NULL UNIQUE,
    php_version_id INTEGER,
    status TEXT NOT NULL DEFAULT 'ready',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (php_version_id) REFERENCES system_runtime_versions(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_system_projects_php_version_id ON system_projects(php_version_id);
