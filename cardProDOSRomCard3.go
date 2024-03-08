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
	bank uint16
	data []uint8
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

			var c CardProDOSRomCard3
			c.data = data
			c.loadRom(data[0x200:0x300])
			c.romC8xx = &c
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
	return c.data[c.translateAddress(address)]
}

func (c *CardProDOSRomCard3) poke(address uint16, value uint8) {
	// Do nothing
}
