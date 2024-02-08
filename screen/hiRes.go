package screen

import (
	"image"
	"image/color"
)

const (
	hiResWidth       = 280
	hiResLineBytes   = hiResWidth / 7
	hiResHeight      = 192
	hiResHeightMixed = 160
)

func snapshotHiRes(vs VideoSource, isSecondPage bool, light color.Color, shiftSupported bool) *image.RGBA {
	data := vs.GetVideoMemory(isSecondPage, false)
	return renderHiRes(data, light, shiftSupported)
}

func getHiResLineOffset(line int) uint16 {
	// See "Understanding the Apple II", page 5-14
	// http://www.applelogic.org/files/UNDERSTANDINGTHEAII.pdf
	section := line >> 6 // Top, middle and bottom
	outerEighth := (line >> 3) & 0x07
	innerEighth := line & 0x07
	return uint16(section*40 + outerEighth*0x80 + innerEighth*0x400)
}

func renderHiRes(data []uint8, light color.Color, shiftSupported bool) *image.RGBA {
	// As described in "Undertanding the Apple II", with half pixel shifts
	size := image.Rect(0, 0, 2*hiResWidth, hiResHeight)
	img := image.NewRGBA(size)

	for y := 0; y < hiResHeight; y++ {
		offset := getHiResLineOffset(y)
		bytes := data[offset : offset+hiResLineBytes]
		x := 0
		var previousColour color.Color = color.Black
		for _, b := range bytes {
			shifted := shiftSupported && b>>7 == 1
			for j := uint(0); j < 7; j++ {
				bit := (b >> j) & 1
				colour := light
				if bit == 0 {
					colour = color.Black
				}

				if shifted {
					// "The general rule of all these HIRES interference patterns is that delayed extends
					// undelayed, and undelayed cuts off delayed"
					img.Set(x, y, previousColour)
				} else {
					img.Set(x, y, colour)
				}
				img.Set(x+1, y, colour)
				previousColour = colour
				x += 2
			}
		}
	}
	return img
}
