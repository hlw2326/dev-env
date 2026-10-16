// Package uninstall applies explicit uninstall choices to managed data only.
package uninstall

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"devenv/internal/infra/sqlite"
	"devenv/internal/shared/paths"
)

type Environment struct {
	Kind    string `json:"kind"`
	Version string `json:"version"`
}

type Options struct {
	AllEnvironments bool          `json:"allEnvironments"`
	Environments    []Environment `json:"environments"`
	DatabaseData    bool          `json:"databaseData"`
	UserData        bool          `json:"userData"`
}

var component = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._+-]*$`)

func ReadOptions(path string) (Options, error) {
	var options Options
	content, err := os.ReadFile(path)
	if err != nil {
		return options, err
	}
	err = json.Unmarshal(content, &options)
	return options, err
}

// Plan validates every recursive target before any data is modified.
func Plan(base string, options Options) ([]string, [][2]string, error) {
	base = filepath.Clean(base)
	if !filepath.IsAbs(base) || !strings.EqualFold(filepath.Base(base), paths.APP_FOLDER) {
		return nil, nil, fmt.Errorf("invalid DevEnv data root: %s", base)
	}

	environments := options.Environments
	softwareRoot := filepath.Join(base, paths.SOFTWARE_FOLDER)
	if options.AllEnvironments {
		environments = nil
		kinds, err := os.ReadDir(softwareRoot)
		if err != nil && !os.IsNotExist(err) {
			return nil, nil, err
		}
		for _, kind := range kinds {
			if !kind.IsDir() {
				continue
			}
			versions, err := os.ReadDir(filepath.Join(softwareRoot, kind.Name()))
			if err != nil {
				return nil, nil, err
			}
			for _, version := range versions {
				if version.IsDir() {
					environments = append(environments, Environment{Kind: kind.Name(), Version: version.Name()})
				}
			}
		}
	}

	targets := make([]string, 0)
	records := make([][2]string, 0)
	seen := make(map[[2]string]bool)
	for _, item := range environments {
		if !component.MatchString(item.Kind) || !component.MatchString(item.Version) {
			return nil, nil, fmt.Errorf("invalid environment: %q %q", item.Kind, item.Version)
		}
		key := [2]string{item.Kind, item.Version}
		if seen[key] {
			continue
		}
		seen[key] = true
		targets = append(targets, paths.Software(base, item.Kind, item.Version))
		records = append(records, key)
	}

	serviceRoot := filepath.Join(base, paths.DATA_FOLDER, paths.SERVICE_FOLDER)
	if options.AllEnvironments {
		targets = append(targets, softwareRoot, paths.Nginx(base), serviceRoot)
	}
	if len(targets) == 0 && !options.DatabaseData && !options.UserData {
		return targets, records, nil
	}
	definitions, err := os.ReadDir(serviceRoot)
	if err != nil && !os.IsNotExist(err) {
		return nil, nil, err
	}
	for _, definition := range definitions {
		if len(records) == 0 && !options.AllEnvironments {
			break
		}
		if definition.IsDir() || filepath.Ext(definition.Name()) != ".json" {
			continue
		}
		definitionPath := filepath.Join(serviceRoot, definition.Name())
		content, err := os.ReadFile(definitionPath)
		if err != nil {
			return nil, nil, err
		}
		var service struct {
			Binary string `json:"binary"`
		}
		if err := json.Unmarshal(content, &service); err != nil {
			return nil, nil, err
		}
		for _, target := range targets {
			prefix := filepath.Clean(target) + string(filepath.Separator)
			if strings.HasPrefix(strings.ToLower(filepath.Clean(service.Binary)), strings.ToLower(prefix)) {
				targets = append(targets, definitionPath)
				break
			}
		}
	}
	if options.DatabaseData {
		targets = append(targets, filepath.Join(base, paths.DBDATA_FOLDER))
	}
	if options.UserData {
		for _, relative := range []string{
			paths.LOG_FOLDER,
			paths.CACHE_FOLDER,
			filepath.Join(paths.DATA_FOLDER, paths.BACKUP_FOLDER),
			filepath.Join(paths.DATA_FOLDER, paths.NGINX_FOLDER),
			filepath.Join(paths.DATA_FOLDER, paths.SERVICE_FOLDER),
		} {
			targets = append(targets, filepath.Join(base, relative))
		}
	}

	for _, target := range append(append([]string{}, targets...), base) {
		if err := paths.CheckWithin(base, target); err != nil {
			return nil, nil, err
		}
		for current := target; ; current = filepath.Dir(current) {
			info, err := os.Lstat(current)
			if err != nil && !os.IsNotExist(err) {
				return nil, nil, err
			}
			if err == nil && info.Mode()&os.ModeSymlink != 0 {
				return nil, nil, fmt.Errorf("refusing linked path: %s", current)
			}
			if current == base {
				break
			}
		}
		if err := filepath.WalkDir(target, func(path string, entry fs.DirEntry, err error) error {
			if os.IsNotExist(err) {
				return nil
			}
			if err != nil {
				return err
			}
			if entry.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("refusing linked path: %s", path)
			}
			return nil
		}); err != nil {
			return nil, nil, err
		}
	}
	return targets, records, nil
}

func Clean(base string, options Options) error {
	targets, records, err := Plan(base, options)
	if err != nil {
		return err
	}
	if len(records) > 0 || options.AllEnvironments || options.UserData {
		if err := sqlite.CleanUninstallRecords(paths.Database(base), records, options.AllEnvironments, options.UserData); err != nil {
			return err
		}
	}
	for _, target := range targets {
		if err := os.RemoveAll(target); err != nil {
			return fmt.Errorf("remove %s: %w", target, err)
		}
	}
	return nil
}
