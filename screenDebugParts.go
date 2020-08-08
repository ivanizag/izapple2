package apple2

import (
	"image"
)

// SnapshotParts the currently visible screen
func (a *Apple2) SnapshotParts() *image.RGBA {
	videoMode := getCurrentVideoMode(a)
	isSecondPage := (videoMode & videoSecondPage) != 0
	videoBase := videoMode & videoBaseMask

	snapScreen := snapshotByMode(a, videoMode)
	snapPage1 := snapshotByMode(a, videoMode&^videoSecondPage)
	snapPage2 := snapshotByMode(a, videoMode|videoSecondPage)
	var snapAux *image.RGBA

	/*
		if videoBase == videoRGBMix {
		_, mask := snapshotDoubleHiResModeMono(a, isSecondPage, true /*isRGBMixMode*/ /*, color.White)
		snapAux = filterMask(mask)
	}*/

	if videoBase == videoRGBText40 {
		snapAux = snapshotText40RGBModeColors(a, isSecondPage)
	} else if (videoMode & videoMixText80) != 0 {
		snapAux = snapshotByMode(a, videoText80)
	} else {
		snapAux = snapshotByMode(a, videoText40)
	}

	return mixFourSnapshots([]*image.RGBA{snapScreen, snapAux, snapPage1, snapPage2})
}

// VideoModeName returns the name of the current video mode
func (a *Apple2) VideoModeName() string {
	videoMode := getCurrentVideoMode(a)
	videoBase := videoMode & videoBaseMask

	var name string
	applyNTSCFilter := a.isColor

	switch videoBase {
	case videoText40:
		name = "TEXT40COL"
		applyNTSCFilter = false
	case videoText80:
		name = "TEXT80COL"
		applyNTSCFilter = false
	case videoRGBText40:
		name = "TEXT40COLRGB"
		applyNTSCFilter = false
	case videoGR:
		name = "GR"
	case videoDGR:
		name = "DGR"
	case videoHGR:
		name = "HGR"
	case videoDHGR:
		name = "DHGR"
	case videoMono560:
		name = "Mono560"
		applyNTSCFilter = false
	case videoRGBMix:
		name = "RGMMIX"
	case videoSHR:
		name = "SHR"
		applyNTSCFilter = false
	default:
		name = "Unknown video mode"
	}

	if (videoMode & videoSecondPage) != 0 {
		name += "-PAGE2"
	}

	if (videoMode & videoMixText40) != 0 {
		name += "-MIX40"
	}

	if (videoMode & videoMixText80) != 0 {
		name += "-MIX80"
	}

	if applyNTSCFilter {
		name += "-NTSC"
	}
	return name
}

func mixFourSnapshots(snaps []*image.RGBA) *image.RGBA {
	size := image.Rect(0, 0, hiResWidth*4, hiResHeight*2)
	out := image.NewRGBA(size)

	for i := 0; i < 4; i++ {
		if snaps[i].Bounds().Dx() < hiResWidth*2 {
			snaps[i] = doubleWidthFilter(snaps[i])
		}
	}

	for y := 0; y < hiResHeight; y++ {
		for x := 0; x < hiResWidth*2; x++ {
			out.Set(x, y, snaps[0].At(x, y))
			out.Set(x+hiResWidth*2, y, snaps[1].At(x, y))
			out.Set(x, y+hiResHeight, snaps[2].At(x, y))
			out.Set(x+hiResWidth*2, y+hiResHeight, snaps[3].At(x, y))
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
