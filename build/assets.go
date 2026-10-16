// Package build provides the shared application branding assets.
package build

import _ "embed"

// Icon is shared by the Windows executable, installer, and system tray.
//
//go:embed windows/icon.ico
var Icon []byte
