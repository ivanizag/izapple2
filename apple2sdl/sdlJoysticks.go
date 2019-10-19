package main

import (
	"fmt"

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
	keys   [3]bool
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

	// Initialize to mid resistance if unplugged
	j.paddle[0] = 128
	j.paddle[1] = 128
	j.paddle[2] = 128
	j.paddle[3] = 128

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

func (j *sdlJoysticks) putKey(keyEvent *sdl.KeyboardEvent) {
	/*
		We will simultate joystick buttons with keyboard keys.
		Actually the Apple//e dis this with the open and solid apple keys.
		   Alt key - button 0 - Open apple
		   AltGr key - button 1- Solid apple
		   //Win key - button 2 (Not in the Apple //e keyboard)
	*/
	pressed := keyEvent.Type == sdl.KEYDOWN
	switch keyEvent.Keysym.Sym {
	case sdl.K_LALT:
		j.keys[0] = pressed
	case sdl.K_RALT:
		j.keys[1] = pressed
		//case sdl.K_LGUI:
		//	j.keys[2] = pressed
	}

}

func (j *sdlJoysticks) ReadButton(i int) bool {
	var value bool
	switch i {
	case 0:
		value = j.button[0] || j.keys[0]
	case 1:
		// It can be secondary of first or primary of second
		value = j.button[1] || j.button[2] || j.keys[1]
	case 2:
		value = j.button[3] || j.keys[2]
	}
	fmt.Printf("Button %v: %v.\n", i, value)
	return value
}

func (j *sdlJoysticks) ReadPaddle(i int) uint8 {
	return j.paddle[i]
}
