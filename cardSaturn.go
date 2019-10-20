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
	activeBank  uint8
	activeBlock uint8
	ramBankA    [saturnBlocks]*memoryRange // First 4kb to map in 0xD000-0xDFFF
	ramBankB    [saturnBlocks]*memoryRange // Second 4kb to map in 0xD000-0xDFFF
	ramUpper    [saturnBlocks]*memoryRange // Upper 8kb to map in 0xE000-0xFFFF
}

const (
	// Write enabling requires two sofstwitch accesses
	saturnWriteDisabled    = 0
	saturnWriteHalfEnabled = 1
	saturnWriteEnabled     = 2
)

const (
	saturnBlocks = 8
)

func (c *cardSaturn) assign(a *Apple2, slot int) {
	c.readState = false
	c.writeState = lcWriteEnabled
	c.activeBank = 1

	for i := 0; i < saturnBlocks; i++ {
		c.ramBankA[i] = newMemoryRange(0xd000, make([]uint8, 0x1000))
		c.ramBankB[i] = newMemoryRange(0xd000, make([]uint8, 0x1000))
		c.ramUpper[i] = newMemoryRange(0xe000, make([]uint8, 0x2000))
	}
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
		err = binary.Write(w, binary.BigEndian, c.activeBank)
		if err != nil {
			return err
		}
		err = binary.Write(w, binary.BigEndian, c.activeBlock)
		if err != nil {
			return err
		}
		err = c.ramBankA[i].save(w)
		if err != nil {
			return err
		}
		err = c.ramBankB[i].save(w)
		if err != nil {
			return err
		}
		err = c.ramUpper[i].save(w)
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
		err = binary.Read(r, binary.BigEndian, &c.activeBank)
		if err != nil {
			return err
		}
		err = binary.Read(r, binary.BigEndian, &c.activeBlock)
		if err != nil {
			return err
		}
		err = c.ramBankA[i].load(r)
		if err != nil {
			return err
		}
		err = c.ramBankB[i].load(r)
		if err != nil {
			return err
		}
		err = c.ramUpper[i].load(r)

		c.applyState()
	}
	return c.cardBase.load(r)
}
