//go:build windows

package tray

import (
	"log"
	"sync"

	"devenv/build"

	"github.com/energye/systray"
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
	systray.SetOnClick(func(menu systray.IMenu) { go m.onClick(menu) })
	systray.SetOnRClick(m.onRightClick)

	show := systray.AddMenuItem("显示", "显示 DevEnv 主窗口")
	hide := systray.AddMenuItem("隐藏", "隐藏 DevEnv 主窗口")
	systray.AddSeparator()
	quit := systray.AddMenuItem("退出", "退出 DevEnv 并停止应用进程")

	show.Click(func() { go m.show() })
	hide.Click(func() { go m.hide() })
	quit.Click(func() { go m.quit() })
}

func (m *manager) onClick(_ systray.IMenu) { m.show() }

func (m *manager) onRightClick(menu systray.IMenu) {
	if err := menu.ShowMenu(); err != nil {
		log.Printf("打开托盘菜单失败: %v", err)
	}
}

func (m *manager) show() {
	if m.callbacks.Show != nil {
		m.callbacks.Show()
	}
}

func (m *manager) hide() {
	if m.callbacks.Hide != nil {
		m.callbacks.Hide()
	}
}

func (m *manager) quit() {
	if m.callbacks.Quit != nil {
		m.callbacks.Quit()
	}
}

func (m *manager) Stop() {
	if !m.started {
		return
	}
	m.stopOnce.Do(systray.Quit)
}
