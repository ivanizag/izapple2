package apple2

import "fmt"

const (
	ioFlagIntCxRom  uint8 = 0x15
	ioFlagSlotC3Rom uint8 = 0x17
)

func addApple2ESoftSwitches(mmu *memoryManager) {
	ss := &mmu.ioPage.softSwitches
	ss[0x06] = getSoftSwitchExt(ioFlagIntCxRom, 0x00, softSwitchIntCxRomOff)
	ss[0x07] = getSoftSwitchExt(ioFlagIntCxRom, 0x80, softSwitchIntCxRomOn)
	ss[0x15] = getStatusSoftSwitchExt(ioFlagIntCxRom)

	ss[0x0A] = getSoftSwitchExt(ioFlagSlotC3Rom, 0x00, softSwitchSlotC3RomOff)
	ss[0x0B] = getSoftSwitchExt(ioFlagSlotC3Rom, 0x80, softSwitchSlotC3RomOn)
	ss[0x17] = getStatusSoftSwitchExt(ioFlagSlotC3Rom)

	ss[0x1c] = getStatusSoftSwitchExt(ioFlagSecondPage)

}

type softSwitchExtAction func(io *ioC0Page)

func getStatusSoftSwitchExt(ioFlag uint8) softSwitch {
	return func(io *ioC0Page, isWrite bool, value uint8) uint8 {
		return io.softSwitchesData[ioFlag]
	}
}

func getSoftSwitchExt(ioFlag uint8, dstValue uint8, action softSwitchExtAction) softSwitch {
	return func(io *ioC0Page, isWrite bool, value uint8) uint8 {
		fmt.Printf("Softswitch 0x%02x %v %v\n", ioFlag, isWrite, dstValue)
		if !isWrite {
			return 0 // Ignore reads
		}
		currentValue := io.softSwitchesData[ioFlag]
		if currentValue == dstValue {
			return 0 // Already switched, ignore
		}
		action(io)
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
	for i := 1; i <= 16; i++ {
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
