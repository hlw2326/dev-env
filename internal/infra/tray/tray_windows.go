//go:build windows

package tray

import (
	"encoding/binary"
	"sync"

	"github.com/getlantern/systray"
)

type manager struct {
	callbacks Callbacks
	stopOnce  sync.Once
	started   bool
}

func NewManager() Manager {
	return &manager{}
}

func (m *manager) Start(callbacks Callbacks) {
	m.callbacks = callbacks
	m.started = true
	systray.Register(m.ready, func() {})
}

func (m *manager) ready() {
	systray.SetIcon(iconData())
	systray.SetTooltip("DevEnv 本地开发环境")

	show := systray.AddMenuItem("显示 DevEnv", "显示 DevEnv 主窗口")
	hide := systray.AddMenuItem("隐藏窗口", "隐藏 DevEnv 主窗口")
	systray.AddSeparator()
	quit := systray.AddMenuItem("退出 DevEnv", "退出 DevEnv 并停止应用进程")

	go func() {
		for {
			select {
			case <-show.ClickedCh:
				if m.callbacks.Show != nil {
					m.callbacks.Show()
				}
			case <-hide.ClickedCh:
				if m.callbacks.Hide != nil {
					m.callbacks.Hide()
				}
			case <-quit.ClickedCh:
				if m.callbacks.Quit != nil {
					m.callbacks.Quit()
				}
				return
			}
		}
	}()
}

func (m *manager) Stop() {
	if !m.started {
		return
	}
	m.stopOnce.Do(systray.Quit)
}

// iconData returns a small self-contained ICO so the tray does not depend on
// a runtime file path or on the frontend asset bundle.
func iconData() []byte {
	const (
		width      = 16
		height     = 16
		directory  = 6
		entry      = 16
		dib        = 40
		pixelBytes = width * height * 4
		maskBytes  = ((width + 31) / 32) * 4 * height
	)
	data := make([]byte, directory+entry+dib+pixelBytes+maskBytes)
	binary.LittleEndian.PutUint16(data[0:2], 0)
	binary.LittleEndian.PutUint16(data[2:4], 1)
	binary.LittleEndian.PutUint16(data[4:6], 1)

	data[6] = width
	data[7] = height
	binary.LittleEndian.PutUint16(data[10:12], 1)
	binary.LittleEndian.PutUint16(data[12:14], 32)
	binary.LittleEndian.PutUint32(data[14:18], dib+pixelBytes+maskBytes)
	binary.LittleEndian.PutUint32(data[18:22], directory+entry)

	binary.LittleEndian.PutUint32(data[22:26], dib)
	binary.LittleEndian.PutUint32(data[26:30], width)
	binary.LittleEndian.PutUint32(data[30:34], height*2)
	binary.LittleEndian.PutUint16(data[34:36], 1)
	binary.LittleEndian.PutUint16(data[36:38], 32)
	binary.LittleEndian.PutUint32(data[42:46], pixelBytes)

	pixels := data[directory+entry+dib : directory+entry+dib+pixelBytes]
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			red, green, blue := byte(0), byte(167), byte(123)
			if isLogoMark(x, y) {
				red, green, blue = 255, 255, 255
			}
			destinationY := height - 1 - y
			index := (destinationY*width + x) * 4
			pixels[index] = blue
			pixels[index+1] = green
			pixels[index+2] = red
			pixels[index+3] = 255
		}
	}
	return data
}

func isLogoMark(x, y int) bool {
	if x == 5 && y >= 4 && y <= 11 {
		return true
	}
	if (y == 4 || y == 11) && x >= 6 && x <= 9 {
		return true
	}
	return x == 10 && y >= 5 && y <= 10
}
