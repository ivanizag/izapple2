//go:build !js

package main

import (
	"image"

	"github.com/ivanizag/izapple2"
	"github.com/ivanizag/izapple2/frontend/shared"

	"github.com/Zyko0/go-sdl3/sdl"
)

/*
The window is divided in as many vertical areas as removable media drives, to
show where a file can be dropped and what each drive has inserted.

Unlike SDL2, SDL3 reports the position of a file dragged over the window, so
the areas are shown while the drag lasts, with the one under the pointer
highlighted. They can also be shown with F8, and after a drop to tell where the
file went.
*/

type sdl3DropTargets struct {
	targets *shared.DropTargets
	window  *sdl.Window

	dragging  bool    // A file is being dragged over the window
	dragKnown bool    // The position of the file has been reported
	dragX     float32 // Where the file is on the window
}

func newSDL3DropTargets(a *izapple2.Apple2, window *sdl.Window) *sdl3DropTargets {
	var d sdl3DropTargets
	d.targets = shared.NewDropTargets(a)
	d.window = window
	return &d
}

// dragStarted is called when a file begins to be dragged over the window,
// before knowing where it is
func (d *sdl3DropTargets) dragStarted() {
	d.dragging = true
	d.dragKnown = false
}

// dragMoved tracks a file being dragged over the window
func (d *sdl3DropTargets) dragMoved(x float32) {
	d.dragging = true
	d.dragKnown = true
	d.dragX = x
}

// dragEnded is called when the file is dropped or leaves the window. The mouse
// motion events do not arrive while a file is being dragged, so the first one
// also means that the drag is over, whatever SDL reported.
func (d *sdl3DropTargets) dragEnded() {
	d.dragging = false
}

// dropped returns the drive that gets a file dropped at the given position of
// the window
func (d *sdl3DropTargets) dropped(x float32) int {
	drive := d.driveAt(x)
	if drive >= 0 {
		d.targets.Dropped(drive)
	}
	return drive
}

// showing returns whether the areas take over the screen, either because a
// file is being dragged, because the user asked for them with F8, or because a
// file has just been dropped
func (d *sdl3DropTargets) showing(requested bool) bool {
	return d.dragging || requested || d.targets.Flashing()
}

// snapshot returns the screen with the areas, highlighting the drive under the
// file being dragged or the one that got the last file
func (d *sdl3DropTargets) snapshot() *image.RGBA {
	selected := -1
	if d.dragging {
		if d.dragKnown {
			selected = d.driveAt(d.dragX)
		}
	} else if d.targets.Flashing() {
		selected = d.targets.FlashDrive()
	}

	return d.targets.Snapshot(selected)
}

func (d *sdl3DropTargets) driveAt(x float32) int {
	width, _, err := d.window.Size()
	if err != nil {
		return -1
	}

	return d.targets.DriveAt(int(x), int(width))
}
