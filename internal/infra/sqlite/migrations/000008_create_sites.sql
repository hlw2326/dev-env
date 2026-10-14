-- Migration: 000008_create_sites
-- Description: Create system_sites table unifying projects and virtual hosts

CREATE TABLE IF NOT EXISTS system_sites (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,                          -- 网站名称
    site_type TEXT NOT NULL,                     -- static | php | nodejs | go | python
    domain_list TEXT NOT NULL DEFAULT '[]',      -- JSON 数组 ["example.com", "www.example.com"]
    root_path TEXT NOT NULL,                     -- 网站根目录

    -- 运行时环境
    runtime_version_id INTEGER,                  -- 关联运行环境（PHP/Node.js/Go/Python 版本）
    service_instance_id INTEGER,                 -- 关联服务实例（PHP-FPM/Node.js 进程）

    -- Git 配置
    git_enabled INTEGER NOT NULL DEFAULT 0,      -- 是否启用 Git
    git_remote_url TEXT,                         -- 远程仓库地址
    git_branch TEXT DEFAULT 'main',              -- 当前分支
    git_last_commit TEXT,                        -- 最后提交 SHA
    git_last_pull_at TEXT,                       -- 最后拉取时间

    -- Nginx 配置
    nginx_config_path TEXT,                      -- 生成的 Nginx 配置文件路径
    listen_port INTEGER NOT NULL DEFAULT 80,     -- 监听端口

    -- SSL 配置
    ssl_enabled INTEGER NOT NULL DEFAULT 0,      -- 是否启用 HTTPS
    ssl_cert_path TEXT,                          -- SSL 证书路径
    ssl_key_path TEXT,                           -- SSL 私钥路径
    ssl_provider TEXT,                           -- letsencrypt | self-signed | custom

    -- 反向代理配置（Node.js/Go/Python）
    proxy_enabled INTEGER NOT NULL DEFAULT 0,    -- 是否启用反向代理
    proxy_target_port INTEGER,                   -- 代理目标端口

    -- 元数据
    metadata_json TEXT NOT NULL DEFAULT '{}',    -- 扩展信息（JSON）
    status TEXT NOT NULL DEFAULT 'stopped',      -- stopped | running | error
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,

    FOREIGN KEY (runtime_version_id) REFERENCES system_runtime_versions(id) ON DELETE SET NULL,
    FOREIGN KEY (service_instance_id) REFERENCES system_service_instances(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_system_sites_type ON system_sites(site_type);
CREATE INDEX IF NOT EXISTS idx_system_sites_status ON system_sites(status);
