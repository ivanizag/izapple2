package apple2

import (
	"fmt"
	"image"
	"image/color"
	"strings"
	"time"
)

const (
	charWidth   = 7
	charHeight  = 8
	textColumns = 40
	textLines   = 24

	textPage1Address = uint16(0x0400)
	textPage2Address = uint16(0x0800)
	textPageSize     = uint16(0x0400)
)

func getTextCharOffset(col int, line int) uint16 {

	// See "Understanding the Apple II", page 5-10
	// http://www.applelogic.org/files/UNDERSTANDINGTHEAII.pdf
	section := line / 8 // Top, middle and bottom
	eighth := line % 8
	return uint16(section*40 + eighth*0x80 + col)
}

func snapshotTextMode(a *Apple2, is80Columns bool, isSecondPage bool, light color.Color) *image.RGBA {
	text, columns, lines := getActiveText(a, is80Columns, isSecondPage)
	return renderTextMode(a, text, columns, lines, light)
}

func getActiveText(a *Apple2, is80Columns bool, isSecondPage bool) ([]uint8, int, int) {
	if !is80Columns {
		text40Columns := getTextFromMemory(a.mmu.physicalMainRAM, isSecondPage)
		return text40Columns, textColumns, textLines
	}

	text40Columns := getTextFromMemory(a.mmu.physicalMainRAM, isSecondPage)
	text40ColumnsAlt := getTextFromMemory(a.mmu.physicalMainRAMAlt, isSecondPage)
	// Merge the two 40 cols to return 80 cols
	text80Columns := make([]uint8, 2*len(text40Columns))
	for i := 0; i < len(text40Columns); i++ {
		text80Columns[2*i] = text40ColumnsAlt[i]
		text80Columns[2*i+1] = text40Columns[i]
	}
	return text80Columns, textColumns * 2, textLines
}

func getTextFromMemory(mem *memoryRange, isSecondPage bool) []uint8 {
	addressStart := textPage1Address
	if isSecondPage {
		addressStart = textPage2Address
	}
	addressEnd := addressStart + textPageSize
	data := mem.subRange(addressStart, addressEnd)

	text := make([]uint8, textLines*textColumns)
	for l := 0; l < textLines; l++ {
		for c := 0; c < textColumns; c++ {
			char := data[getTextCharOffset(c, l)]
			text[textColumns*l+c] = char
		}
	}
	return text
}

func renderTextMode(a *Apple2, text []uint8, columns int, lines int, light color.Color) *image.RGBA {
	// Flash mode is 2Hz (host time)
	isFlashedFrame := time.Now().Nanosecond() > (1 * 1000 * 1000 * 1000 / 2)

	width := columns * charWidth
	height := lines * charHeight
	size := image.Rect(0, 0, width, height)
	img := image.NewRGBA(size)

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			line := y / charHeight
			col := x / charWidth
			rowInChar := y % charHeight
			colInChar := x % charWidth
			char := text[line*columns+col]
			var pixel bool
			if a.isApple2e {
				isAltText := a.io.isSoftSwitchActive(ioFlagAltChar)
				vid6 := (char & 0x40) != 0
				vid7 := (char & 0x80) != 0

				char := char & 0x3f
				if vid6 && (vid7 || isAltText) {
					char += 0x40
				}
				if vid7 || (vid6 && isFlashedFrame && !isAltText) {
					char += 0x80
				}
				pixel = !a.cg.getPixel(char, rowInChar, colInChar)
			} else {
				pixel = a.cg.getPixel(char, rowInChar, colInChar)
				topBits := char >> 6
				isInverse := topBits == 0
				isFlash := topBits == 1

				pixel = pixel != (isInverse || (isFlash && isFlashedFrame))
			}
			var colour color.Color
			if pixel {
				colour = light
			} else {
				colour = color.Black
			}
			img.Set(x, y, colour)
		}
	}

	return img
}

// DumpTextModeAnsi returns the text mode contents using ANSI escape codes
// for reverse and flash
func DumpTextModeAnsi(a *Apple2) string {
	is80Columns := a.io.isSoftSwitchActive(ioFlag80Col)
	isSecondPage := a.io.isSoftSwitchActive(ioFlagSecondPage) && !a.mmu.store80Active

	text, columns, lines := getActiveText(a, is80Columns, isSecondPage)
	content := "\n"
	content += fmt.Sprintln(strings.Repeat("#", columns+4))

	isAltText := a.isApple2e && a.io.isSoftSwitchActive(ioFlagAltChar)

	for l := 0; l < lines; l++ {
		line := ""
		for c := 0; c < columns; c++ {
			char := text[l*columns+c]
			line += textMemoryByteToString(char, isAltText)
		}
		content += fmt.Sprintf("# %v #\n", line)
	}

	content += fmt.Sprintln(strings.Repeat("#", columns+4))
	return content
}

func textMemoryByteToString(value uint8, isAltCharSet bool) string {
	// See https://en.wikipedia.org/wiki/Apple_II_character_set
	// Supports the new lowercase characters in the Apple2e
	// Only ascii from 0x20 to 0x5F is visible
	topBits := value >> 6
	isInverse := topBits == 0
	isFlash := topBits == 1
	if isFlash && isAltCharSet {
		// On the Apple2e with lowercase chars there is not flash mode.
		isFlash = false
		isInverse = true
	}

	if isAltCharSet {
		value = value & 0x7F
	} else {
		value = value & 0x3F
	}

	if value < 0x20 {
		value += 0x40
	}

	if value == 0x7f {
		// DEL is full box
		value = '_'
	}

	if isFlash {
		if value == ' ' {
			// Flashing space in Apple is the full box. It can't be done with ANSI codes
			value = '_'
		}
		return fmt.Sprintf("\033[5m%v\033[0m", string(value))
	} else if isInverse {
		return fmt.Sprintf("\033[7m%v\033[0m", string(value))
	} else {
		return string(value)
	}
}
