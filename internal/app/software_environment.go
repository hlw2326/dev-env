package app

import (
	"context"
	"slices"
	"strings"
	"sync"
	"time"

	"devenv/internal/domain/software"
	appErrors "devenv/internal/shared/errors"
	"devenv/internal/shared/paths"
)

func (s *SoftwareService) InspectEnvironment(ctx context.Context, kind, executable string) (software.CommandEnvironment, error) {
	result := software.CommandEnvironment{Kind: kind, Name: software.Name(kind), Entries: []software.CommandLocation{}, Warnings: []string{}}
	probes := software.InspectionProbes(kind)
	if len(probes) == 0 {
		return result, appErrors.New(appErrors.INVALID_ARGUMENT, "当前软件不支持环境检测")
	}
	result.Command = probes[0].Name
	item, err := s.repo.GetMain(ctx, kind)
	if err != nil {
		return result, err
	}
	if item != nil && item.State == software.STATE_INSTALLED {
		result.DefaultVersion = item.Version
	}
	scan, err := s.commands.Inspect(ctx, kind, executable)
	if err != nil {
		return result, appErrors.Wrap(err, appErrors.PATH_NOT_ACCESSIBLE, "检测系统已有环境失败")
	}
	result.Warnings = append(result.Warnings, scan.Warnings...)
	activeFound := false
	exposed := false
	for _, definition := range software.GlobalCommandDefinitions(kind) {
		if definition.Name == result.Command {
			exposed = true
		}
	}
	for _, entry := range scan.Entries {
		entry.Managed = paths.CheckWithin(paths.Commands(s.base), entry.Path) == nil ||
			paths.CheckWithin(paths.Software(s.base, kind, ""), entry.Path) == nil
		if entry.Active && entry.Command == result.Command && entry.Role == probes[0].Role {
			activeFound = true
			if result.DefaultVersion != "" && exposed {
				result.Conflict = !entry.Managed || !slices.Contains(entry.Versions, result.DefaultVersion) || entry.Error != ""
			}
		}
		result.Entries = append(result.Entries, entry)
	}
	if executable == "" && !activeFound && exposed && result.DefaultVersion != "" {
		result.Conflict = true
	}
	return result, nil
}

const (
	ENVIRONMENT_SCAN_TIMEOUT = 30 * time.Second
	ENVIRONMENT_SCAN_WORKERS = 3
)

// ScanEnvironments is invoked by an explicit user action, never by market listing.
func (s *SoftwareService) ScanEnvironments(ctx context.Context, kind, executable string) ([]software.CommandEnvironment, error) {
	if kind != "" {
		item, err := s.InspectEnvironment(ctx, kind, executable)
		if err != nil {
			return []software.CommandEnvironment{}, err
		}
		return []software.CommandEnvironment{item}, nil
	}
	if executable != "" {
		return []software.CommandEnvironment{}, appErrors.New(appErrors.INVALID_ARGUMENT, "选择程序前请指定软件类型")
	}
	ctx, cancel := context.WithTimeout(ctx, ENVIRONMENT_SCAN_TIMEOUT)
	defer cancel()
	kinds := software.AllKinds()
	result := make([]software.CommandEnvironment, len(kinds))
	jobs := make(chan int, len(kinds))
	for index := range kinds {
		jobs <- index
	}
	close(jobs)
	var workers sync.WaitGroup
	for worker := 0; worker < ENVIRONMENT_SCAN_WORKERS; worker++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for index := range jobs {
				if ctx.Err() != nil {
					return
				}
				item, err := s.InspectEnvironment(ctx, kinds[index], "")
				if err != nil {
					item.Error = err.Error()
				}
				result[index] = item
			}
		}()
	}
	workers.Wait()
	if ctx.Err() != nil {
		return []software.CommandEnvironment{}, ctx.Err()
	}
	failures := make([]string, 0)
	failed := 0
	for _, item := range result {
		if item.Error == "" {
			continue
		}
		failed++
		if !slices.Contains(failures, item.Error) {
			failures = append(failures, item.Error)
		}
	}
	if failed == len(result) {
		return []software.CommandEnvironment{}, appErrors.New(appErrors.PATH_NOT_ACCESSIBLE, strings.Join(failures, "\n"))
	}
	return result, nil
}
