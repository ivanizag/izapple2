package apple2

import (
	"fmt"
	"image"
	"image/color"
	"strings"
	"time"
)

const (
	charWidth        = 7
	charHeight       = 8
	textColumns      = 40
	textLines        = 24
	textLinesMix     = 4
	textPage1Address = uint16(0x0400)
	textPage2Address = uint16(0x0800)
)

func getTextCharOffset(col int, line int) uint16 {

	// See "Understanding the Apple II", page 5-10
	// http://www.applelogic.org/files/UNDERSTANDINGTHEAII.pdf
	section := line / 8 // Top, middle and bottom
	eigth := line % 8
	return uint16(section*40 + eigth*0x80 + col)
}

func getTextChar(a *Apple2, col int, line int, page int) uint8 {
	address := textPage1Address
	if page == 1 {
		address = textPage2Address
	}
	address += getTextCharOffset(col, line)
	return a.mmu.physicalMainRAM.subRange(address, address+1)[0]
}

func snapshotTextMode(a *Apple2, page int, mixMode bool, light color.Color) *image.RGBA {
	// Flash mode is 2Hz
	isFlashedFrame := time.Now().Nanosecond() > (1 * 1000 * 1000 * 1000 / 2)

	lineStart := 0
	if mixMode {
		lineStart = textLines - textLinesMix
	}

	width := textColumns * charWidth
	height := (textLines - lineStart) * charHeight
	size := image.Rect(0, 0, width, height)
	img := image.NewRGBA(size)

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			line := y/charHeight + lineStart
			col := x / charWidth
			rowInChar := y % charHeight
			colInChar := x % charWidth
			char := getTextChar(a, col, line, page)
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
	content := "\n"
	content += fmt.Sprintln(strings.Repeat("#", textColumns+4))

	pageIndex := 0
	if a.io.isSoftSwitchActive(ioFlagSecondPage) {
		pageIndex = 1
	}
	isAltText := a.isApple2e && a.io.isSoftSwitchActive(ioFlagAltChar)

	for l := 0; l < textLines; l++ {
		line := ""
		for c := 0; c < textColumns; c++ {
			char := getTextChar(a, c, l, pageIndex)
			line += textMemoryByteToString(char, isAltText)
		}
		content += fmt.Sprintf("# %v #\n", line)
	}

	content += fmt.Sprintln(strings.Repeat("#", textColumns+4))
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
