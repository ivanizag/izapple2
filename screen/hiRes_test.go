package screen

import (
	"testing"
)

func TestGetHiResLineOffest(t *testing.T) {
	scenarios := map[int]uint16{
		0:   0x2000,
		1:   0x2400,
		8:   0x2080,
		63:  0x3f80,
		64:  0x2028,
		128: 0x2050,
		191: 0x3fd0,
	}

	for in, want := range scenarios {
		got := 0x2000 + getHiResLineOffset(in)
		if want != got {
			t.Errorf("expected %x but got %x for line %v", want, got, in)
		}
	}
}
