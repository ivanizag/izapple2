package apple2

import (
	"fmt"
)

/*
	Copam BASE64A uses paginated ROM
*/

// Base64aROM Models the paginated ROM on a BASE64A clone
type base64aROM struct {
	romBanks [4]*memoryRange
	romBank  uint8
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
	var r base64aROM
	for j := range romBanksBytes {
		r.romBanks[j] = newMemoryRange(0xd000, romBanksBytes[j])
	}

	// Start with first bank active
	r.changeBank(a.mmu, 0)

	// Add rom soft switches. They use the annunciator 0 and 1
	a.io.addSoftSwitchRW(0x58, func(*ioC0Page) uint8 {
		r.changeBank(a.mmu, r.romBank&2)
		return 0
	})
	a.io.addSoftSwitchRW(0x59, func(*ioC0Page) uint8 {
		r.changeBank(a.mmu, r.romBank|1)
		return 0
	})
	a.io.addSoftSwitchRW(0x5A, func(*ioC0Page) uint8 {
		r.changeBank(a.mmu, r.romBank&1)
		return 0
	})
	a.io.addSoftSwitchRW(0x5B, func(*ioC0Page) uint8 {
		r.changeBank(a.mmu, r.romBank|2)
		return 0
	})

	return nil
}

func (r *base64aROM) changeBank(mmu *memoryManager, bank uint8) {
	r.romBank = bank
	//fmt.Printf("Change to ROM bank #%v\n", r.romBank)
	mmu.physicalROM = r.romBanks[r.romBank]
	mmu.resetRomPaging() // If rom was not active. This is going too far?
}
