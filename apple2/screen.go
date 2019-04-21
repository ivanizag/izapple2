package apple2

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

func snapshot(a *Apple2) image.Image {
	isTextMode := a.io.isSoftSwitchActive(ioFlagGraphics)
	is80ColMode := a.io.isSoftSwitchActive(ioFlag80Col)
	pageIndex := 0
	if a.io.isSoftSwitchActive(ioFlagSecondPage) {
		pageIndex = 1
	}

	if isTextMode && !is80ColMode {
		//Text mode
		return snapshotTextMode(a, pageIndex)
	}
	fmt.Printf("t: %v, 8: %v\n", isTextMode, is80ColMode)
	return nil
	//panic("Screen mode not supported")
}

func saveSnapshot(a *Apple2) {
	img := snapshot(a)
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
	charWidth        = 7
	charHeight       = 8
	textColumns      = 40
	textLines        = 24
	textPage1Address = uint16(0x400)
	textPage2Address = uint16(0x400)
)

func getTextChar(a *Apple2, col int, line int, page int) uint8 {
	address := textPage1Address
	if page == 1 {
		address = textPage2Address
	}

	// See "Understand the Apple II", page 5-10
	// http://www.applelogic.org/files/UNDERSTANDINGTHEAII.pdf
	section := line / 8 // Top, middle and bottom
	eigth := line % 8
	address += uint16(section*40 + eigth*0x80 + col)
	return a.mmu.internalPeek(address)
}

func snapshotTextMode(a *Apple2, page int) image.Image {
	width := textColumns * charWidth
	height := textLines * charHeight
	size := image.Rect(0, 0, width, height)
	bwPalette := []color.Color{color.Black, color.White}
	img := image.NewPaletted(size, bwPalette)

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			//yRev := height - y
			line := y / charHeight
			col := x / charWidth
			rowInChar := y % charHeight
			colInChar := x % charWidth
			char := getTextChar(a, col, line, page)
			pixel := a.cg.getPixel(char, rowInChar, colInChar)
			color := uint8(0)
			if pixel {
				color = 1
			}
			img.SetColorIndex(x, y, color)
		}

	}

	return img
}
