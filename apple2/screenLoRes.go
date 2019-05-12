package apple2

import (
	"image"
)

const (
	loResPixelWidth  = charWidth
	loResPixelHeight = charHeight / 2

	loResWidth       = textColumns
	loResHeight      = textLines * 2
	loResHeightMixed = (textLines - textLinesMix) * 2
	loRes
	loResPage1Address = textPage1Address
	loResPage2Address = textPage2Address
)

func snapshotLoResModeReferenceColor(a *Apple2, page int, mixedMode bool) *image.RGBA {
	// As defined on "Apple II Reference Manual"

	height := loResHeight
	if mixedMode {
		height = loResHeightMixed
	}

	size := image.Rect(0, 0, loResWidth, height)
	img := image.NewRGBA(size)

	// Lores colors correspond to the NTSC 4 bit patterns reversed
	colorMap := getNTSCColorMap()
	reversedNibble := []uint8{0, 8, 4, 12, 2, 10, 6, 14, 1, 9, 5, 13, 3, 11, 7, 15}

	for y := 0; y < height; y = y + 2 {
		for x := 0; x < loResWidth; x++ {
			// Each text mode char encodes two pixels
			char := getTextChar(a, x, y/2, page)
			bottom := char >> 4
			top := char & 0xf
			img.Set(x, y, colorMap[reversedNibble[top]])
			img.Set(x, y+1, colorMap[reversedNibble[bottom]])
		}
	}

	return img
}

/*
func getLoResLine(a *Apple2, line int, page int) []uint8 {
	address := loResPage1Address
	if page == 1 {
		address = loResPage2Address
	}

	// Every text line encodes two lores lines
	address += getTextCharOffset(0, line/2)
	data := make([]uint8, 0, textColumns)
	lower := (line % 2) == 1
	for i := uint16(0); i < textColumns; i++ {
		// Two pixels are encoded on each text page char position
		v := a.mmu.internalPeek(address + i)
		if lower {
			// The four nost significant bits store the odd lines
			v >>= 4
		} else {
			// The four least significant bits store the even lines
			v &= 0xf
		}
		data = append(data)
	}
	return data
}

func snapshotLoResModeMonoShift(a *Apple2, page int, mixedMode bool, light color.Color) *image.RGBA {
	// As described in "Undertanding the Apple II", with half pixel shifts

	height := loResHeight
	if mixedMode {
		height = loResHeightMixed
	}

	size := image.Rect(0, 0, 2*loResWidth*loResPixelWidth, height*loResPixelHeight)
	img := image.NewRGBA(size)

	for y := 0; y < height; y++ {
		bytes := getLoResLine(a, y, page)
		x := 0
		for i, v := range bytes {
			// For each loRes 4bit pixel we have to complete 7*2 half mono pixels
			for j := 0; j < 14; i++ {

			}
		}

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
*/
