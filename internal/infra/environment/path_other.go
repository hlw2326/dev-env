//go:build !windows

package environment

import (
	"errors"

	"devenv/internal/ports"
)

type userPathManager struct{}

func NewUserPathManager() ports.UserPathManager {
	return userPathManager{}
}

func (userPathManager) Ensure(string) error {
	return errors.New("user PATH binding is only available on Windows")
}

func (userPathManager) Remove(string) error {
	return errors.New("user PATH binding is only available on Windows")
}
