package sqlite

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	appErrors "devenv/internal/shared/errors"
)

var migrationFilePattern = regexp.MustCompile(`^(\d{6})_([a-z0-9_]+)\.sql$`)

type migrationFile struct {
	version  int
	name     string
	checksum string
	content  string
}

type recordedMigration struct {
	version    int
	name       string
	checksum   string
	startedAt  string
	finishedAt string
}

func openWithMigrations(path string, migrationsFS fs.FS, resetOnDrift bool, allowReset bool) (*DB, error) {
	files, err := loadMigrationFiles(migrationsFS)
	if err != nil {
		return nil, err
	}

	raw, err := openRaw(path)
	if err != nil {
		return nil, err
	}

	if err := ensureMigrationTable(raw); err != nil {
		_ = raw.Close()
		return nil, err
	}

	recorded, driftReasons, err := inspectMigrations(raw, files)
	if err != nil {
		_ = raw.Close()
		return nil, err
	}

	if len(driftReasons) > 0 {
		if !resetOnDrift || !allowReset {
			_ = raw.Close()
			return nil, appErrors.New(
				appErrors.CodeDatabaseSchemaMismatch,
				"数据库结构与迁移脚本不一致: "+strings.Join(driftReasons, "; "),
			)
		}
		if err := backupDatabase(raw, path, "pre-reset"); err != nil {
			_ = raw.Close()
			return nil, err
		}
		_ = raw.Close()
		if err := removeDatabaseFiles(path); err != nil {
			return nil, err
		}
		return openWithMigrations(path, migrationsFS, false, false)
	}

	var pending []migrationFile
	for _, f := range files {
		if _, ok := recorded[f.version]; !ok {
			pending = append(pending, f)
		}
	}

	if len(pending) > 0 {
		if len(recorded) > 0 {
			if err := backupDatabase(raw, path, "pre-migrate"); err != nil {
				_ = raw.Close()
				return nil, err
			}
		}
		for _, m := range pending {
			if err := applyMigration(raw, m); err != nil {
				_ = raw.Close()
				return nil, err
			}
		}
	}

	if err := verifyForeignKeys(raw); err != nil {
		_ = raw.Close()
		return nil, err
	}

	return &DB{DB: raw, path: path}, nil
}

func loadMigrationFiles(migrationsFS fs.FS) ([]migrationFile, error) {
	entries, err := fs.ReadDir(migrationsFS, ".")
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeMigrationFailed, "读取迁移目录失败")
	}

	seenVersions := make(map[int]string)
	var files []migrationFile

	for _, entry := range entries {
		if entry.IsDir() {
			return nil, appErrors.New(appErrors.CodeMigrationFailed, fmt.Sprintf("迁移目录中不允许包含子目录: %s", entry.Name()))
		}
		matches := migrationFilePattern.FindStringSubmatch(entry.Name())
		if matches == nil {
			return nil, appErrors.New(appErrors.CodeMigrationFailed, fmt.Sprintf("非法迁移文件名: %s", entry.Name()))
		}
		version, err := strconv.Atoi(matches[1])
		if err != nil || version <= 0 {
			return nil, appErrors.New(appErrors.CodeMigrationFailed, fmt.Sprintf("非法迁移序号: %s", entry.Name()))
		}
		if prev, dup := seenVersions[version]; dup {
			return nil, appErrors.New(appErrors.CodeMigrationFailed, fmt.Sprintf("迁移序号重复: %s 与 %s", prev, entry.Name()))
		}
		seenVersions[version] = entry.Name()

		rawBytes, err := fs.ReadFile(migrationsFS, entry.Name())
		if err != nil {
			return nil, appErrors.Wrap(err, appErrors.CodeMigrationFailed, fmt.Sprintf("读取迁移文件失败: %s", entry.Name()))
		}
		normalized := strings.ReplaceAll(string(rawBytes), "\r\n", "\n")
		sum := sha256.Sum256([]byte(normalized))

		files = append(files, migrationFile{
			version:  version,
			name:     matches[2],
			checksum: hex.EncodeToString(sum[:]),
			content:  normalized,
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].version < files[j].version
	})

	return files, nil
}

func ensureMigrationTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS system_schema_migrations (
			version     INTEGER PRIMARY KEY,
			name        TEXT    NOT NULL,
			checksum    TEXT    NOT NULL,
			started_at  TEXT    NOT NULL,
			finished_at TEXT    NOT NULL
		);
	`)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeMigrationFailed, "创建迁移记录表失败")
	}
	return nil
}

func inspectMigrations(db *sql.DB, files []migrationFile) (map[int]recordedMigration, []string, error) {
	rows, err := db.Query(`SELECT version, name, checksum, started_at, finished_at FROM system_schema_migrations ORDER BY version ASC`)
	if err != nil {
		return nil, []string{"迁移记录表结构与当前版本不兼容"}, nil
	}
	defer rows.Close()

	recorded := make(map[int]recordedMigration)
	maxRecorded := 0
	for rows.Next() {
		var rec recordedMigration
		if scanErr := rows.Scan(&rec.version, &rec.name, &rec.checksum, &rec.startedAt, &rec.finishedAt); scanErr != nil {
			return nil, []string{"迁移记录表字段类型不兼容"}, nil
		}
		recorded[rec.version] = rec
		if rec.version > maxRecorded {
			maxRecorded = rec.version
		}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, appErrors.Wrap(err, appErrors.CodeMigrationFailed, "读取迁移记录失败")
	}

	fileByVersion := make(map[int]migrationFile, len(files))
	for _, f := range files {
		fileByVersion[f.version] = f
	}

	var drifts []string
	for ver, rec := range recorded {
		f, exists := fileByVersion[ver]
		if !exists {
			drifts = append(drifts, fmt.Sprintf("已记录的迁移 %06d (%s) 文件已缺失", ver, rec.name))
			continue
		}
		if rec.checksum != f.checksum {
			drifts = append(drifts, fmt.Sprintf("迁移 %06d (%s) 校验和不匹配", ver, f.name))
		}
	}

	for _, f := range files {
		if _, applied := recorded[f.version]; !applied && f.version < maxRecorded {
			drifts = append(drifts, fmt.Sprintf("未执行的迁移 %06d (%s) 小于已记录最大序号 %06d", f.version, f.name, maxRecorded))
		}
	}

	sort.Strings(drifts)
	return recorded, drifts, nil
}

func applyMigration(db *sql.DB, m migrationFile) error {
	started := time.Now().UTC().Format(time.RFC3339)
	tx, err := db.Begin()
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeMigrationFailed, fmt.Sprintf("开启迁移事务失败 (%06d)", m.version))
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(m.content); err != nil {
		return appErrors.Wrap(err, appErrors.CodeMigrationFailed, fmt.Sprintf("执行迁移 %06d_%s 失败", m.version, m.name))
	}

	finished := time.Now().UTC().Format(time.RFC3339)
	if _, err := tx.Exec(
		`INSERT INTO system_schema_migrations (version, name, checksum, started_at, finished_at) VALUES (?, ?, ?, ?, ?)`,
		m.version, m.name, m.checksum, started, finished,
	); err != nil {
		return appErrors.Wrap(err, appErrors.CodeMigrationFailed, fmt.Sprintf("写入迁移记录 %06d_%s 失败", m.version, m.name))
	}

	if err := tx.Commit(); err != nil {
		return appErrors.Wrap(err, appErrors.CodeMigrationFailed, fmt.Sprintf("提交迁移 %06d_%s 失败", m.version, m.name))
	}
	return nil
}

func verifyForeignKeys(db *sql.DB) error {
	rows, err := db.Query(`PRAGMA foreign_key_check`)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeMigrationFailed, "执行外键完整性检查失败")
	}
	defer rows.Close()

	if rows.Next() {
		var table string
		var rowid int64
		var parent string
		var fkid int
		_ = rows.Scan(&table, &rowid, &parent, &fkid)
		return appErrors.New(appErrors.CodeMigrationFailed, fmt.Sprintf("外键约束检查未通过: 表 %s 行 %d 引用 %s", table, rowid, parent))
	}
	return rows.Err()
}

func backupDatabase(db *sql.DB, dbPath, reason string) error {
	backupDir := filepath.Join(filepath.Dir(dbPath), "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "创建数据库备份目录失败")
	}
	stamp := time.Now().UTC().Format("20060102150405")
	target := filepath.Join(backupDir, fmt.Sprintf("devenv-%s-%s.db", stamp, reason))
	for i := 1; ; i++ {
		if _, err := os.Stat(target); os.IsNotExist(err) {
			break
		}
		target = filepath.Join(backupDir, fmt.Sprintf("devenv-%s-%d-%s.db", stamp, i, reason))
	}

	escaped := strings.ReplaceAll(filepath.ToSlash(target), "'", "''")
	if _, err := db.Exec(fmt.Sprintf("VACUUM INTO '%s'", escaped)); err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "备份数据库失败")
	}

	return pruneBackups(backupDir, 5)
}

func pruneBackups(backupDir string, keep int) error {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return err
	}
	type fileEntry struct {
		path    string
		modTime time.Time
	}
	var backups []fileEntry
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".db") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		backups = append(backups, fileEntry{
			path:    filepath.Join(backupDir, entry.Name()),
			modTime: info.ModTime(),
		})
	}
	if len(backups) <= keep {
		return nil
	}
	sort.Slice(backups, func(i, j int) bool {
		if backups[i].modTime.Equal(backups[j].modTime) {
			return backups[i].path > backups[j].path
		}
		return backups[i].modTime.After(backups[j].modTime)
	})
	for _, old := range backups[keep:] {
		_ = os.Remove(old.path)
	}
	return nil
}

func removeDatabaseFiles(dbPath string) error {
	for _, suffix := range []string{"", "-wal", "-shm"} {
		target := dbPath + suffix
		if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
			return appErrors.Wrap(err, appErrors.CodeDatabaseError, "删除旧数据库文件失败")
		}
	}
	return nil
}
