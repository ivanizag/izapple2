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
