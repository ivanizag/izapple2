package main

/*
#include "libretro.h"
#include "shim.h"
*/
import "C"

/*
joysticks maps the first libretro controller port to the Apple II paddles.

The methods are called from the emulation, which under RunCycles runs on the
same thread as retro_run, so the input state can be read straight from the
frontend as the machine asks for it.
*/
type joysticks struct {
	// The open and closed apple keys, that the Apple II reads as the buttons
	// of the first joystick
	openApple   bool
	closedApple bool
}

func newJoysticks() *joysticks {
	return &joysticks{}
}

func (j *joysticks) ReadButton(i int) bool {
	switch i {
	case 0:
		return j.openApple || joypadPressed(C.RETRO_DEVICE_ID_JOYPAD_A)
	case 1:
		return j.closedApple || joypadPressed(C.RETRO_DEVICE_ID_JOYPAD_B)
	case 2:
		return joypadPressed(C.RETRO_DEVICE_ID_JOYPAD_X)
	}
	return false
}

func (j *joysticks) ReadPaddle(i int) (uint8, bool) {
	var negative, positive, axis C.uint
	switch i {
	case 0:
		negative = C.RETRO_DEVICE_ID_JOYPAD_LEFT
		positive = C.RETRO_DEVICE_ID_JOYPAD_RIGHT
		axis = C.RETRO_DEVICE_ID_ANALOG_X
	case 1:
		negative = C.RETRO_DEVICE_ID_JOYPAD_UP
		positive = C.RETRO_DEVICE_ID_JOYPAD_DOWN
		axis = C.RETRO_DEVICE_ID_ANALOG_Y
	default:
		// Only one joystick, the second one reads centered
		return 128, false
	}

	// The d-pad takes precedence, for the controllers with no analog stick
	if joypadPressed(negative) {
		return 0, true
	}
	if joypadPressed(positive) {
		return 255, true
	}

	value := C.shim_input_state(0, C.RETRO_DEVICE_ANALOG,
		C.RETRO_DEVICE_INDEX_ANALOG_LEFT, axis)

	// The analog axis goes from -32768 to 32767, the paddle from 0 to 255
	return uint8((int32(value) + 32768) >> 8), true
}

func joypadPressed(id C.uint) bool {
	return C.shim_input_state(0, C.RETRO_DEVICE_JOYPAD, 0, id) != 0
}
