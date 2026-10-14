-- Migration: 000006_create_app_settings
-- Description: Create system_app_settings table for application settings

CREATE TABLE IF NOT EXISTS system_app_settings (
    key TEXT PRIMARY KEY NOT NULL,
    value_json TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
