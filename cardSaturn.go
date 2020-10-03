package izapple2

/*
RAM card with 128Kb. It's like 8 language cards.

See:
	http://www.applelogic.org/files/SATURN128MAN.pdf
*/

type cardSaturn struct {
	cardBase
	readState   bool
	writeState  uint8
	altBank     bool
	activeBlock uint8
}

const (
	saturnBlocks = 8
)

func (c *cardSaturn) assign(a *Apple2, slot int) {
	c.readState = false
	c.writeState = lcWriteEnabled
	c.altBank = true
	c.activeBlock = 0
	a.mmu.initLanguageRAM(saturnBlocks)

	for i := uint8(0x0); i <= 0xf; i++ {
		iCopy := i
		c.addCardSoftSwitchR(iCopy, func(*ioC0Page) uint8 {
			c.ssAction(iCopy)
			return 0
		}, "SATURNR")
		c.addCardSoftSwitchW(iCopy, func(*ioC0Page, uint8) {
			c.ssAction(iCopy)
		}, "SATURNW")
	}
	c.cardBase.assign(a, slot)
	c.applyState()
}

func (c *cardSaturn) ssAction(ss uint8) {
	switch ss {
	case 0:
		// RAM read, no writes
		c.altBank = false
		c.readState = true
		c.writeState = lcWriteDisabled
	case 1:
		// ROM read, RAM write
		c.altBank = false
		c.readState = false
		c.writeState++
	case 2:
		// ROM read, no writes
		c.altBank = false
		c.readState = false
		c.writeState = lcWriteDisabled
	case 3:
		//RAM read, RAM write
		c.altBank = false
		c.readState = true
		c.writeState++
	case 4:
		c.activeBlock = 0
	case 5:
		c.activeBlock = 1
	case 6:
		c.activeBlock = 2
	case 7:
		c.activeBlock = 3
	case 8:
		// RAM read, no writes
		c.altBank = true
		c.readState = true
		c.writeState = lcWriteDisabled
	case 9:
		// ROM read, RAM write
		c.altBank = true
		c.readState = false
		c.writeState++
	case 10:
		// ROM read, no writes
		c.altBank = true
		c.readState = false
		c.writeState = lcWriteDisabled
	case 11:
		//RAM read, RAM write
		c.altBank = true
		c.readState = true
		c.writeState++
	case 12:
		c.activeBlock = 4
	case 13:
		c.activeBlock = 5
	case 14:
		c.activeBlock = 6
	case 15:
		c.activeBlock = 7
	}

	if c.writeState > lcWriteEnabled {
		c.writeState = lcWriteEnabled
	}

	c.applyState()
}

func (c *cardSaturn) applyState() {
	c.a.mmu.setLanguageRAMActiveBlock(c.activeBlock)
	c.a.mmu.setLanguageRAM(c.readState, c.writeState == lcWriteEnabled, c.altBank)
}
