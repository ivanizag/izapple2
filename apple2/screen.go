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
	isTextMode := a.io.isSoftSwitchActive(ioFlagText)
	isHiResMode := a.io.isSoftSwitchActive(ioFlagHiRes)
	isMixMode := a.io.isSoftSwitchActive(ioFlagMixed)
	// Todo: isMixMode
	pageIndex := 0
	if a.io.isSoftSwitchActive(ioFlagSecondPage) {
		pageIndex = 1
	}

	var snap *image.RGBA
	if isTextMode {
		// Color for typical Apple ][ period green phosphor monitors
		// See: https://superuser.com/questions/361297/what-colour-is-the-dark-green-on-old-fashioned-green-screen-computer-displays
		p1GreenPhosphorColor := color.RGBA{65, 255, 0, 255}

		snap = snapshotTextMode(a, pageIndex, false, p1GreenPhosphorColor)
		snap = linesSeparatedFilter(snap)
	} else {
		if isHiResMode {
			//snap = snapshotHiResModeReferenceMono(a, pageIndex, isMixMode)
			//snap = snapshotHiResModeReferenceColor(a, pageIndex, isMixMode)
			//snap = snapshotHiResModeReferenceColorSolid(a, pageIndex, isMixMode)
			snap = snapshotHiResModeMonoShift(a, pageIndex, isMixMode)
		} else {
			// Lo res mode not supported
			return nil
		}
		if isMixMode {
			snapText := snapshotTextMode(a, pageIndex, isHiResMode, color.White)
			snap = mixSnapshots(snap, snapText)
		}
		//snap = filterNTSCColorStatic(snap)
		snap = filterNTSCColorMoving(false /*blacker*/, snap)
		snap = linesSeparatedFilter(snap)
	}
	return snap
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

	// Copy bottom, applyng the factor
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
