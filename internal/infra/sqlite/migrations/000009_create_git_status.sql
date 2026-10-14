-- Migration: 000009_create_git_status
-- Description: Create system_git_status table for tracking site Git repository status

CREATE TABLE IF NOT EXISTS system_git_status (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    site_id INTEGER NOT NULL,
    current_branch TEXT NOT NULL,
    current_commit TEXT NOT NULL,
    is_clean INTEGER NOT NULL DEFAULT 1,         -- 工作区是否干净
    ahead_count INTEGER NOT NULL DEFAULT 0,      -- 领先远程多少提交
    behind_count INTEGER NOT NULL DEFAULT 0,     -- 落后远程多少提交
    modified_files TEXT NOT NULL DEFAULT '[]',   -- JSON 数组，已修改文件列表
    untracked_files TEXT NOT NULL DEFAULT '[]',  -- JSON 数组，未跟踪文件列表
    last_check_at TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (site_id) REFERENCES system_sites(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_system_git_status_site ON system_git_status(site_id);
