package screen

import (
	"image"
)

// SnapshotParts the currently visible screen
func SnapshotParts(vs VideoSource, screenMode int) *image.RGBA {
	videoMode := vs.GetCurrentVideoMode()
	isSecondPage := (videoMode & VideoSecondPage) != 0
	videoBase := videoMode & VideoBaseMask
	mixMode := videoMode & VideoMixTextMask
	modifiers := videoMode & VideoModifiersMask

	snapScreen := snapshotByMode(vs, videoMode, screenMode)
	snapPage1 := snapshotByMode(vs, videoMode&^VideoSecondPage, screenMode)
	snapPage2 := snapshotByMode(vs, videoMode|VideoSecondPage, screenMode)
	var snapAux *image.RGBA

	/*
		if videoBase == videoRGBMix {
		_, mask := snapshotDoubleHiResModeMono(a, isSecondPage, true /*isRGBMixMode*/ /*, color.White)
		snapAux = filterMask(mask)
	}*/

	if videoBase == VideoText40RGB {
		snapAux = snapshotText40RGBColors(vs, isSecondPage)
	} else {
		switch mixMode {
		case VideoMixText80:
			snapAux = snapshotByMode(vs, VideoText80|modifiers, screenMode)
		case VideoMixText40RGB:
			snapAux = snapshotByMode(vs, VideoText40RGB|modifiers, screenMode)
		default:
			snapAux = snapshotByMode(vs, VideoText40|modifiers, screenMode)
		}
	}

	return mixFourSnapshots([]*image.RGBA{snapScreen, snapAux, snapPage1, snapPage2})
}

// VideoModeName returns the name of the current video mode
func VideoModeName(vs VideoSource) string {
	videoMode := vs.GetCurrentVideoMode()
	videoBase := videoMode & VideoBaseMask
	mixMode := videoMode & VideoMixTextMask

	var name string

	switch videoBase {
	case VideoText40:
		name = "TEXT40COL"
	case VideoText80:
		name = "TEXT80COL"
	case VideoText40RGB:
		name = "TEXT40COLRGB"
	case VideoGR:
		name = "GR"
	case VideoDGR:
		name = "DGR"
	case VideoHGR:
		name = "HGR"
	case VideoDHGR:
		name = "DHGR"
	case VideoMono560:
		name = "Mono560"
	case VideoRGBMix:
		name = "RGBMIX"
	case VideoRGB160:
		name = "RGB160"
	case VideoSHR:
		name = "SHR"
	default:
		name = "Unknown video mode"
	}

	if (videoMode & VideoSecondPage) != 0 {
		name += "-PAGE2"
	}

	switch mixMode {
	case VideoMixText40:
		name += "-MIX40"
	case VideoMixText80:
		name += "-MIX80"
	case VideoMixText40RGB:
		name += "-MIX40RGB"
	}

	return name
}

func mixFourSnapshots(snaps []*image.RGBA) *image.RGBA {
	width := snaps[0].Rect.Dx()
	height := snaps[0].Rect.Dy()
	size := image.Rect(0, 0, width*2, height*2)
	out := image.NewRGBA(size)

	for i := 1; i < 4; i++ {
		if snaps[i].Bounds().Dx() < width {
			snaps[i] = doubleWidthFilter(snaps[i])
		}
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			out.Set(x, y, snaps[0].At(x, y))
			out.Set(x+width, y, snaps[1].At(x, y))
			out.Set(x, y+height, snaps[2].At(x, y))
			out.Set(x+width, y+height, snaps[3].At(x, y))
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
