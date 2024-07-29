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
	paddle       [4]uint8
	hasPaddle    [4]bool
	button       [4]bool
	keys         [3]bool
	mousebuttons [3]bool
	useMouse     bool
}

func newSDLJoysticks(useMouseAlt bool) *sdlJoysticks {
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
		j.hasPaddle[iJoy*2] = true
		j.hasPaddle[iJoy*2+1] = true
	}

	// Initialize to max resistance if unplugged
	j.paddle[0] = 255
	j.paddle[1] = 255
	j.paddle[2] = 255
	j.paddle[3] = 255

	if useMouseAlt && !j.hasPaddle[0] {
		// Use the mouse as joystick
		j.useMouse = true
		j.hasPaddle[0] = true
		j.hasPaddle[1] = true
		j.paddle[0] = 127
		j.paddle[1] = 127
	}

	// To enter Apple IIe on self test mode
	// j.keys[1] = true

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
	if e.Which >= 2 {
		// Process only the buttons of the first two joysticks
		return
	}

	j.button[uint8(e.Which)*2+(e.Button%2)] = (e.State != 0)
}

func mouseToJoyCentered(x int32, w int32) uint8 {
	r := x - (w / 2) + 127
	if r >= 255 {
		r = 255
	}
	if r < 0 {
		r = 0
	}
	return uint8(r)

}

func (j *sdlJoysticks) putMouseMotionEvent(e *sdl.MouseMotionEvent, width int32, height int32) {
	if j.useMouse {
		// The mouse moves on all the window
		// j.paddle[0] = mouseToJoyFull(e.X, width)
		// j.paddle[1] = mouseToJoyFull(e.Y, height)

		// The mouse moves around the center of the window
		j.paddle[0] = mouseToJoyCentered(e.X, width)
		j.paddle[1] = mouseToJoyCentered(e.Y, height)
	}
}

func (j *sdlJoysticks) putMouseButtonEvent(e *sdl.MouseButtonEvent) {
	pressed := e.State == sdl.PRESSED
	switch e.Button {
	case 1: // BUTTON_LEFT
		j.mousebuttons[0] = pressed
	case 3: // BUTTON_RIGHT
		j.mousebuttons[1] = pressed
	case 2: // BUTTON_MIDDLE
		j.mousebuttons[2] = pressed
	}
}

func (j *sdlJoysticks) putKey(keyEvent *sdl.KeyboardEvent) {
	/*
		We will simultate joystick buttons with keyboard keys.
		Actually the Apple//e does this with the open and solid apple keys.
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
		// case sdl.K_LGUI:
		//   j.keys[2] = pressed
	}

}

func (j *sdlJoysticks) ReadButton(i int) bool {
	var value bool
	switch i {
	case 0:
		value = j.button[0] || j.keys[0] || j.mousebuttons[0]
	case 1:
		// It can be secondary of first or primary of second
		value = j.button[1] || j.button[2] || j.keys[1] || j.mousebuttons[1]
	case 2:
		value = j.button[3] || j.keys[2] || j.mousebuttons[2]
	}
	return value
}

func (j *sdlJoysticks) ReadPaddle(i int) (uint8, bool) {
	return j.paddle[i], j.hasPaddle[i]
}
