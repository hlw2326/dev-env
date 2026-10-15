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

const MAX_BACKUPS = 5

var migrationPattern = regexp.MustCompile(`^(\d{6})_([a-z0-9_]+)\.sql$`)

type migrationFile struct {
	version  int
	name     string
	checksum string
	content  string
}

type migrationRecord struct {
	version  int
	name     string
	checksum string
	started  string
	finished string
}

func openDB(path string, migrations fs.FS, reset bool, allowReset bool) (*DB, error) {
	files, err := loadFiles(migrations)
	if err != nil {
		return nil, err
	}

	raw, err := openRaw(path)
	if err != nil {
		return nil, err
	}

	if err := createTable(raw); err != nil {
		closeRaw(raw)
		return nil, err
	}

	recorded, drifts, err := inspectMigrations(raw, files)
	if err != nil {
		closeRaw(raw)
		return nil, err
	}

	if len(drifts) > 0 {
		if !reset || !allowReset {
			closeRaw(raw)
			return nil, appErrors.New(
				appErrors.DATABASE_SCHEMA_MISMATCH,
				"数据库结构与迁移脚本不一致: "+strings.Join(drifts, "; "),
			)
		}
		if err := backupDatabase(raw, path, "pre-reset"); err != nil {
			closeRaw(raw)
			return nil, err
		}
		closeRaw(raw)
		if err := removeFiles(path); err != nil {
			return nil, err
		}
		return openDB(path, migrations, false, false)
	}

	var pending []migrationFile
	for _, file := range files {
		if _, ok := recorded[file.version]; !ok {
			pending = append(pending, file)
		}
	}

	if len(pending) > 0 {
		if len(recorded) > 0 {
			if err := backupDatabase(raw, path, "pre-migrate"); err != nil {
				closeRaw(raw)
				return nil, err
			}
		}
		for _, item := range pending {
			if err := applyMigration(raw, item); err != nil {
				closeRaw(raw)
				return nil, err
			}
		}
	}

	if err := checkKeys(raw); err != nil {
		closeRaw(raw)
		return nil, err
	}

	return &DB{DB: raw, path: path}, nil
}

func loadFiles(migrations fs.FS) ([]migrationFile, error) {
	entries, err := fs.ReadDir(migrations, ".")
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.MIGRATION_FAILED, "读取迁移目录失败")
	}

	seen := make(map[int]string)
	var files []migrationFile

	for _, entry := range entries {
		if entry.IsDir() {
			return nil, appErrors.New(appErrors.MIGRATION_FAILED, fmt.Sprintf("迁移目录中不允许包含子目录: %s", entry.Name()))
		}
		matches := migrationPattern.FindStringSubmatch(entry.Name())
		if matches == nil {
			return nil, appErrors.New(appErrors.MIGRATION_FAILED, fmt.Sprintf("非法迁移文件名: %s", entry.Name()))
		}
		version, err := strconv.Atoi(matches[1])
		if err != nil || version <= 0 {
			return nil, appErrors.New(appErrors.MIGRATION_FAILED, fmt.Sprintf("非法迁移序号: %s", entry.Name()))
		}
		if prev, dup := seen[version]; dup {
			return nil, appErrors.New(appErrors.MIGRATION_FAILED, fmt.Sprintf("迁移序号重复: %s 与 %s", prev, entry.Name()))
		}
		seen[version] = entry.Name()

		rawBytes, err := fs.ReadFile(migrations, entry.Name())
		if err != nil {
			return nil, appErrors.Wrap(err, appErrors.MIGRATION_FAILED, fmt.Sprintf("读取迁移文件失败: %s", entry.Name()))
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

func createTable(db *sql.DB) error {
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
		return appErrors.Wrap(err, appErrors.MIGRATION_FAILED, "创建迁移记录表失败")
	}
	return nil
}

func inspectMigrations(db *sql.DB, files []migrationFile) (map[int]migrationRecord, []string, error) {
	rows, err := db.Query(`SELECT version, name, checksum, started_at, finished_at FROM system_schema_migrations ORDER BY version ASC`)
	if err != nil {
		return nil, []string{"迁移记录表结构与当前版本不兼容"}, nil
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			return
		}
	}()

	recorded := make(map[int]migrationRecord)
	maxVersion := 0
	for rows.Next() {
		var record migrationRecord
		if scanErr := rows.Scan(&record.version, &record.name, &record.checksum, &record.started, &record.finished); scanErr != nil {
			return nil, []string{"迁移记录表字段类型不兼容"}, nil
		}
		recorded[record.version] = record
		if record.version > maxVersion {
			maxVersion = record.version
		}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, appErrors.Wrap(err, appErrors.MIGRATION_FAILED, "读取迁移记录失败")
	}

	fileMap := make(map[int]migrationFile, len(files))
	for _, file := range files {
		fileMap[file.version] = file
	}

	var drifts []string
	for version, record := range recorded {
		file, exists := fileMap[version]
		if !exists {
			drifts = append(drifts, fmt.Sprintf("已记录的迁移 %06d (%s) 文件已缺失", version, record.name))
			continue
		}
		if record.checksum != file.checksum {
			drifts = append(drifts, fmt.Sprintf("迁移 %06d (%s) 校验和不匹配", version, file.name))
		}
	}

	for _, file := range files {
		if _, applied := recorded[file.version]; !applied && file.version < maxVersion {
			drifts = append(drifts, fmt.Sprintf("未执行的迁移 %06d (%s) 小于已记录最大序号 %06d", file.version, file.name, maxVersion))
		}
	}

	sort.Strings(drifts)
	return recorded, drifts, nil
}

