package shared

import "testing"

func TestViewToggleDropTargets(t *testing.T) {
	v := NewView()
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
