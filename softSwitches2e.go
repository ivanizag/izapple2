package apple2

/*
 See:
   https://www.apple.asimov.net/documentation/hardware/machines/APPLE%20IIe%20Auxiliary%20Memory%20Softswitches.pdf
*/

const (
	ioFlagAltChar uint8 = 0x1E
	ioFlag80Col   uint8 = 0x1F
)

const (
	screenDrawCycles         = uint64(12480 + 4550)
	screenVertBlankingCycles = uint64(4550)
)

func addApple2ESoftSwitches(io *ioC0Page) {
	// New MMU read softswithes
	mmu := io.apple2.mmu
	io.addSoftSwitchesMmu(0x02, 0x03, 0x13, &mmu.altMainRAMActiveRead, "RAMRD")
	io.addSoftSwitchesMmu(0x04, 0x05, 0x14, &mmu.altMainRAMActiveWrite, "RAMWRT")
	io.addSoftSwitchesMmu(0x06, 0x07, 0x15, &mmu.intCxROMActive, "INTCXROM")
	io.addSoftSwitchesMmu(0x08, 0x09, 0x16, &mmu.altZeroPage, "ALTZP")
	io.addSoftSwitchesMmu(0x0a, 0x0b, 0x17, &mmu.slotC3ROMActive, "SLOTC3ROM")
	io.addSoftSwitchesMmu(0x00, 0x01, 0x18, &mmu.store80Active, "80STORE")

	// New IOU read softswithes
	io.addSoftSwitchesIou(0x0c, 0x0d, 0x1f, ioFlag80Col, "80COL")
	io.addSoftSwitchesIou(0x0e, 0x0f, 0x1e, ioFlagAltChar, "ALTCHARSET")

	// Previous read softswithes
	io.addSoftSwitchR(0x1A, getStatusSoftSwitch(ioFlagText), "TEXT")
	io.addSoftSwitchR(0x1B, getStatusSoftSwitch(ioFlagMixed), "MIXED")
	io.addSoftSwitchR(0x1C, getStatusSoftSwitch(ioFlagSecondPage), "PAGE2")
	io.addSoftSwitchR(0x1D, getStatusSoftSwitch(ioFlagHiRes), "HIRES")

	io.addSoftSwitchR(0x11, func(_ *ioC0Page) uint8 {
		return ssFromBool(mmu.lcAltBank)
	}, "BSRBANK2")
	io.addSoftSwitchR(0x12, func(_ *ioC0Page) uint8 {
		return ssFromBool(mmu.lcActiveRead)
	}, "BSRREADRAM")

	io.addSoftSwitchR(0x19, func(_ *ioC0Page) uint8 {
		// See "Inside Apple IIe", page 268
		// See http://rich12345.tripod.com/aiivideo/vbl.html
		// For each screen draw:
		//      12480 cycles drawing lines, VERTBLANK = $00
		//       4550 cycles doing the return to position (0,0), VERTBLANK = $80
		// Vert blank takes 12480 cycles every page redraw
		cycles := io.apple2.cpu.GetCycles() % screenDrawCycles
		if cycles <= screenVertBlankingCycles {
			return ssOn
		}
		return ssOff
	}, "VERTBLANK")

	//io.softSwitchesData[ioFlagAltChar] = ssOn // Not sure about this.

}