func applyMigration(db *sql.DB, item migrationFile) error {
	started := time.Now().UTC().Format(time.RFC3339)
	tx, err := db.Begin()
	if err != nil {
		return appErrors.Wrap(err, appErrors.MIGRATION_FAILED, fmt.Sprintf("开启迁移事务失败 (%06d)", item.version))
	}
	defer rollbackTx(tx)

	if _, err := tx.Exec(item.content); err != nil {
		return appErrors.Wrap(err, appErrors.MIGRATION_FAILED, fmt.Sprintf("执行迁移 %06d_%s 失败", item.version, item.name))
	}

	finished := time.Now().UTC().Format(time.RFC3339)
	if _, err := tx.Exec(
		`INSERT INTO system_schema_migrations (version, name, checksum, started_at, finished_at) VALUES (?, ?, ?, ?, ?)`,
		item.version, item.name, item.checksum, started, finished,
	); err != nil {
		return appErrors.Wrap(err, appErrors.MIGRATION_FAILED, fmt.Sprintf("写入迁移记录 %06d_%s 失败", item.version, item.name))
	}

	if err := tx.Commit(); err != nil {
		return appErrors.Wrap(err, appErrors.MIGRATION_FAILED, fmt.Sprintf("提交迁移 %06d_%s 失败", item.version, item.name))
	}
	return nil
}

func checkKeys(db *sql.DB) error {
	rows, err := db.Query(`PRAGMA foreign_key_check`)
	if err != nil {
		return appErrors.Wrap(err, appErrors.MIGRATION_FAILED, "执行外键完整性检查失败")
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			return
		}
	}()

	if rows.Next() {
		var table string
		var rowid int64
		var parent string
		var fkid int
		if scanErr := rows.Scan(&table, &rowid, &parent, &fkid); scanErr != nil {
			return appErrors.Wrap(scanErr, appErrors.MIGRATION_FAILED, "读取外键检查结果失败")
		}
		return appErrors.New(appErrors.MIGRATION_FAILED, fmt.Sprintf("外键约束检查未通过: 表 %s 行 %d 引用 %s", table, rowid, parent))
	}
	return rows.Err()
}

func backupDatabase(db *sql.DB, dbPath, reason string) error {
	folder := filepath.Join(filepath.Dir(dbPath), "backups")
	if err := os.MkdirAll(folder, 0o755); err != nil {
		return appErrors.Wrap(err, appErrors.DATABASE_ERROR, "创建数据库备份目录失败")
	}
	stamp := time.Now().UTC().Format("20060102150405")
	target := filepath.Join(folder, fmt.Sprintf("devenv-%s-%s.db", stamp, reason))
	for i := 1; ; i++ {
		if _, err := os.Stat(target); os.IsNotExist(err) {
			break
		}
		target = filepath.Join(folder, fmt.Sprintf("devenv-%s-%d-%s.db", stamp, i, reason))
	}

	escaped := strings.ReplaceAll(filepath.ToSlash(target), "'", "''")
	if _, err := db.Exec(fmt.Sprintf("VACUUM INTO '%s'", escaped)); err != nil {
		return appErrors.Wrap(err, appErrors.DATABASE_ERROR, "备份数据库失败")
	}

	return pruneBackups(folder, MAX_BACKUPS)
}

func pruneBackups(folder string, keep int) error {
	entries, err := os.ReadDir(folder)
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
			path:    filepath.Join(folder, entry.Name()),
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
		if err := os.Remove(old.path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func removeFiles(dbPath string) error {
	for _, suffix := range []string{"", "-wal", "-shm"} {
		target := dbPath + suffix
		if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
			return appErrors.Wrap(err, appErrors.DATABASE_ERROR, "删除旧数据库文件失败")
		}
	}
	return nil
}
