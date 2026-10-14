//go:build !windows

package autostart

func Set(_ bool) error {
	return nil
}

func Enabled() bool {
	return false
}
