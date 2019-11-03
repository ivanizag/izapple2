package apple2

import (
	"fmt"
)

/*
	Copam BASE64A uses paginated ROM
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
	// Load the 6 PROM dumps
	romBanksBytes := make([][]uint8, base64aRomBankCount)
	for j := range romBanksBytes {
		romBanksBytes[j] = make([]uint8, 0, base64aRomBankSize)
	}

	for i := 0; i < base64aRomChipCount; i++ {
		filename := fmt.Sprintf("<internal>/BASE64A_%X.BIN", 0xd0+i*0x08)
		data, err := loadResource(filename)
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
		a.mmu.physicalROM[j] = newMemoryRange(0xd000, romBanksBytes[j])
	}

	// Start with first bank active
	a.mmu.setActiveROMPage(0)

	// Add rom soft switches. They use the annunciator 0 and 1
	a.io.addSoftSwitchRW(0x58, func(*ioC0Page) uint8 {
		p := a.mmu.getActiveROMPage()
		a.mmu.setActiveROMPage(p & 2)
		return 0
	}, "ANN0OFF-ROM")
	a.io.addSoftSwitchRW(0x59, func(*ioC0Page) uint8 {
		p := a.mmu.getActiveROMPage()
		a.mmu.setActiveROMPage(p | 1)
		return 0
	}, "ANN0ON-ROM")
	a.io.addSoftSwitchRW(0x5A, func(*ioC0Page) uint8 {
		p := a.mmu.getActiveROMPage()
		a.mmu.setActiveROMPage(p & 1)
		return 0
	}, "ANN1OFF-ROM")
	a.io.addSoftSwitchRW(0x5B, func(*ioC0Page) uint8 {
		p := a.mmu.getActiveROMPage()
		a.mmu.setActiveROMPage(p | 2)
		return 0
	}, "ANN1ON-ROM")

	return nil
}
