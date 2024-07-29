package screen

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
	// ScreenModeGreen to render as a green phosphor monitor
	ScreenModeGreen = iota
	// ScreenModePlain to render in color with filled areas
	ScreenModePlain
	// ScreenModeNTSC shows spaces between pixels
	ScreenModeNTSC
)

// Snapshot the currently visible screen
func Snapshot(vs VideoSource, screenMode int) *image.RGBA {
	videoMode := vs.GetCurrentVideoMode()
	snap := snapshotByMode(vs, videoMode, screenMode)

	if screenMode != ScreenModePlain && snap.Bounds().Dy() == hiResHeight {
		// Apply the filter to regular CRT snapshots with 192 lines. Not to SHR
		snap = linesSeparatedFilter(snap)
	}

	return snap
}

// SnapshotPaletted, snapshot of the currently visible screen as a paletted image
func SnapshotPaletted(vs VideoSource, screenMode int) *image.Paletted {
	img := Snapshot(vs, screenMode)
	return palletedFilter(img)
}

// Color for typical Apple ][ period green P1 phosphor monitors
// See: https://superuser.com/questions/361297/what-colour-is-the-dark-green-on-old-fashioned-green-screen-computer-displays
var greenPhosphorColor = color.RGBA{65, 255, 0, 255}

func snapshotByMode(vs VideoSource, videoMode uint32, screenMode int) *image.RGBA {
	videoBase := videoMode & VideoBaseMask
	mixMode := videoMode & VideoMixTextMask
	isSecondPage := (videoMode & VideoSecondPage) != 0
	isAltText := (videoMode & VideoAltText) != 0
	isRGBCard := (videoMode & VideoRGBCard) != 0
	shiftSupported := (videoMode & VideoFourColors) == 0
	hasAltOrder := (videoMode & VideoText80AltOrder) != 0

	var lightColor color.Color = color.White
	if screenMode == ScreenModeGreen {
		lightColor = greenPhosphorColor
	}

	applyNTSCFilter := screenMode != ScreenModeGreen
	var snap *image.RGBA
	var ntscMask *image.Alpha
	switch videoBase {
	case VideoText40:
		snap = snapshotText40(vs, isSecondPage, isAltText, lightColor)
		applyNTSCFilter = false
	case VideoText80:
		snap = snapshotText80(vs, isSecondPage, isAltText, hasAltOrder, lightColor)
		applyNTSCFilter = false
	case VideoText40RGB:
		snap = snapshotText40RGB(vs, isSecondPage, isAltText)
		applyNTSCFilter = false
	case VideoGR:
		snap = snapshotLoRes(vs, isSecondPage, lightColor)
	case VideoDGR:
		snap = snapshotMeRes(vs, isSecondPage, lightColor)
	case VideoHGR:
		snap = snapshotHiRes(vs, isSecondPage, lightColor, shiftSupported)
	case VideoDHGR:
		snap, _ = snapshotDoubleHiRes(vs, isSecondPage, false /*isRGBMixMode*/, lightColor)
	case VideoMono560:
		snap, _ = snapshotDoubleHiRes(vs, isSecondPage, false /*isRGBMixMode*/, lightColor)
		applyNTSCFilter = false
	case VideoRGBMix:
		snap, ntscMask = snapshotDoubleHiRes(vs, isSecondPage, true /*isRGBMixMode*/, lightColor)
	case VideoRGB160:
		snap = snapshotDoubleHiRes160(vs, isSecondPage, lightColor)
	case VideoSHR:
		snap = snapshotSuperHiRes(vs)
		applyNTSCFilter = false
	case VideoVidex:
		snap = vs.GetCardImage(lightColor)
		applyNTSCFilter = false
	}

	if applyNTSCFilter {
		snap = filterNTSCColor(snap, ntscMask, screenMode)
	}

	if mixMode != 0 {
		var bottom *image.RGBA
		applyNTSCFilter := screenMode != ScreenModeGreen && !isRGBCard
		switch mixMode {
		case VideoMixText40:
			bottom = snapshotText40(vs, isSecondPage, isAltText, lightColor)
		case VideoMixText80:
			bottom = snapshotText80(vs, isSecondPage, isAltText, hasAltOrder, lightColor)
		case VideoMixText40RGB:
			bottom = snapshotText40RGB(vs, isSecondPage, isAltText)
			applyNTSCFilter = false
		}
		if applyNTSCFilter {
			bottom = filterNTSCColor(bottom, ntscMask, screenMode)
		}
		snap = mixSnapshots(snap, bottom)
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
func SaveSnapshot(vs VideoSource, screenMode int, filename string) error {
	img := Snapshot(vs, screenMode)
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

func palletedFilter(in *image.RGBA) *image.Paletted {
	bounds := in.Bounds()
	outBounds := image.Rect(0, 0, bounds.Dx()*2, bounds.Dy())
	palette := []color.Color{color.Black, color.White, greenPhosphorColor}
	palette = append(palette, ntscColorMap[:]...)
	palette = append(palette, attenuatedColorMap[:]...)
	paletted := image.NewPaletted(outBounds, palette)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := in.At(x, y)
			paletted.Set(x*2, y, c)
			paletted.Set(x*2+1, y, c)
		}
	}
	return paletted
}
