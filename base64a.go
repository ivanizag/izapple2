package izapple2

import (
	"fmt"

	"github.com/ivanizag/izapple2/core6502"
	"github.com/ivanizag/izapple2/storage"
)

/*
	Copam BASE64A adaptation.
*/

func setBase64a(a *Apple2) {
	a.Name = "Base 64A"
	a.cpu = core6502.NewNMOS6502(a.mmu)
	addApple2SoftSwitches(a.io)
	addBase64aSoftSwitches(a.io)
}

const (
	// There are 6 ROM chips. Each can have 4Kb or 8Kb. They can fill
	// 2 or 4 banks with 2kb windows.
	base64aRomBankSize   = 12 * 1024
	base64aRomBankCount  = 4
	base64aRomWindowSize = 2 * 1024
	base64aRomChipCount  = 6
)

func loadBase64aRom(a *Apple2) error {
	// Load the 6 PROM dumps
	romBanksBytes := make([][]uint8, base64aRomBankCount)
	for j := range romBanksBytes {
		romBanksBytes[j] = make([]uint8, 0, base64aRomBankSize)
	}

	for i := 0; i < base64aRomChipCount; i++ {
		filename := fmt.Sprintf("<internal>/BASE64A_%X.BIN", 0xd0+i*0x08)
		data, _, err := storage.LoadResource(filename)
		if err != nil {
			return err
		}
		for j := range romBanksBytes {
			start := (j * base64aRomWindowSize) % len(data)
			romBanksBytes[j] = append(romBanksBytes[j], data[start:start+base64aRomWindowSize]...)
		}
	}

	// Create banks
	for j := range romBanksBytes {
		a.mmu.physicalROM[j] = newMemoryRange(0xd000, romBanksBytes[j], fmt.Sprintf("Base64 ROM page %v", j))
	}

	// Start with first bank active
	a.mmu.setActiveROMPage(0)

	return nil
}

func addBase64aSoftSwitches(io *ioC0Page) {
	// Other softswitches, not implemented but called from the ROM
	io.addSoftSwitchW(0x0C, notImplementedSoftSwitchW, "80COLOFF")
	io.addSoftSwitchW(0x0E, notImplementedSoftSwitchW, "ALTCHARSETOFF")

	// Write on the speaker. That is a double access and should do nothing
	// but works somehow on the BASE64A
	io.addSoftSwitchW(0x30, func(io *ioC0Page, value uint8) {
		speakerSoftSwitch(io)
	}, "SPEAKER")

	// ROM pagination softswitches. They use the annunciator 0 and 1
	mmu := io.apple2.mmu
	io.addSoftSwitchRW(0x58, func(*ioC0Page) uint8 {
		p := mmu.getActiveROMPage()
		mmu.setActiveROMPage(p & 2)
		return 0
	}, "ANN0OFF-ROM")
	io.addSoftSwitchRW(0x59, func(*ioC0Page) uint8 {
		p := mmu.getActiveROMPage()
		mmu.setActiveROMPage(p | 1)
		return 0
	}, "ANN0ON-ROM")
	io.addSoftSwitchRW(0x5A, func(*ioC0Page) uint8 {
		p := mmu.getActiveROMPage()
		mmu.setActiveROMPage(p & 1)
		return 0
	}, "ANN1OFF-ROM")
	io.addSoftSwitchRW(0x5B, func(*ioC0Page) uint8 {
		p := mmu.getActiveROMPage()
		mmu.setActiveROMPage(p | 2)
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
