/*
A libretro core for izapple2.

Unlike the other frontends, that let the emulation run free on its own
goroutine, the libretro frontend owns the frame timing: it calls retro_run()
once per frame and the core has to advance the machine by exactly one frame,
hand over one video frame and the matching audio samples, and return. That is
what Apple2.RunCycles is for, and everything here happens on the thread that
calls retro_run().

Kept deliberately small: no super hi-res, no Videx video, no RGB card and no
savestates. See README.md.
*/
package main

/*
#include <stdlib.h>
#include "libretro.h"
#include "shim.h"
*/
import "C"

import (
	"unsafe"

	"github.com/ivanizag/izapple2"
)

const (
	// cyclesPerFrame is a full NTSC video frame, 65 cycles for each of the
	// 262 lines
	cyclesPerFrame = 17030

	/*
		fastModeFactor is how many frames worth of cycles are run in a single
		frame while a device asks for fast mode, which the disk drive does
		while the motor is on. Without it a DOS 3.3 boot, that takes some 37
		million cycles of mostly waiting for the drive, would need more than
		half a minute. The other frontends let the emulation run free instead,
		but a libretro core has to return in time for the next frame, so the
		speed up is capped. Sound is choppy while it lasts, as it is on the
		other frontends.
	*/
	fastModeFactor = 16
)

// core is the state of the loaded core. All of it is touched only from the
// thread that calls retro_run().
type core struct {
	a         *izapple2.Apple2
	video     *videoOutput
	audio     *audioOutput
	keyboard  *keyboard
	joysticks *joysticks
	disks     *diskControl

	// Media to reload when the machine is rebuilt on reset
	filenames []string
}

var theCore = core{disks: newDiskControl()}

// build creates the machine from the current core options and the media
func (c *core) build() bool {
	c.close()

	options := readOptions()

	a, err := izapple2.CreateAppleFromModel(
		options.model, buildOverrides(options.model), c.filenames)
	if err != nil {
		errorf("could not create the machine: %v", err)
		return false
	}

	c.a = a
	c.video = newVideoOutput(options.screenMode)
	c.audio = newAudioOutput(a.GetClockMhz())
	c.keyboard = newKeyboard(a)
	c.joysticks = newJoysticks()

	for _, source := range a.GetAudioSources() {
		source.SetAudioSink(c.audio.mixer.NewSource())
	}
	a.SetJoysticksProvider(c.joysticks)
	a.SetMouseProvider(newMouse())

	a.Init()
	return true
}

func (c *core) close() {
	c.a = nil
	c.video = nil
	c.audio = nil
	c.keyboard = nil
	c.joysticks = nil
}

/*
retro_api_version and retro_get_system_info are in shim.c, in C. The frontend
asks them on its main thread while it is loading the core, and an exported Go
function waits there for the Go runtime to finish starting, which deadlocks.
*/

//export retro_get_system_av_info
func retro_get_system_av_info(info *C.struct_retro_system_av_info) {
	clockMhz := izapple2.CPUClockMhz
	if theCore.a != nil {
		clockMhz = theCore.a.GetClockMhz()
	}

	info.geometry.base_width = frameWidth
	info.geometry.base_height = frameHeight
	info.geometry.max_width = frameWidth
	info.geometry.max_height = frameHeight
	// The image has two pixels per Apple II pixel, the display is 4:3
	info.geometry.aspect_ratio = 4.0 / 3.0

	info.timing.fps = C.double(clockMhz * 1_000_000 / cyclesPerFrame)
	info.timing.sample_rate = C.double(sampleRate)
}

//export retro_set_environment
func retro_set_environment(cb C.retro_environment_t) {
	C.shim_set_environment(cb)

	// Take the log of the frontend before anything can need it
	setupLog()

	// The core boots to BASIC with no media
	noGame := C.bool(true)
	C.shim_environment(C.RETRO_ENVIRONMENT_SET_SUPPORT_NO_GAME, unsafe.Pointer(&noGame))

	setVariables()
}

//export retro_set_video_refresh
func retro_set_video_refresh(cb C.retro_video_refresh_t) {
	C.shim_set_video_refresh(cb)
}

//export retro_set_audio_sample
func retro_set_audio_sample(cb C.retro_audio_sample_t) {
	// Not used, the core delivers the samples in batches
}

