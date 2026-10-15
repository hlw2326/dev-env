package windowsservice

import "os"

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

// CurrentExecutable returns the executable that should host a registered
// service. The service manager starts this same binary with a definition path.
func CurrentExecutable() (string, error) {
	return os.Executable()
}
