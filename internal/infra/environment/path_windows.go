//go:build windows

package environment

import (
	"fmt"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"devenv/internal/ports"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const (
	userEnvironmentKey = `Environment`
	hwndBroadcast      = 0xffff
	wmSettingChange    = 0x001a
	smtoAbortIfHung    = 0x0002
)

type userPathManager struct{}

func NewUserPathManager() ports.UserPathManager {
	return userPathManager{}
}

func (userPathManager) Ensure(path string) error {
	target := filepath.Clean(strings.TrimSpace(path))
	if target == "" {
		return fmt.Errorf("PATH entry is empty")
	}

	key, _, err := registry.CreateKey(registry.CURRENT_USER, userEnvironmentKey, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("open user environment registry: %w", err)
	}
	defer key.Close()

	current, valueType, err := readPathValue(key)
	if err != nil {
		return err
	}
	entries := splitPath(current)
	filtered := make([]string, 0, len(entries)+1)
	for _, entry := range entries {
		if !samePath(entry, target) {
			filtered = append(filtered, entry)
		}
	}
	// Prefer DevEnv within the user PATH. Machine PATH entries can still win;
	// command inspection reports that precedence without modifying them.
	updated := append([]string{target}, filtered...)
	if err := writePathValue(key, strings.Join(updated, ";"), valueType); err != nil {
		return err
	}
	return broadcastEnvironmentChange()
}

func (userPathManager) Remove(path string) error {
	target := filepath.Clean(strings.TrimSpace(path))
	if target == "" {
		return fmt.Errorf("PATH entry is empty")
	}

	key, err := registry.OpenKey(registry.CURRENT_USER, userEnvironmentKey, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		if err == registry.ErrNotExist {
			return nil
		}
		return fmt.Errorf("open user environment registry: %w", err)
	}
	defer key.Close()

	current, valueType, err := readPathValue(key)
	if err != nil {
		return err
	}
	entries := splitPath(current)
	updated := make([]string, 0, len(entries))
	changed := false
	for _, entry := range entries {
		if samePath(entry, target) {
			changed = true
			continue
		}
		updated = append(updated, entry)
	}
	if !changed {
		return nil
	}
	if err := writePathValue(key, strings.Join(updated, ";"), valueType); err != nil {
		return err
	}
	return broadcastEnvironmentChange()
}

func readPathValue(key registry.Key) (string, uint32, error) {
	value, valueType, err := key.GetStringValue("Path")
	if err != nil {
		if err == registry.ErrNotExist {
			return "", registry.SZ, nil
		}
		return "", 0, fmt.Errorf("read user PATH from registry: %w", err)
	}
	return value, valueType, nil
}

func writePathValue(key registry.Key, value string, valueType uint32) error {
	if valueType == registry.EXPAND_SZ {
		if err := key.SetExpandStringValue("Path", value); err != nil {
			return fmt.Errorf("write user PATH to registry: %w", err)
		}
		return nil
	}
	if err := key.SetStringValue("Path", value); err != nil {
		return fmt.Errorf("write user PATH to registry: %w", err)
	}
	return nil
}

func splitPath(value string) []string {
	parts := strings.Split(value, ";")
	entries := make([]string, 0, len(parts))
	for _, part := range parts {
		entry := strings.TrimSpace(part)
		if entry != "" {
			entries = append(entries, entry)
		}
	}
	return entries
}

func samePath(left, right string) bool {
	return strings.EqualFold(filepath.Clean(strings.TrimSpace(left)), filepath.Clean(strings.TrimSpace(right)))
}

func broadcastEnvironmentChange() error {
	user32 := windows.NewLazySystemDLL("user32.dll")
	sendMessage := user32.NewProc("SendMessageTimeoutW")
	environment := windows.StringToUTF16Ptr("Environment")
	var result uintptr
	ret, _, callErr := sendMessage.Call(
		hwndBroadcast,
		wmSettingChange,
		0,
		uintptr(unsafe.Pointer(environment)),
		smtoAbortIfHung,
		5000,
		uintptr(unsafe.Pointer(&result)),
	)
	if ret == 0 {
		if callErr == syscall.Errno(0) {
			return fmt.Errorf("broadcast environment change failed")
		}
		return fmt.Errorf("broadcast environment change failed: %w", callErr)
	}
	return nil
}