//export retro_set_audio_sample_batch
func retro_set_audio_sample_batch(cb C.retro_audio_sample_batch_t) {
	C.shim_set_audio_sample_batch(cb)
}

//export retro_set_input_poll
func retro_set_input_poll(cb C.retro_input_poll_t) {
	C.shim_set_input_poll(cb)
}

//export retro_set_input_state
func retro_set_input_state(cb C.retro_input_state_t) {
	C.shim_set_input_state(cb)
}

//export retro_init
func retro_init() {
}

//export retro_deinit
func retro_deinit() {
	theCore.close()
}

//export retro_set_controller_port_device
func retro_set_controller_port_device(port C.unsigned, device C.unsigned) {
}

/*
The four entry points that take a const pointer are exported to C under
another name and reached through a wrapper in shim.c: cgo can not put const in
the declarations it generates, and they would clash with libretro.h.
*/

//export izapple2LoadGame
func izapple2LoadGame(info *C.struct_retro_game_info) C.bool {
	format := C.enum_retro_pixel_format(C.RETRO_PIXEL_FORMAT_XRGB8888)
	if !C.shim_environment(C.RETRO_ENVIRONMENT_SET_PIXEL_FORMAT, unsafe.Pointer(&format)) {
		errorf("the frontend does not support the XRGB8888 pixel format")
		return false
	}

	theCore.filenames = nil
	theCore.disks = newDiskControl()

	if info != nil && info.path != nil {
		images, err := buildImageList(C.GoString(info.path))
		if err != nil {
			errorf("could not read the playlist: %v", err)
			return false
		}
		theCore.disks.images = images

		// Start with the diskette that was in use the last time
		if initialIndex < uint(len(images)) {
			theCore.disks.index = initialIndex
		}
		if path, ok := theCore.disks.current(); ok {
			theCore.filenames = []string{path}
		}
	}

	if !theCore.build() {
		return false
	}

	/*
		The keyboard callback is given to the frontend only once there is a
		machine to receive the keys, and not in retro_set_environment. Handing
		it over earlier lets the frontend call into Go from its own threads
		while the core is still being loaded, and the Go runtime does not
		survive that: it dies with a morestack on g0, running on a stack it has
		no record of.

		The cost is that the game focus of RetroArch set to Detect does not see
		the keyboard in time. Set it to On instead.
	*/
	registerKeyboardCallback()

	registerDiskControl()
	setInputDescriptors()
	return true
}

//export izapple2LoadGameSpecial
func izapple2LoadGameSpecial(gameType C.unsigned, info *C.struct_retro_game_info, numInfo C.size_t) C.bool {
	return false
}

//export retro_unload_game
func retro_unload_game() {
	theCore.close()
	theCore.filenames = nil
	theCore.disks = newDiskControl()
}

//export retro_reset
func retro_reset() {
	// Rebuild the machine instead of sending a reset to the 6502, so that a
	// change of the model option is picked up with a cold boot
	theCore.build()
}

//export retro_run
func retro_run() {
	if theCore.a == nil {
		return
	}

	C.shim_input_poll()
	theCore.video.setScreenMode(readOptions().screenMode)

	cycles := uint64(cyclesPerFrame)
	if theCore.a.IsFastModeRequested() {
		cycles *= fastModeFactor
	}
	theCore.a.RunCycles(cycles)

	theCore.video.render(theCore.a.GetVideoSource())
	theCore.audio.render()
}

//export retro_get_region
func retro_get_region() C.unsigned {
	return C.RETRO_REGION_NTSC
}

//export retro_get_memory_data
func retro_get_memory_data(id C.unsigned) unsafe.Pointer {
	return nil
}

//export retro_get_memory_size
func retro_get_memory_size(id C.unsigned) C.size_t {
	return 0
}

// The core does not implement savestates, so it reports a size of zero and
// the frontend disables rewind, run ahead and netplay

//export retro_serialize_size
func retro_serialize_size() C.size_t {
	return 0
}

//export retro_serialize
func retro_serialize(data unsafe.Pointer, size C.size_t) C.bool {
	return false
}

//export izapple2Unserialize
func izapple2Unserialize(data unsafe.Pointer, size C.size_t) C.bool {
	return false
}

//export retro_cheat_reset
func retro_cheat_reset() {
}

//export izapple2CheatSet
func izapple2CheatSet(index C.unsigned, enabled C.bool, code *C.char) {
}

func main() {}
