//go:build !js

package main

import (
	"github.com/ivanizag/izapple2/frontend/shared"

	"github.com/Zyko0/go-sdl3/sdl"
)

// sdl3Joysticks turns the joystick, mouse and apple key events of SDL3 into
// the paddles and buttons the machine reads
type sdl3Joysticks struct {
	paddles *shared.Paddles

	// SDL3 identifies joysticks with instance ids that are not indexes:
	// they start at 1 and grow as devices are plugged. The events carry
	// the instance id, so we keep the slot each one was assigned to.
	slots map[sdl.JoystickID]int
}

func newSDL3Joysticks(useMouseAlt bool) *sdl3Joysticks {
	var j sdl3Joysticks
	j.slots = make(map[sdl.JoystickID]int)

	sdl.SetJoystickEventsEnabled(true)

	// Init up to two joysticks
	ids, err := sdl.GetJoysticks()
	if err == nil {
		for _, id := range ids {
			if len(j.slots) >= 2 {
				break
			}
			_, err := id.OpenJoystick()
			if err != nil {
				continue
			}
			j.slots[id] = len(j.slots)
		}
	}

	j.paddles = shared.NewPaddles(len(j.slots), useMouseAlt)
	return &j
}

func (j *sdl3Joysticks) putAxisEvent(e *sdl.JoyAxisEvent) {
	j.paddles.SetAxis(j.slotOf(e.Which), int(e.Axis), e.Value)
}

func (j *sdl3Joysticks) putButtonEvent(e *sdl.JoyButtonEvent) {
	j.paddles.SetButton(j.slotOf(e.Which), int(e.Button), e.Down)
}

func (j *sdl3Joysticks) putMouseMotionEvent(e *sdl.MouseMotionEvent, width int32, height int32) {
	j.paddles.SetMousePosition(int(e.X), int(e.Y), int(width), int(height))
}

func (j *sdl3Joysticks) putMouseButtonEvent(e *sdl.MouseButtonEvent) {
	j.paddles.SetMouseButton(sdl3MouseButton(e.Button), e.Down)
}

func (j *sdl3Joysticks) putKey(keyEvent *sdl.KeyboardEvent) {
	/*
		We will simultate joystick buttons with keyboard keys.
		Actually the Apple//e does this with the open and solid apple keys.
		   Alt key - button 0 - Open apple
		   AltGr key - button 1- Solid apple

		The apple keys are a place on the keyboard, not a symbol, so we look at
		the scancode. Unlike SDL2, SDL3 resolves the keycode through the current
		keymap: on layouts where the left alt key is Meta_L instead of Alt_L the
		keycode is not K_LALT and the open apple would never be pressed.
	*/
	switch keyEvent.Scancode {
	case sdl.SCANCODE_LALT:
		j.paddles.SetOpenApple(keyEvent.Down)
	case sdl.SCANCODE_RALT:
		j.paddles.SetClosedApple(keyEvent.Down)
	}
}

// slotOf returns the joystick an instance id was assigned to, -1 for the ones
// left out
func (j *sdl3Joysticks) slotOf(id sdl.JoystickID) int {
	slot, ok := j.slots[id]
	if !ok {
		return -1
	}

	return slot
}

// sdl3MouseButton translates a button of the mouse of SDL3
func sdl3MouseButton(button uint8) int {
	switch sdl.MouseButtonFlags(button) {
	case sdl.BUTTON_LEFT:
		return shared.MouseButtonLeft
	case sdl.BUTTON_RIGHT:
		return shared.MouseButtonRight
	case sdl.BUTTON_MIDDLE:
		return shared.MouseButtonMiddle
	}

	return -1
}
