package apple2

import "fmt"

/*
	Copam BASE64A adaptation.
*/

// Base64a extends an Apple2
type Base64a struct {
	a        *Apple2
	romBanks [4]*memoryRange
	romBank  uint8
}

// NewBase64a instantiates an apple2
func NewBase64a(a *Apple2) *Base64a {

	var b Base64a
	b.a = a
	b.loadRom()

	return &b
}

const (
	// There are 6 ROM chips. Each can have 4Kb or 8Kb. They can fill
	// 2 or 4 banks with 2kb windows.
	base64aRomBankSize   = 12 * 1024
	base64aRomBankCount  = 4
	base64aRomWindowSize = 2 * 1024
	base64aRomChipCount  = 6
)

func (b *Base64a) loadRom() {
	// Load the 6 PROM dumps
	romBanksBytes := make([][]uint8, base64aRomBankCount)
	for j := range romBanksBytes {
		romBanksBytes[j] = make([]uint8, 0, base64aRomBankSize)
	}

	for i := 0; i < base64aRomChipCount; i++ {
		filename := fmt.Sprintf("<internal>/BASE64A_%X.BIN", 0xd0+i*0x08)
		data := loadResource(filename)
		for j := range romBanksBytes {
			start := (j * base64aRomWindowSize) % len(data)
			romBanksBytes[j] = append(romBanksBytes[j], data[start:start+base64aRomWindowSize]...)
		}
	}

	for j := range romBanksBytes {
		b.romBanks[j] = newMemoryRange(0xd000, romBanksBytes[j])
	}

	// Start with first bank active
	b.changeRomBank(0)

	// Add rom soft switches. They use the annunciator 0 and 1
	b.a.io.addSoftSwitchRW(0x58, func(*ioC0Page) uint8 {
		b.changeRomBank(b.romBank & 2)
		return 0
	})
	b.a.io.addSoftSwitchRW(0x59, func(*ioC0Page) uint8 {
		b.changeRomBank(b.romBank | 1)
		return 0
	})
	b.a.io.addSoftSwitchRW(0x5A, func(*ioC0Page) uint8 {
		b.changeRomBank(b.romBank & 1)
		return 0
	})
	b.a.io.addSoftSwitchRW(0x5B, func(*ioC0Page) uint8 {
		b.changeRomBank(b.romBank | 2)
		return 0
	})

	// Other softswitches
	b.a.io.addSoftSwitchW(0x0C, notImplementedSoftSwitchW) // 80 columns off?
	b.a.io.addSoftSwitchW(0x0E, notImplementedSoftSwitchW) // Alt char off?

	// Write on the speaker. That is a double access and should do nothing
	// but works somehow on the BASE64A
	b.a.io.addSoftSwitchW(0x30, func(io *ioC0Page, value uint8) {
		getSpeakerSoftSwitch(io)
	})

}

func (b *Base64a) changeRomBank(bank uint8) {
	b.romBank = bank
	fmt.Printf("Change to ROM bank #%v\n", b.romBank)
	b.a.mmu.physicalROM = b.romBanks[b.romBank]
	b.a.mmu.resetRomPaging() // If rom was not active. This is going to far?
}
