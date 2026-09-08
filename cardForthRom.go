package izapple2

import "fmt"

/*

Forth ROM Card by Offete Industries:

See:
	https://www.ultratechnology.com/offete.html
	https://www.forth.org/fd/FD-V04N4.pdf, page 6

It's a ROM card 600 not fully populated and without the switch.

THe ROM has a version of 6502 fig-Forth

The card has 8Kb of ROM replacing the ROM from $D000 to $EFFF.
It is designed to be inserted in slot 0.
It freezes if there is a disk2 card in slot 6, it is designed to be used with tape.

*/

const (
	forthRomCardBase  uint16 = 0xd000
	forthRomCardLimit uint16 = 0xefff
	forthRomCardSize         = int(forthRomCardLimit-forthRomCardBase) + 1 // 8Kb
)

// CardForthRom implements the Forth ROM Card
type CardForthRom struct {
	cardBase
	forthRom *memoryRangeROM
}

func newCardForthRomBuilder() *cardBuilder {
	return &cardBuilder{
		name:        "Forth ROM card",
		description: "Firmware card with Forth. It replaces the Basic ROM",
		defaultParams: &[]paramSpec{
			{"rom", "ROM file to load", "<internal>/Apple2_Forth.rom"},
		},

		buildFunc: func(params map[string]string) (Card, error) {
			var c CardForthRom

			romFile := paramsGetPath(params, "rom")
			data, _, err := LoadResource(romFile)
			if err != nil {
				return nil, err
			}
			if len(data) < forthRomCardSize {
				return nil, fmt.Errorf("the ROM file for the Forth ROM card must be at least 8Kb, %s has %v bytes", romFile, len(data))
			}

			c.forthRom = newMemoryRangeROM(forthRomCardBase, data[:forthRomCardSize], "Forth ROM")
			return &c, nil
		},
	}
}

func (c *CardForthRom) assign(a *Apple2, slot int) {
	c.cardBase.assign(a, slot)
	// The card inhibits the motherboard ROM, and with it the language card RAM
	c.a.mmu.inhibitROM(c)
}

func (c *CardForthRom) peek(address uint16) uint8 {
	if address <= forthRomCardLimit {
		// Only $D000 to $EFFF is asked to the mmu when the ROM is inhibited
		return c.forthRom.peek(address)
	}

	// The card does not cover the monitor ROM, the motherboard answers
	return c.a.mmu.physicalROM.peek(address)
}

func (c *CardForthRom) poke(address uint16, value uint8) {
	// It's a ROM card, writes are ignored
}
