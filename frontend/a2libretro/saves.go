package main

/*
#include "libretro.h"
#include "shim.h"
*/
import "C"

import (
	"unsafe"
)

/*
Where the games save.

The emulation writes to a diskette as the software saves, and by default it
writes straight into the file the diskette was loaded from. That is what the
command line does, but not what somebody loading a game in a libretro frontend
expects: the content should be left alone and the saves should go to the save
directory, where the frontend backs them up.

So the core points the emulator at the save directory and the changes go to an
overlay file next to the other saves, one per diskette. Only the tracks the
software writes are stored, so an overlay is a few kilobytes rather than a copy
of the diskette, and the image loaded is never touched.

It also means that the diskettes that could not be written at all, the ones
inside a zip or a gzip, the ones loaded from an URL and the DOS 3.3 built into
the core, can now be saved to.
*/

/*
saveOverrides is the configuration that sends the changes written to the disks
to the save directory. It is a setting of the machine, applied when it is built,
so it is given as a configuration override and not set afterwards: the disks are
inserted while the machine is being built.
*/
func saveOverrides() map[string]string {
	directory, ok := saveDirectory()
	if !ok {
		logf("the frontend gave no save directory, the disks will be written where they are")
		return nil
	}

	return map[string]string{"saveDir": directory}
}

// saveDirectory returns the directory the frontend keeps the saves in
func saveDirectory() (string, bool) {
	var directory *C.char
	ok := C.shim_environment(C.RETRO_ENVIRONMENT_GET_SAVE_DIRECTORY, unsafe.Pointer(&directory))
	if !ok || directory == nil {
		return "", false
	}

	path := C.GoString(directory)
	if path == "" {
		return "", false
	}
	return path, true
}
