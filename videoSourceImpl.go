package izapple2

import (
	"image"
	"image/color"

	"github.com/ivanizag/izapple2/screen"
)

const (
	textPage1Address  = uint16(0x0400)
	textPage2Address  = uint16(0x0800)
	textPageSize      = uint16(0x0400)
	hiResPage1Address = uint16(0x2000)
	hiResPage2Address = uint16(0x4000)
	hiResPageSize     = uint16(0x2000)
	shResPageAddress  = uint16(0x2000)
	shResPageSize     = uint16(0x8000)
)

// GetCurrentVideoMode returns the active video mode
func (a *Apple2) GetCurrentVideoMode() uint16 {
	isTextMode := a.io.isSoftSwitchActive(ioFlagText)
	isHiResMode := a.io.isSoftSwitchActive(ioFlagHiRes)
	is80Columns := a.io.isSoftSwitchActive(ioFlag80Col)
	isStore80Active := a.mmu.store80Active
	isDoubleResMode := !isTextMode && is80Columns && !a.io.isSoftSwitchActive(ioFlagAnnunciator3)
	isSuperHighResMode := a.io.isSoftSwitchActive(ioDataNewVideo)
	isVidex := a.softVideoSwitch.isActive()

	isRGBCard := a.io.isSoftSwitchActive(ioFlagRGBCardActive)
	rgbFlag1 := a.io.isSoftSwitchActive(ioFlag1RGBCard)
	rgbFlag2 := a.io.isSoftSwitchActive(ioFlag2RGBCard)
	isMono560 := isDoubleResMode && !rgbFlag1 && !rgbFlag2
	isRGBMixMode := isDoubleResMode && !rgbFlag1 && rgbFlag2
	isRGB160Mode := isDoubleResMode && rgbFlag1 && !rgbFlag2

	isMixMode := a.io.isSoftSwitchActive(ioFlagMixed)
	isSecondPage := a.io.isSoftSwitchActive(ioFlagSecondPage) && !a.mmu.store80Active
	isAltText := a.isApple2e && a.io.isSoftSwitchActive(ioFlagAltChar)

	var mode uint16
	if isSuperHighResMode {
		mode = screen.VideoSHR
		isMixMode = false
	} else if isVidex {
		mode = screen.VideoVidex
		isMixMode = false
	} else if isTextMode {
		if is80Columns {
			mode = screen.VideoText80
		} else if isRGBCard && isStore80Active {
			mode = screen.VideoText40RGB
		} else {
			mode = screen.VideoText40
		}
		isMixMode = false
	} else if isHiResMode {
		if !isDoubleResMode {
			mode = screen.VideoHGR
		} else if isMono560 {
			mode = screen.VideoMono560
		} else if isRGBMixMode {
			mode = screen.VideoRGBMix
		} else if isRGB160Mode {
			mode = screen.VideoRGB160
		} else {
			mode = screen.VideoDHGR
		}
	} else if isDoubleResMode {
		mode = screen.VideoDGR
	} else {
		mode = screen.VideoGR
	}

	// Modifiers
	if isMixMode {
		if is80Columns {
			mode |= screen.VideoMixText80
		} else /* if isStore80Active {
			mode |= screen.VideoMixText40RGB
		}  else */{
			mode |= screen.VideoMixText40
		}
	}
	if isSecondPage {
		mode |= screen.VideoSecondPage
	}
	if isAltText {
		mode |= screen.VideoAltText
	}
	if isRGBCard {
		mode |= screen.VideoRGBCard
	}
	if a.isFourColors {
		mode |= screen.VideoFourColors
	}

	return mode
}

// GetTextMemory returns a slice to the text memory pages
func (a *Apple2) GetTextMemory(secondPage bool, ext bool) []uint8 {
	mem := a.mmu.getVideoRAM(ext)
	addressStart := textPage1Address
	if secondPage {
		addressStart = textPage2Address
	}
	return mem.subRange(addressStart, addressStart+textPageSize)
}

// GetVideoMemory returns a slice to the video memory pages
func (a *Apple2) GetVideoMemory(secondPage bool, ext bool) []uint8 {
	mem := a.mmu.getVideoRAM(ext)
	addressStart := hiResPage1Address
	if secondPage {
		addressStart = hiResPage2Address
	}
	return mem.subRange(addressStart, addressStart+hiResPageSize)
}

// GetSuperVideoMemory returns a slice to the SHR video memory
func (a *Apple2) GetSuperVideoMemory() []uint8 {
	mem := a.mmu.getVideoRAM(true)
	return mem.subRange(shResPageAddress, shResPageAddress+shResPageSize)
}

// GetCharacterPixel returns the pixel as output by the character generator
func (a *Apple2) GetCharacterPixel(char uint8, rowInChar int, colInChar int, isAltText bool, isFlashedFrame bool) bool {
	var pixel bool
	if a.isApple2e {
		vid6 := (char & 0x40) != 0
		vid7 := (char & 0x80) != 0
		char := char & 0x3f
		if vid6 && (vid7 || isAltText) {
			char += 0x40
		}
		if vid7 || (vid6 && isFlashedFrame && !isAltText) {
			char += 0x80
		}
		pixel = !a.cg.getPixel(char, rowInChar, colInChar)
	} else {
		pixel = a.cg.getPixel(char, rowInChar, colInChar)
		topBits := char >> 6
		isInverse := topBits == 0
		isFlash := topBits == 1

		pixel = pixel != (isInverse || (isFlash && isFlashedFrame))
	}
	return pixel
}

// GetCardImage returns an image provided by a card, like the videx card
func (a *Apple2) GetCardImage(light color.Color) *image.RGBA {
	return a.softVideoSwitch.BuildAlternateImage(light)
}

// SupportsLowercase returns true if the video source supports lowercase
func (a *Apple2) SupportsLowercase() bool {
	return a.hasLowerCase
}

// DumpTextModeAnsi returns the text mode contents using ANSI escape codes for reverse and flash
func DumpTextModeAnsi(a *Apple2) string {
	is80Columns := a.io.isSoftSwitchActive(ioFlag80Col)
	isSecondPage := a.io.isSoftSwitchActive(ioFlagSecondPage) && !a.mmu.store80Active
	isAltText := a.isApple2e && a.io.isSoftSwitchActive(ioFlagAltChar)
	return screen.RenderTextModeAnsi(a, is80Columns, isSecondPage, isAltText, a.isApple2e)
}
