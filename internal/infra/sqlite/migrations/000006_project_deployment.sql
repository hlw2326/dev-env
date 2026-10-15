CREATE TABLE system_project_deployment (
    project INTEGER PRIMARY KEY REFERENCES system_project(id) ON DELETE CASCADE,
    config TEXT NOT NULL,
    updated TEXT NOT NULL
);
CREATE TABLE system_project_deployment_log (
    id TEXT PRIMARY KEY,
    project INTEGER NOT NULL REFERENCES system_project(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    state TEXT NOT NULL,
    message TEXT NOT NULL,
    target TEXT NOT NULL,
    completed INTEGER NOT NULL,
    total INTEGER NOT NULL,
    created TEXT NOT NULL,
    finished TEXT NOT NULL
);
CREATE INDEX idx_system_project_deployment_log_project ON system_project_deployment_log(project, created);
