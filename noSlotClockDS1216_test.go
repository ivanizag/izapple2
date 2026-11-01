package izapple2

import (
	"testing"
)

/*
nscReadProgram talks to the No-Slot Clock under the main ROM as the drivers
do, through the internal $C3xx page: a read of $C304 resets the pattern
matching, the 64 bits of the pattern go on A0 with reads of $C300 and $C301,
and the 64 bits of the time come back on bit 0 of reads of $C304. They are
kept on $0380 to $0387, least significant byte first. The pattern is on
$0360.
*/
var nscReadProgram = []uint8{
	0x8d, 0x0a, 0xc0, // STA $C00A, the internal ROM on $C3xx
	0xad, 0x04, 0xc3, // LDA $C304, reset
	0xa2, 0x00, // LDX #0
	0xbd, 0x60, 0x03, // LDA $0360,X, a byte of the pattern
	0xa0, 0x08, // LDY #8
	0x4a,       // LSR A
	0x90, 0x05, // BCC to BIT $C300
	0x2c, 0x01, 0xc3, // BIT $C301, a one
	0x80, 0x03, // BRA to DEY
	0x2c, 0x00, 0xc3, // BIT $C300, a zero
	0x88,       // DEY
	0xd0, 0xf2, // BNE to LSR A
	0xe8,       // INX
	0xe0, 0x08, // CPX #8
	0xd0, 0xe8, // BNE to LDA $0360,X
	0xa2, 0x00, // LDX #0
	0xa0, 0x08, // LDY #8
	0xad, 0x04, 0xc3, // LDA $C304, a bit of the time
	0x4a,             // LSR A
	0x7e, 0x80, 0x03, // ROR $0380,X
	0x88,       // DEY
	0xd0, 0xf6, // BNE to LDA $C304
	0xe8,       // INX
	0xe0, 0x08, // CPX #8
	0xd0, 0xef, // BNE to LDY #8
	0x80, 0xfe, // BRA *
}

var nscPattern = []uint8{0xc5, 0x3a, 0xa3, 0x5c, 0xc5, 0x3a, 0xa3, 0x5c}

func TestNoSlotClockReadByAProgram(t *testing.T) {
	a, err := CreateAppleFromModel("2enh", map[string]string{
		"s6":        "empty",
		"nsc":       "main",
		"startTime": "1985-03-15T10:30:00",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	a.Init()
	a.RunCycles(1_000_000)
	a.mmu.pokeRange(0x0300, nscReadProgram)
	a.mmu.pokeRange(0x0360, nscPattern)
	a.cpu.SetPC(0x0300)
	a.RunCycles(100_000)

	// 85-03-15, the 6th day of the week, 10:30, the seconds go on
	expected := []uint8{0x85, 0x03, 0x15, 0x06, 0x10, 0x30}
	for i, value := range expected {
		if got := a.mmu.Peek(0x0387 - uint16(i)); got != value {
			t.Errorf("expected $%02x on the byte %v of the time, got $%02x", value, i, got)
		}
	}
}
