package shared

import (
	"strings"
	"testing"
	"time"

	"github.com/ivanizag/izapple2"
)

func TestDropTargetIndex(t *testing.T) {
	cases := []struct {
		x        int
		width    int
		drives   int
		expected int
	}{
		{0, 400, 2, 0},
		{199, 400, 2, 0},
		{200, 400, 2, 1},
		{399, 400, 2, 1},
		{0, 400, 4, 0},
		{250, 400, 4, 2},
		{-10, 400, 4, 0},  // Out of the window on the left
		{500, 400, 4, 3},  // Out of the window on the right
		{100, 400, 0, -1}, // No drives to drop a file on
		{100, 0, 2, -1},   // Window not sized yet
	}

	for _, c := range cases {
		actual := DropTargetIndex(c.x, c.width, c.drives)
		if actual != c.expected {
			t.Errorf("DropTargetIndex(%v, %v, %v) is %v, expected %v",
				c.x, c.width, c.drives, actual, c.expected)
		}
	}
}

// TestDropTargetsDrag follows a file dragged over the window and dropped,
// checking when the areas are shown and which drive they highlight
func TestDropTargetsDrag(t *testing.T) {
	d := NewDropTargets(nil)

	steps := []struct {
		name string
		do   func()

		showing  bool // With the areas not asked for
		selected int  // With the pointer on drive 2
	}{
		{"at rest", func() {}, false, 2},
		{"a file enters the window", d.DragStarted, true, -1},
		{"the file moves over drive 1", func() { d.DragMoved(1) }, true, 1},
		{"the file is dropped on drive 1", func() { d.Dropped(1); d.DragEnded() }, true, 1},
		{"the flash is over", func() { d.flashUntil = time.Now() }, false, 2},
	}

	for _, step := range steps {
		step.do()

		if actual := d.Showing(false); actual != step.showing {
			t.Errorf("%v: Showing(false) is %v, expected %v",
				step.name, actual, step.showing)
		}
		if actual := d.selected(2); actual != step.selected {
			t.Errorf("%v: selected(2) is %v, expected %v",
				step.name, actual, step.selected)
		}
	}

	// Nothing is going on by now, only the user asking shows the areas
	if !d.Showing(true) {
		t.Error("the areas asked for should be shown")
	}
}

func TestSplitInLines(t *testing.T) {
	cases := []struct {
		text     string
		columns  int
		maxLines int
		expected []string
	}{
		{"dos33.dsk", 9, 3, []string{"dos33.dsk"}},
		{"dos33.dsk", 5, 3, []string{"dos33", ".dsk"}},                        // No spaces to break on
		{"Total Replay v5.dsk", 11, 3, []string{"Total", "Replay", "v5.dsk"}}, // Broken on the spaces
		{"Total Replay v5.dsk", 11, 2, []string{"Total", "Replay v..."}},      // What is left does not fit
		{"", 8, 3, []string{""}},
	}

	for _, c := range cases {
		actual := splitInLines(c.text, c.columns, c.maxLines)
		if strings.Join(actual, "|") != strings.Join(c.expected, "|") {
			t.Errorf("splitInLines(%q, %v, %v) is %q, expected %q",
				c.text, c.columns, c.maxLines, actual, c.expected)
		}
		if len(actual) > c.maxLines {
			t.Errorf("splitInLines(%q, %v, %v) returned %v lines",
				c.text, c.columns, c.maxLines, len(actual))
		}
		for _, line := range actual {
			if len(line) > c.columns {
				t.Errorf("splitInLines(%q, %v, %v) returned the long line %q",
					c.text, c.columns, c.maxLines, line)
			}
		}
	}
}

func TestMediaDisplayName(t *testing.T) {
	cases := map[string]string{
		"dos33.dsk":                  "dos33.dsk",
		"/home/user/disks/dos33.dsk": "dos33.dsk",
		"C:\\disks\\dos33.dsk":       "dos33.dsk",
		"<internal>/dos33.dsk":       "dos33.dsk",
		"":                           "",
	}

	for media, expected := range cases {
		actual := mediaDisplayName(media)
		if actual != expected {
			t.Errorf("mediaDisplayName(%q) is %q, expected %q", media, actual, expected)
		}
	}
}

// TestDropTargetsMessageFits verifies that the screen built is not wider or
// taller than what SnapshotMessageGenerator shows, whatever the drives are
func TestDropTargetsMessageFits(t *testing.T) {
	drives := []izapple2.DriveInfo{
		{Label: "S5D1", Media: "/home/user/disks/Total Replay v5.0.2mg"},
		{Label: "S5D2", Media: ""},
		{Label: "S6D1", Media: "dos33.dsk"},
		{Label: "S6D2", Media: strings.Repeat("long", 40) + ".dsk"},
	}

	for count := 1; count <= len(drives); count++ {
		for selected := -1; selected < count; selected++ {
			message := dropTargetsMessage(drives[:count], selected)
			lines := strings.Split(message, "\n")
			if len(lines) > dropTargetsLines {
				t.Errorf("%v drives, selected %v: %v lines", count, selected, len(lines))
			}
			for i, line := range lines {
				if len(line) > dropTargetsColumns {
					t.Errorf("%v drives, selected %v: line %v is %v columns",
						count, selected, i, len(line))
				}
			}
			for _, drive := range drives[:count] {
				if !strings.Contains(message, drive.Label) {
					t.Errorf("%v drives, selected %v: %v is missing",
						count, selected, drive.Label)
				}
			}
		}
	}
}

func TestDropTargetsMessageWithoutDrives(t *testing.T) {
	message := dropTargetsMessage(nil, -1)
	if !strings.Contains(message, "no drives") {
		t.Errorf("the screen without drives should say so, it is %q", message)
	}
}
