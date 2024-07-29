package izapple2

import "fmt"

/*
Terence Boldt's ProDOS-ROM-Drive: A bootable 1 MB solid state disk for Apple ][ computers

Emulates version 4.0+

See:
	https://github.com/tjboldt/ProDOS-ROM-Drive
	https://github.com/Alex-Kw/ProDOS-ROM-Drive-Images

*/

// CardMemoryExpansion is a Memory Expansion card
type CardProDOSRomDrive struct {
	cardBase
	address uint16
	data    []uint8
}

const proDOSRomDriveMask = 0xf_ffff // 1 MB mask

func newCardProDOSRomDriveBuilder() *cardBuilder {
	return &cardBuilder{
		name:        "ProDOS ROM Drive",
		description: "A bootable 1 MB solid state disk by Terence Boldt",
		defaultParams: &[]paramSpec{
			// {"image", "ROM image with the ProDOS volume", "https://github.com/tjboldt/ProDOS-ROM-Drive/raw/v4.0/Firmware/GamesWithFirmware.po"},
			{"image", "ROM image with the ProDOS volume", "https://github.com/Alex-Kw/ProDOS-ROM-Drive-Images/raw/main/ProDOS_2.4.3_TJ.po"},
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

			var c CardProDOSRomDrive
			c.data = data
			c.loadRom(data[0x300:0x400], cardRomSimple)
			return &c, nil
		},
	}
}

func (c *CardProDOSRomDrive) assign(a *Apple2, slot int) {

	// Set pointer position
	c.addCardSoftSwitchW(0, func(value uint8) {
		c.address = uint16(value) | c.address&0xff00
	}, "LATCHLO")
	c.addCardSoftSwitchW(1, func(value uint8) {
		c.address = uint16(value)<<8 | c.address&0xff
	}, "LATCHHI")

	// Read data
	for i := uint8(0x0); i <= 0xf; i++ {
		iCopy := i
		c.addCardSoftSwitchR(iCopy, func() uint8 {
			offset := uint32(c.address)<<4 + uint32(iCopy)
			offset &= proDOSRomDriveMask
			return c.data[offset]
		}, fmt.Sprintf("READ%X", iCopy))
	}

	c.cardBase.assign(a, slot)
}
