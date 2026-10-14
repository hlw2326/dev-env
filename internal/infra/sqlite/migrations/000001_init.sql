CREATE TABLE system_software (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    kind      TEXT    NOT NULL,
    version   TEXT    NOT NULL,
    path      TEXT    NOT NULL,
    state     TEXT    NOT NULL,
    progress  INTEGER NOT NULL DEFAULT 0,
    error     TEXT    NOT NULL DEFAULT '',
    main      INTEGER NOT NULL DEFAULT 0,
    port      INTEGER NOT NULL DEFAULT 0,
    autostart INTEGER NOT NULL DEFAULT 0,
    created   TEXT    NOT NULL,
    updated   TEXT    NOT NULL,
    UNIQUE (kind, version)
);

CREATE INDEX idx_system_software_kind_state ON system_software (kind, state);

CREATE TABLE system_sites (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    name      TEXT    NOT NULL UNIQUE,
    kind      TEXT    NOT NULL,
    root      TEXT    NOT NULL,
    run       TEXT    NOT NULL DEFAULT '',
    software  INTEGER NULL REFERENCES system_software(id) ON DELETE RESTRICT,
    command   TEXT    NOT NULL DEFAULT '',
    port      INTEGER NOT NULL DEFAULT 0,
    autostart INTEGER NOT NULL DEFAULT 1,
    created   TEXT    NOT NULL,
    updated   TEXT    NOT NULL
);

CREATE TABLE system_site_domains (
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    site    INTEGER NOT NULL REFERENCES system_sites(id) ON DELETE CASCADE,
    host    TEXT    NOT NULL UNIQUE,
    created TEXT    NOT NULL
);

CREATE INDEX idx_system_site_domains_site ON system_site_domains (site);

CREATE TABLE system_app_settings (
    key     TEXT PRIMARY KEY,
    value   TEXT NOT NULL,
    updated TEXT NOT NULL
);
