package apple2

const (
	ioFlagIntCxRom  uint8 = 0x15
	ioFlagSlotC3Rom uint8 = 0x17
	ioFlag80Store   uint8 = 0x18
	ioFlagAltChar   uint8 = 0x19
	ioFlag80Col     uint8 = 0x1F
)

func addApple2ESoftSwitches(io *ioC0Page) {

	io.addSoftSwitchW(0x00, getSoftSwitchExt(ioFlag80Store, ssOff, nil))
	io.addSoftSwitchW(0x01, getSoftSwitchExt(ioFlag80Store, ssOn, nil))
	io.addSoftSwitchW(0x06, getSoftSwitchExt(ioFlagIntCxRom, ssOff, softSwitchIntCxRomOff))
	io.addSoftSwitchW(0x07, getSoftSwitchExt(ioFlagIntCxRom, ssOn, softSwitchIntCxRomOn))
	io.addSoftSwitchW(0x0A, getSoftSwitchExt(ioFlagSlotC3Rom, ssOff, softSwitchSlotC3RomOff))
	io.addSoftSwitchW(0x0B, getSoftSwitchExt(ioFlagSlotC3Rom, ssOn, softSwitchSlotC3RomOn))
	io.addSoftSwitchW(0x0C, getSoftSwitchExt(ioFlag80Col, ssOff, nil))
	io.addSoftSwitchW(0x0D, getSoftSwitchExt(ioFlag80Col, ssOn, nil))
	io.addSoftSwitchW(0x0E, getSoftSwitchExt(ioFlagAltChar, ssOff, nil))
	io.addSoftSwitchW(0x0F, getSoftSwitchExt(ioFlagAltChar, ssOn, nil))
	io.softSwitchesData[ioFlagAltChar] = ssOn // Not sure about this.

	io.addSoftSwitchR(0x15, getStatusSoftSwitch(ioFlagIntCxRom))
	io.addSoftSwitchR(0x17, getStatusSoftSwitch(ioFlagSlotC3Rom))
	io.addSoftSwitchR(0x18, getStatusSoftSwitch(ioFlag80Store))
	io.addSoftSwitchR(0x1C, getStatusSoftSwitch(ioFlagSecondPage))
	io.addSoftSwitchR(0x1F, getStatusSoftSwitch(ioFlag80Col))
}

type softSwitchExtAction func(io *ioC0Page)

func getSoftSwitchExt(ioFlag uint8, dstValue uint8, action softSwitchExtAction) softSwitchW {
	return func(io *ioC0Page, _ uint8) {
		currentValue := io.softSwitchesData[ioFlag]
		if currentValue == dstValue {
			return // Already switched, ignore
		}
		if action != nil {
			action(io)
		}
		io.softSwitchesData[ioFlag] = dstValue
	}
}

func softSwitchIntCxRomOn(io *ioC0Page) {
	mmu := io.apple2.mmu
	for i := uint8(1); i < 16; i++ {
		mmu.setPage(uint8(0xc0+i), &mmu.physicalROMe[i])
	}
}

func softSwitchIntCxRomOff(io *ioC0Page) {
	// TODO restore all the ROM from the slot for 0xc1 to 0xc7
	mmu := io.apple2.mmu
	for i := 1; i < 16; i++ {
		mmu.setPage(uint8(0xc0+i), &mmu.unassignedExpansionROM[i])
	}
}

func softSwitchSlotC3RomOn(io *ioC0Page) {
	if io.isSoftSwitchExtActive(ioFlagIntCxRom) {
		return // Ignore if allt the Apple2 shadow ROM is active
	}
	// TODO restore the slot 3 ROM
	mmu := io.apple2.mmu
	mmu.setPage(0xC3, &mmu.unassignedExpansionROM[3])
}

func softSwitchSlotC3RomOff(io *ioC0Page) {
	if io.isSoftSwitchExtActive(ioFlagIntCxRom) {
		return // Ignore if allt the Apple2 shadow ROM is active
	}
	mmu := io.apple2.mmu
	mmu.setPage(0xC3, &mmu.physicalROMe[3])
}
