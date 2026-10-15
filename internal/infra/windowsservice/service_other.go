//go:build !windows

package windowsservice

import "errors"

func Run(string) error {
	return errors.New("Windows service mode is only available on Windows")
}
