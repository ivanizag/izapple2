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

type video struct {
	a *Apple2
}

var _ screen.VideoSource = (*video)(nil)

func newVideo(a *Apple2) *video {
	return &video{a}
}

// GetCurrentVideoMode returns the active video mode
func (v *video) GetCurrentVideoMode() uint32 {
	isTextMode := v.a.io.isSoftSwitchActive(ioFlagText)
	isHiResMode := v.a.io.isSoftSwitchActive(ioFlagHiRes)
	is80Columns := v.a.io.isSoftSwitchActive(ioFlag80Col)
	isStore80Active := v.a.mmu.store80Active
	isDoubleResMode := !isTextMode && is80Columns && !v.a.io.isSoftSwitchActive(ioFlagAnnunciator3)
	isSuperHighResMode := v.a.io.isSoftSwitchActive(ioDataNewVideo)
	isVidex := v.a.isSoftVideoSwitchActive()

	isRGBCard := v.a.io.isSoftSwitchActive(ioFlagRGBCardActive)
	rgbFlag1 := v.a.io.isSoftSwitchActive(ioFlag1RGBCard)
	rgbFlag2 := v.a.io.isSoftSwitchActive(ioFlag2RGBCard)
	isMono560 := isDoubleResMode && !rgbFlag1 && !rgbFlag2
	isRGBMixMode := isDoubleResMode && !rgbFlag1 && rgbFlag2
	isRGB160Mode := isDoubleResMode && rgbFlag1 && !rgbFlag2

	isMixMode := v.a.io.isSoftSwitchActive(ioFlagMixed)
	isSecondPage := v.a.io.isSoftSwitchActive(ioFlagSecondPage) && !v.a.mmu.store80Active
	isAltText := v.a.io.isSoftSwitchActive(ioFlagAltChar)

	var mode uint32
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
	if v.a.isFourColors {
		mode |= screen.VideoFourColors
	}

	return mode
}

// GetTextMemory returns a slice to the text memory pages
func (v *video) GetTextMemory(secondPage bool, ext bool) []uint8 {
	mem := v.a.mmu.getVideoRAM(ext)
	addressStart := textPage1Address
	if secondPage {
		addressStart = textPage2Address
	}
	return mem.subRange(addressStart, addressStart+textPageSize)
}

// GetVideoMemory returns a slice to the video memory pages
func (v *video) GetVideoMemory(secondPage bool, ext bool) []uint8 {
	mem := v.a.mmu.getVideoRAM(ext)
	addressStart := hiResPage1Address
	if secondPage {
		addressStart = hiResPage2Address
	}
	return mem.subRange(addressStart, addressStart+hiResPageSize)
}

// GetSuperVideoMemory returns a slice to the SHR video memory
func (v *video) GetSuperVideoMemory() []uint8 {
	mem := v.a.mmu.getVideoRAM(true)
	return mem.subRange(shResPageAddress, shResPageAddress+shResPageSize)
}

// GetCharacterPixel returns the pixel as output by the character generator
func (v *video) GetCharacterPixel(char uint8, rowInChar int, colInChar int, isAltText bool, isFlashedFrame bool) bool {
	var pixel bool
	if v.a.isApple2e {
		vid6 := (char & 0x40) != 0
		vid7 := (char & 0x80) != 0
		char := char & 0x3f
		if vid6 && (vid7 || isAltText) {
			char += 0x40
		}
		if vid7 || (vid6 && isFlashedFrame && !isAltText) {
			char += 0x80
		}
		pixel = !v.a.cg.getPixel(char, rowInChar, colInChar)
	} else {
		pixel = v.a.cg.getPixel(char, rowInChar, colInChar)
		topBits := char >> 6
		isInverse := topBits == 0
		isFlash := topBits == 1

		pixel = pixel != (isInverse || (isFlash && isFlashedFrame))
	}
	return pixel
}

// GetCardImage returns an image provided by a card, like the videx card
func (v *video) GetCardImage(light color.Color) *image.RGBA {
	return v.a.softVideoSwitch.buildImage(light)
}

// SupportsLowercase returns true if the video source supports lowercase
func (v *video) SupportsLowercase() bool {
	return v.a.hasLowerCase
}

// DumpTextModeAnsi returns the text mode contents using ANSI escape codes for reverse and flash
func DumpTextModeAnsi(a *Apple2) string {
	is80Columns := a.io.isSoftSwitchActive(ioFlag80Col)
	isSecondPage := a.io.isSoftSwitchActive(ioFlagSecondPage) && !a.mmu.store80Active
	isAltText := a.io.isSoftSwitchActive(ioFlagAltChar)
	supportsLowercase := a.hasLowerCase
	return screen.RenderTextModeAnsi(a.video, is80Columns, isSecondPage, isAltText, supportsLowercase, false)
}
