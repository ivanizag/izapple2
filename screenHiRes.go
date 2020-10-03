package izapple2

import (
	"image"
	"image/color"
)

const (
	hiResWidth        = 280
	hiResLineBytes    = hiResWidth / 7
	hiResHeight       = 192
	hiResHeightMixed  = 160
	hiResPage1Address = uint16(0x2000)
	hiResPage2Address = uint16(0x4000)
)

func getHiResLineOffset(line int) uint16 {

	// See "Understanding the Apple II", page 5-14
	// http://www.applelogic.org/files/UNDERSTANDINGTHEAII.pdf
	section := line >> 6 // Top, middle and bottom
	outerEighth := (line >> 3) & 0x07
	innerEighth := line & 0x07
	return uint16(section*40 + outerEighth*0x80 + innerEighth*0x400)
}

func getHiResLine(a *Apple2, line int, isSecondPage bool, auxMem bool) []uint8 {
	address := hiResPage1Address
	if isSecondPage {
		address = hiResPage2Address
	}

	address += getHiResLineOffset(line)
	return a.mmu.getVideoRAM(auxMem).subRange(address, address+hiResLineBytes)
}

func snapshotHiResModeMono(a *Apple2, isSecondPage bool, light color.Color) *image.RGBA {
	// As described in "Undertanding the Apple II", with half pixel shifts
	size := image.Rect(0, 0, 2*hiResWidth, hiResHeight)
	img := image.NewRGBA(size)

	for y := 0; y < hiResHeight; y++ {
		bytes := getHiResLine(a, y, isSecondPage, false /*auxMem*/)
		x := 0
		var previousColour color.Color = color.Black
		for _, b := range bytes {
			shifted := b>>7 == 1
			for j := uint(0); j < 7; j++ {
				bit := (b >> j) & 1
				colour := light
				if bit == 0 {
					colour = color.Black
				}

				if shifted {
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
