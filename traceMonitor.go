package izapple2

import (
	"fmt"
)

/*

	Trace the inpit and output of using the wozmon calls.

*/

type traceMonitor struct {
	a             *Apple2
	closingBuffer bool
	buffer        string
}

const (
	//wozmonPrompt uint16 = 0x0033

	wozmonGETLNZ      uint16 = 0xfd67
	woamonGETLN       uint16 = 0xfd6a
	wozmonGETLNReturn uint16 = 0xfd90
	wozmonRDCHAR      uint16 = 0xfd35
	wozmonRDKEY       uint16 = 0xfd0c
	wozmonKEYIN       uint16 = 0xfd1b

	wozmonCOUT  uint16 = 0xfded
	wozmonCOUT1 uint16 = 0xfdf0
	wozmonCOUTZ uint16 = 0xfdf6
)

func newTraceMonitor() *traceMonitor {
	var t traceMonitor
	return &t
}

func (t *traceMonitor) connect(a *Apple2) {
	t.a = a
}

func (t *traceMonitor) inspect() {
	if t.a.dmaActive {
		return
	}

	if t.a.mmu.altMainRAMActiveRead {
		// We want to trace only the activity on the ROM
		return
	}

	pc, _ := t.a.cpu.GetPCAndSP()
	a, _, _, _ := t.a.cpu.GetAXYP()

	desc := ""
	switch pc {
	case wozmonGETLNZ:
		desc = "GETLNZ"
	case woamonGETLN:
		fmt.Printf("Wozmon output: %s\n", t.buffer)
		t.buffer = ""
		desc = "GETLN"
	case wozmonGETLNReturn:
		t.closingBuffer = true
		//desc = "GETLN return"
	case wozmonRDKEY:
		//desc = "RDKEY"
	case wozmonKEYIN:
		//desc = "KEYIN"
	case wozmonCOUT:
		//desc = fmt.Sprintf("COUT  0x%02x %c", a, toAscii(a))
		if t.closingBuffer {
			fmt.Printf("Wozmon input: %s\n", t.buffer)
			t.buffer = ""
			t.closingBuffer = false
			desc = fmt.Sprintf("GETLN returns <<%s>>", t.getInputBuffer())
		}
	case wozmonCOUT1:
		t.buffer += string(toAscii(a))
		//desc = fmt.Sprintf("COUT1 0x%02x %c", a, toAscii(a))
	case wozmonCOUTZ:
		desc = "COUTZ"
	}

	if desc != "" {
		fmt.Printf("Wozmon call to $%04x %s\n", pc, desc)
	}
}

func toAscii(b uint8) rune {
	b = b & 0x7f
	if b < 0x20 {
		return rune(uint16(b) + 0x2400)
	}
	return rune(b)
}

func (t *traceMonitor) getInputBuffer() string {
	buffer := ""
	for address := uint16(0x200); address < 0x300; address++ {
		b := t.a.mmu.Peek(address)
		buffer += string(toAscii(b))
		if b == 0x8d {
			break
		}
	}

	return buffer
}
