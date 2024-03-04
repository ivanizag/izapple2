package izapple2

/*
	MultiROM card for Apple II

See: https://www.applefritter.com/content/multiple-image-rom-card


-------
28C256  (Also for 27256 on Rev 1.1 cards)
-------
0000-07FF - CPM boot   (User 2)                  Bank 3
0800-0FFF - Freeze     (User 3)                  Bank 2
1000-37FF - IntBasic including programmers aid	 BankBasic 0
3800-3FFF - Monitor						         Bank 6
4000-47FF - Lockbuster (User 1)                  Bank 4
4800-4FFF - Dead boot diagnostic                 Bank 5
5000-77FF - Applesoft                            BankBasic 1
7800-7FFF - Autostart                            Bank 7


*/

// MultiRomCard represents a Multiple Image ROM Card
type MultiRomCard struct {
	cardBase
	rom       []uint8
	basicBank int
	f8Bank    int
}

func newMultiRomCardBuilder() *cardBuilder {
	return &cardBuilder{
		name:        "MultiROM",
		description: "Multiple Image ROM card",
		defaultParams: &[]paramSpec{
			{"rom", "ROM file to load", "<internal>/MultiRom(SP boot)-Prog aid-28C256.BIN"},
			{"basic", "Bank for D000 to F7FF", "1"},
			{"bank", "Bank for F8", "7"}},
		buildFunc: func(params map[string]string) (Card, error) {
			var c MultiRomCard
			var err error
			c.basicBank, err = paramsGetInt(params, "basic")
			if err != nil {
				return nil, err
			}
			c.f8Bank, err = paramsGetInt(params, "bank")
			if err != nil {
				return nil, err
			}

			romFile := paramsGetPath(params, "rom")
			data, _, err := LoadResource(romFile)
			if err != nil {
				return nil, err
			}

			c.rom = data
			return &c, nil
		},
	}
}

func (c *MultiRomCard) assign(a *Apple2, slot int) {
	c.cardBase.assign(a, slot)
	a.mmu.inhibitROM(c)
}

func (c *MultiRomCard) translateAddress(address uint16) uint16 {
	var baseAddress uint16
	// Basic part
	if address < 0xf800 {
		switch c.basicBank {
		case 0:
			baseAddress = 0x1000
		default:
			baseAddress = 0x5000
		}
		return address - 0xd000 + baseAddress
	}

	// F8 part

	switch c.f8Bank {
	case 2:
		baseAddress = 0x0800
	case 3:
		baseAddress = 0x0000
	case 4:
		baseAddress = 0x4000
	case 5:
		baseAddress = 0x4800
	case 6:
		baseAddress = 0x3800
	default:
		baseAddress = 0x7800
	}
	return address - 0xf800 + baseAddress
}

func (c *MultiRomCard) peek(address uint16) uint8 {
	return c.rom[c.translateAddress(address)]
}

func (c *MultiRomCard) poke(address uint16, value uint8) {
	// Nothing
}
