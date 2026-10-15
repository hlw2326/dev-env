package sqlite

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	appErrors "devenv/internal/shared/errors"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var embeddedMigrations embed.FS

type Options struct {
	ResetOnDrift bool
	Migrations   fs.FS
}

type DB struct {
	*sql.DB
	path string
}

// OpenReadOnly opens an existing database without running migrations or reset.
func OpenReadOnly(path string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?mode=ro&_pragma=busy_timeout(5000)", filepath.ToSlash(path))
	raw, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.DATABASE_ERROR, "只读打开数据库失败")
	}
	raw.SetMaxOpenConns(1)
	if err := raw.Ping(); err != nil {
		_ = raw.Close()
		return nil, appErrors.Wrap(err, appErrors.DATABASE_ERROR, "只读连接数据库失败")
	}
	return raw, nil
}

func Open(path string, options Options) (*DB, error) {
	migrations := options.Migrations
	if migrations == nil {
		sub, err := fs.Sub(embeddedMigrations, "migrations")
		if err != nil {
			return nil, appErrors.Wrap(err, appErrors.MIGRATION_FAILED, "读取内置迁移脚本失败")
		}
		migrations = sub
	}
	return openDB(path, migrations, options.ResetOnDrift, true)
}

func openRaw(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, appErrors.Wrap(err, appErrors.DATABASE_ERROR, "创建数据库目录失败")
	}
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)", filepath.ToSlash(path))
	raw, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.DATABASE_ERROR, "打开数据库失败")
	}
	raw.SetMaxOpenConns(1)
	if err := raw.Ping(); err != nil {
		if closeErr := raw.Close(); closeErr != nil {
			return nil, appErrors.Wrap(err, appErrors.DATABASE_ERROR, "连接数据库失败")
		}
		return nil, appErrors.Wrap(err, appErrors.DATABASE_ERROR, "连接数据库失败")
	}
	return raw, nil
}

func closeRaw(raw *sql.DB) {
	if raw == nil {
		return
	}
	if err := raw.Close(); err != nil {
		return
	}
}

func rollbackTx(tx *sql.Tx) {
	if tx == nil {
		return
	}
	if err := tx.Rollback(); err != nil {
		return
	}
}
