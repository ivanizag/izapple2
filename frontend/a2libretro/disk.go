package main

/*
#include <stdlib.h>
#include "libretro.h"
#include "shim.h"
*/
import "C"

import (
	"os"
	"path/filepath"
	"strings"
	"unsafe"
)

/*
The disk control interface, what lets the frontend swap diskettes from its own
menu so that the games spanning several disks can be finished.

The list of diskettes comes from an .m3u playlist, the libretro convention for
multi disk content: a text file with one image per line. Loading a single disk
image gives a list of one.

The Apple II has no way to know that a diskette was taken out, and the drive
has no eject, so ejecting only bookkeeps. The diskette of the selected index is
inserted when the frontend closes the tray again, which is the order it always
uses: eject, select, insert.
*/

const m3uExtension = ".m3u"

type diskImage struct {
	path  string
	label string
}

type diskControl struct {
	images  []diskImage
	index   uint
	ejected bool
}

// initialIndex is what the frontend asks for before the content is loaded, to
// restore the diskette that was in use the last time
var initialIndex uint

func newDiskControl() *diskControl {
	return &diskControl{}
}

// registerDiskControl offers the interface to the frontend, falling back to the
// older one that has no labels if the extended one is not known
func registerDiskControl() {
	if !C.shim_set_disk_control_interface() {
		logf("the frontend does not support swapping diskettes")
	}
}

// buildImageList returns the diskettes of the content, expanding a playlist
func buildImageList(path string) ([]diskImage, error) {
	if !strings.EqualFold(filepath.Ext(path), m3uExtension) {
		return []diskImage{newDiskImage(path)}, nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	dir := filepath.Dir(path)
	images := make([]diskImage, 0)
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !filepath.IsAbs(line) {
			line = filepath.Join(dir, line)
		}
		images = append(images, newDiskImage(line))
	}
	return images, nil
}

func newDiskImage(path string) diskImage {
	name := filepath.Base(path)
	return diskImage{
		path:  path,
		label: strings.TrimSuffix(name, filepath.Ext(name)),
	}
}

// current returns the path of the selected diskette
func (d *diskControl) current() (string, bool) {
	if d.index >= uint(len(d.images)) {
		return "", false
	}
	return d.images[d.index].path, true
}

// insert puts the selected diskette in the first drive
func (d *diskControl) insert() bool {
	path, ok := d.current()
	if !ok {
		return false
	}
	if theCore.a == nil {
		return false
	}

	theCore.a.SendLoadDisk(0, path)

	// A reset rebuilds the machine, keep it booting what is in the drive
	theCore.filenames = []string{path}
	return true
}

//export izapple2DiskSetEjectState
func izapple2DiskSetEjectState(ejected C.int) C.int {
	d := theCore.disks
	wasEjected := d.ejected
	d.ejected = ejected != 0

	if wasEjected && !d.ejected {
		// The tray is closed again, load whatever was selected meanwhile.
		// With no diskette selected there is nothing to load, and that is not
		// a failure: the frontend is just taking the diskette out.
		d.insert()
	}
	return 1
}

//export izapple2DiskGetEjectState
func izapple2DiskGetEjectState() C.int {
	return boolToC(theCore.disks.ejected)
}

//export izapple2DiskGetImageIndex
func izapple2DiskGetImageIndex() C.uint {
	return C.uint(theCore.disks.index)
}

//export izapple2DiskSetImageIndex
func izapple2DiskSetImageIndex(index C.uint) C.int {
	d := theCore.disks

	// Any index from the number of diskettes up means taking the diskette out
	// without putting another one in, so every value is accepted and it is
	// current() that tells whether there is something to insert
	d.index = uint(index)

	// The frontends eject before selecting, but swap right away for the ones
	// that do not, instead of leaving the drive with the previous diskette
	if !d.ejected {
		d.insert()
	}
	return 1
}

//export izapple2DiskGetNumImages
func izapple2DiskGetNumImages() C.uint {
	return C.uint(len(theCore.disks.images))
}

//export izapple2DiskReplaceImageIndex
func izapple2DiskReplaceImageIndex(index C.uint, info *C.struct_retro_game_info) C.int {
	d := theCore.disks
	if uint(index) >= uint(len(d.images)) {
		return 0
	}

	if info == nil || info.path == nil {
		// Remove the entry
		d.images = append(d.images[:index], d.images[index+1:]...)
		if d.index >= uint(len(d.images)) && len(d.images) > 0 {
			d.index = uint(len(d.images)) - 1
		}
		return 1
	}

	d.images[index] = newDiskImage(C.GoString(info.path))
	return 1
}

//export izapple2DiskAddImageIndex
func izapple2DiskAddImageIndex() C.int {
	theCore.disks.images = append(theCore.disks.images, diskImage{})
	return 1
}

//export izapple2DiskSetInitialImage
func izapple2DiskSetInitialImage(index C.uint, path *C.char) C.int {
	initialIndex = uint(index)
	return 1
}

//export izapple2DiskGetImagePath
func izapple2DiskGetImagePath(index C.uint, buffer *C.char, length C.size_t) C.int {
	d := theCore.disks
	if uint(index) >= uint(len(d.images)) {
		return 0
	}
	return copyToBuffer(buffer, length, d.images[index].path)
}

//export izapple2DiskGetImageLabel
func izapple2DiskGetImageLabel(index C.uint, buffer *C.char, length C.size_t) C.int {
	d := theCore.disks
	if uint(index) >= uint(len(d.images)) {
		return 0
	}
	return copyToBuffer(buffer, length, d.images[index].label)
}

// copyToBuffer writes a null terminated string in a buffer owned by the
// frontend, truncating it if it does not fit
func copyToBuffer(buffer *C.char, length C.size_t, value string) C.int {
	if buffer == nil || length == 0 {
		return 0
	}

	destination := unsafe.Slice((*byte)(unsafe.Pointer(buffer)), int(length))
	written := copy(destination[:len(destination)-1], value)
	destination[written] = 0
	return 1
}

func boolToC(value bool) C.int {
	if value {
		return 1
	}
	return 0
}
