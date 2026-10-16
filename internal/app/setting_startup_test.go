package app

import (
	"errors"
	"testing"
)

type startupStub struct {
	enabled bool
	failure error
}

func (s *startupStub) Enabled() (bool, error) { return s.enabled, s.failure }
func (s *startupStub) Set(enabled bool) error {
	if s.failure != nil {
		return s.failure
	}
	s.enabled = enabled
	return nil
}

func TestSettingApplicationStartup(t *testing.T) {
	startup := &startupStub{}
	service := &SettingService{startup: startup}
	for _, desired := range []bool{true, false} {
		if err := service.SetStartupEnabled(desired); err != nil {
			t.Fatal(err)
		}
		actual, err := service.StartupEnabled()
		if err != nil || actual != desired {
			t.Fatalf("enabled=%v error=%v", actual, err)
		}
	}
	startup.failure = errors.New("access denied")
	if err := service.SetStartupEnabled(true); !errors.Is(err, startup.failure) {
		t.Fatalf("lost startup error: %v", err)
	}
}
