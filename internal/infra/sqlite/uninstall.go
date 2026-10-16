package sqlite

import (
	"context"
	"os"
	"strings"
)

// CleanUninstallRecords removes selected installations without discarding retained software.
func CleanUninstallRecords(path string, environments [][2]string, allEnvironments, userData bool) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	db, err := Open(path, Options{})
	if err != nil {
		return err
	}
	defer db.Close()
	if _, err := db.Exec("PRAGMA secure_delete = ON"); err != nil {
		return err
	}
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if userData {
		if _, err := tx.Exec("DELETE FROM system_project; DELETE FROM system_app_setting;"); err != nil {
			return err
		}
		// Reuse the original seed values rather than maintaining another set of defaults.
		seed, err := embeddedMigrations.ReadFile("migrations/000002_seed_settings.sql")
		if err != nil {
			return err
		}
		if _, err := tx.Exec(strings.ReplaceAll(string(seed), "system_app_settings", "system_app_setting")); err != nil {
			return err
		}
	}
	for _, item := range environments {
		if _, err := tx.Exec(`UPDATE system_project SET software = NULL, autostart = 0 WHERE software IN
			(SELECT id FROM system_software WHERE kind = ? AND version = ?)`, item[0], item[1]); err != nil {
			return err
		}
		if _, err := tx.Exec("DELETE FROM system_software WHERE kind = ? AND version = ?", item[0], item[1]); err != nil {
			return err
		}
	}
	if allEnvironments {
		if _, err := tx.Exec("UPDATE system_project SET software = NULL, autostart = 0; DELETE FROM system_software;"); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	if userData {
		if _, err := db.Exec("VACUUM"); err != nil {
			return err
		}
		if _, err := db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
			return err
		}
	}
	return nil
}
