-- Migration: 000003_create_service_instances
-- Description: Create system_service_instances table with autoincrement ID

CREATE TABLE IF NOT EXISTS system_service_instances (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    runtime_version_id INTEGER NOT NULL,
    service_type TEXT NOT NULL,
    port INTEGER,
    config_path TEXT,
    data_path TEXT,
    auto_start INTEGER NOT NULL DEFAULT 0,
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (runtime_version_id) REFERENCES system_runtime_versions(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_system_service_instances_type ON system_service_instances(service_type);
CREATE INDEX IF NOT EXISTS idx_system_service_instances_runtime ON system_service_instances(runtime_version_id);
