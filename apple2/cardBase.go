package apple2

type cardBase struct {
	a    *Apple2
	rom  *memoryRange
	slot int
	ssr  [16]softSwitchR
	ssw  [16]softSwitchW
}

func (c *cardBase) insert(a *Apple2, slot int) {
	c.a = a
	c.slot = slot
	if slot != 0 && c.rom != nil {
		c.rom.base = uint16(0xc000 + slot*0x100)
		a.mmu.setPagesRead(uint8(0xc0+slot), uint8(0xc0+slot), c.rom)
	}

	for i := 0; i < 0x10; i++ {
		a.io.addSoftSwitchR(uint8(0xC80+slot*0x10+i), c.ssr[i])
		a.io.addSoftSwitchW(uint8(0xC80+slot*0x10+i), c.ssw[i])
	}
}
