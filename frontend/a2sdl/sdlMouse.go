package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

type sdlMouse struct {
	x       uint16
	y       uint16
	pressed bool
}

func newSDLMouse() *sdlMouse {
	var m sdlMouse
	return &m
}

func (m *sdlMouse) putMouseMotionEvent(e *sdl.MouseMotionEvent, width int32, height int32) {
	m.x = uint16(65536 * e.X / width)
	m.y = uint16(65536 * e.Y / height)
}

func (m *sdlMouse) putMouseButtonEvent(e *sdl.MouseButtonEvent) {
	if e.Button == 1 { // BUTTTON_LEFT
		m.pressed = e.State == sdl.PRESSED
	}
}

func (m *sdlMouse) ReadMouse() (x uint16, y uint16, pressed bool) {
	return m.x, m.y, m.pressed
}

// TODO: SDL_WarpMouseInWindow
