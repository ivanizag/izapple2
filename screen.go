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

const (
	videoText40 uint8 = 0x01
	videoGR     uint8 = 0x02
	videoHGR    uint8 = 0x03

	videoText80 uint8 = 0x08
	videoDGR    uint8 = 0x09
	videoDHGR   uint8 = 0x0a

	videoRGBText40 uint8 = 0x10
	videoMono560   uint8 = 0x11
	videoRGBMix    uint8 = 0x12
	videoSHR       uint8 = 0x13

	// Modifiers
	videoBaseMask   uint8 = 0x1f
	videoSecondPage uint8 = 0x20
	videoMixText40  uint8 = 0x40
	videoMixText80  uint8 = 0x80
)

func getCurrentVideoMode(a *Apple2) uint8 {
	isTextMode := a.io.isSoftSwitchActive(ioFlagText)
	isHiResMode := a.io.isSoftSwitchActive(ioFlagHiRes)
	is80Columns := a.io.isSoftSwitchActive(ioFlag80Col)
	isDoubleResMode := !isTextMode && is80Columns && !a.io.isSoftSwitchActive(ioFlagAnnunciator3)
	isSuperHighResMode := a.io.isSoftSwitchActive(ioDataNewVideo)

	rgbFlag1 := a.io.isSoftSwitchActive(ioFlag1RGBCard)
	rgbFlag2 := a.io.isSoftSwitchActive(ioFlag2RGBCard)
	isMono560 := isDoubleResMode && !rgbFlag1 && !rgbFlag2
	isRGBMixMode := isDoubleResMode && !rgbFlag1 && rgbFlag2
	isMixMode := a.io.isSoftSwitchActive(ioFlagMixed)

	mode := uint8(0)
	if isSuperHighResMode {
		mode = videoSHR
		isMixMode = false
	} else if isTextMode {
		if is80Columns {
			mode = videoText80
		} else {
			if a.mmu.store80Active {
				mode = videoRGBText40
			} else {
				mode = videoText40
			}
		}
		isMixMode = false
	} else if isHiResMode {
		if !isDoubleResMode {
			mode = videoHGR
		} else if isMono560 {
			mode = videoMono560
		} else if isRGBMixMode {
			mode = videoRGBMix
		} else {
			mode = videoDHGR
		}
	} else if isDoubleResMode {
		mode = videoDGR
	} else {
		mode = videoGR
	}

	// Modifiers
	if isMixMode {
		if is80Columns {
			mode |= videoMixText80
		} else {
			mode |= videoMixText40
		}
	}
	isSecondPage := a.io.isSoftSwitchActive(ioFlagSecondPage) && !a.mmu.store80Active
	if isSecondPage {
		mode |= videoSecondPage
	}

	return mode
}

func snapshotByMode(a *Apple2, videoMode uint8) *image.RGBA {
	videoBase := videoMode & videoBaseMask
	isSecondPage := (videoMode & videoSecondPage) != 0
	isMixMode := (videoMode & (videoMixText40 | videoMixText80)) != 0

	var lightColor color.Color
	if a.isColor {
		lightColor = color.White
	} else {
		// Color for typical Apple ][ period green P1 phosphor monitors
		// See: https://superuser.com/questions/361297/what-colour-is-the-dark-green-on-old-fashioned-green-screen-computer-displays
		lightColor = color.RGBA{65, 255, 0, 255}

	}

	applyNTSCFilter := a.isColor
	var snap *image.RGBA
	var ntscMask *image.Alpha
	switch videoBase {
	case videoText40:
		snap = snapshotText40Mode(a, isSecondPage, lightColor)
		applyNTSCFilter = false
	case videoText80:
		snap = snapshotText80Mode(a, isSecondPage, lightColor)
		applyNTSCFilter = false
	case videoRGBText40:
		snap = snapshotText40RGBMode(a, isSecondPage)
		applyNTSCFilter = false
	case videoGR:
		snap = snapshotLoResModeMono(a, isSecondPage, lightColor)
	case videoDGR:
		snap = snapshotMeResModeMono(a, isSecondPage, lightColor)
	case videoHGR:
		snap = snapshotHiResModeMono(a, isSecondPage, lightColor)
	case videoDHGR:
		snap, _ = snapshotDoubleHiResModeMono(a, isSecondPage, false /*isRGBMixMode*/, lightColor)
	case videoMono560:
		snap, _ = snapshotDoubleHiResModeMono(a, isSecondPage, false /*isRGBMixMode*/, lightColor)
		applyNTSCFilter = false
	case videoRGBMix:
		snap, ntscMask = snapshotDoubleHiResModeMono(a, isSecondPage, true /*isRGBMixMode*/, lightColor)
	case videoSHR:
		snap = snapshotSuperHiResMode(a)
		applyNTSCFilter = false
	}

	if isMixMode {
		var bottom *image.RGBA
		if (videoMode & videoMixText40) != 0 {
			bottom = snapshotText40Mode(a, isSecondPage, lightColor)
		} else {
			bottom = snapshotText80Mode(a, isSecondPage, lightColor)
		}
		snap = mixSnapshots(snap, bottom)
	}

	if applyNTSCFilter {
		snap = filterNTSCColor(snap, ntscMask)
	}
	return snap
}

// Snapshot the currently visible screen
func (a *Apple2) Snapshot() *image.RGBA {
	videoMode := getCurrentVideoMode(a)
	snap := snapshotByMode(a, videoMode)

	if snap.Bounds().Dy() == hiResHeight {
		// Apply the filter to regular CRT snapshots with 192 lines. Not to SHR
		snap = linesSeparatedFilter(snap)
	}

	return snap
}

func mixSnapshots(top, bottom *image.RGBA) *image.RGBA {
	topWidth := top.Bounds().Dx()
	bottomWidth := bottom.Bounds().Dx()
	factor := topWidth / bottomWidth

	// Copy bottom's bottom on top's bottom, applying the factor
	for y := hiResHeightMixed; y < hiResHeight; y++ {
		for x := 0; x < topWidth; x++ {
			c := bottom.At(x, y)
			for f := 0; f < factor; f++ {
				top.Set(x*factor+f, y, c)
			}
		}
	}

	return top
}

// SaveSnapshot saves a snapshot of the screen to a png file
func SaveSnapshot(a *Apple2, filename string) error {
	img := a.Snapshot()
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
