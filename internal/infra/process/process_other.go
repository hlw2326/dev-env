//go:build !windows

package process

// ExecutablePath is only available on Windows because DevEnv manages Windows services.
func ExecutablePath(pid int) string {
	return ""
}
