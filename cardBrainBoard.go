package izapple2

import "fmt"

/*
	Brain board card for Apple II

See:
	http://www.willegal.net/appleii/brainboard.htm
	http://www.willegal.net/appleii/bb-v5_3.1.pdf

The Brain Board card has 2 banks of ROM to replace the main ROM

A 27c256 ROM (32k) is used, with the following mapping:
	0x0000-0x00ff: lower card rom maps to $Csxx
	0x1000-0x37ff: lower bank rom maps to $D000 to $F7FF
	0x3800-0x3fff: F8 lower bank rom maps to $F800 to $FFFF
	0x4000-0x40ff: upper card rom maps to $Csxx
	0x5000-0x77ff: upper bank rom maps to $D000 to $F7FF
	0x7800-0x7fff: F8 upper bank rom maps to $F800 to $FFFF

DIP SWitches_
	1-ON : The range F8 can be replaced
	1-OFF: The range F8 is not replaced

	3-ON ,4-OFF: The motherboard ROM can be used
	3-OFF,4-ON : The motherboard ROM is always replaced

	5-ON ,6-OFF: The lower bank is used and mapped to Bank A
	5-OFF,6-ON : The lower bank is not used (A can be motherboards or uppers)

	7-ON ,8-OFF: The upper bank is used and mapper to bank B
	7-OFF,8-ON : The upper bank is not used (B can be motherboards or lowers)


Switches and softswitches:
	Up,   $COsO - SS clear: A bank selected
	Down, $COs1 - SS set: B bank selected

*/

// CardBrainBoard represents a Brain Board card
type CardBrainBoard struct {
	cardBase
	isBankB                 bool
	isMotherboardRomEnabled bool

	dip1_replaceF8          bool
	dip34_useMotherboardRom bool
	dip56_lowerInA          bool
	dip78_upperInB          bool

	rom []uint8
}

func newCardBrainBoardBuilder() *cardBuilder {
	return &cardBuilder{
		name:        "Brain Board",
		description: "Firmware card. It has two ROM banks",
		defaultParams: &[]paramSpec{
			{"rom", "ROM file to load", "<internal>/wozaniam_integer.rom"},
			{"dips", "DIP switches, leftmost is DIP 1", "1-011010"},
			{"switch", "Bank selected at boot, 'up' or 'down'", "up"},
		},

		buildFunc: func(params map[string]string) (Card, error) {
			var c CardBrainBoard
			var err error

			bank := paramsGetString(params, "switch")
			if bank == "up" {
				c.isBankB = false
			} else if bank == "down" {
				c.isBankB = true
			} else {
				return nil, fmt.Errorf("invalid bank '%s', must be up or down", bank)
			}

			dips, err := paramsGetDIPs(params, "dips", 8)
			if err != nil {
				return nil, err
			}
			c.dip1_replaceF8 = dips[1]
			if dips[3] == dips[4] {
				return nil, fmt.Errorf("DIP switches 3 and 4 must be different")
			}
			if dips[5] == dips[6] {
				return nil, fmt.Errorf("DIP switches 5 and 6 must be different")
			}
			if dips[7] == dips[8] {
				return nil, fmt.Errorf("DIP switches 7 and 8 must be different")
			}

			c.dip34_useMotherboardRom = dips[3]
			c.dip56_lowerInA = dips[5]
			c.dip78_upperInB = dips[7]

			romFile := paramsGetPath(params, "rom")
			data, _, err := LoadResource(romFile)
			if err != nil {
				return nil, err
			}
			if len(data) != 0x8000 {
				return nil, fmt.Errorf("the ROM file for the Brainboard must be 32k")
			}

			c.isMotherboardRomEnabled = true
			c.rom = data
			c.romCxxx = &c
			return &c, nil
		},
	}
}

func (c *CardBrainBoard) updateState() {
	isMotherboardRomEnabled := c.dip34_useMotherboardRom &&
		((!c.dip56_lowerInA && !c.isBankB) || (!c.dip78_upperInB && c.isBankB))

	if isMotherboardRomEnabled && !c.isMotherboardRomEnabled {
		fmt.Print("ROM: main")
		c.a.mmu.inhibitROM(nil)
	} else if !isMotherboardRomEnabled && c.isMotherboardRomEnabled {
		fmt.Print("ROM: brain")
		c.a.mmu.inhibitROM(c)
	}

	c.isMotherboardRomEnabled = isMotherboardRomEnabled
}

func (c *CardBrainBoard) assign(a *Apple2, slot int) {
	c.addCardSoftSwitchRW(0, func() uint8 {
		c.isBankB = false
		c.updateState()
		return 0x55
	}, "BRAINCLEAR")

	c.addCardSoftSwitchRW(1, func() uint8 {
		c.isBankB = true
		c.updateState()
		return 0x55
	}, "BRAINSET")

	c.cardBase.assign(a, slot)
	c.updateState()
}

func (c *CardBrainBoard) translateAddress(address uint16) uint16 {
	if c.isBankB {
		return address - 0xc000 + 0x4000
	} else {
		return address - 0xc000
	}
}

func (c *CardBrainBoard) peek(address uint16) uint8 {
	return c.rom[c.translateAddress(address)]
}

func (c *CardBrainBoard) poke(address uint16, value uint8) {
	// Nothing
}

func (c *CardBrainBoard) setBase(base uint16) {
	// Nothing
}
