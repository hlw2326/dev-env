//go:build windows

package tray

import "testing"

type menuStub struct{ shown int }

func (m *menuStub) ShowMenu() error { m.shown++; return nil }

func TestMouseActionsRemainSeparate(t *testing.T) {
	shown, hidden, quit := 0, 0, 0
	m := &manager{callbacks: Callbacks{
		Show: func() { shown++ },
		Hide: func() { hidden++ },
		Quit: func() { quit++ },
	}}
	menu := &menuStub{}
	m.onClick(menu)
	if shown != 1 || menu.shown != 0 || hidden != 0 || quit != 0 {
		t.Fatal("left click must only show the window")
	}
	m.onRightClick(menu)
	if shown != 1 || menu.shown != 1 || hidden != 0 || quit != 0 {
		t.Fatal("right click must only show the menu")
	}
	m.hide()
	m.quit()
	if hidden != 1 || quit != 1 {
		t.Fatal("menu actions not dispatched")
	}
}

func TestUnsetCallbacks(t *testing.T) {
	m := &manager{}
	m.onClick(&menuStub{})
	m.hide()
	m.quit()
}
