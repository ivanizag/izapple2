//go:build !js

package main

import (
	"github.com/Zyko0/go-sdl3/sdl"
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

type sdl3Joysticks struct {
	paddle       [4]uint8
	hasPaddle    [4]bool
	button       [4]bool
	keys         [3]bool
	mousebuttons [3]bool
	useMouse     bool

	// SDL3 identifies joysticks with instance ids that are not indexes:
	// they start at 1 and grow as devices are plugged. The events carry
	// the instance id, so we keep the slot each one was assigned to.
	slots map[sdl.JoystickID]uint8
}

func newSDL3Joysticks(useMouseAlt bool) *sdl3Joysticks {
	var j sdl3Joysticks
	j.slots = make(map[sdl.JoystickID]uint8)

	sdl.SetJoystickEventsEnabled(true)

	// Init up to two joysticks
	ids, err := sdl.GetJoysticks()
	if err == nil {
		for slot, id := range ids {
			if slot >= 2 {
				break
			}
			_, err := id.OpenJoystick()
			if err != nil {
				continue
			}
			j.slots[id] = uint8(slot)
			j.hasPaddle[slot*2] = true
			j.hasPaddle[slot*2+1] = true
		}
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

func (j *sdl3Joysticks) putAxisEvent(e *sdl.JoyAxisEvent) {
	slot, ok := j.slots[e.Which]
	if !ok || e.Axis >= 2 {
		// Process only the first two axis of the first two joysticks
		return
	}

	j.paddle[slot*2+e.Axis] = uint8((e.Value >> 8) + 128)
}

func (j *sdl3Joysticks) putButtonEvent(e *sdl.JoyButtonEvent) {
	slot, ok := j.slots[e.Which]
	if !ok {
		// Process only the buttons of the first two joysticks
		return
	}

	j.button[slot*2+(e.Button%2)] = e.Down
}

func mouseToJoyCentered(x int32, w int32) uint8 {
	r := max(min(x-(w/2)+127, 255), 0)
	return uint8(r)

}

func (j *sdl3Joysticks) putMouseMotionEvent(e *sdl.MouseMotionEvent, width int32, height int32) {
	if j.useMouse {
		// The mouse moves on all the window
		// j.paddle[0] = mouseToJoyFull(int32(e.X), width)
		// j.paddle[1] = mouseToJoyFull(int32(e.Y), height)

		// The mouse moves around the center of the window
		j.paddle[0] = mouseToJoyCentered(int32(e.X), width)
		j.paddle[1] = mouseToJoyCentered(int32(e.Y), height)
	}
}

func (j *sdl3Joysticks) putMouseButtonEvent(e *sdl.MouseButtonEvent) {
	if j.useMouse {
		switch e.Button {
		case 1: // BUTTON_LEFT
			j.mousebuttons[0] = e.Down
		case 3: // BUTTON_RIGHT
			j.mousebuttons[1] = e.Down
		case 2: // BUTTON_MIDDLE
			j.mousebuttons[2] = e.Down
		}
	}
}

func (j *sdl3Joysticks) putKey(keyEvent *sdl.KeyboardEvent) {
	/*
		We will simultate joystick buttons with keyboard keys.
		Actually the Apple//e does this with the open and solid apple keys.
		   Alt key - button 0 - Open apple
		   AltGr key - button 1- Solid apple
		   //Win key - button 2 (Not in the Apple //e keyboard)

		The apple keys are a place on the keyboard, not a symbol, so we look at
		the scancode. Unlike SDL2, SDL3 resolves the keycode through the current
		keymap: on layouts where the left alt key is Meta_L instead of Alt_L the
		keycode is not K_LALT and the open apple would never be pressed.
	*/
	switch keyEvent.Scancode {
	case sdl.SCANCODE_LALT:
		j.keys[0] = keyEvent.Down
	case sdl.SCANCODE_RALT:
		j.keys[1] = keyEvent.Down
		// case sdl.SCANCODE_LGUI:
		//   j.keys[2] = keyEvent.Down
	}

}

// releaseKeys forgets the apple keys being pressed. The key up event of a key
// held while the window loses the focus goes to whoever took it, leaving the
// apple key pressed forever.
func (j *sdl3Joysticks) releaseKeys() {
	j.keys = [3]bool{}
}

func (j *sdl3Joysticks) ReadButton(i int) bool {
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

func (j *sdl3Joysticks) ReadPaddle(i int) (uint8, bool) {
	return j.paddle[i], j.hasPaddle[i]
}
