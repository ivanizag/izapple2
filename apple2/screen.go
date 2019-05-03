package apple2

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"time"
)

/*
References:
 - "Understanding the Apple II", http://www.applelogic.org/files/UNDERSTANDINGTHEAII.pdf
 - "Apple II Reference Manual"
 - "More Colors for your Apple", https://archive.org/details/byte-magazine-1979-06/page/n61
*/

// Snapshot the currently visible screen
func Snapshot(a *Apple2) *image.RGBA {
	isTextMode := a.io.isSoftSwitchActive(ioFlagText)
	isHiResMode := a.io.isSoftSwitchActive(ioFlagHiRes)
	// Todo: isMixMode
	pageIndex := 0
	if a.io.isSoftSwitchActive(ioFlagSecondPage) {
		pageIndex = 1
	}

	if isTextMode {
		//return snapshotTextMode(a, pageIndex)
		return linesSeparatedFilter(snapshotTextMode(a, pageIndex))
	} else {
		if isHiResMode {
			//return snapshotHiResModeReferenceMono(a, pageIndex)
			//return linesSeparatedFilter(snapshotHiResModeMonoShift(a, pageIndex))
			return linesSeparatedFilter(filterNTSCColorMoving(false, snapshotHiResModeMonoShift(a, pageIndex)))
			//return linesSeparatedFilter(filterNTSCColorStatic(snapshotHiResModeMonoShift(a, pageIndex)))

			//return snapshotHiResModeReferenceColor(a, pageIndex)
			//return snapshotHiResModeReferenceColorSolid(a, pageIndex)
		} else {
			// Lo res mode not supported
		}
	}

	//fmt.Printf("g: %v, h: %v\n", isTextMode, isHiResMode)
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

func linesSeparatedFilter(in *image.RGBA) *image.RGBA {
	b := in.Bounds()
	size := image.Rect(0, 0, b.Dx(), 4*b.Dy())
	out := image.NewRGBA(size)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := in.At(x, y)
			out.Set(x, 4*y, c)
			out.Set(x, 4*y+1, c)
			out.Set(x, 4*y+2, c)
			out.Set(x, 4*y+3, color.Black)
		}
	}
	return out
}

func getNTSCColorMap() []color.Color {
	// RGB values from https://mrob.com/pub/xapple2/colors.html
	black := color.RGBA{0, 0, 0, 255} /*COLOR=0 */ /*  0   0    0*/

	dkBlue := color.RGBA{96, 78, 189, 255} /*COLOR=2 */ /*  0  60   25*/
	red := color.RGBA{227, 30, 96, 255}    /*COLOR=1 */ /* 90  60   25*/
	brown := color.RGBA{96, 114, 3, 255}   /*COLOR=8 */ /*180  60   25*/
	dkGreen := color.RGBA{0, 163, 96, 255} /*COLOR=4 */ /*270  60   25*/

	grey := color.RGBA{156, 156, 156, 255}  /*COLOR=10*/ /*  0   0   50*/
	purple := color.RGBA{255, 68, 253, 255} /*HCOLOR=2*/ /* 45 100   50*/
	orange := color.RGBA{255, 106, 60, 255} /*HCOLOR=5*/ /*135 100   50*/
	green := color.RGBA{20, 245, 60, 255}   /*HCOLOR=1*/ /*225 100   50*/
	blue := color.RGBA{20, 207, 253, 255}   /*HCOLOR=6*/ /*315 100   50*/

	//purple := color.RGBA{255, 68, 253, 255} /*COLOR=3 */ /* 45 100   50*/
	//orange := color.RGBA{255, 106, 60, 255} /*COLOR=9 */ /*135 100   50*/
	//green := color.RGBA{20, 245, 60, 255} /*COLOR=12*/ /*225 100   50*/
	//blue := color.RGBA{20, 207, 253, 255} /*COLOR=6 */ /*315 100   50*/

	ltBlue := color.RGBA{208, 195, 255, 255} /*COLOR=7 */ /*  0  60   75*/
	pink := color.RGBA{255, 160, 208, 255}   /*COLOR=11*/ /* 90  60   75*/
	yellow := color.RGBA{208, 221, 141, 255} /*COLOR=13*/ /*180  60   75*/
	aqua := color.RGBA{114, 255, 208, 255}   /*COLOR=14*/ /*270  60   75*/

	white := color.RGBA{255, 255, 255, 255} /*COLOR=15*/ /*  0   0  100*/

	colorMap := []color.Color{
		/* 0000 */ black,
		/* 0001 */ brown,
		/* 0010 */ dkGreen,
		/* 0011 */ green,
		/* 0100 */ dkBlue,
		/* 0101 */ grey,
		/* 0110 */ blue,
		/* 0111 */ aqua,
		/* 1000 */ red,
		/* 1001 */ orange,
		/* 1010 */ grey,
		/* 1011 */ yellow,
		/* 1100 */ purple,
		/* 1101 */ pink,
		/* 1110 */ ltBlue,
		/* 1111 */ white,
	}
	return colorMap
}

