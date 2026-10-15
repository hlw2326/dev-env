package autostart

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

const runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`

// Set registers a command to run for the current Windows user at login.
func Set(name, command string) error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("open Windows startup registry: %w", err)
	}
	defer key.Close()
	if err := key.SetStringValue(name, command); err != nil {
		return fmt.Errorf("write Windows startup registry: %w", err)
	}
	return nil
}

// Remove removes a command from the current user's Windows startup entries.
func Remove(name string) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		if err == registry.ErrNotExist {
			return nil
		}
		return fmt.Errorf("open Windows startup registry: %w", err)
	}
	defer key.Close()
	if err := key.DeleteValue(name); err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("delete Windows startup registry: %w", err)
	}
	return nil
}
