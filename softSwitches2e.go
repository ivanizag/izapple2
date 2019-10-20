package apple2

/*
 See:
   https://www.apple.asimov.net/documentation/hardware/machines/APPLE%20IIe%20Auxiliary%20Memory%20Softswitches.pdf
*/

const (
	ioFlagRamRd     uint8 = 0x13
	ioFlagRamWrt    uint8 = 0x14
	ioFlagIntCxRom  uint8 = 0x15
	ioFlagAltZp     uint8 = 0x16
	ioFlagSlotC3Rom uint8 = 0x17
	ioFlag80Store   uint8 = 0x18
	ioFlagAltChar   uint8 = 0x1E
	ioFlag80Col     uint8 = 0x1F
	// ??? ioVertBlank uin8 = 0x19
)

func addApple2ESoftSwitches(io *ioC0Page) {
	// New MMU read softswithes
	io.addSoftSwitchW(0x02, getSoftSwitchExt(ioFlagRamRd, ssOff, nil), "RAMRDOFF")
	io.addSoftSwitchW(0x03, getSoftSwitchExt(ioFlagRamWrt, ssOn, nil), "RAMRDON")
	io.addSoftSwitchR(0x13, getStatusSoftSwitch(ioFlagRamWrt), "RAMRD")

	io.addSoftSwitchW(0x04, getSoftSwitchExt(ioFlagRamWrt, ssOff, nil), "RAMWRTOFF")
	io.addSoftSwitchW(0x05, getSoftSwitchExt(ioFlagRamWrt, ssOn, nil), "RAMWRTON")
	io.addSoftSwitchR(0x14, getStatusSoftSwitch(ioFlagRamWrt), "RAMWRT")

	io.addSoftSwitchW(0x06, getSoftSwitchExt(ioFlagIntCxRom, ssOff, softSwitchIntCxRomOff), "INTCXROMOFF")
	io.addSoftSwitchW(0x07, getSoftSwitchExt(ioFlagIntCxRom, ssOn, softSwitchIntCxRomOn), "INTCXROMON")
	io.addSoftSwitchR(0x15, getStatusSoftSwitch(ioFlagIntCxRom), "INTCXROM")

	io.addSoftSwitchW(0x08, getSoftSwitchExt(ioFlagAltZp, ssOff, nil), "ALTZPOFF")
	io.addSoftSwitchW(0x09, getSoftSwitchExt(ioFlagAltZp, ssOn, nil), "ALTZPON")
	io.addSoftSwitchR(0x16, getStatusSoftSwitch(ioFlagAltZp), "ALTZP")

	io.addSoftSwitchW(0x0A, getSoftSwitchExt(ioFlagSlotC3Rom, ssOff, softSwitchSlotC3RomOff), "SLOTC3ROMOFF")
	io.addSoftSwitchW(0x0B, getSoftSwitchExt(ioFlagSlotC3Rom, ssOn, softSwitchSlotC3RomOn), "SLOTC3ROMON")
	io.addSoftSwitchR(0x17, getStatusSoftSwitch(ioFlagSlotC3Rom), "SLOTC3ROM")

	// Previous read softswithes
	io.addSoftSwitchR(0x1A, getStatusSoftSwitch(ioFlagText), "TEXT")
	io.addSoftSwitchR(0x1B, getStatusSoftSwitch(ioFlagMixed), "MIXED")
	io.addSoftSwitchR(0x1C, getStatusSoftSwitch(ioFlagSecondPage), "PAGE2")
	io.addSoftSwitchR(0x1D, getStatusSoftSwitch(ioFlagHiRes), "HIRES")

	// New IOU read softswithes
	io.addSoftSwitchW(0x00, getSoftSwitchExt(ioFlag80Store, ssOff, nil), "80STOREOFF")
	io.addSoftSwitchW(0x01, getSoftSwitchExt(ioFlag80Store, ssOn, nil), "80STOREON")
	io.addSoftSwitchR(0x18, getStatusSoftSwitch(ioFlag80Store), "80STORE")

	io.addSoftSwitchW(0x0C, getSoftSwitchExt(ioFlag80Col, ssOff, nil), "80COLOFF")
	io.addSoftSwitchW(0x0D, getSoftSwitchExt(ioFlag80Col, ssOn, nil), "80COLON")
	io.addSoftSwitchR(0x1F, getStatusSoftSwitch(ioFlag80Col), "80COL")

	io.addSoftSwitchW(0x0E, getSoftSwitchExt(ioFlagAltChar, ssOff, nil), "ALTCHARSETOFF")
	io.addSoftSwitchW(0x0F, getSoftSwitchExt(ioFlagAltChar, ssOn, nil), "ALTCHARSETON")
	io.addSoftSwitchR(0x1E, getStatusSoftSwitch(ioFlagAltChar), "ALTCHARSET")

	// TOOD:
	// AKD read on 0x10
	// VBL read on 0x19

	//io.softSwitchesData[ioFlagAltChar] = ssOn // Not sure about this.

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
	io.apple2.mmu.setPagesRead(0xc1, 0xcf, io.apple2.mmu.physicalROMe)
}

func softSwitchIntCxRomOff(io *ioC0Page) {
	// TODO restore all the ROM from the slots for 0xc1 to 0xc7
	io.apple2.mmu.setPages(0xc1, 0xc7, nil)
}

func softSwitchSlotC3RomOn(io *ioC0Page) {
	// TODO restore the slot 3 ROM
	io.apple2.mmu.setPages(0xc3, 0xc3, nil)
}

func softSwitchSlotC3RomOff(io *ioC0Page) {
	io.apple2.mmu.setPagesRead(0xc3, 0xc3, io.apple2.mmu.physicalROMe)
}

// TODO: apply state after persistance load
