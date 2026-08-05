package main

/*
#include <stdlib.h>
#include "libretro.h"
#include "shim.h"
*/
import "C"

import (
	"fmt"
	"os"
	"unsafe"
)

// messageFrames is how long a message stays on screen, about three seconds
const messageFrames = 180

// setupLog takes the log of the frontend, if it offers one. Everything goes to
// stderr until it does, and if it never does.
func setupLog() {
	var callback C.struct_retro_log_callback
	if C.shim_environment(C.RETRO_ENVIRONMENT_GET_LOG_INTERFACE, unsafe.Pointer(&callback)) {
		C.shim_set_log(callback.log)
	}
}

// logf reports something worth knowing to whoever is watching the log
func logf(format string, args ...any) {
	writeLog(C.RETRO_LOG_INFO, fmt.Sprintf(format, args...))
}

// errorf reports a failure the user has to know about: it goes to the log and
// on screen, as nobody reads the log of a frontend
func errorf(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	writeLog(C.RETRO_LOG_ERROR, message)
	showMessage(message)
}

func writeLog(level C.int, message string) {
	if !C.shim_has_log() {
		fmt.Fprintf(os.Stderr, "[izapple2] %v\n", message)
		return
	}

	cMessage := C.CString(message)
	defer C.free(unsafe.Pointer(cMessage))
	C.shim_log(level, cMessage)
}

// showMessage puts a notice on the screen of the frontend
func showMessage(message string) {
	cMessage := C.CString(message)
	defer C.free(unsafe.Pointer(cMessage))

	notice := C.struct_retro_message{
		msg:    cMessage,
		frames: messageFrames,
	}
	C.shim_environment(C.RETRO_ENVIRONMENT_SET_MESSAGE, unsafe.Pointer(&notice))
}

// setInputDescriptors names the controls in the menus of the frontend, so that
// remapping them shows what they do on the Apple II
func setInputDescriptors() {
	type descriptor struct {
		device      C.uint
		index       C.uint
		id          C.uint
		description string
	}

	descriptors := []descriptor{
		{C.RETRO_DEVICE_JOYPAD, 0, C.RETRO_DEVICE_ID_JOYPAD_LEFT, "Paddle 0 left"},
		{C.RETRO_DEVICE_JOYPAD, 0, C.RETRO_DEVICE_ID_JOYPAD_RIGHT, "Paddle 0 right"},
		{C.RETRO_DEVICE_JOYPAD, 0, C.RETRO_DEVICE_ID_JOYPAD_UP, "Paddle 1 up"},
		{C.RETRO_DEVICE_JOYPAD, 0, C.RETRO_DEVICE_ID_JOYPAD_DOWN, "Paddle 1 down"},
		{C.RETRO_DEVICE_JOYPAD, 0, C.RETRO_DEVICE_ID_JOYPAD_A, "Button 0 (open apple)"},
		{C.RETRO_DEVICE_JOYPAD, 0, C.RETRO_DEVICE_ID_JOYPAD_B, "Button 1 (closed apple)"},
		{C.RETRO_DEVICE_JOYPAD, 0, C.RETRO_DEVICE_ID_JOYPAD_X, "Button 2"},
		{C.RETRO_DEVICE_ANALOG, C.RETRO_DEVICE_INDEX_ANALOG_LEFT, C.RETRO_DEVICE_ID_ANALOG_X, "Paddle 0"},
		{C.RETRO_DEVICE_ANALOG, C.RETRO_DEVICE_INDEX_ANALOG_LEFT, C.RETRO_DEVICE_ID_ANALOG_Y, "Paddle 1"},
	}

	// The frontend keeps the array until the content is unloaded, so it is
	// built in C memory and ends with a null description
	count := len(descriptors) + 1
	size := C.size_t(count) * C.size_t(unsafe.Sizeof(C.struct_retro_input_descriptor{}))
	array := (*C.struct_retro_input_descriptor)(C.malloc(size))
	entries := unsafe.Slice(array, count)

	for i, d := range descriptors {
		entries[i].port = 0
		entries[i].device = d.device
		entries[i].index = d.index
		entries[i].id = d.id
		entries[i].description = C.CString(d.description)
	}
	entries[count-1].description = nil

	C.shim_environment(C.RETRO_ENVIRONMENT_SET_INPUT_DESCRIPTORS, unsafe.Pointer(array))
}
