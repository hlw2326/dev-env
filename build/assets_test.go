package build

import (
	"bytes"
	"encoding/binary"
	"image/png"
	"os"
	"testing"
)

func TestSharedIcon(t *testing.T) {
	if len(Icon) < 6 || binary.LittleEndian.Uint16(Icon[2:4]) != 1 {
		t.Fatal("invalid application ICO")
	}
	count := int(binary.LittleEndian.Uint16(Icon[4:6]))
	if count != 9 || len(Icon) < 6+16*count {
		t.Fatal("missing application icon sizes")
	}
	application, err := os.ReadFile("appicon.png")
	if err != nil {
		t.Fatal(err)
	}
	for index := 0; index < count; index++ {
		entry := Icon[6+16*index : 6+16*(index+1)]
		size := int(binary.LittleEndian.Uint32(entry[8:12]))
		offset := int(binary.LittleEndian.Uint32(entry[12:16]))
		if offset < 6+16*count || size <= 0 || offset+size > len(Icon) {
			t.Fatal("invalid icon image bounds")
		}
		content := Icon[offset : offset+size]
		config, err := png.DecodeConfig(bytes.NewReader(content))
		if err != nil {
			t.Fatal(err)
		}
		width := int(entry[0])
		if width == 0 {
			width = 256
		}
		if config.Width != width || config.Height != width {
			t.Fatal("icon dimensions do not match directory")
		}
		if width == 256 && !bytes.Equal(application, content) {
			t.Fatal("application logo and tray icon differ; run generate-icon.ps1")
		}
	}
}
