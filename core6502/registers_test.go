package core6502

import "testing"

func TestRegA(t *testing.T) {
	var r registers
	data := uint8(200)
	r.setA(data)
	if r.getA() != data {
		t.Error("Error storing and loading A")
	}
}
func TestRegPC(t *testing.T) {
	var r registers
	data := uint16(0xc600)
	r.setPC(data)
	if r.getPC() != data {
		t.Error("Error storing and loading PC")
	}
}

func TestFlags(t *testing.T) {
	var r registers
	r.setP(0x23)
	if r.getP() != 0x23 {
		t.Error("Error storing and loading P")
	}

	r.setP(0)
	r.setFlag(flagD)
	if !r.getFlag(flagD) {
		t.Error("Error setting and getting flag")
	}

	r.clearFlag(flagD)
	if r.getFlag(flagD) {
		t.Error("Error clearing flag")
	}

	r.updateFlag(flagD, true)
	if !r.getFlag(flagD) {
		t.Error("Error update flag to true")
	}

	r.updateFlag(flagD, false)
	if r.getFlag(flagD) {
		t.Error("Error updating flag to false")
	}
}

func TestUpdateFlagZN(t *testing.T) {
	var r registers
	r.updateFlagZN(0)
	if r.getP() != flagZ {
		t.Error("Error update flags ZN with 0")
	}

	r.updateFlagZN(0x10)
	if r.getP() != 0 {
		t.Error("Error update flags ZN with 0x10")
	}

	r.updateFlagZN(0xF2)
	if r.getP() != flagN {
		t.Error("Error update flags ZN with 0xF2")
	}
}
