package screen

import (
	"image"
	"image/color"
)

// Base Video Modes
const (
	VideoBaseMask  uint16 = 0x1f
	VideoText40    uint16 = 0x01
	VideoGR        uint16 = 0x02
	VideoHGR       uint16 = 0x03
	VideoText80    uint16 = 0x08
	VideoDGR       uint16 = 0x09
	VideoDHGR      uint16 = 0x0a
	VideoText40RGB uint16 = 0x10
	VideoMono560   uint16 = 0x11
	VideoRGBMix    uint16 = 0x12
	VideoRGB160    uint16 = 0x13
	VideoSHR       uint16 = 0x14
	VideoVidex     uint16 = 0x15
)

// Mix text video mdes modifiers
const (
	VideoMixTextMask  uint16 = 0x0f00
	VideoMixText40    uint16 = 0x0100
	VideoMixText80    uint16 = 0x0200
	VideoMixText40RGB uint16 = 0x0300
)

// Other video mode modifiers
const (
	VideoModifiersMask uint16 = 0xf000
	VideoSecondPage    uint16 = 0x1000
	VideoAltText       uint16 = 0x2000
	VideoRGBCard       uint16 = 0x4000
	VideoFourColors    uint16 = 0x8000
)

// VideoSource provides the info to build the video output
type VideoSource interface {
	// GetCurrentVideoMode returns the active video mode
	GetCurrentVideoMode() uint16
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
