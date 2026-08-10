package main

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/ivanizag/izapple2"
	"github.com/ivanizag/izapple2/frontend/shared"

	"github.com/hajimehoshi/ebiten/v2"
)

/*
The window is divided in as many vertical areas as removable media drives, to
show where a file can be dropped and what each drive has inserted.

Ebiten reports the files dropped but not where they were dropped, and the
cursor position is not updated while another application drags a file over the
window. The area used is the one the pointer was last seen on, the areas are
shown with F8 to know them beforehand, and they are shown again after a drop to
tell where the file went.

Ebitengine hands the files dropped as a file system that hides their paths, but
it opens the real files, so the handle of a file tells its path back and the
diskette is loaded the same way as on the other frontends. Only the files with
a path can be loaded, which leaves out the browser.
*/

type ebitenDropTargets struct {
	a       *izapple2.Apple2
	targets *shared.DropTargets
}

func newEbitenDropTargets(a *izapple2.Apple2) *ebitenDropTargets {
	var d ebitenDropTargets
	d.a = a
	d.targets = shared.NewDropTargets(a)
	return &d
}

// update loads the files dropped on the window, if any
func (d *ebitenDropTargets) update() {
	files := ebiten.DroppedFiles()
	if files == nil {
		return
	}

	entries, err := fs.ReadDir(files, ".")
	if err != nil {
		fmt.Printf("Could not read the files dropped: %v\n", err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only the first file dropped is loaded, on the drive it was dropped on
		if d.load(files, entry.Name()) {
			return
		}
	}
}

// load inserts a file dropped on the window. It returns false to try with the
// next file dropped, if there is one.
func (d *ebitenDropTargets) load(files fs.FS, name string) bool {
	file, err := files.Open(name)
	if err != nil {
		fmt.Printf("Could not open '%v': %v\n", name, err)
		return false
	}
	defer file.Close()

	drive := d.dropped()
	if drive < 0 {
		fmt.Printf("There are no drives to load '%v' on\n", name)
		return true
	}

	// The file dropped is a real file that Ebitengine has just opened. Its
	// handle knows the path it came from, and the path is what the emulator
	// wants: it takes the compressed images too, and the changes go back to
	// the file or to the save directory.
	realFile, ok := file.(*os.File)
	if !ok {
		fmt.Printf("Could not find where '%v' came from\n", name)
		return false
	}

	path := realFile.Name()
	fmt.Printf("Loading '%s' in drive %v\n", path, drive+1)
	d.a.SendLoadDisk(drive, path)
	return true
}

// dropped returns the drive that gets a file dropped now
func (d *ebitenDropTargets) dropped() int {
	drive := d.pointedDrive()
	if drive >= 0 {
		d.targets.Dropped(drive)
	}
	return drive
}

// pointedDrive returns the drive the mouse pointer is on. The positions are on
// the virtual screen the game is laid out on, not on the window.
func (d *ebitenDropTargets) pointedDrive() int {
	x, y := ebiten.CursorPosition()
	if x < 0 || x >= virtualWidth || y < 0 || y >= virtualHeight {
		return -1
	}

	return d.targets.DriveAt(x, virtualWidth)
}
