package shared

import (
	"testing"

	"github.com/ivanizag/izapple2/screen"
)

func TestViewToggleDropTargets(t *testing.T) {
	v := NewView(screen.ScreenModeColorScanlines)
	v.ToggleHelp()

	v.ToggleDropTargets()
	if !v.ShowDropTargets {
		t.Error("the drop targets should be shown")
	}
	if v.ShowHelp {
		t.Error("the help should be out of the way of the drop targets")
	}

	v.ToggleDropTargets()
	if v.ShowDropTargets {
		t.Error("the drop targets should be hidden")
	}
}
