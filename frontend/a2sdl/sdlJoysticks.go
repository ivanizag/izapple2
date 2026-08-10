//go:build !js

package main

import (
	"github.com/ivanizag/izapple2/frontend/shared"
	"github.com/veandco/go-sdl2/sdl"
)

// sdlJoysticks turns the joystick, mouse and apple key events of SDL2 into the
// paddles and buttons the machine reads
type sdlJoysticks struct {
	paddles *shared.Paddles
}

func newSDLJoysticks(useMouseAlt bool) *sdlJoysticks {
	err := sdl.InitSubSystem(sdl.INIT_JOYSTICK)
	if err != nil {
		panic(err)
	}

	// Init up to two joysticks
	sdl.JoystickEventState(sdl.ENABLE)
	joyCount := min(sdl.NumJoysticks(), 2)
	for iJoy := range joyCount {
		sdl.JoystickOpen(iJoy)
	}

	var j sdlJoysticks
	j.paddles = shared.NewPaddles(joyCount, useMouseAlt)
	return &j
}

func (j *sdlJoysticks) putAxisEvent(e *sdl.JoyAxisEvent) {
	j.paddles.SetAxis(int(e.Which), int(e.Axis), e.Value)
}

func (j *sdlJoysticks) putButtonEvent(e *sdl.JoyButtonEvent) {
	j.paddles.SetButton(int(e.Which), int(e.Button), e.State != 0)
}

func (j *sdlJoysticks) putMouseMotionEvent(e *sdl.MouseMotionEvent, width int32, height int32) {
	j.paddles.SetMousePosition(int(e.X), int(e.Y), int(width), int(height))
}

func (j *sdlJoysticks) putMouseButtonEvent(e *sdl.MouseButtonEvent) {
	j.paddles.SetMouseButton(sdlMouseButton(e.Button), e.State == sdl.PRESSED)
}

func (j *sdlJoysticks) putKey(keyEvent *sdl.KeyboardEvent) {
	/*
		We will simultate joystick buttons with keyboard keys.
		Actually the Apple//e does this with the open and solid apple keys.
		   Alt key - button 0 - Open apple
		   AltGr key - button 1- Solid apple

		The apple keys are a place on the keyboard, not a symbol, so we look at
		the scancode. The keycode is resolved through the current keymap: on
		layouts where the left alt key is Meta_L instead of Alt_L it is not
		K_LALT and the open apple would never be pressed.
	*/
	pressed := keyEvent.Type == sdl.KEYDOWN
	switch keyEvent.Keysym.Scancode {
	case sdl.SCANCODE_LALT:
		j.paddles.SetOpenApple(pressed)
	case sdl.SCANCODE_RALT:
		j.paddles.SetClosedApple(pressed)
	}
}

// sdlMouseButton translates a button of the mouse of SDL2
func sdlMouseButton(button uint8) int {
	switch button {
	case sdl.BUTTON_LEFT:
		return shared.MouseButtonLeft
	case sdl.BUTTON_RIGHT:
		return shared.MouseButtonRight
	case sdl.BUTTON_MIDDLE:
		return shared.MouseButtonMiddle
	}

	return -1
}
