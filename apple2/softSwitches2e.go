package apple2

const (
	ioFlagIntCxRom  uint8 = 0x15
	ioFlagSlotC3Rom uint8 = 0x17
	ioFlag80Store   uint8 = 0x18
	ioFlag80Col     uint8 = 0x1F
)

func addApple2ESoftSwitches(mmu *memoryManager) {
	ss := &mmu.ioPage.softSwitches

	ss[0x00] = getSoftSwitchExt(ioFlag80Store, ssOff, nil)
	ss[0x01] = getSoftSwitchExt(ioFlag80Store, ssOn, nil)
	ss[0x06] = getSoftSwitchExt(ioFlagIntCxRom, ssOff, softSwitchIntCxRomOff)
	ss[0x07] = getSoftSwitchExt(ioFlagIntCxRom, ssOn, softSwitchIntCxRomOn)
	ss[0x0A] = getSoftSwitchExt(ioFlagSlotC3Rom, ssOff, softSwitchSlotC3RomOff)
	ss[0x0B] = getSoftSwitchExt(ioFlagSlotC3Rom, ssOn, softSwitchSlotC3RomOn)
	ss[0x0C] = getSoftSwitchExt(ioFlag80Col, ssOff, nil)
	ss[0x0D] = getSoftSwitchExt(ioFlag80Col, ssOn, nil)

	ss[0x15] = getStatusSoftSwitch(ioFlagIntCxRom)
	ss[0x17] = getStatusSoftSwitch(ioFlagSlotC3Rom)
	ss[0x18] = getStatusSoftSwitch(ioFlag80Store)
	ss[0x1C] = getStatusSoftSwitch(ioFlagSecondPage)
	ss[0x1F] = getStatusSoftSwitch(ioFlag80Col)
}

type softSwitchExtAction func(io *ioC0Page)

func getSoftSwitchExt(ioFlag uint8, dstValue uint8, action softSwitchExtAction) softSwitch {
	return func(io *ioC0Page, isWrite bool, value uint8) uint8 {
		//fmt.Printf("Softswitch 0x%02x %v %v\n", ioFlag, isWrite, dstValue)
		if !isWrite {
			return 0 // New Apple2e softswitches ignore reads
		}
		currentValue := io.softSwitchesData[ioFlag]
		if currentValue == dstValue {
			return 0 // Already switched, ignore
		}
		if action != nil {
			action(io)
		}
		io.softSwitchesData[ioFlag] = value
		return 0
	}
}

func softSwitchIntCxRomOn(io *ioC0Page) {
	for i := uint8(1); i < 16; i++ {
		io.mmu.activeMemory.SetPage(uint8(0xc0+i), &io.mmu.physicalROMe[i])
	}
}

func softSwitchIntCxRomOff(io *ioC0Page) {
	// TODO restore all the ROM from the slot for 0xc1 to 0xc7
	for i := 1; i < 16; i++ {
		io.mmu.activeMemory.SetPage(uint8(0xc0+i), &io.mmu.unassignedExpansionROM[i])
	}
}

func softSwitchSlotC3RomOn(io *ioC0Page) {
	if io.isSoftSwitchExtActive(ioFlagIntCxRom) {
		return // Ignore if allt the Apple2 shadow ROM is active
	}
	// TODO restore the slot 3 ROM
	io.mmu.activeMemory.SetPage(0xC3, &io.mmu.unassignedExpansionROM[3])
}

func softSwitchSlotC3RomOff(io *ioC0Page) {
	if io.isSoftSwitchExtActive(ioFlagIntCxRom) {
		return // Ignore if allt the Apple2 shadow ROM is active
	}
	io.mmu.activeMemory.SetPage(0xC3, &io.mmu.physicalROMe[3])
}
