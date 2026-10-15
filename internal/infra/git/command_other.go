//go:build !windows

package git

import "os/exec"

func prepareCommand(_ *exec.Cmd) {}
