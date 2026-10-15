package windowsservice

import (
	"io"
	"os"
	"path/filepath"

	"devenv/internal/shared/paths"
)

const serviceDefinitionArgument = "--devenv-service-config"

// ConfigPathFromArguments returns the service definition path when the
// executable was launched by the Windows Service Control Manager.
func ConfigPathFromArguments(arguments []string) (string, bool) {
	for index := 0; index < len(arguments)-1; index++ {
		if arguments[index] == serviceDefinitionArgument && arguments[index+1] != "" {
			return arguments[index+1], true
		}
	}
	return "", false
}

// ServiceExecutable returns the stable service host shared by installed and
// development builds. It is created once from the currently running program.
func ServiceExecutable(base string) (string, error) {
	target := paths.ServiceHost(base)
	if info, err := os.Stat(target); err == nil && !info.IsDir() {
		return target, nil
	} else if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	source, err := os.Executable()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return "", err
	}
	input, err := os.Open(source)
	if err != nil {
		return "", err
	}
	defer func() { _ = input.Close() }()

	temporaryPath := target + ".tmp"
	output, err := os.OpenFile(temporaryPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(output, input); err != nil {
		_ = output.Close()
		_ = os.Remove(temporaryPath)
		return "", err
	}
	if err := output.Close(); err != nil {
		_ = os.Remove(temporaryPath)
		return "", err
	}
	if err := os.Rename(temporaryPath, target); err != nil {
		_ = os.Remove(temporaryPath)
		return "", err
	}
	return target, nil
}
