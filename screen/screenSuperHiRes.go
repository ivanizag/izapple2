package screen

import (
	"image"
	"image/color"
)

const (
	shrWidth      = 640
	shrWidthBytes = 640 / 4
	shrHeight     = 200
	palettesCount = 256

	shrScanLineControlOffset = uint16(0x7d00)
	shrColorPalettesOffset   = uint16(0x7e00)
)

func snapshotSuperHiRes(vs VideoSource) *image.RGBA {
	data := vs.GetSuperVideoMemory()
	return renderSuperHiRes(data)
}

func renderSuperHiRes(data []uint8) *image.RGBA {
	// See "Apple IIGS Hardware Reference", chapter 4, page 91
	// http://www.applelogic.org/files/GSHARDWAREREF.pdf
	size := image.Rect(0, 0, shrWidth, shrHeight)
	img := image.NewRGBA(size)

	// Load the palettes
	colors := make([]color.Color, palettesCount)
	iMem := uint16(0)
	for i := 0; i < palettesCount; i++ {
		b0 := data[iMem+shrColorPalettesOffset]
		iMem++
		b1 := data[iMem+shrColorPalettesOffset]
		iMem++

		red := (b1 & 0x0f) << 4
		green := b0 & 0xf0
		blue := (b0 & 0x0f) << 4

		colors[i] = color.RGBA{red, green, blue, 255}
	}

	// See "Apple IIGS Hardware Reference", table 4-21
	palettesSelectionTable := []uint8{0x4, 0x0, 0xc, 0x8}

	// Build the lines
	for y := 0; y < shrHeight; y++ {
		controlByte := data[uint16(y)+shrScanLineControlOffset]
		is640Wide := (controlByte & 0x80) != 0
		isColorFill := (controlByte & 0x20) != 0
		paletteIndex := (controlByte & 0x0f) << 4

		lineAddress := uint16(shrWidthBytes * y)
		lineBytes := data[lineAddress : lineAddress+shrWidthBytes]
		if is640Wide {
			// Line is 640 pixels, two bits per pixel
			x := 0
			for i := 0; i < shrWidthBytes; i++ {
				b := lineBytes[i]
				for j := 3; j >= 0; j-- {
					p := (b >> (uint(j) * 2)) & 0x03
					offset := palettesSelectionTable[j]
					color := colors[paletteIndex+offset+p]
					img.Set(x, y, color)
					x++
				}
			}
		} else {
			// Line is 320 pixels, two pixels per byte
			x := 0
			previousColor := uint8(0)
			for i := 0; i < shrWidthBytes; i++ {
				p0 := (lineBytes[i] & 0xf0) >> 4
				if isColorFill && p0 == 0 {
					p0 = previousColor
				} else {
					previousColor = p0
				}
				p1 := lineBytes[i] & 0x0f
				if isColorFill && p1 == 0 {
					p1 = previousColor
				} else {
					previousColor = p1
				}
				img.Set(x, y, colors[paletteIndex+p0])
				img.Set(x+1, y, colors[paletteIndex+p0])
				img.Set(x+2, y, colors[paletteIndex+p1])
				img.Set(x+3, y, colors[paletteIndex+p1])
				x += 4
			}
		}
	}

	return img
}
