package apple2

/*

RAM card with 128Kb. It's like 8 language cards.

http://www.applelogic.org/files/SATURN128MAN.pdf

*/

type cardSaturn struct {
	cardBase
	readState   bool
	writeState  int
	activeBank  int
	activeBlock int
	ramBankA    [8]*memoryRange // First 4kb to map in 0xD000-0xDFFF
	ramBankB    [8]*memoryRange // Second 4kb to map in 0xD000-0xDFFF
	ramUpper    [8]*memoryRange // Upper 8kb to map in 0xE000-0xFFFF
}

const (
	// Write enabling requires two sofstwitch accesses
	saturnWriteDisabled    = 0
	saturnWriteHalfEnabled = 1
	saturnWriteEnabled     = 2
)

func newCardSaturn() *cardSaturn {
	var c cardSaturn
	c.readState = false
	c.writeState = lcWriteEnabled
	c.activeBank = 1

	for i := 0; i < 8; i++ {
		c.ramBankA[i] = newMemoryRange(0xd000, make([]uint8, 0x1000))
		c.ramBankB[i] = newMemoryRange(0xd000, make([]uint8, 0x1000))
		c.ramUpper[i] = newMemoryRange(0xe000, make([]uint8, 0x2000))
	}
	for i := 0x0; i <= 0xf; i++ {
		iCopy := i
		c.ssr[i] = func(*ioC0Page) uint8 {
			c.ssAction(iCopy)
			return 0
		}
		c.ssw[i] = func(*ioC0Page, uint8) {
			// Writing resets write count (from A2AUDIT)
			c.writeState = lcWriteDisabled
		}
	}
	return &c
}

func (c *cardSaturn) ssAction(ss int) {
	switch ss {
	case 0:
		// RAM read, no writes
		c.activeBank = 0
		c.readState = true
		c.writeState = lcWriteDisabled
	case 1:
		// ROM read, RAM write
		c.activeBank = 0
		c.readState = false
		c.writeState++
	case 2:
		// ROM read, no writes
		c.activeBank = 0
		c.readState = false
		c.writeState = lcWriteDisabled
	case 3:
		//RAM read, RAM write
		c.activeBank = 0
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
		c.activeBank = 1
		c.readState = true
		c.writeState = lcWriteDisabled
	case 9:
		// ROM read, RAM write
		c.activeBank = 1
		c.readState = false
		c.writeState++
	case 10:
		// ROM read, no writes
		c.activeBank = 1
		c.readState = false
		c.writeState = lcWriteDisabled
	case 11:
		//RAM read, RAM write
		c.activeBank = 1
		c.readState = true
		c.writeState++
	case 12:
		c.activeBlock = 0
	case 13:
		c.activeBlock = 1
	case 14:
		c.activeBlock = 2
	case 15:
		c.activeBlock = 3
	}

	if c.writeState > lcWriteEnabled {
		c.writeState = lcWriteEnabled
	}

	c.applyState()
}

func (c *cardSaturn) getActiveBank() [8]*memoryRange {
	if c.activeBank == 0 {
		return c.ramBankA
	}
	return c.ramBankB
}

func (c *cardSaturn) applyState() {
	mmu := c.a.mmu
	block := c.activeBlock

	if c.readState {
		mmu.setPagesRead(0xd0, 0xdf, c.getActiveBank()[block])
		mmu.setPagesRead(0xe0, 0xff, c.ramUpper[block])
	} else {
		mmu.setPagesRead(0xd0, 0xff, mmu.physicalROM)
	}

	if c.writeState == lcWriteEnabled {
		mmu.setPagesWrite(0xd0, 0xdf, c.getActiveBank()[block])
		mmu.setPagesWrite(0xe0, 0xff, c.ramUpper[block])
	} else {
		mmu.setPagesWrite(0xd0, 0xff, nil)
	}

}
