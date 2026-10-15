//go:build !windows

package environment

import (
	"context"
	"errors"

	"devenv/internal/domain/software"
	"devenv/internal/ports"
)

type commandInspector struct{}

func NewCommandInspector() ports.CommandInspector { return commandInspector{} }

func (commandInspector) Inspect(context.Context, string, string) (software.EnvironmentScan, error) {
	return software.EnvironmentScan{Entries: []software.CommandLocation{}, Warnings: []string{}}, errors.New("全局命令检测仅支持 Windows")
}
