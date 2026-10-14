//go:build !windows

package autostart

func Set(_ bool) error {
	return nil
}

func Enabled() bool {
	return false
}

func SetService(_ string, _ bool, _ string, _ []string, _ string) error {
	return nil
}

func ServiceEnabled(_ string) bool {
	return false
}
