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
	writeState uint8
	altBank    bool // false is bank1, true is bank2
}

const (
	// Write enabling requires two softswitch accesses
	lcWriteDisabled    = 0
	lcWriteHalfEnabled = 1
	lcWriteEnabled     = 2
)

func (c *cardLanguage) assign(a *Apple2, slot int) {
	c.readState = false
	c.writeState = lcWriteEnabled
	c.altBank = true // Start on bank2

	if a.isApple2e {
		// The Apple //e with 128kb has two blocks of language upper RAM
		a.mmu.initLanguageRAM(2)
	} else {
		a.mmu.initLanguageRAM(1)
	}
	for i := uint8(0x0); i <= 0xf; i++ {
		iCopy := i
		c.addCardSoftSwitchR(iCopy, func(*ioC0Page) uint8 {
			c.ssAction(iCopy, false)
			return 0
		}, "LANGCARDR")
		c.addCardSoftSwitchW(iCopy, func(*ioC0Page, uint8) {
			c.ssAction(iCopy, true)
		}, "LANGCARDW")
	}

	c.cardBase.assign(a, slot)
	c.applyState()
}

func (c *cardLanguage) ssAction(ss uint8, write bool) {
	c.altBank = ((ss >> 3) & 1) == 0
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

	// Writing to the softswtich disables writes.
	if write {
		c.writeState = lcWriteDisabled
	}

	c.applyState()
}

func (c *cardLanguage) applyState() {
	c.a.mmu.setLanguageRAM(c.readState, c.writeState == lcWriteEnabled, c.altBank)
}

func (c *cardLanguage) save(w io.Writer) error {
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
	return c.cardBase.save(w)
}

func (c *cardLanguage) load(r io.Reader) error {
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
	c.applyState()
	return c.cardBase.load(r)
}
