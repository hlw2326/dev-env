-- Migration: 000002_create_runtime_versions
-- Description: Create system_runtime_versions table with autoincrement ID

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
