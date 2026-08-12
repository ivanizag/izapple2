package shared

import "testing"

func TestMouse(t *testing.T) {
	m := NewMouse()

	m.SetPosition(320, 120, 640, 480)
	x, y, pressed := m.ReadMouse()
	if x != 32768 || y != 16384 || pressed {
		t.Errorf("the mouse is at (%v, %v, %v), expected (32768, 16384, false)",
			x, y, pressed)
	}

	m.SetButton(true)
	if _, _, pressed := m.ReadMouse(); !pressed {
		t.Error("the button of the mouse should be pressed")
	}

	// A window not sized yet leaves the pointer where it was
	m.SetPosition(10, 10, 0, 0)
	if x, y, _ := m.ReadMouse(); x != 32768 || y != 16384 {
		t.Errorf("the mouse moved to (%v, %v) on a window without size", x, y)
	}

	// The pointer outside the window stays on the edge
	m.SetPosition(-10, 999, 640, 480)
	if x, y, _ := m.ReadMouse(); x != 0 || y != 65535 {
		t.Errorf("the mouse is at (%v, %v), expected (0, 65535)", x, y)
	}
}
