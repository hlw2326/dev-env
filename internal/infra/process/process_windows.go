//go:build windows

package process

import (
	"path/filepath"

	"golang.org/x/sys/windows"
)

const processPathBufferSize = 32768

// ExecutablePath returns the executable path for a live Windows process.
func ExecutablePath(pid int) string {
	if pid <= 0 {
		return ""
	}
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return ""
	}
	defer func() { _ = windows.CloseHandle(handle) }()

	buffer := make([]uint16, processPathBufferSize)
	size := uint32(len(buffer))
	if err := windows.QueryFullProcessImageName(handle, 0, &buffer[0], &size); err != nil {
		return ""
	}
	return filepath.Clean(windows.UTF16ToString(buffer[:size]))
}
