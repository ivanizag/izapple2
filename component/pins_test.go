package component

import (
	"testing"
)

func TestByteToPins(t *testing.T) {
	v := uint8(0b10010110)
	a := [8]bool{
		false, true, true, false,
		true, false, false, true,
	}

	if a != ByteToPins(v) {
		t.Error("Error on byte to pins")
	}
}

func TestPinsToByte(t *testing.T) {
	v := uint8(0b10010110)
	a := [8]bool{
		false, true, true, false,
		true, false, false, true,
	}

	if v != PinsToByte(a) {
		t.Error("Error on pins to byte")
	}
}
