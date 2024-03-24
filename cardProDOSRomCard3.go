package izapple2

import "fmt"

/*
Ralle Palaveev's ProDOS-Romcard3

See:
	https://github.com/rallepaqlaveev/ProDOS-Romcard3

Note that this card disables the C800-CFFF range only on writes to CFFF, not as most other cards that disable on reads and writes.

*/

// CardProDOSRomCard3 is a Memory Expansion card
type CardProDOSRomCard3 struct {
	cardBase
	bank          uint16
	data          []uint8
	nvram         bool
	secondROMPage bool
}

func newCardProDOSRomCard3Builder() *cardBuilder {
	return &cardBuilder{
		name:        "ProDOS ROM Card 3",
		description: "A bootable 4 MB ROM card by Ralle Palaveev",
		defaultParams: &[]paramSpec{
			{"image", "ROM image with the ProDOS volume", "https://github.com/rallepalaveev/ProDOS-Romcard3/raw/main/ProDOS-ROMCARD3_4MB_A2D.v1.4_v37.po"},
		},
		buildFunc: func(params map[string]string) (Card, error) {
			image := paramsGetPath(params, "image")
			if image == "" {
				return nil, fmt.Errorf("image required for the ProDOS ROM drive")
			}

			data, _, err := LoadResource(image)
			if err != nil {
				return nil, err
			}

			if len(data) != 4*1024*1024 {
				return nil, fmt.Errorf("NVRAM image must be 4MB")
			}

			var c CardProDOSRomCard3
			c.data = data
			c.loadRom(data[0x200:0x300], cardRomSimple)
			c.romC8xx = &c
			return &c, nil
		},
	}
}

func newCardProDOSNVRAMDriveBuilder() *cardBuilder {
	return &cardBuilder{
		name:        "ProDOS 4MB NVRAM DRive",
		description: "A bootable 4 MB NVRAM card by Ralle Palaveev",
		defaultParams: &[]paramSpec{
			{"image", "ROM image with the ProDOS volume", ""},
		},
		buildFunc: func(params map[string]string) (Card, error) {
			image := paramsGetPath(params, "image")
			if image == "" {
				return nil, fmt.Errorf("image required for the ProDOS ROM drive")
			}

			data, _, err := LoadResource(image)
			if err != nil {
				return nil, err
			}

			if len(data) != 4*1024*1024 && len(data) != 512*1024 {
				return nil, fmt.Errorf("NVRAM image must be 512KB or 4MB")
			}

			var c CardProDOSRomCard3
			c.data = data
			c.loadRom(data[0x200:0x400], cardRomSimple)
			c.romC8xx = &c
			c.nvram = true
			return &c, nil
		},
	}
}

func (c *CardProDOSRomCard3) assign(a *Apple2, slot int) {

	// Set pointer position
	c.addCardSoftSwitchW(0, func(value uint8) {
		c.bank = uint16(value) | c.bank&0xff00
	}, "BANKLO")
	c.addCardSoftSwitchW(1, func(value uint8) {
		c.bank = uint16(value)<<8 | c.bank&0xff
	}, "BANKHI")

	if c.nvram {
		c.addCardSoftSwitchW(2, func(value uint8) {
			if c.secondROMPage {
				c.romCsxx.setPage(0)
			} else {
				c.romCsxx.setPage(1)
			}
		}, "?????")
	}

	c.cardBase.assign(a, slot)
}

func (c *CardProDOSRomCard3) translateAddress(address uint16) int {
	// An address from 0xC800 to 0xCFFF is mapped to the corresponding bank of the ROM
	// There are 0x800 (2048) banks with 0x0800 (2048) bytes each
	offset := address - 0xC800
	pageAddress := int(c.bank&0x7FF) * 0x0800

	//fmt.Printf("CardProDOSRomCard3.translateAddress: address=%04X, bank=%04X, offset=%04X, pageAddress=%08X\n", address, c.bank, offset, pageAddress)

	return pageAddress + int(offset)
}

func (c *CardProDOSRomCard3) peek(address uint16) uint8 {
	if address&0xff == 0 {
		fmt.Printf("CardProDOSRomCard3.peek: address=%04X\n", address)
	}
	return c.data[c.translateAddress(address)]
}

func (c *CardProDOSRomCard3) poke(address uint16, value uint8) {
	fmt.Printf("CardProDOSRomCard3.poke: address=%04X, value=%02X\n", address, value)
	if c.nvram && address != 0xcfff {
		c.data[c.translateAddress(address)] = value
	}
}
