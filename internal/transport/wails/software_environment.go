package wails

import "devenv/internal/domain/software"

type CommandLocationDTO struct {
	Path         string   `json:"path"`
	Command      string   `json:"command"`
	Role         string   `json:"role"`
	Source       string   `json:"source"`
	Versions     []string `json:"versions"`
	Error        string   `json:"error"`
	Managed      bool     `json:"managed"`
	Active       bool     `json:"active"`
	ServiceName  string   `json:"serviceName"`
	ServiceState string   `json:"serviceState"`
}

type CommandEnvironmentDTO struct {
	Kind           string               `json:"kind"`
	Name           string               `json:"name"`
	Error          string               `json:"error"`
	Command        string               `json:"command"`
	Entries        []CommandLocationDTO `json:"entries"`
	DefaultVersion string               `json:"defaultVersion"`
	Conflict       bool                 `json:"conflict"`
	Warnings       []string             `json:"warnings"`
}

func (a *SoftwareAPI) ScanEnvironments(kind, executable string) ([]CommandEnvironmentDTO, error) {
	items, err := a.service.ScanEnvironments(a.ctx, kind, executable)
	result := make([]CommandEnvironmentDTO, 0, len(items))
	if err != nil {
		return result, err
	}
	for _, item := range items {
		result = append(result, convertEnvironment(item))
	}
	return result, nil
}

func convertEnvironment(item software.CommandEnvironment) CommandEnvironmentDTO {
	result := CommandEnvironmentDTO{Entries: []CommandLocationDTO{}, Warnings: []string{}}
	result.Kind, result.Name, result.Error = item.Kind, item.Name, item.Error
	result.Command = item.Command
	result.DefaultVersion = item.DefaultVersion
	result.Conflict = item.Conflict
	result.Warnings = append(result.Warnings, item.Warnings...)
	for _, entry := range item.Entries {
		result.Entries = append(result.Entries, CommandLocationDTO{
			Path: entry.Path, Command: entry.Command, Role: entry.Role, Source: entry.Source,
			Versions: append([]string{}, entry.Versions...), Error: entry.Error, Managed: entry.Managed, Active: entry.Active,
			ServiceName: entry.ServiceName, ServiceState: entry.ServiceState,
		})
	}
	return result
}
