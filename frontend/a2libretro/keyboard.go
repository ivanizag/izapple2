package main

/*
#include "libretro.h"
#include "shim.h"
*/
import "C"

import (
	"unsafe"

	"github.com/ivanizag/izapple2"
)

// keyboard turns the key events of the frontend into the codes the Apple II
// keyboard would produce
type keyboard struct {
	channel *izapple2.KeyboardChannel
}

func newKeyboard(a *izapple2.Apple2) *keyboard {
	return &keyboard{
		channel: izapple2.NewKeyboardChannel(a),
	}
}

/*
registerKeyboardCallback asks the frontend to report the key events. The
callback is a C function that forwards to izapple2KeyboardEvent.

It is given as soon as the environment is set, so that the frontend knows this
is a machine with a keyboard before it loads anything, which is what it looks at
to turn its game focus on by itself. The C side drops the keys until
setKeyboardReady opens it, so nothing reaches Go before there is a machine.
*/
func registerKeyboardCallback() {
	callback := C.struct_retro_keyboard_callback{
		callback: C.retro_keyboard_event_t(C.shim_keyboard_callback),
	}
	C.shim_environment(C.RETRO_ENVIRONMENT_SET_KEYBOARD_CALLBACK, unsafe.Pointer(&callback))
}

// setKeyboardReady lets the key events reach the machine, or stops them
func setKeyboardReady(ready bool) {
	value := C.int(0)
	if ready {
		value = 1
	}
	C.shim_set_keyboard_ready(value)
}

//export izapple2KeyboardEvent
func izapple2KeyboardEvent(down C.int, keycode C.uint, character C.uint32_t, keyModifiers C.uint16_t) {
	if theCore.keyboard == nil {
		return
	}

	// The Apple II reads the open and closed apple keys as joystick buttons,
	// so they are a state and not a keystroke
	switch keycode {
	case C.RETROK_LALT:
		theCore.joysticks.openApple = down != 0
		return
	case C.RETROK_RALT:
		theCore.joysticks.closedApple = down != 0
		return
	}

	if down == 0 {
		return
	}

	if key, ok := specialKey(uint32(keycode), uint16(keyModifiers)); ok {
		theCore.keyboard.channel.PutChar(key)
		return
	}

	// PutRune drops what is not printable ASCII and applies the upper case of
	// the models without lower case
	theCore.keyboard.channel.PutRune(rune(character))
}

// specialKey returns the code of the keys that do not come as a printable
// character, and of the control combinations
func specialKey(keycode uint32, modifiers uint16) (uint8, bool) {
	switch keycode {
	case C.RETROK_RETURN, C.RETROK_KP_ENTER:
		return 13, true
	case C.RETROK_TAB:
		return 9, true
	case C.RETROK_ESCAPE:
		return 27, true
	case C.RETROK_BACKSPACE, C.RETROK_LEFT:
		return 8, true
	case C.RETROK_RIGHT:
		return 21, true
	case C.RETROK_UP:
		return 11, true
	case C.RETROK_DOWN:
		return 10, true
	case C.RETROK_DELETE:
		return 127, true
	}

	if modifiers&C.RETROKMOD_CTRL != 0 &&
		keycode >= C.RETROK_a && keycode <= C.RETROK_z {
		return uint8(keycode-C.RETROK_a) + 1, true
	}

	return 0, false
}
