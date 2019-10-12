package apple2

import (
	"github.com/ivanizag/apple2/core6502"
)

/*
	Copam BASE64A adaptation.
*/

// newBase64a instantiates an apple2
func newBase64a() *Apple2 {
	var a Apple2

	a.Name = "Base 64A"
	a.mmu = newMemoryManager(&a)
	a.cpu = core6502.NewNMOS6502(a.mmu)

	// Set the io in 0xc000
	a.io = newIoC0Page(&a)
	a.mmu.setPages(0xc0, 0xc0, a.io)
	addApple2SoftSwitches(a.io)
	addBase64aSoftSwitches(a.io)

	return &a
}

func addBase64aSoftSwitches(io *ioC0Page) {
	// Other softswitches
	io.addSoftSwitchW(0x0C, notImplementedSoftSwitchW) // 80 columns off?
	io.addSoftSwitchW(0x0E, notImplementedSoftSwitchW) // Alt char off?

	// Write on the speaker. That is a double access and should do nothing
	// but works somehow on the BASE64A
	io.addSoftSwitchW(0x30, func(io *ioC0Page, value uint8) {
		speakerSoftSwitch(io)
	})
}

func charGenColumnsMapBase64a(column int) int {
	bit := column + 2
	// Weird positions
	if column == 6 {
		bit = 2
	} else if column == 0 {
		bit = 1
	}
	return bit
}
