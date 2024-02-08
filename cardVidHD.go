package izapple2

/*
Simulates just what is needed to make Total Replay use the GS modes if the VidHD card is found

See:
	https://github.com/a2-4am/4cade/blob/master/src/hw.vidhd.a
	http://www.applelogic.org/files/GSHARDWAREREF.pdf, page 89
*/

// CardVidHD represents a VidHD card
type CardVidHD struct {
	cardBase
}

func newCardVidHDBuilder() *cardBuilder {
	return &cardBuilder{
		name:        "VidHD Card - limited",
		description: "Firmware signature of the VidHD card to trick Total Replay to use the SHR mode",
		buildFunc: func(params map[string]string) (Card, error) {
			var c CardVidHD
			c.loadRom(buildVidHDRom())
			return &c, nil
		},
	}
}

func buildVidHDRom() []uint8 {
	data := make([]uint8, 256)

	data[0] = 0x24
	data[1] = 0xEA
	data[2] = 0x4C

	return data
}

const (
	ioDataNewVideo uint8 = 0x29
)

func (c *CardVidHD) assign(a *Apple2, slot int) {
	// The softswitches are outside the card reserved ss
	a.io.addSoftSwitchR(0x22, buildNotImplementedSoftSwitchR(a.io), "VIDHD-TBCOLOR")
	a.io.addSoftSwitchW(0x22, buildNotImplementedSoftSwitchW(a.io), "VIDHD-TBCOLOR")
	a.io.addSoftSwitchR(0x29, getStatusSoftSwitch(a.io, ioDataNewVideo), "VIDHD-NEWVIDEO")
	a.io.addSoftSwitchW(0x29, setStatusSoftSwitch(a.io, ioDataNewVideo), "VIDHD-NEWVIDEO")
	a.io.addSoftSwitchR(0x34, buildNotImplementedSoftSwitchR(a.io), "VIDHD-CLOCKCTL")
	a.io.addSoftSwitchW(0x34, buildNotImplementedSoftSwitchW(a.io), "VIDHD-CLOCKCTL")
	a.io.addSoftSwitchR(0x35, buildNotImplementedSoftSwitchR(a.io), "VIDHD-SHADOW")
	a.io.addSoftSwitchW(0x35, buildNotImplementedSoftSwitchW(a.io), "VIDHD-SHADOW")

	c.cardBase.assign(a, slot)
}
