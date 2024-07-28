package izapple2

/*
	Copam BASE64A adaptation.
*/

const (
	// There are 6 ROM chips. Each can have 4Kb or 8Kb. They can fill
	// 2 or 4 banks with 2kb windows.
	base64aRomBankSize   = 12 * 1024
	base64aRomBankCount  = 4
	base64aRomWindowSize = 2 * 1024
	base64aRomChipCount  = 6
)

func loadBase64aRom(a *Apple2) error {
	return loadMultiPageRom(a, []string{
		"<internal>/BASE64A_D0.BIN",
		"<internal>/BASE64A_D8.BIN",
		"<internal>/BASE64A_E0.BIN",
		"<internal>/BASE64A_E8.BIN",
		"<internal>/BASE64A_F0.BIN",
		"<internal>/BASE64A_F8.BIN",
	})
}

func addBase64aSoftSwitches(io *ioC0Page) {
	// Other softswitches, not implemented but called from the ROM
	io.addSoftSwitchW(0x0C, buildNotImplementedSoftSwitchW(io), "80COLOFF")
	io.addSoftSwitchW(0x0E, buildNotImplementedSoftSwitchW(io), "ALTCHARSETOFF")

	// ROM pagination softswitches. They use the annunciator 0 and 1
	mmu := io.apple2.mmu
	io.addSoftSwitchRW(0x58, func() uint8 {
		if rom, ok := mmu.physicalROM.(*memoryRangeROM); ok {
			p := rom.getPage()
			rom.setPage(p & 2)
		}
		return 0
	}, "ANN0OFF-ROM")
	io.addSoftSwitchRW(0x59, func() uint8 {
		if rom, ok := mmu.physicalROM.(*memoryRangeROM); ok {
			p := rom.getPage()
			rom.setPage(p | 1)
		}
		return 0
	}, "ANN0ON-ROM")
	io.addSoftSwitchRW(0x5A, func() uint8 {
		if rom, ok := mmu.physicalROM.(*memoryRangeROM); ok {
			p := rom.getPage()
			rom.setPage(p & 1)
		}
		return 0
	}, "ANN1OFF-ROM")
	io.addSoftSwitchRW(0x5B, func() uint8 {
		if rom, ok := mmu.physicalROM.(*memoryRangeROM); ok {
			p := rom.getPage()
			rom.setPage(p | 2)
		}
		return 0
	}, "ANN1ON-ROM")

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
