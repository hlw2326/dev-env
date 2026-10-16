//go:build windows

package tray

import (
	"sync"

	"devenv/build"

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
	systray.SetIcon(build.Icon)
	systray.SetTooltip("DevEnv 本地开发环境")

	show := systray.AddMenuItem("显示", "显示 DevEnv 主窗口")
	hide := systray.AddMenuItem("隐藏", "隐藏 DevEnv 主窗口")
	systray.AddSeparator()
	quit := systray.AddMenuItem("退出", "退出 DevEnv 并停止应用进程")

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
