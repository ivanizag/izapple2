package screen

import (
	"image"
	"image/color"
)

// Base Video Modes
const (
	VideoBaseMask  uint32 = 0x1f
	VideoText40    uint32 = 0x01
	VideoGR        uint32 = 0x02
	VideoHGR       uint32 = 0x03
	VideoText80    uint32 = 0x08
	VideoDGR       uint32 = 0x09
	VideoDHGR      uint32 = 0x0a
	VideoText40RGB uint32 = 0x10
	VideoMono560   uint32 = 0x11
	VideoRGBMix    uint32 = 0x12
	VideoRGB160    uint32 = 0x13
	VideoSHR       uint32 = 0x14
	VideoVidex     uint32 = 0x15
)

// Mix text video mdes modifiers
const (
	VideoMixTextMask  uint32 = 0x0f00
	VideoMixText40    uint32 = 0x0100
	VideoMixText80    uint32 = 0x0200
	VideoMixText40RGB uint32 = 0x0300
)

// Other video mode modifiers
const (
	VideoModifiersMask  uint32 = 0xf000
	VideoSecondPage     uint32 = 0x1000
	VideoAltText        uint32 = 0x2000
	VideoRGBCard        uint32 = 0x4000
	VideoFourColors     uint32 = 0x8000
	VideoText80AltOrder uint32 = 0x10000
)

// VideoSource provides the info to build the video output
type VideoSource interface {
	// GetCurrentVideoMode returns the active video mode
	GetCurrentVideoMode() uint32
	// GetTextMemory returns a slice to the text memory pages
	GetTextMemory(secondPage bool, ext bool) []uint8
	// GetVideoMemory returns a slice to the video memory pages
	GetVideoMemory(secondPage bool, ext bool) []uint8
	// GetCharactePixel returns the pixel as output by the character generator
	GetCharacterPixel(char uint8, rowInChar int, colInChar int, isAltText bool, isFlashedFrame bool) bool
	// GetSuperVideoMemory returns a slice to the SHR video memory
	GetSuperVideoMemory() []uint8
	// GetCardImage returns an image provided by a card, like the videx card
	GetCardImage(light color.Color) *image.RGBA
	// SupportsLowercase returns true if the video source supports lowercase
	SupportsLowercase() bool
}
