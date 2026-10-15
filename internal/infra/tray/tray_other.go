//go:build !windows

package tray

type manager struct{}

func NewManager() Manager {
	return &manager{}
}

func (*manager) Start(Callbacks) {}

func (*manager) Stop() {}
