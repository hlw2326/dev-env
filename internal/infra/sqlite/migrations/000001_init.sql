-- Migration: 000001_init
-- Description: Initial schema with system_ prefix and concise column names

CREATE TABLE IF NOT EXISTS system_runtime_versions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    runtime_type TEXT NOT NULL,
    version TEXT NOT NULL,
    architecture TEXT NOT NULL,
    install_path TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'available',
    metadata_json TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    UNIQUE (runtime_type, version, architecture, install_path)
);

CREATE INDEX IF NOT EXISTS idx_system_runtime_versions_type_status ON system_runtime_versions(runtime_type, status);

CREATE TABLE IF NOT EXISTS system_service_instances (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    runtime_id INTEGER NOT NULL,
    service_type TEXT NOT NULL,
    port INTEGER,
    config_path TEXT,
    data_path TEXT,
    auto_start INTEGER NOT NULL DEFAULT 0,
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (runtime_id) REFERENCES system_runtime_versions(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_system_service_instances_type ON system_service_instances(service_type);
CREATE INDEX IF NOT EXISTS idx_system_service_instances_runtime ON system_service_instances(runtime_id);

CREATE TABLE IF NOT EXISTS system_sites (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    kind TEXT NOT NULL,
    domains TEXT NOT NULL DEFAULT '[]',
    root TEXT NOT NULL,
    runtime_id INTEGER,
    service_id INTEGER,
    git INTEGER NOT NULL DEFAULT 0,
    remote TEXT NOT NULL DEFAULT '',
    branch TEXT NOT NULL DEFAULT 'main',
    commit_hash TEXT NOT NULL DEFAULT '',
    pulled_at TEXT NOT NULL DEFAULT '',
    conf TEXT NOT NULL DEFAULT '',
    port INTEGER NOT NULL DEFAULT 80,
    tls INTEGER NOT NULL DEFAULT 0,
    cert TEXT NOT NULL DEFAULT '',
    key TEXT NOT NULL DEFAULT '',
    issuer TEXT NOT NULL DEFAULT '',
    proxy INTEGER NOT NULL DEFAULT 0,
    upstream INTEGER,
    meta TEXT NOT NULL DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'stopped',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (runtime_id) REFERENCES system_runtime_versions(id) ON DELETE SET NULL,
    FOREIGN KEY (service_id) REFERENCES system_service_instances(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_system_sites_kind ON system_sites(kind);
CREATE INDEX IF NOT EXISTS idx_system_sites_status ON system_sites(status);

CREATE TABLE IF NOT EXISTS system_git_status (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    site_id INTEGER NOT NULL,
    branch TEXT NOT NULL,
    commit_hash TEXT NOT NULL,
    clean INTEGER NOT NULL DEFAULT 1,
    ahead INTEGER NOT NULL DEFAULT 0,
    behind INTEGER NOT NULL DEFAULT 0,
    modified TEXT NOT NULL DEFAULT '[]',
    untracked TEXT NOT NULL DEFAULT '[]',
    checked_at TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (site_id) REFERENCES system_sites(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_system_git_status_site ON system_git_status(site_id);

CREATE TABLE IF NOT EXISTS system_tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'queued',
    progress INTEGER NOT NULL DEFAULT 0,
    message TEXT,
    error_code TEXT,
    error_message TEXT,
    payload_json TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    finished_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_system_tasks_status ON system_tasks(status);
CREATE INDEX IF NOT EXISTS idx_system_tasks_type ON system_tasks(task_type);

CREATE TABLE IF NOT EXISTS system_app_settings (
    key TEXT PRIMARY KEY NOT NULL,
    value_json TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
