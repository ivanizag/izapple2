package apple2

import (
	"image"
	"image/color"
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

func getLoResPixel(a *Apple2, x int, y int, page int) uint8 {
	// Each text mode char encodes two pixels
	char := getTextChar(a, x, y/2, page)
	if y%2 == 0 {
		// Top pixel in char
		return char & 0xf
	}
	// Bottom pixel in char
	return char >> 4
}

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

	for x := 0; x < loResWidth; x++ {
		for y := 0; y < height; y++ {
			v := getLoResPixel(a, x, y, page)
			img.Set(x, y, colorMap[reversedNibble[v]])
		}
	}

	return img
}

func getColorPatterns(light color.Color) [16][16]color.Color {
	/*
		For each lores pixel we have to fill 14 half mono pixels with
		the 4 bits of the color repeated. We will need to shift by 2 bits
		on the odd columns. Lets prepare 14+2 values for each color.
	*/

	var data [16][16]color.Color

	for ci := 0; ci < 16; ci++ {
		for cb := uint8(0); cb < 4; cb++ {
			bit := (ci >> cb) & 1
			var colour color.Color
			if bit == 0 {
				colour = color.Black
			} else {
				colour = light
			}
			for i := uint8(0); i < 4; i++ {
				data[ci][cb+4*i] = colour
			}
		}
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

	patterns := getColorPatterns(light)

	for x := 0; x < loResWidth; x++ {
		for y := 0; y < height; y++ {
			offset := (x % 2) * 2 // 2 pixel offset for odd lores pixels, 0 for even pixels
			c := getLoResPixel(a, x, y, page)
			// Insert the 14 half pixels required
			for i := 0; i < loResPixelWidth*2; i++ {
				v := patterns[c][i+offset]
				// Repeat the same color for 4 lines
				for r := 0; r < loResPixelHeight; r++ {
					img.Set(x*loResPixelWidth*2+i, y*4+r, v)
				}
			}
		}
	}
	return img
}
