package apple2

import (
	"image"
	"image/color"
	"time"
)

const (
	charWidth         = 7
	charHeight        = 8
	textColumns       = 40
	textLines         = 24
	textPage1Address  = uint16(0x0400)
	textPage2Address  = uint16(0x0800)
	graphWidth        = 280
	graphHeight       = 192
	graphPage1Address = uint16(0x2000)
	graphPage2Address = uint16(0x4000)
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
	return a.mmu.internalPeek(address)
}

func snapshotTextMode(a *Apple2, page int) *image.RGBA {
	// Color for typical Apple ][ period green phosphor monitors
	// See: https://superuser.com/questions/361297/what-colour-is-the-dark-green-on-old-fashioned-green-screen-computer-displays
	p1GreenPhosphorColor := color.RGBA{65, 255, 0, 255}

	// Flash mode is 2Hz
	isFlashedFrame := time.Now().Nanosecond() > (1 * 1000 * 1000 * 1000 / 2)

	width := textColumns * charWidth
	height := textLines * charHeight
	size := image.Rect(0, 0, width, height)
	img := image.NewRGBA(size)

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			line := y / charHeight
			col := x / charWidth
			rowInChar := y % charHeight
			colInChar := x % charWidth
			char := getTextChar(a, col, line, page)
			topBits := char >> 6
			isInverse := topBits == 0
			isFlash := topBits == 1

			pixel := a.cg.getPixel(char, rowInChar, colInChar)
			pixel = pixel != (isInverse || (isFlash && isFlashedFrame))
			var colour color.Color
			if pixel {
				colour = p1GreenPhosphorColor
			} else {
				colour = color.Black
			}
			img.Set(x, y, colour)
		}
	}

	return img
}
