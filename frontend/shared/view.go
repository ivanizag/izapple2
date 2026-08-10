package shared

import (
	"fmt"
	"image"

	"github.com/ivanizag/izapple2"
	"github.com/ivanizag/izapple2/screen"
)

/*
The frontends with a window show the screen of the machine, and on demand one
of the screens that take over it: the help, the areas of the drives a file can
be dropped on, the character generator or the pages of video memory. What is
shown and what the window is titled is the same on all of them, only the way
the image reaches the window changes.
*/

// View is the screen a frontend shows and the choices the user made about it
type View struct {
	ShowHelp        bool
	ShowPages       bool
	ShowCharGen     bool
	ShowAltText     bool
	ShowDropTargets bool
	ScreenMode      int
}

// NewView creates the view a frontend starts with
func NewView() *View {
	var v View
	v.ScreenMode = screen.ScreenModeColorScanlines
	return &v
}

// ToggleHelp shows or hides the help
func (v *View) ToggleHelp() {
	v.ShowHelp = !v.ShowHelp
}

// ToggleDropTargets shows or hides the areas of the drives a file can be
// dropped on
func (v *View) ToggleDropTargets() {
	v.ShowDropTargets = !v.ShowDropTargets
	if v.ShowDropTargets {
		// The help is shown on top of the drop targets, get it out of the way
		v.ShowHelp = false
	}
}

// NextScreenMode moves to the next way of rendering the screen
func (v *View) NextScreenMode() {
	v.ScreenMode = screen.NextScreenMode(v.ScreenMode)
}

/*
Snapshot returns the image to put on the window and the title it should have,
empty when the title does not change.

The drive the pointer is on highlights the areas of the drop targets while they
are shown. Pass -1 where the pointer is outside the window or its position is
not known.
*/
func (v *View) Snapshot(a *izapple2.Apple2, d *DropTargets, pointed int) (*image.RGBA, string) {
	vs := a.GetVideoSource()

	switch {
	case v.ShowHelp:
		return screen.SnapshotMessageGenerator(vs, HelpMessage, false /*is80Columns*/), ""

	case d.Showing(v.ShowDropTargets):
		return d.SnapshotAt(pointed), ""

	case v.ShowCharGen:
		cgPage, cgPages := a.GetCgPageInfo()
		return screen.SnapshotCharacterGenerator(vs, v.ShowAltText),
			fmt.Sprintf("%v character map, page %v/%v", a.Name, cgPage+1, cgPages)

	case v.ShowPages:
		img := screen.SnapshotParts(vs, v.ScreenMode)
		return img, fmt.Sprintf("%v %v %vx%v", a.Name, screen.VideoModeName(vs),
			img.Rect.Dx()/2, img.Rect.Dy()/2)
	}

	return screen.Snapshot(vs, v.ScreenMode), ""
}

// HelpMessage is the text shown with F1, the keys the frontends with a window
// have in common
const HelpMessage = `
          F1: Show/Hide help
     Ctrl-F2: Reset
      F1, F2: Reset
          F4: Show/Hide CPU trace
          F5: Fast/Normal speed
     Ctrl-F5: Show speed
          F6: Next screen mode
          F7: Show/Hide pages
          F8: Show/Hide drop targets
         F10: Next character set
    Ctrl-F10: Show/Hide character set
   Shift-F10: Show/Hide alternate text
         F12: Save screen snapshot
       Pause: Pause the emulation

  Left alt or option key: Open-Apple
 Right alt or option key: Closed-Apple

Drop a file on a drive area to load it

 Run izapple2 -h for more options
   https://github.com/ivanizag/izapple2
`
