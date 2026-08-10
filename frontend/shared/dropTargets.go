// Package shared has the code the frontends have in common: the parts that are
// not the emulated machine, and would be copied from one frontend to the next
// otherwise.
package shared

import (
	"image"
	"strings"
	"time"

	"github.com/ivanizag/izapple2"
	"github.com/ivanizag/izapple2/screen"
)

/*
A file dragged on the window can be dropped on any of the removable media
drives of the machine. The window is divided in as many vertical areas as
drives, and this builds the screen that shows them, each one with the name of
its drive and the media it has inserted.

It is a text screen of 80 columns rendered with the character generator of the
machine, the same way the frontends show their help.
*/

const (
	dropTargetsColumns = 80
	dropTargetsLines   = 24

	dropTargetsTitle  = "DROP A FILE ON A DRIVE"
	dropTargetsFooter = "The file goes to the drive of the area it is dropped on"
	dropTargetsEmpty  = "EMPTY"

	dropTargetsTitleLine  = 1  // Title, centered on the screen
	dropTargetsTopLine    = 3  // Top of the areas
	dropTargetsRuleTop    = 4  // Rule over the drive selected
	dropTargetsLabelLine  = 6  // Name of the drive
	dropTargetsMediaLine  = 8  // First line with the media inserted
	dropTargetsMediaLines = 6  // Lines available for the media inserted
	dropTargetsRuleBottom = 15 // Rule under the drive selected
	dropTargetsBottomLine = 17 // Bottom of the areas, not included
	dropTargetsFooterLine = 21 // Footer, centered on the screen

	// DropTargetsFlashDuration is how long the areas are shown after a drop,
	// to tell where the file went
	DropTargetsFlashDuration = 1500 * time.Millisecond
)

// DropTargets shows where a file dragged on the window can be dropped
type DropTargets struct {
	a *izapple2.Apple2

	flashUntil time.Time // The areas are shown for a while after a drop
	flashDrive int
}

func NewDropTargets(a *izapple2.Apple2) *DropTargets {
	var d DropTargets
	d.a = a
	d.flashDrive = -1
	return &d
}

// Count returns how many drives a file can be dropped on
func (d *DropTargets) Count() int {
	return len(d.a.GetRemovableMediaDrives())
}

// DriveAt returns the drive that would get a file dropped at the x position of
// a window of the given width. It returns -1 if there are no drives to drop a
// file on.
func (d *DropTargets) DriveAt(x int, width int) int {
	return DropTargetIndex(x, width, d.Count())
}

// Dropped notes the drive a file has just gone to, to show the areas for a
// moment with that drive highlighted
func (d *DropTargets) Dropped(drive int) {
	d.flashDrive = drive
	d.flashUntil = time.Now().Add(DropTargetsFlashDuration)
}

// Flashing returns whether the areas are being shown after a drop
func (d *DropTargets) Flashing() bool {
	return time.Now().Before(d.flashUntil)
}

// FlashDrive returns the drive the last file dropped went to
func (d *DropTargets) FlashDrive() int {
	return d.flashDrive
}

// Snapshot returns the screen with the areas of the drives, with the one
// passed highlighted. Pass -1 to highlight none of them.
func (d *DropTargets) Snapshot(selected int) *image.RGBA {
	drives := d.a.GetRemovableMediaDrives()
	return screen.SnapshotMessageGenerator(d.a.GetVideoSource(),
		dropTargetsMessage(drives, selected), true /*is80Columns*/)
}

// DropTargetIndex returns the drive that would get a file dropped at the x
// position of a window of the given width. It returns -1 if there are no
// drives to drop a file on.
func DropTargetIndex(x int, width int, driveCount int) int {
	if driveCount <= 0 || width <= 0 {
		return -1
	}

	index := x * driveCount / width
	if index < 0 {
		return 0
	}
	if index >= driveCount {
		return driveCount - 1
	}
	return index
}

// dropTargetsMessage builds the text screen with an area per drive
func dropTargetsMessage(drives []izapple2.DriveInfo, selected int) string {
	canvas := newTextCanvas(dropTargetsColumns, dropTargetsLines)
	if len(drives) == 0 {
		canvas.putTextCentered(0, dropTargetsColumns, dropTargetsLabelLine,
			"There are no drives to drop a file on")
		return canvas.String()
	}

	canvas.putTextCentered(0, dropTargetsColumns, dropTargetsTitleLine, dropTargetsTitle)
	canvas.putTextCentered(0, dropTargetsColumns, dropTargetsFooterLine, dropTargetsFooter)

	for i, drive := range drives {
		left := i * dropTargetsColumns / len(drives)
		right := (i + 1) * dropTargetsColumns / len(drives)

		// Divide the screen, as the window is divided
		inLeft := left + 1
		if i != 0 {
			canvas.putVerticalLine(left, dropTargetsTopLine, dropTargetsBottomLine, '|')
			// Leave the same margin on both sides of the division
			inLeft = left + 2
		}
		inRight := right - 1

		canvas.putTextCentered(inLeft, inRight, dropTargetsLabelLine, drive.Label)

		media := mediaDisplayName(drive.Media)
		if media == "" {
			media = dropTargetsEmpty
		}
		lines := splitInLines(media, inRight-inLeft, dropTargetsMediaLines)
		for j, line := range lines {
			canvas.putTextCentered(inLeft, inRight, dropTargetsMediaLine+j, line)
		}

		if i == selected {
			canvas.putHorizontalLine(inLeft, inRight, dropTargetsRuleTop, '=')
			canvas.putHorizontalLine(inLeft, inRight, dropTargetsRuleBottom, '=')
		}
	}

	return canvas.String()
}

// mediaDisplayName keeps the last segment of the name of a media, the part
// that tells the images apart
func mediaDisplayName(media string) string {
	if i := strings.LastIndexAny(media, "/\\"); i != -1 {
		return media[i+1:]
	}
	return media
}

// splitInLines cuts a text in lines of at most the given columns, breaking on
// the spaces where it can. What does not fit in the lines available is
// replaced by an ellipsis.
func splitInLines(text string, columns int, maxLines int) []string {
	if columns < 1 || maxLines < 1 {
		return nil
	}

	lines := make([]string, 0, maxLines)
	for len(text) > columns {
		if len(lines) == maxLines-1 {
			return append(lines, ellipsize(text, columns))
		}

		// The first space that would be left out is where the line is cut.
		// Without one, the word is longer than the line and has to be broken.
		cut := strings.LastIndex(text[:columns+1], " ")
		if cut <= 0 {
			lines = append(lines, text[:columns])
			text = text[columns:]
			continue
		}

		lines = append(lines, strings.TrimRight(text[:cut], " "))
		text = strings.TrimLeft(text[cut:], " ")
	}
	return append(lines, text)
}

// ellipsize cuts a text to the columns given, marking that there was more
func ellipsize(text string, columns int) string {
	if len(text) <= columns {
		return text
	}
	if columns > 3 {
		return text[:columns-3] + "..."
	}
	return text[:columns]
}
