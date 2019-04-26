package apple2

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

// Snapshot the currently visible screen
func Snapshot(a *Apple2) *image.RGBA {
	isTextMode := a.io.isSoftSwitchActive(ioFlagText)
	is80ColMode := a.io.isSoftSwitchActive(ioFlag80Col)
	isHiResMode := a.io.isSoftSwitchActive(ioFlagHiRes)
	pageIndex := 0
	if a.io.isSoftSwitchActive(ioFlagSecondPage) {
		pageIndex = 1
	}

	if isTextMode {
		if is80ColMode {
			// Not supported
		} else {
			return snapshotTextMode(a, pageIndex)
		}
	} else {
		if isHiResMode {
			return snapshotGraphMode(a, pageIndex)
		} else {
			// Not supported
		}
	}

	//fmt.Printf("g: %v, 8: %v, h: %v\n", isTextMode, is80ColMode, isHiResMode)
	return nil
	//panic("Screen mode not supported")
}

func saveSnapshot(a *Apple2) {
	img := Snapshot(a)
	if img == nil {
		return
	}

	f, err := os.Create("snapshot.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	fmt.Println("Saving snapshot")

	png.Encode(f, img)
}

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

	// See "Understand the Apple II", page 5-10
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
	// TODO: Missing inverse and flash modes

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
			pixel := a.cg.getPixel(char, rowInChar, colInChar)
			colour := color.Black
			if pixel {
				colour = color.White
			}
			img.Set(x, y, colour)
		}
	}

	return img
}

func getGraphLineOffset(line int) uint16 {

	// See "Understand the Apple II", page 5-14
	// http://www.applelogic.org/files/UNDERSTANDINGTHEAII.pdf
	section := line >> 6 // Top, middle and bottom
	outerEigth := (line >> 3) & 0x07
	innerEigth := line & 0x07
	return uint16(section*40 + outerEigth*0x80 + innerEigth*0x400)
}

func getGraphLine(a *Apple2, line int, page int) []uint8 {
	address := graphPage1Address
	if page == 1 {
		address = graphPage2Address
	}

	address += getGraphLineOffset(line)
	hi := uint8(address >> 8)
	lo := uint8(address)

	memPage := a.mmu.internalPage(hi)
	//fmt.Printf("line: %v, lo: %x\n", line, lo)
	return memPage[lo : lo+40]
}

func snapshotGraphMode(a *Apple2, page int) *image.RGBA {
	size := image.Rect(0, 0, graphWidth, graphHeight)
	img := image.NewRGBA(size)

	for y := 0; y < graphHeight; y++ {
		bytes := getGraphLine(a, y, page)
		x := 0
		for _, b := range bytes {
			for j := uint(0); j < 7; j++ {
				bit := (b >> j) & 1
				colour := color.Black
				if bit == 1 {
					colour = color.White
				}
				img.Set(x, y, colour)
				x++
			}
		}
	}

	return img
}
