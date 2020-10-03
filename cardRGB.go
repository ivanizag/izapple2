package izapple2

/*
Extended 80-Column Text AppleColor Card or Video7 RGB-SL7 card
See:
	https://mirrors.apple2.org.za/Apple%20II%20Documentation%20Project/Interface%20Cards/Apple%20IIe/Apple%20IIe%20Extended%2080%20Column%20RGB%20Card/Manuals/Apple%20Ext80ColumnAppleColorCardHR%20Manual.pdf
	https://apple2online.com/web_documents/Video-7%20Manual%20KB.pdf
	https://mirrors.apple2.org.za/ftp.apple.asimov.net/documentation/hardware/video/DIGICARD%2064K%20Extended%2080%20Column%20RGB%20Card%20for%20Apple%20IIe%20Instruction%20Manual.pdf

Diagnostics disk:
	https://mirrors.apple2.org.za/ftp.apple.asimov.net/images/hardware/video/Video-7%20Apple%20II%20RGB%20Demo%20%28Video-7%2C%20Inc.%29%281984%29.dsk

It goes to the 80 column slot.

To set the state it AN3 in graphics mode has to go off-on-off-on. Each pair off-on record the state of 80col:
	on step 0, an ANN3OFF moves to step 1
	on step 1, an ANN3ON moves to step 2, and the value of 80COL is copied to RGB flag 1
	on step 2, an ANN3OFF moves to step 3
	on step 3, an ANN3ON moves to step 4, and the value of 80COL is copied to RGB flag 2

Modes by RGB flags 1 and 2:
	0-0: 560*192 mono
	1-1: 140*192 ntsc
	0-1: Mixed mode
	1-0: 160*192 ntsc

*/

type cardRGB struct {
	// cardBase, not a regular card
	step uint8
}

func setupRGBCard(a *Apple2) *cardRGB {
	var c cardRGB
	c.step = 0

	a.io.softSwitchesData[ioFlagRGBCardActive] = ssOn

	// Does not have ROM or private softswitches. It spies on the softswitches
	a.io.addSoftSwitchRW(0x50, func(io *ioC0Page) uint8 {
		io.softSwitchesData[ioFlagText] = ssOff
		// Reset RGB modes when entering graphics mode
		c.step = 0
		io.softSwitchesData[ioFlag1RGBCard] = ssOn
		io.softSwitchesData[ioFlag2RGBCard] = ssOn
		return 0
	}, "TEXTOFF")

	a.io.addSoftSwitchRW(0x5e, func(io *ioC0Page) uint8 {
		io.softSwitchesData[ioFlagAnnunciator3] = ssOff
		switch c.step {
		case 0:
			c.step++
		case 2:
			c.step++
		}

		return 0
	}, "ANN3OFF-RGB")

	a.io.addSoftSwitchRW(0x5f, func(io *ioC0Page) uint8 {
		io.softSwitchesData[ioFlagAnnunciator3] = ssOn
		switch c.step {
		case 1:
			io.softSwitchesData[ioFlag1RGBCard] = io.softSwitchesData[ioFlag80Col]
			c.step++
		case 3:
			io.softSwitchesData[ioFlag2RGBCard] = io.softSwitchesData[ioFlag80Col]
			c.step++
		}

		return 0
	}, "ANN3ON-RGB")

	return &c
}
