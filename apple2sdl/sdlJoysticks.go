package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

/*
  Apple 2 supports four paddles and 3 pushbuttons. The first two paddles are
the X, Y axis of the first joystick. The second two correspond the the second
joystick.
  Button 0 is the primary button of joystick 0.
  Button 1 is the secondary button of joystick 0 but also the primary button of
joystick 1.
  Button 2 is the secondary button of Joystick 1.
*/

type sdlJoysticks struct {
	paddle [4]uint8
	button [4]bool
}

func newSDLJoysticks() *sdlJoysticks {
	var j sdlJoysticks

	err := sdl.InitSubSystem(sdl.INIT_JOYSTICK)
	if err != nil {
		panic(err)
	}

	// Init up to two joysticks
	sdl.JoystickEventState(sdl.ENABLE)
	joyCount := sdl.NumJoysticks()
	for iJoy := 0; iJoy < joyCount && iJoy < 2; iJoy++ {
		/*joystick := */ sdl.JoystickOpen(iJoy)
	}

	// Initialize to maximum resistance if unpugged
	j.paddle[0] = 255
	j.paddle[1] = 255
	j.paddle[2] = 255
	j.paddle[3] = 255

	return &j
}

func (j *sdlJoysticks) putAxisEvent(e *sdl.JoyAxisEvent) {
	if e.Which >= 2 || e.Axis >= 2 {
		// Process only the first two axis of the first two joysticks
		return
	}

	j.paddle[uint8(e.Which)*2+e.Axis] = uint8((e.Value >> 8) + 128)
}

func (j *sdlJoysticks) putButtonEvent(e *sdl.JoyButtonEvent) {
	if e.Which >= 2 || e.Button >= 2 {
		// Process only the first two buttons of the first two joysticks
		return
	}

	j.button[uint8(e.Which)*2+e.Button] = (e.State != 0)
}

func (j *sdlJoysticks) ReadButton(i int) bool {
	switch i {
	case 0:
		return j.button[0]
	case 1:
		// It can be secondary of first or primary of second
		return j.button[1] || j.button[2]
	case 2:
		return j.button[3]
	default:
		return false
	}
}

func (j *sdlJoysticks) ReadPaddle(i int) uint8 {
	return j.paddle[i]
}
