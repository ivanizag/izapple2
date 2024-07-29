package izapple2

/*
Language card with 16 extra kb for the Apple ][ and  ][+
Manual: http://www.applelogic.org/files/LANGCARDMAN.pdf

The language card doesn't have ROM for Cx00. It would not
be used in slot 0 anyway.

Note also that language cards for the Apple ][ had ROM on
board to replace the main board F8 ROM with Autostart. That
was not used/needed on the Apple ][+. As this emulates the
Apple ][+, it is not considered. For the Plus it is often
referred as Language card but it is really a 16 KB Ram card,


"When RAM is deselected, the ROM on the Language card is selected for
the top 2K ($F800-$FFFF), and the ROM on the main board is selected
for $D000-$F7FF.

Power on RESET initializes ROM to read mode and RAM to write mode,
and selects the second 4K bank to map $D000-$DFFF."

Writing to the softswitch disables writing in LC? Saw that
somewhere but doing so fails IIe self check.


*/

// CardLanguage is an Language Card
type CardLanguage struct {
	cardBase
	readState  bool
	writeState uint8
	altBank    bool // false is bank1, true is bank2
}

func newCardLanguageBuilder() *cardBuilder {
	return &cardBuilder{
		name:        "16 KB Language Card",
		description: "Language card with 16 extra KB for the Apple ][ and ][+",
		buildFunc: func(params map[string]string) (Card, error) {
			return &CardLanguage{}, nil
		},
	}
}

const (
	// Write enabling requires two softswitch accesses
	lcWriteDisabled    = 0
	lcWriteHalfEnabled = 1
	lcWriteEnabled     = 2
)

func (c *CardLanguage) reset() {
	if c.a.isApple2e {
		// UtA2e 1-3, 5-23
		c.readState = false
		c.writeState = lcWriteEnabled
		c.altBank = true // Start on bank2
		c.applyState()
	}

}

func (c *CardLanguage) assign(a *Apple2, slot int) {
	c.readState = false
	c.writeState = lcWriteEnabled
	c.altBank = true // Start on bank2

	a.mmu.initLanguageRAM(1)
	for i := uint8(0x0); i <= 0xf; i++ {
		iCopy := i
		c.addCardSoftSwitchR(iCopy, func() uint8 {
			c.ssAction(iCopy, false)
			return 0
		}, "LANGCARDR")
		c.addCardSoftSwitchW(iCopy, func(uint8) {
			c.ssAction(iCopy, true)
		}, "LANGCARDW")
	}

	c.cardBase.assign(a, slot)
	c.applyState()
}

func (c *CardLanguage) ssAction(ss uint8, write bool) {
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
		if !write {
			c.writeState++
		}
	case 2:
		// ROM read, no writes
		c.readState = false
		c.writeState = lcWriteDisabled
	case 3:
		// RAM read, RAM write
		c.readState = true
		if !write {
			c.writeState++
		}
	}

	if write && c.writeState == lcWriteHalfEnabled {
		// UtA2e, 5-23. It is reset by even read access or any write acccess in the $C08x range
		// And https://github.com/zellyn/a2audit/issues/3
		c.writeState = lcWriteDisabled
	}

	if c.writeState > lcWriteEnabled {
		c.writeState = lcWriteEnabled
	}

	c.applyState()
}

func (c *CardLanguage) applyState() {
	c.a.mmu.setLanguageRAM(c.readState, c.writeState == lcWriteEnabled, c.altBank)
}
