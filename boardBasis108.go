package izapple2

import "github.com/ivanizag/izapple2/screen"

/*
	Basis 108 clone

	Manual: https://www.applefritter.com/files/Basis%201982%20basis%20108%20instruction%20manual.pdf

	ROM: Two pages of 12 KB each. Page 0 sets the 80 column mode. Page 1 starts in 40 column mode.

	Character ROM: Four pages, the inverse and flash characters are built from the normal ones. Pages:
		0: Apple II characters (no lowercase)
		1: German ASCII
		2: ASCII (default)
		3: APL symbols

	Memory: Has a full 64KB extra RAM replacing both main and LC RAM. It can be mapped on 8kb blocks with
	new softswitches.

	Video: 80 columns are made by having a sideways static RAM.

	The keyboard can generate interrupts.

	Missing: second 64kb block, keyboard interrupts, Z80 emulation, Parallel an Serial.
*/

func loadBasis108Rom(a *Apple2) error {
	return loadMultiPageRom(a, []string{
		"<internal>/Basis108_D83_D0.BIN",
		"<internal>/Basis108_D70_D8.BIN",
		"<internal>/Basis108_D56_E0.BIN",
		"<internal>/Basis108_D40_E8.BIN",
		"<internal>/Basis108_D39_F0.BIN",
		"<internal>/Basis108_D25_F8.BIN",
	})
}

type videoBasis108 struct {
	video
	ram   *memoryRangeBasis108
	col80 bool
}

func newVideoBasis108(a *Apple2, ram *memoryRangeBasis108) *videoBasis108 {
	var v videoBasis108
	v.video = *newVideo(a)
	v.ram = ram
	return &v
}

// GetCurrentVideoMode returns the active video mode
func (v *videoBasis108) GetCurrentVideoMode() uint32 {
	if v.col80 {
		mode := screen.VideoText80AltOrder

		isTextMode := v.a.io.isSoftSwitchActive(ioFlagText)
		isHiResMode := v.a.io.isSoftSwitchActive(ioFlagHiRes)

		if isTextMode {
			mode |= screen.VideoText80
		} else if isHiResMode {
			mode |= screen.VideoHGR
		} else {
			mode |= screen.VideoDGR
		}

		isSecondPage := v.a.io.isSoftSwitchActive(ioFlagSecondPage)
		if isSecondPage {
			mode |= screen.VideoSecondPage
		}

		return mode
	}

	return v.video.GetCurrentVideoMode()
}

// GetTextMemory returns a slice to the text memory pages
func (v *videoBasis108) GetTextMemory(secondPage bool, ext bool) []uint8 {
	return v.ram.getTextMemory(secondPage, ext)
}

func addBasis108SoftSwitches(io *ioC0Page, ram *memoryRangeBasis108, video *videoBasis108, cg *CharacterGenerator) {

	// Character generator softswitches
	io.addSoftSwitchW(0x00, buildNotImplementedSoftSwitchW(io), "BASIS108-CG-SW0-OFF") // Inverse?
	io.addSoftSwitchW(0x01, buildNotImplementedSoftSwitchW(io), "BASIS108-CG-SW0-ON")  // Flash?
	io.addSoftSwitchW(0x02, func(_ uint8) { cg.setPage(cg.page & 0x02) }, "BASIS108-CG-SW2-OFF")
	io.addSoftSwitchW(0x03, func(_ uint8) { cg.setPage(cg.page | 0x01) }, "BASIS108-CG-SW2-ON")
	io.addSoftSwitchW(0x04, func(_ uint8) { cg.setPage(cg.page & 0x01) }, "BASIS108-CG-SW1-OFF")
	io.addSoftSwitchW(0x05, func(_ uint8) { cg.setPage(cg.page | 0x02) }, "BASIS108-CG-SW1-ON")
	io.addSoftSwitchW(0x06, buildNotImplementedSoftSwitchW(io), "BASIS108-CG-SW0-OFF")
	io.addSoftSwitchW(0x07, buildNotImplementedSoftSwitchW(io), "BASIS108-CG-SW0-ON")

	// Keyboard interrupts
	io.addSoftSwitchW(0x08, buildNotImplementedSoftSwitchW(io), "BASIS108-KBDINT-OFF")
	io.addSoftSwitchW(0x09, buildNotImplementedSoftSwitchW(io), "BASIS108-KBDINT-ON")

	// 80 column softswitches
	io.addSoftSwitchW(0x0A, func(_ uint8) { video.col80 = false }, "BASIS108-80COL-OFF")
	io.addSoftSwitchW(0x0B, func(_ uint8) { video.col80 = true }, "BASIS108-80COL-ON")
	io.addSoftSwitchW(0x0C, func(_ uint8) { ram.staticRam = false }, "BASIS108-STATICRAM-OFF")
	io.addSoftSwitchW(0x0D, func(_ uint8) { ram.staticRam = true }, "BASIS108-STATICRAM-ON")

	// Language card configuration
	io.addSoftSwitchW(0x0E, buildNotImplementedSoftSwitchW(io), "BASIS108-LANG-ON")
	io.addSoftSwitchW(0x0F, buildNotImplementedSoftSwitchW(io), "BASIS108-LANG-OFF")

	// RAM bank softswitches
	io.addSoftSwitchW(0x60, buildNotImplementedSoftSwitchW(io), "BASIS108-RAM0000-BANK0")
	io.addSoftSwitchW(0x61, buildNotImplementedSoftSwitchW(io), "BASIS108-RAM0000-BANK1")
	io.addSoftSwitchW(0x62, buildNotImplementedSoftSwitchW(io), "BASIS108-RAM2000-BANK0")
	io.addSoftSwitchW(0x63, buildNotImplementedSoftSwitchW(io), "BASIS108-RAM2000-BANK1")
	io.addSoftSwitchW(0x64, buildNotImplementedSoftSwitchW(io), "BASIS108-RAM4000-BANK0")
	io.addSoftSwitchW(0x65, buildNotImplementedSoftSwitchW(io), "BASIS108-RAM4000-BANK1")
	io.addSoftSwitchW(0x66, buildNotImplementedSoftSwitchW(io), "BASIS108-RAM6000-BANK0")
	io.addSoftSwitchW(0x67, buildNotImplementedSoftSwitchW(io), "BASIS108-RAM6000-BANK1")
	io.addSoftSwitchW(0x68, buildNotImplementedSoftSwitchW(io), "BASIS108-RAM8000-BANK0")
	io.addSoftSwitchW(0x69, buildNotImplementedSoftSwitchW(io), "BASIS108-RAM8000-BANK1")
	io.addSoftSwitchW(0x6A, buildNotImplementedSoftSwitchW(io), "BASIS108-RAMA000-BANK0")
	io.addSoftSwitchW(0x6B, buildNotImplementedSoftSwitchW(io), "BASIS108-RAMA000-BANK1")
	io.addSoftSwitchW(0x6C, buildNotImplementedSoftSwitchW(io), "BASIS108-RAMD000-BANK0")
	io.addSoftSwitchW(0x6D, buildNotImplementedSoftSwitchW(io), "BASIS108-RAMD000-BANK1")
	io.addSoftSwitchW(0x6E, buildNotImplementedSoftSwitchW(io), "BASIS108-RAME000-BANK0")
	io.addSoftSwitchW(0x6F, buildNotImplementedSoftSwitchW(io), "BASIS108-RAME000-BANK1")
}
