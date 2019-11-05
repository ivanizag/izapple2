package apple2

/*
 See:
   https://www.apple.asimov.net/documentation/hardware/machines/APPLE%20IIe%20Auxiliary%20Memory%20Softswitches.pdf
*/

const (
	ioFlag80Store uint8 = 0x18
	ioFlagAltChar uint8 = 0x1E
	ioFlag80Col   uint8 = 0x1F
	// ??? ioVertBlank uin8 = 0x19
)

func addApple2ESoftSwitches(io *ioC0Page) {
	// New MMU read softswithes
	mmu := io.apple2.mmu
	addSoftSwitchesMmu(io, 0x02, 0x03, 0x13, &mmu.altMainRAMActiveRead, "RAMRD")
	addSoftSwitchesMmu(io, 0x04, 0x05, 0x14, &mmu.altMainRAMActiveWrite, "RAMWRT")
	addSoftSwitchesMmu(io, 0x06, 0x07, 0x15, &mmu.intCxROMActive, "INTCXROM")
	addSoftSwitchesMmu(io, 0x08, 0x09, 0x16, &mmu.altZeroPage, "ALTZP")
	addSoftSwitchesMmu(io, 0x0a, 0x0b, 0x17, &mmu.slotC3ROMActive, "SLOTC3ROM")

	// New IOU read softswithes
	addSoftSwitchesIou(io, 0x00, 0x01, 0x18, ioFlag80Store, "80STORE")
	addSoftSwitchesIou(io, 0x0c, 0x0d, 0x1f, ioFlag80Col, "80COL")
	addSoftSwitchesIou(io, 0x0e, 0x0f, 0x1e, ioFlagAltChar, "ALTCHARSET")

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

	// TOOD:
	// AKD read on 0x10
	// VBL read on 0x19

	//io.softSwitchesData[ioFlagAltChar] = ssOn // Not sure about this.

}

func addSoftSwitchesMmu(io *ioC0Page, addressClear uint8, addressSet uint8, AddressGet uint8, flag *bool, name string) {
	io.addSoftSwitchW(addressClear, func(_ *ioC0Page, _ uint8) {
		*flag = false
	}, name+"OFF")

	io.addSoftSwitchW(addressSet, func(_ *ioC0Page, _ uint8) {
		*flag = true
	}, name+"ON")

	io.addSoftSwitchR(AddressGet, func(_ *ioC0Page) uint8 {
		return ssFromBool(*flag)
	}, name)
}

func addSoftSwitchesIou(io *ioC0Page, addressClear uint8, addressSet uint8, AddressGet uint8, ioFlag uint8, name string) {
	io.addSoftSwitchW(addressClear, func(_ *ioC0Page, _ uint8) {
		io.softSwitchesData[ioFlag] = ssOff
	}, name+"OFF")

	io.addSoftSwitchW(addressSet, func(_ *ioC0Page, _ uint8) {
		io.softSwitchesData[ioFlag] = ssOn
	}, name+"ON")

	io.addSoftSwitchR(AddressGet, func(_ *ioC0Page) uint8 {
		return io.softSwitchesData[ioFlag]
	}, name)
}