func filterNTSCColorStatic(in *image.RGBA) *image.RGBA {
	colorMap := getNTSCColorMap()

	b := in.Bounds()
	size := image.Rect(0, 0, b.Dx()/4, b.Dy())
	out := image.NewRGBA(size)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x += 4 {
			v := 0
			for i := 0; i < 4; i++ {
				cIn := in.At(x+i, y)
				r, _, _, _ := cIn.RGBA()
				v = v << 1
				if r != 0 {
					v++
				}
			}
			cOut := colorMap[v]
			out.Set(x/4, y, cOut)
		}
	}
	return out
}

func filterNTSCColorMoving(blacker bool, in *image.RGBA) *image.RGBA {
	colorMap := getNTSCColorMap()

	b := in.Bounds()
	size := image.Rect(0, 0, b.Dx()+3, b.Dy())
	out := image.NewRGBA(size)

	// We store the last four bits. We start will 0000
	v := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Dx(); x++ {
			cIn := in.At(x, y)
			r, _, _, _ := cIn.RGBA()

			pos := 1 << (3 - uint(x%4))
			var cOut color.Color
			if r != 0 {
				v |= pos
				cOut = colorMap[v]
			} else {
				v &^= pos
				if blacker {
					// If there is no luminance, let's have black anyway
					cOut = colorMap[0]
				} else {
					cOut = colorMap[v]
				}
			}
			out.Set(x, y, cOut)
		}

		// We fade for the last three positions
		for x := b.Dx(); x < b.Max.X; x++ {
			v = (v << 1) & 0xF
			cOut := colorMap[v]
			out.Set(x, y, cOut)
		}
	}
	return out
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

func snapshotHiResModeReferenceMono(a *Apple2, page int) *image.RGBA {
	// As defined on "Apple II Reference Manual", page 19
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

func snapshotHiResModeMonoShift(a *Apple2, page int) *image.RGBA {
	// As described in "Undertanding the Apple II", with half pixel shifts
	size := image.Rect(0, 0, 2*graphWidth, graphHeight)
	img := image.NewRGBA(size)

	for y := 0; y < graphHeight; y++ {
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

func snapshotHiResModeReferenceColor(a *Apple2, page int) *image.RGBA {
	// As defined on "Apple II Reference Manual", page 19
	size := image.Rect(0, 0, graphWidth, graphHeight)
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

	for y := 0; y < graphHeight; y++ {
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

func snapshotHiResModeReferenceColorSolid(a *Apple2, page int) *image.RGBA {
	// As defined on "Apple II Reference Manual", page 19
	// but with more solid colors
	size := image.Rect(0, 0, graphWidth, graphHeight)
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

	for y := 0; y < graphHeight; y++ {
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
					img.Set(x-1, y, colour)
					img.Set(x, y, colour)
				}
				x++
			}
		}
	}

	return img
}
