package apple2

/*
Simulates just what is needed to make Total Replay use the GS modes if the VidHD card is found

See:
	https://github.com/a2-4am/4cade/blob/master/src/hw.vidhd.a
*/

type cardVidHD struct {
	cardBase
}

func buildVidHDRom() []uint8 {
	data := make([]uint8, 256)

	data[0] = 0x24
	data[1] = 0xEA
	data[2] = 0x4C

	return data
}

func (c *cardVidHD) assign(a *Apple2, slot int) {
	c.cardBase.assign(a, slot)
}
