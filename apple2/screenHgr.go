package apple2

import (
	"image"
	"image/color"
)

const (
	graphWidth        = 280
	graphHeight       = 192
	graphHeightMixed  = 160
	graphPage1Address = uint16(0x2000)
	graphPage2Address = uint16(0x4000)
)

func getGraphLineOffset(line int) uint16 {

	// See "Understanding the Apple II", page 5-14
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
	return memPage[lo : lo+40]
}

func snapshotHiResModeReferenceMono(a *Apple2, page int, mixedMode bool) *image.RGBA {
	// As defined on "Apple II Reference Manual", page 19

	height := graphHeight
	if mixedMode {
		height = graphHeightMixed
	}

	size := image.Rect(0, 0, graphWidth, height)
	img := image.NewRGBA(size)

	for y := 0; y < height; y++ {
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

func snapshotHiResModeMonoShift(a *Apple2, page int, mixedMode bool) *image.RGBA {
	// As described in "Undertanding the Apple II", with half pixel shifts

	height := graphHeight
	if mixedMode {
		height = graphHeightMixed
	}

	size := image.Rect(0, 0, 2*graphWidth, height)
	img := image.NewRGBA(size)

	for y := 0; y < height; y++ {
		bytes := getGraphLine(a, y, page)
		x := 0
		previousColour := color.Black
		for _, b := range bytes {
			shifted := b>>7 == 1
			for j := uint(0); j < 7; j++ {
				bit := (b >> j) & 1
				colour := color.Black
				if bit == 1 {
					colour = color.White
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

func snapshotHiResModeReferenceColor(a *Apple2, page int, mixedMode bool) *image.RGBA {
	// As defined on "Apple II Reference Manual", page 19

	height := graphHeight
	if mixedMode {
		height = graphHeightMixed
	}

	size := image.Rect(0, 0, graphWidth, height)
	img := image.NewRGBA(size)

	// RGB values from https://mrob.com/pub/xapple2/colors.html
	black := color.RGBA{0, 0, 0, 255}
	violet := color.RGBA{255, 68, 253, 255}
	red := color.RGBA{255, 106, 60, 255}
	green := color.RGBA{20, 246, 60, 255}
	blue := color.RGBA{20, 207, 253, 255}
	white := color.RGBA{255, 255, 255, 255}
	colorMap := [][][]color.Color{
		{
			/* 00 */ {black, black},
			/* 01 */ {black, green},
			/* 10 */ {violet, black},
			/* 11 */ {white, white},
		},
		{
			/* 00 */ {black, black},
			/* 01 */ {black, red},
			/* 10 */ {blue, black},
			/* 11 */ {white, white},
		},
	}

	for y := 0; y < height; y++ {
		bytes := getGraphLine(a, y, page)
		x := 0
		previous := uint8(0)
		for _, b := range bytes {
			shift := b >> 7
			for j := uint(0); j < 7; j++ {
				bit := (b >> j) & 1
				even := x%2 == 0
				if even {
					previous = bit
				} else {
					pair := colorMap[shift][(previous<<1)+bit]
					img.Set(x-1, y, pair[0])
					img.Set(x, y, pair[1])
				}
				x++
			}
		}
	}

	return img
}

func snapshotHiResModeReferenceColorSolid(a *Apple2, page int, mixedMode bool) *image.RGBA {
	// As defined on "Apple II Reference Manual", page 19
	// but with more solid colors and half the resolution

	height := graphHeight
	if mixedMode {
		height = graphHeightMixed
	}

	size := image.Rect(0, 0, graphWidth/2, height)
	img := image.NewRGBA(size)

	// RGB values from https://mrob.com/pub/xapple2/colors.html
	black := color.RGBA{0, 0, 0, 255}
	violet := color.RGBA{255, 68, 253, 255}
	red := color.RGBA{255, 106, 60, 255}
	green := color.RGBA{20, 246, 60, 255}
	blue := color.RGBA{20, 207, 253, 255}
	white := color.RGBA{255, 255, 255, 255}
	colorMap := [][]color.Color{
		{
			/* 00 */ black,
			/* 01 */ green,
			/* 10 */ violet,
			/* 11 */ white,
		},
		{
			/* 00 */ black,
			/* 01 */ red,
			/* 10 */ blue,
			/* 11 */ white,
		},
	}

	for y := 0; y < height; y++ {
		bytes := getGraphLine(a, y, page)
		x := 0
		previous := uint8(0)
		for _, b := range bytes {
			shift := b >> 7
			for j := uint(0); j < 7; j++ {
				bit := (b >> j) & 1
				even := x%2 == 0
				if even {
					previous = bit
				} else {
					colour := colorMap[shift][(previous<<1)+bit]
					img.Set(x/2, y, colour)
				}
				x++
			}
		}
	}

	return img
}
