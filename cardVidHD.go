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

// NewCardVidHD creates a new VidHD card
func NewCardVidHD() *CardVidHD {
	var c CardVidHD
	c.name = "VidHD Card"
	return &c
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
	c.loadRom(buildVidHDRom())

	// The softswitches are outside the card reserved ss
	a.io.addSoftSwitchR(0x22, notImplementedSoftSwitchR, "VIDHD-TBCOLOR")
	a.io.addSoftSwitchW(0x22, notImplementedSoftSwitchW, "VIDHD-TBCOLOR")
	a.io.addSoftSwitchR(0x29, getStatusSoftSwitch(ioDataNewVideo), "VIDHD-NEWVIDEO")
	a.io.addSoftSwitchW(0x29, setStatusSoftSwitch(ioDataNewVideo), "VIDHD-NEWVIDEO")
	a.io.addSoftSwitchR(0x34, notImplementedSoftSwitchR, "VIDHD-CLOCKCTL")
	a.io.addSoftSwitchW(0x34, notImplementedSoftSwitchW, "VIDHD-CLOCKCTL")
	a.io.addSoftSwitchR(0x35, notImplementedSoftSwitchR, "VIDHD-SHADOW")
	a.io.addSoftSwitchW(0x35, notImplementedSoftSwitchW, "VIDHD-SHADOW")

	c.cardBase.assign(a, slot)
}
