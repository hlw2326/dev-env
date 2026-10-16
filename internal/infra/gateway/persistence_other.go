//go:build !windows

package gateway

import "devenv/internal/ports"

type persistentService struct {
	name      string
	options   ports.Options
	autoStart bool
}

type serviceAccess struct {
	path       string
	permission string
	direct     bool
}

func configurePersistentService(string, persistentService) error { return nil }
func startPersistentService(string) error                        { return nil }
func stopPersistentService(string) error                         { return nil }
func persistentServiceExists(string) bool                        { return false }
func persistentServiceOwnsProcess(string, int, string) bool      { return false }
