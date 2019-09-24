package apple2

import (
	"encoding/binary"
	"io"
)

/*
Language card with 16 extra kb for the Apple ][ and  ][+
Manual: http://www.applelogic.org/files/LANGCARDMAN.pdf

The language card doesn't have ROM for Cx00. It would not
be used in slot 0 anyway.

Note also that language cards for the Apple ][ had ROM on
board to replace the main board F8 ROM with Autostart. That
was not used/needed on the Apple ][+. As this emulates the
Apple ][+, it is not considered. For the Plus it is often
refered as Language card but it is really a 16 KB Ram card,


"When RAM is deselected, the ROM on the Language card is selected for
the top 2K ($F800-$FFFF), and the ROM on the main board is selected
for $D000-$F7FF.

Power on RESET initializes ROM to read mode and RAM to write mode,
and selects the second 4K bank to map $D000-$DFFF."

*/

type cardLanguage struct {
	cardBase
	readState  bool
	writeState int
	activeBank int
	ramBankA   *memoryRange // First 4kb to map in 0xD000-0xDFFF
	ramBankB   *memoryRange // Second 4kb to map in 0xD000-0xDFFF
	ramUpper   *memoryRange // Upper 8kb to map in 0xE000-0xFFFF
}

const (
	// Write enabling requires two sofstwitch accesses
	lcWriteDisabled    = 0
	lcWriteHalfEnabled = 1
	lcWriteEnabled     = 2
)

func (c *cardLanguage) assign(a *Apple2, slot int) {
	c.readState = false
	c.writeState = lcWriteEnabled
	c.activeBank = 1

	c.ramBankA = newMemoryRange(0xd000, make([]uint8, 0x1000))
	c.ramBankB = newMemoryRange(0xd000, make([]uint8, 0x1000))
	c.ramUpper = newMemoryRange(0xe000, make([]uint8, 0x2000))

	for i := 0x0; i <= 0xf; i++ {
		iCopy := i
		c.ssr[iCopy] = func(*ioC0Page) uint8 {
			c.ssAction(iCopy)
			return 0
		}
		c.ssw[iCopy] = func(*ioC0Page, uint8) {
			c.ssAction(iCopy)

			// Writing shoud reset write count per A2AUDIT
			// but doing that makes ProDos to not load.
			// c.writeState = lcWriteDisabled
		}
	}

	c.cardBase.assign(a, slot)
	c.applyState()
}

func (c *cardLanguage) ssAction(ss int) {
	c.activeBank = (ss >> 3) & 1
	action := ss & 0x3
	switch action {
	case 0:
		// RAM read, no writes
		c.readState = true
		c.writeState = lcWriteDisabled
	case 1:
		// ROM read, RAM write
		c.readState = false
		c.writeState++
	case 2:
		// ROM read, no writes
		c.readState = false
		c.writeState = lcWriteDisabled
	case 3:
		//RAM read, RAM write
		c.readState = true
		c.writeState++
	}

	if c.writeState > lcWriteEnabled {
		c.writeState = lcWriteEnabled
	}

	c.applyState()
}

func (c *cardLanguage) getActiveBank() *memoryRange {
	if c.activeBank == 0 {
		return c.ramBankA
	}
	return c.ramBankB
}

func (c *cardLanguage) applyState() {
	mmu := c.a.mmu

	if c.readState {
		mmu.setPagesRead(0xd0, 0xdf, c.getActiveBank())
		mmu.setPagesRead(0xe0, 0xff, c.ramUpper)
	} else {
		mmu.setPagesRead(0xd0, 0xff, mmu.physicalROM)
	}

	if c.writeState == lcWriteEnabled {
		mmu.setPagesWrite(0xd0, 0xdf, c.getActiveBank())
		mmu.setPagesWrite(0xe0, 0xff, c.ramUpper)
	} else {
		mmu.setPagesWrite(0xd0, 0xff, nil)
	}

}

func (c *cardLanguage) save(w io.Writer) {
	binary.Write(w, binary.BigEndian, c.readState)
	binary.Write(w, binary.BigEndian, c.writeState)
	binary.Write(w, binary.BigEndian, c.activeBank)
	c.ramBankA.save(w)
	c.ramBankB.save(w)
	c.ramUpper.save(w)

	c.cardBase.save(w)
}

func (c *cardLanguage) load(r io.Reader) {
	binary.Read(r, binary.BigEndian, &c.readState)
	binary.Read(r, binary.BigEndian, &c.writeState)
	binary.Read(r, binary.BigEndian, &c.activeBank)
	c.ramBankA.load(r)
	c.ramBankB.load(r)
	c.ramUpper.load(r)

	c.applyState()
	c.cardBase.load(r)
}
