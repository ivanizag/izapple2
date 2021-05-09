package component

import (
	"testing"
)

func TestSN74LS323Reset(t *testing.T) {
	var o SN74LS323
	o.value = 0x89
	o.Update(0x12, false, false, false, false, false)
	if o.Output() != 0 {
		t.Error("Value should reset to 0")
	}
}

func TestSN74LS323ShiftLeft(t *testing.T) {
	var o SN74LS323
	o.value = 0x11
	o.Update(0x12, false, true, true, false, true)
	if o.Output() != 0x88 {
		t.Error("Bad shift left")
	}
}

func TestSN74LS323ShiftRight(t *testing.T) {
	var o SN74LS323
	o.value = 0x11
	o.Update(0x12, true, false, true, true, false)
	if o.Output() != 0x23 {
		t.Error("Bad shift right")
	}
}

func TestSN74LS323Load(t *testing.T) {
	var o SN74LS323
	o.value = 0x11
	o.Update(0x12, true, true, true, true, false)
	if o.Output() != 0x12 {
		t.Error("Bad load")
	}
}
