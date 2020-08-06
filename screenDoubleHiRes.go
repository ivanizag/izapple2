package apple2

import (
	"image"
	"image/color"
)

const (
	doubleHiResWidth = 2 * hiResWidth
)

func snapshotDoubleHiResModeMono(a *Apple2, isSecondPage bool, mixedMode bool, getNTSCMask bool, light color.Color) (*image.RGBA, *image.Alpha) {
	// As described in "Inside the Apple IIe"
	height := hiResHeight
	if mixedMode {
		height = hiResHeightMixed
	}
	size := image.Rect(0, 0, doubleHiResWidth, height)

	// To support RGB-mode14 we will have a mask to mark where we should not have the NTSC filter applied
	// See: https://apple2online.com/web_documents/Video-7%20Manual%20KB.pdf
	var ntscMask *image.Alpha
	if getNTSCMask {
		ntscMask = image.NewAlpha(size)
	}

	img := image.NewRGBA(size)
	for y := 0; y < height; y++ {
		lineParts := [][]uint8{
			getHiResLine(a, y, isSecondPage, true),
			getHiResLine(a, y, isSecondPage, false),
		}
		x := 0
		// For the NTSC filter to work we have to insert an initial black pixel and skip the last one
		img.Set(x, y, color.Black)
		if getNTSCMask {
			ntscMask.Set(x, y, color.Opaque)
		}
		x++
		for iByte := 0; iByte < hiResLineBytes; iByte++ {
			for iPart := 0; iPart < 2; iPart++ {
				b := lineParts[iPart][iByte]
				mask := color.Transparent // Apply the NTSC filter
				if getNTSCMask && b&0x80 == 0 {
					mask = color.Opaque // Do not apply the NTSC filter
				}
				for j := uint(0); j < 7; j++ {
					// Set color
					bit := (b >> j) & 1
					colour := light
					if bit == 0 {
						colour = color.Black
					}
					img.Set(x, y, colour)

					// Set mask if requested
					if getNTSCMask {
						ntscMask.Set(x, y, mask)
					}
					x++
				}
			}
		}
	}
	return img, ntscMask
}
