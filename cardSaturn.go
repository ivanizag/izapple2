package izapple2

/*
RAM card with 128Kb. It's like 8 language cards.

See:
	http://www.applelogic.org/files/SATURN128MAN.pdf
*/

// CardSaturn is a Saturn128 card
type CardSaturn struct {
	cardBase
	readState   bool
	writeState  uint8
	altBank     bool
	activeBlock uint8
}

func newCardSaturnBuilder() *cardBuilder {
	return &cardBuilder{
		name:        "Saturn 128KB Ram Card",
		description: "RAM card with 128Kb, it's like 8 language cards",
		buildFunc: func(params map[string]string) (Card, error) {
			return &CardSaturn{}, nil
		},
	}
}

const (
	saturnBlocks = 8
)

func (c *CardSaturn) assign(a *Apple2, slot int) {
	c.readState = false
	c.writeState = lcWriteEnabled
	c.altBank = true
	c.activeBlock = 0
	a.mmu.initLanguageRAM(saturnBlocks)
	c.addCardSoftSwitches(func(address uint8, data uint8, write bool) uint8 {
		c.ssAction(address)
		return 0
	}, "SATURN")

	c.cardBase.assign(a, slot)
	c.applyState()
}

func (c *CardSaturn) ssAction(ss uint8) {
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
		// RAM read, RAM write
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
		// RAM read, RAM write
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

func (c *CardSaturn) applyState() {
	c.a.mmu.setLanguageRAMActiveBlock(c.activeBlock)
	c.a.mmu.setLanguageRAM(c.readState, c.writeState == lcWriteEnabled, c.altBank)
}
