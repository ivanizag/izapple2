package apple2

import (
	"fmt"
)

/*
Logger card. It never existed, I use it to trace accesses to the card.
*/

type cardLogger struct {
	cardBase
}

func (c *cardLogger) assign(a *Apple2, slot int) {
	for i := 0x0; i <= 0xf; i++ {
		iCopy := i
		c.ssr[i] = func(*ioC0Page) uint8 {
			fmt.Printf("[cardLogger] Read access to softswith 0x%x for slot %v.\n", iCopy, slot)
			return 0
		}
		c.ssw[i] = func(_ *ioC0Page, value uint8) {
			fmt.Printf("[cardLogger] Write access to softswith 0x%x for slot %v, value 0x%v.\n", iCopy, slot, value)
		}
	}

	if slot != 0 {
		a.mmu.setPagesRead(uint8(0xc0+slot), uint8(0xc0+slot), c)
	}
	c.cardBase.assign(a, slot)
}

// MemoryHandler implementation
func (c *cardLogger) peek(address uint16) uint8 {
	fmt.Printf("[cardLogger] Read in %x.\n", address)
	c.a.dumpDebugInfo()

	return 0xf3
}

func (*cardLogger) poke(address uint16, value uint8) {
	fmt.Printf("[cardLogger] Write %x in %x.\n", value, address)
}
