package apple2

import (
	"encoding/binary"
	"io"
)

/*

RAM card with 128Kb. It's like 8 language cards.

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
			c.ssAction(iCopy, false)
			return 0
		}, "SATURNR")
		c.addCardSoftSwitchW(iCopy, func(*ioC0Page, uint8) {
			c.ssAction(iCopy, true)
		}, "SATURNW")
	}
	c.cardBase.assign(a, slot)
	c.applyState()
}

func (c *cardSaturn) ssAction(ss uint8, write bool) {
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

	// Writing to the softswtich disables writes.
	if write {
		c.writeState = lcWriteDisabled
	}

	c.applyState()
}

func (c *cardSaturn) applyState() {
	c.a.mmu.setLanguageRAMBlock(c.activeBlock)
	c.a.mmu.setLanguageRAM(c.readState, c.writeState == lcWriteEnabled, c.altBank)
}

func (c *cardSaturn) save(w io.Writer) error {
	for i := 0; i < saturnBlocks; i++ {
		err := binary.Write(w, binary.BigEndian, c.readState)
		if err != nil {
			return err
		}
		err = binary.Write(w, binary.BigEndian, c.writeState)
		if err != nil {
			return err
		}
		err = binary.Write(w, binary.BigEndian, c.altBank)
		if err != nil {
			return err
		}
		err = binary.Write(w, binary.BigEndian, c.activeBlock)
		if err != nil {
			return err
		}
	}
	return c.cardBase.save(w)
}

func (c *cardSaturn) load(r io.Reader) error {
	for i := 0; i < saturnBlocks; i++ {
		err := binary.Read(r, binary.BigEndian, &c.readState)
		if err != nil {
			return err
		}
		err = binary.Read(r, binary.BigEndian, &c.writeState)
		if err != nil {
			return err
		}
		err = binary.Read(r, binary.BigEndian, &c.altBank)
		if err != nil {
			return err
		}
		err = binary.Read(r, binary.BigEndian, &c.activeBlock)
		if err != nil {
			return err
		}
		c.applyState()
	}
	return c.cardBase.load(r)
}
