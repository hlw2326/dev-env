package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type DB struct {
	*sql.DB
	path string
}

func Open(dbPath string) (*DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(1) // SQLite single writer optimization
	db.SetMaxIdleConns(1)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	sdb := &DB{
		DB:   db,
		path: dbPath,
	}

	if err := sdb.runMigrations(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return sdb, nil
}

func (d *DB) runMigrations(ctx context.Context) error {
	// Ensure legacy schema prior to runtime_id / system_sites refactor is upgraded
	var hasRuntimeID int
	_ = d.QueryRowContext(ctx, `SELECT COUNT(*) FROM pragma_table_info('system_service_instances') WHERE name = 'runtime_id'`).Scan(&hasRuntimeID)
	var hasSitesTable int
	_ = d.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'system_sites'`).Scan(&hasSitesTable)
	if hasRuntimeID == 0 || hasSitesTable == 0 {
		_, _ = d.ExecContext(ctx, `
			PRAGMA foreign_keys = OFF;
			DROP TABLE IF EXISTS system_git_repo_statuses;
			DROP TABLE IF EXISTS system_git_status;
			DROP TABLE IF EXISTS system_virtual_hosts;
			DROP TABLE IF EXISTS system_projects;
			DROP TABLE IF EXISTS system_sites;
			DROP TABLE IF EXISTS system_service_instances;
			DROP TABLE IF EXISTS system_schema_migrations;
			PRAGMA foreign_keys = ON;
		`)
	}

	// Create system_schema_migrations table if not exists
	_, err := d.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS system_schema_migrations (
			version TEXT PRIMARY KEY NOT NULL,
			name TEXT NOT NULL,
			applied_at TEXT NOT NULL
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to ensure system_schema_migrations table: %w", err)
	}

	// Read migration files
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		filename := entry.Name()
		parts := strings.SplitN(filename, "_", 2)
		if len(parts) < 2 {
			continue
		}
		version := parts[0]
		name := strings.TrimSuffix(parts[1], ".sql")

		var count int
		err := d.QueryRowContext(ctx, "SELECT COUNT(*) FROM system_schema_migrations WHERE version = ?", version).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration %s: %w", version, err)
		}
		if count > 0 {
			continue
		}

		content, err := migrationsFS.ReadFile("migrations/" + filename)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		tx, err := d.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction for %s: %w", filename, err)
		}

		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		now := time.Now().UTC().Format(time.RFC3339)
		if _, err := tx.ExecContext(ctx, "INSERT INTO system_schema_migrations (version, name, applied_at) VALUES (?, ?, ?)", version, name, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", filename, err)
		}
	}

	return nil
}
