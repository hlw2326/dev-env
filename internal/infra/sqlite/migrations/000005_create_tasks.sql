-- Migration: 000005_create_tasks
-- Description: Create system_tasks table with autoincrement ID

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
