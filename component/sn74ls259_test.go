package component

import (
	"testing"
)

func TestSN74LS259WriteTrue(t *testing.T) {
	var o SN74LS259
	o.Write(2, true, true)
	if !o.Q(2) {
		t.Error("Wrote true but got false")
	}
}

func TestSN74LS259WriteFalse(t *testing.T) {
	var o SN74LS259
	o.Write(1, false, true)
	if o.Q(1) {
		t.Error("Wrote false but got true")
	}
}

func TestSN74LS259WriteTrueThenFalse(t *testing.T) {
	var o SN74LS259
	o.Write(5, true, true)
	o.Write(5, false, true)
	if o.Q(5) {
		t.Error("Wrote true the false but got true")
	}
}
