package izapple2

import (
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
	// Base Video Mode
	videoBaseMask  uint16 = 0x1f
	videoText40    uint16 = 0x01
	videoGR        uint16 = 0x02
	videoHGR       uint16 = 0x03
	videoText80    uint16 = 0x08
	videoDGR       uint16 = 0x09
	videoDHGR      uint16 = 0x0a
	videoText40RGB uint16 = 0x10
	videoMono560   uint16 = 0x11
	videoRGBMix    uint16 = 0x12
	videoRGB160    uint16 = 0x13
	videoSHR       uint16 = 0x14

	// Mix text modifiers
	videoMixTextMask  uint16 = 0x0f00
	videoMixText40    uint16 = 0x0100
	videoMixText80    uint16 = 0x0200
	videoMixText40RGB uint16 = 0x0300

	// Other modifiers
	videoModifiersMask uint16 = 0xf000
	videoSecondPage    uint16 = 0x1000
)

func getCurrentVideoMode(a *Apple2) uint16 {
	isTextMode := a.io.isSoftSwitchActive(ioFlagText)
	isHiResMode := a.io.isSoftSwitchActive(ioFlagHiRes)
	is80Columns := a.io.isSoftSwitchActive(ioFlag80Col)
	isStore80Active := a.mmu.store80Active
	isDoubleResMode := !isTextMode && is80Columns && !a.io.isSoftSwitchActive(ioFlagAnnunciator3)
	isSuperHighResMode := a.io.isSoftSwitchActive(ioDataNewVideo)

	isRGBCard := a.io.isSoftSwitchActive(ioFlagRGBCardActive)
	rgbFlag1 := a.io.isSoftSwitchActive(ioFlag1RGBCard)
	rgbFlag2 := a.io.isSoftSwitchActive(ioFlag2RGBCard)
	isMono560 := isDoubleResMode && !rgbFlag1 && !rgbFlag2
	isRGBMixMode := isDoubleResMode && !rgbFlag1 && rgbFlag2
	isRGB160Mode := isDoubleResMode && rgbFlag1 && !rgbFlag2

	isMixMode := a.io.isSoftSwitchActive(ioFlagMixed)

	var mode uint16
	if isSuperHighResMode {
		mode = videoSHR
		isMixMode = false
	} else if isTextMode {
		if is80Columns {
			mode = videoText80
		} else if isRGBCard && isStore80Active {
			mode = videoText40RGB
		} else {
			mode = videoText40
		}
		isMixMode = false
	} else if isHiResMode {
		if !isDoubleResMode {
			mode = videoHGR
		} else if isMono560 {
			mode = videoMono560
		} else if isRGBMixMode {
			mode = videoRGBMix
		} else if isRGB160Mode {
			mode = videoRGB160
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
		} else /* if isStore80Active {
			mode |= videoMixText40RGB
		}  else */{
			mode |= videoMixText40
		}
	}
	isSecondPage := a.io.isSoftSwitchActive(ioFlagSecondPage) && !a.mmu.store80Active
	if isSecondPage {
		mode |= videoSecondPage
	}

	return mode
}

func snapshotByMode(a *Apple2, videoMode uint16) *image.RGBA {
	videoBase := videoMode & videoBaseMask
	mixMode := videoMode & videoMixTextMask
	isSecondPage := (videoMode & videoSecondPage) != 0

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
	case videoText40RGB:
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
	case videoRGB160:
		snap = snapshotDoubleHiRes160ModeMono(a, isSecondPage, lightColor)
	case videoSHR:
		snap = snapshotSuperHiResMode(a)
		applyNTSCFilter = false
	}

	if applyNTSCFilter {
		snap = filterNTSCColor(snap, ntscMask)
	}

	if mixMode != 0 {
		var bottom *image.RGBA
		applyNTSCFilter := a.isColor
		switch mixMode {
		case videoMixText40:
			bottom = snapshotText40Mode(a, isSecondPage, lightColor)
		case videoMixText80:
			bottom = snapshotText80Mode(a, isSecondPage, lightColor)
		case videoMixText40RGB:
			bottom = snapshotText40RGBMode(a, isSecondPage)
			applyNTSCFilter = false
		}
		if applyNTSCFilter {
			bottom = filterNTSCColor(bottom, ntscMask)
		}
		snap = mixSnapshots(snap, bottom)
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
	bottomWidth := bottom.Bounds().Dx()

	// Copy bottom's bottom on top's bottom
	for y := hiResHeightMixed; y < hiResHeight; y++ {
		for x := 0; x < bottomWidth; x++ {
			c := bottom.At(x, y)
			top.Set(x, y, c)
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
