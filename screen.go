package apple2

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

/*
References:
 - "Understanding the Apple II", http://www.applelogic.org/files/UNDERSTANDINGTHEAII.pdf
 - "Apple II Reference Manual"
 - "More Colors for your Apple", https://archive.org/details/byte-magazine-1979-06/page/n61
*/

// Snapshot the currently visible screen
func Snapshot(a *Apple2) *image.RGBA {
	return activeSnapshot(a, false)
}

func activeSnapshot(a *Apple2, raw bool) *image.RGBA {
	isColor := a.isColor
	isTextMode := a.io.isSoftSwitchActive(ioFlagText)
	isHiResMode := a.io.isSoftSwitchActive(ioFlagHiRes)
	isMixMode := a.io.isSoftSwitchActive(ioFlagMixed)
	is80Columns := a.io.isSoftSwitchActive(ioFlag80Col)
	isDoubleResMode := !isTextMode && is80Columns && !a.io.isSoftSwitchActive(ioFlagAnnunciator3)
	isSecondPage := a.io.isSoftSwitchActive(ioFlagSecondPage) && !a.mmu.store80Active
	isSuperHighResMode := a.io.isSoftSwitchActive(ioDataNewVideo)

	rgbFlag1 := a.io.isSoftSwitchActive(ioFlag1RGBCard)
	rgbFlag2 := a.io.isSoftSwitchActive(ioFlag2RGBCard)
	isMono560 := isDoubleResMode && !rgbFlag1 && !rgbFlag2
	isRGBMixMode := isDoubleResMode && !rgbFlag1 && rgbFlag2

	var lightColor color.Color
	if isColor {
		lightColor = color.White
	} else {
		// Color for typical Apple ][ period green P1 phosphor monitors
		// See: https://superuser.com/questions/361297/what-colour-is-the-dark-green-on-old-fashioned-green-screen-computer-displays
		lightColor = color.RGBA{65, 255, 0, 255}

	}

	var snap *image.RGBA
	var ntscMask *image.Alpha
	if isSuperHighResMode { // Has to be first and disables the rest
		snap = snapshotSuperHiResMode(a)
	} else if isTextMode {
		snap = snapshotTextMode(a, is80Columns, isSecondPage, false /*isMixMode*/, lightColor)
	} else {
		if isHiResMode {
			if isDoubleResMode {
				snap, ntscMask = snapshotDoubleHiResModeMono(a, isSecondPage, isMixMode, isRGBMixMode, lightColor)
			} else {
				snap = snapshotHiResModeMono(a, isSecondPage, isMixMode, lightColor)
			}
		} else {
			snap = snapshotLoResModeMono(a, isDoubleResMode, isSecondPage, isMixMode, lightColor)
		}

		if isMixMode {
			snapText := snapshotTextMode(a, is80Columns, false /*isSecondPage*/, true /*isMixMode*/, lightColor)
			snap = mixSnapshots(snap, snapText)
		}
		if isColor && !(raw || isMono560) {
			snap = filterNTSCColor(snap, ntscMask)
		}
	}

	if !raw && !isSuperHighResMode {
		snap = linesSeparatedFilter(snap)
	}
	return snap
}

// SnapshotHGRModes to get all modes mixed
func SnapshotHGRModes(a *Apple2) *image.RGBA {
	bwSnap := activeSnapshot(a, true)
	if bwSnap.Bounds().Dx() == hiResWidth {
		bwSnap = doubleWidthFilter(bwSnap)
	}
	colorSnap := filterNTSCColor(bwSnap, nil)
	page1Snap := filterNTSCColor(snapshotHiResModeMono(a, false /*2nd page*/, false /*mix*/, color.White), nil) // HGR 1
	page2Snap := filterNTSCColor(snapshotHiResModeMono(a, true /*2nd page*/, false /*mix*/, color.White), nil)  // HGR 2

	size := image.Rect(0, 0, hiResWidth*4, hiResHeight*2)
	out := image.NewRGBA(size)

	for y := 0; y < hiResHeight; y++ {
		for x := 0; x < hiResWidth*2; x++ {
			out.Set(x, y, colorSnap.At(x, y))
			out.Set(x+hiResWidth*2, y, bwSnap.At(x, y))
			out.Set(x, y+hiResHeight, page1Snap.At(x, y))
			out.Set(x+hiResWidth*2, y+hiResHeight, page2Snap.At(x, y))
		}
	}

	return out
}

func mixSnapshots(top, bottom *image.RGBA) *image.RGBA {
	topBounds := top.Bounds()
	topWidth := topBounds.Dx()
	topHeight := topBounds.Dy()

	bottomBounds := bottom.Bounds()
	bottomWidth := bottomBounds.Dx()
	bottomHeight := bottomBounds.Dy()

	factor := topWidth / bottomWidth

	size := image.Rect(0, 0, topWidth, topHeight+bottomHeight)
	out := image.NewRGBA(size)

	// Copy top
	for y := topBounds.Min.Y; y < topBounds.Max.Y; y++ {
		for x := topBounds.Min.X; x < topBounds.Max.X; x++ {
			c := top.At(x, y)
			out.Set(x, y, c)
		}
	}

	// Copy bottom, applying the factor
	for y := bottomBounds.Min.Y; y < bottomBounds.Max.Y; y++ {
		for x := bottomBounds.Min.X; x < bottomBounds.Max.X; x++ {
			c := bottom.At(x, y)
			for f := 0; f < factor; f++ {
				out.Set(x*factor+f, topHeight+y, c)
			}
		}
	}

	return out
}

// SaveSnapshot saves a snapshot of the screen to a png file
func SaveSnapshot(a *Apple2, filename string) error {
	img := Snapshot(a)
	img = squarishPixelsFilter(img)

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	png.Encode(f, img)
	return nil
}

func squarishPixelsFilter(in *image.RGBA) *image.RGBA {
	b := in.Bounds()
	factor := 1200 / b.Dx()
	fmt.Println(factor)
	size := image.Rect(0, 0, factor*b.Dx(), b.Dy())
	out := image.NewRGBA(size)
	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			c := in.At(x, y)
			for i := 0; i < factor; i++ {
				out.Set(factor*x+i, y, c)
			}
		}
	}
	return out
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

func doubleWidthFilter(in *image.RGBA) *image.RGBA {
	b := in.Bounds()
	size := image.Rect(0, 0, 2*b.Dx(), b.Dy())
	out := image.NewRGBA(size)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := in.At(x, y)
			out.Set(2*x, y, c)
			out.Set(2*x+1, y, c)
		}
	}
	return out
}
