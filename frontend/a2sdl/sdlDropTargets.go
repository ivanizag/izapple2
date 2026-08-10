//go:build !js

package main

import (
	"image"

	"github.com/ivanizag/izapple2"
	"github.com/ivanizag/izapple2/frontend/shared"
	"github.com/veandco/go-sdl2/sdl"
)

/*
The window is divided in as many vertical areas as removable media drives, to
show where a file can be dropped and what each drive has inserted.

SDL2 does not tell that a file is being dragged over the window: the drop
events arrive only when the file is released, and SDL_DropEvent has no
position. The areas are shown with F8 then, and after a drop to tell where the
file went. SDL3 does report the drag, see the a2sdl3 frontend.
*/

type sdlDropTargets struct {
	targets *shared.DropTargets
	window  *sdl.Window
}

func newSDLDropTargets(a *izapple2.Apple2, window *sdl.Window) *sdlDropTargets {
	var d sdlDropTargets
	d.targets = shared.NewDropTargets(a)
	d.window = window
	return &d
}

// dropped returns the drive that gets a file dropped now
func (d *sdlDropTargets) dropped() int {
	width, _ := d.window.GetSize()

	// The mouse motion events do not arrive while another application drags a
	// file, so the position of the pointer on the window is stale and the one
	// on the desktop is asked for instead. Where that is not available, as on
	// Wayland, it lands outside the window and the stale one is all there is.
	globalX, _, _ := sdl.GetGlobalMouseState()
	windowX, _ := d.window.GetPosition()
	x := globalX - windowX
	if x < 0 || x >= width {
		x, _, _ = sdl.GetMouseState()
	}

	drive := d.targets.DriveAt(int(x), int(width))
	if drive >= 0 {
		d.targets.Dropped(drive)
	}
	return drive
}

// showing returns whether the areas take over the screen, either because the
// user asked for them with F8 or because a file has just been dropped
func (d *sdlDropTargets) showing(requested bool) bool {
	return requested || d.targets.Flashing()
}

// snapshot returns the screen with the areas, highlighting the drive that got
// the last file while the flash lasts, or the one under the pointer
func (d *sdlDropTargets) snapshot() *image.RGBA {
	selected := d.targets.FlashDrive()
	if !d.targets.Flashing() {
		selected = d.pointedDrive()
	}
	return d.targets.Snapshot(selected)
}

// pointedDrive returns the drive the mouse pointer is on, -1 when it is
// outside the window
func (d *sdlDropTargets) pointedDrive() int {
	if sdl.GetMouseFocus() != d.window {
		return -1
	}

	mouseX, _, _ := sdl.GetMouseState()
	width, _ := d.window.GetSize()
	if mouseX < 0 || mouseX >= width {
		return -1
	}

	return d.targets.DriveAt(int(mouseX), int(width))
}
