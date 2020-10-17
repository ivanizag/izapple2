package screen

import (
	"image"
	"image/color"
)

const (
	charWidth     = 7
	charHeight    = 8
	text40Columns = 40
	textLines     = 24
)

func snapshotText40(vs VideoSource, isSecondPage bool, light color.Color) *image.RGBA {
	text := getTextFromMemory(vs, isSecondPage, false)
	return renderText(vs, text, nil /*colorMap*/, light)
}

func snapshotText80(vs VideoSource, isSecondPage bool, light color.Color) *image.RGBA {
	text := getText80FromMemory(vs, isSecondPage)
	return renderText(vs, text, nil /*colorMap*/, light)
}

func snapshotText40RGB(vs VideoSource, isSecondPage bool) *image.RGBA {
	text := getTextFromMemory(vs, isSecondPage, false)
	colorMap := getTextFromMemory(vs, isSecondPage, true)
	return renderText(vs, text, colorMap, nil)
}

func snapshotText40RGBColors(vs VideoSource, isSecondPage bool) *image.RGBA {
	colorMap := getTextFromMemory(vs, isSecondPage, true)
	return renderText(vs, nil /*text*/, colorMap, nil)
}

func getText80FromMemory(vs VideoSource, isSecondPage bool) []uint8 {
	text40Columns := getTextFromMemory(vs, isSecondPage, false)
	text40ColumnsAlt := getTextFromMemory(vs, isSecondPage, true)

	// Merge the two 40 cols to return 80 cols
	text80Columns := make([]uint8, 2*len(text40Columns))
	for i := 0; i < len(text40Columns); i++ {
		text80Columns[2*i] = text40ColumnsAlt[i]
		text80Columns[2*i+1] = text40Columns[i]
	}

	return text80Columns
}

func getTextFromMemory(vs VideoSource, isSecondPage bool, isExt bool) []uint8 {
	data := vs.GetTextMemory(isSecondPage, isExt)

	text := make([]uint8, textLines*text40Columns)
	for l := 0; l < textLines; l++ {
		for c := 0; c < text40Columns; c++ {
			char := data[getTextCharOffset(c, l)]
			text[text40Columns*l+c] = char
		}
	}
	return text
}

func getTextCharOffset(col int, line int) uint16 {
	// See "Understanding the Apple II", page 5-10
	// http://www.applelogic.org/files/UNDERSTANDINGTHEAII.pdf
	section := line / 8 // Top, middle and bottom
	eighth := line % 8
	return uint16(section*40 + eighth*0x80 + col)
}

func getRGBTextColor(pixel bool, colorKey uint8) color.Color {
	if pixel {
		colorKey >>= 4
	}
	colorKey &= 0x0f
	return ntscColorMap[colorKey]

}

func renderText(vs VideoSource, text []uint8, colorMap []uint8, light color.Color) *image.RGBA {
	columns := len(text) / textLines
	if text == nil {
		columns = text40Columns
	}
	width := columns * charWidth
	height := textLines * charHeight

	size := image.Rect(0, 0, 2*hiResWidth, hiResHeight)
	img := image.NewRGBA(size)

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			line := y / charHeight
			col := x / charWidth
			rowInChar := y % charHeight
			colInChar := x % charWidth
			charIndex := line*columns + col
			var char uint8
			if text != nil {
				char = text[charIndex]
			} else {
				char = 79 + 128 // Debug screen filed with O
			}

			pixel := vs.GetCharacterPixel(char, rowInChar, colInChar)

			var colour color.Color
			if colorMap != nil {
				colour = getRGBTextColor(pixel, colorMap[charIndex])
			} else if pixel {
				colour = light
			} else {
				colour = color.Black
			}

			if columns == text40Columns {
				img.Set(x*2, y, colour)
				img.Set(x*2+1, y, colour)
			} else {
				img.Set(x, y, colour)
			}
		}
	}

	return img
}
