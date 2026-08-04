//go:build !js

package main

import (
	"github.com/Zyko0/go-sdl3/sdl"
)

type sdl3Mouse struct {
	x       uint16
	y       uint16
	pressed bool
}

func newSDL3Mouse() *sdl3Mouse {
	var m sdl3Mouse
	return &m
}

func (m *sdl3Mouse) putMouseMotionEvent(e *sdl.MouseMotionEvent, width int32, height int32) {
	m.x = uint16(65536 * int32(e.X) / width)
	m.y = uint16(65536 * int32(e.Y) / height)
}

func (m *sdl3Mouse) putMouseButtonEvent(e *sdl.MouseButtonEvent) {
	if e.Button == 1 { // BUTTTON_LEFT
		m.pressed = e.Down
	}
}

func (m *sdl3Mouse) ReadMouse() (x uint16, y uint16, pressed bool) {
	return m.x, m.y, m.pressed
}

// TODO: SDL_WarpMouseInWindow
