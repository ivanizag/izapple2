package apple2

import (
	"io"
)

type card interface {
	loadRom(data []uint8)
	assign(a *Apple2, slot int)
	persistent
}

type cardBase struct {
	a        *Apple2
	rom      *memoryRange
	romExtra *memoryRange
	slot     int
	_ssr     [16]softSwitchR
	_ssw     [16]softSwitchW
	_ssrName [16]string
	_sswName [16]string
}

func (c *cardBase) loadRom(data []uint8) {
	if c.a != nil {
		panic("Assert failed. Rom must be loaded before inserting the card in the slot")
	}
	if len(data) >= 0x100 {
		c.rom = newMemoryRange(0, data)
	}
	if len(data) >= 0x800 {
		c.romExtra = newMemoryRange(0, data)
	}
}

func (c *cardBase) assign(a *Apple2, slot int) {
	c.a = a
	c.slot = slot
	if slot != 0 && c.rom != nil {
		c.rom.base = uint16(0xc000 + slot*0x100)
		a.mmu.setCardROM(slot, c.rom)
	}

	if slot != 0 && c.romExtra != nil {
		c.romExtra.base = uint16(0xc800)
		a.mmu.setCardROMExtra(slot, c.romExtra)
	}

	for i := 0; i < 0x10; i++ {
		a.io.addSoftSwitchR(uint8(0xC80+slot*0x10+i), c._ssr[i], c._ssrName[i])
		a.io.addSoftSwitchW(uint8(0xC80+slot*0x10+i), c._ssw[i], c._sswName[i])
	}
}

func (c *cardBase) addCardSoftSwitchR(address uint8, ss softSwitchR, name string) {
	c._ssr[address] = ss
	c._ssrName[address] = name
}

func (c *cardBase) addCardSoftSwitchW(address uint8, ss softSwitchW, name string) {
	c._ssw[address] = ss
	c._sswName[address] = name
}

func (c *cardBase) save(w io.Writer) error {
	// Empty
	return nil
}

func (c *cardBase) load(r io.Reader) error {
	// Empty
	return nil
}
