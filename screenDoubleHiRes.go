package apple2

import (
	"image"
	"image/color"
)

const (
	doubleHiResWidth = 2 * hiResWidth
)

func snapshotDoubleHiResModeMono(a *Apple2, isSecondPage bool, mixedMode bool, light color.Color) *image.RGBA {
	// As described in "Inside the Apple IIe"

	height := hiResHeight
	if mixedMode {
		height = hiResHeightMixed
	}

	size := image.Rect(0, 0, doubleHiResWidth, height)
	img := image.NewRGBA(size)
	for y := 0; y < height; y++ {
		lineParts := [][]uint8{
			getHiResLine(a, y, isSecondPage, true),
			getHiResLine(a, y, isSecondPage, false),
		}
		x := 0
		// For the NTSC filter to work we have to insert an initial black pixel and skip the last one
		img.Set(x, y, color.Black)
		x++
		for iByte := 0; iByte < hiResLineBytes-1; iByte++ {
			for iPart := 0; iPart < 2; iPart++ {
				b := lineParts[iPart][iByte]
				for j := uint(0); j < 7; j++ {
					bit := (b >> j) & 1
					colour := light
					if bit == 0 {
						colour = color.Black
					}
					img.Set(x, y, colour)
					x++
				}
			}
		}
	}
	return img
}
