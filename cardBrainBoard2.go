package izapple2

/*
	Brain board II card for Apple II

See:
	http://www.willegal.net/appleii/brainboard.htm
	http://www.willegal.net/appleii/bb-v5_3.1.pdf
	https://github.com/Alex-Kw/Brain-Board-II

The Brain Board II card has 4 banks instead of the 2 of the
original Brain Board:
	Bank 0: SS clear, DIP2 OFF (wozaniam)
	Bank 1: SS set, DIP2 OFF (applesoft)
	Bank 2: SS clear, DIP2 ON (wozaniam)
	Bank 3: SS set, DIP2 ON (integer)

The card is emulated as having the DIP switches as follows:
	1 - ON: The range F8 can be replaced
	2 - ON: Select the two top banks
	3 - OFF: The motherboard ROM is always replaced
	4 - ON: The softswitch selects low or high bank

Softswitches:
	$COsO - SS clear: Low bank selected
	$COs1 - SS set: High bank selected

Operation:
	The card boots on wozaniam. Use CAPS LOCK for the commands
	to work. Starts with left arrow.
	To siwtch to Integer BASIC, type:
		1000:AD 91 C0 6C FC FF
		R
	To return to wozaniam, type:
		CALL -151
		1000:AD 90 C0 6C FD FF
		1000G
*/

// CardBrainBoardII represents a Brain Board II card
type CardBrainBoardII struct {
	cardBase
	highBank bool
	dip2     bool
	rom      []uint8
}

// NewCardBrainBoardII creates a new CardBrainBoardII
func NewCardBrainBoardII() *CardBrainBoardII {
	var c CardBrainBoardII
	c.name = "Brain Board II"

	c.highBank = false // Start with wozaniam by default
	c.dip2 = true      // Use the wozaniam+integer banks

	// The ROM has:xaa-wozaniam xab-applesoft xac-wozaniam xad-integer
	data, _, err := LoadResource("<internal>/ApplesoftInteger.BIN")
	if err != nil {
		// The resource should be internal and never fail
		panic(err)
	}
	c.rom = data

	// The ROM of the card is paged as the rest of the ROMs
	c.romCxxx = &c
	return &c
}

func (c *CardBrainBoardII) assign(a *Apple2, slot int) {
	c.addCardSoftSwitchRW(0, func() uint8 {
		c.highBank = false
		return 0x55
	}, "BRAINCLEAR")

	c.addCardSoftSwitchRW(1, func() uint8 {
		c.highBank = true
		return 0x55
	}, "BRAINSET")

	c.cardBase.assign(a, slot)
	a.mmu.inhibitROM(c)
}

func (c *CardBrainBoardII) translateAddress(address uint16) uint16 {
	translated := address - 0xc000
	if c.highBank {
		translated += 0x4000
	}
	if c.dip2 {
		translated += 0x8000
	}
	return translated
}

func (c *CardBrainBoardII) peek(address uint16) uint8 {
	return c.rom[c.translateAddress(address)]
}

func (c *CardBrainBoardII) poke(address uint16, value uint8) {
	// Nothing
}

func (c *CardBrainBoardII) setBase(base uint16) {
	// Nothing
}
