package tray

// Callbacks contains the actions exposed by the application tray menu.
type Callbacks struct {
	Show func()
	Hide func()
	Quit func()
}

// Manager controls the process tray icon and its native menu.
type Manager interface {
	Start(Callbacks)
	Stop()
}
