//go:build !windows

package deployment

import "os/exec"

func prepareCommand(command *exec.Cmd) {}
