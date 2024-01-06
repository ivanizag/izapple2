package screen

import (
	"fmt"
	"image"
	"image/color"
)

/*
See:

	https://mrob.com/pub/xapple2/colors.html
	https://archive.org/details/IIgs_2523063_Master_Color_Values
*/
var ntscColorMap = [16]color.Color{
	color.RGBA{0, 0, 0, 255},       // Black
	color.RGBA{227, 30, 96, 255},   // Magenta
	color.RGBA{96, 78, 189, 255},   // Dark Blue
	color.RGBA{255, 68, 253, 255},  // Purple
	color.RGBA{0, 163, 96, 255},    // Dark Green
	color.RGBA{156, 156, 156, 255}, // Grey 1
	color.RGBA{20, 207, 253, 255},  // Medium Blue
	color.RGBA{208, 195, 255, 255}, // Light Blue
	color.RGBA{96, 114, 3, 255},    // Brown
	color.RGBA{255, 106, 60, 255},  // Orange
	color.RGBA{156, 156, 156, 255}, // Grey 2
	color.RGBA{255, 160, 208, 255}, // Pink
	color.RGBA{20, 245, 60, 255},   // Green
	color.RGBA{208, 221, 141, 255}, // Yellow
	color.RGBA{114, 255, 208, 255}, // Aquamarine
	color.RGBA{255, 255, 255, 255}, // White
}

var attenuatedColorMap = buildAttenuatedColorMap(ntscColorMap)

func buildAttenuatedColorMap(colorMap [16]color.Color) [16]color.Color {
	colors := [16]color.Color{}
	for i := 0; i < len(colorMap); i++ {
		r, g, b, _ := colorMap[i].RGBA()
		colors[i] = color.RGBA64{
			uint16(r / 2), uint16(g / 2), uint16(b / 2),
			65535,
		}
	}
	return colors
}

/*
var rgbColorMap = [16]color.Color{
	color.RGBA{0, 0, 0, 255},       // Black
	color.RGBA{221, 0, 51, 255},    // Magenta
	color.RGBA{0, 0, 153, 255},     // Dark Blue
	color.RGBA{221, 34, 221, 255},  // Purple
	color.RGBA{0, 119, 34, 255},    // Dark Green
	color.RGBA{85, 85, 85, 255},    // Grey 1
	color.RGBA{34, 34, 255, 255},   // Medium Blue
	color.RGBA{102, 170, 255, 255}, // Light Blue
	color.RGBA{136, 85, 0, 255},    // Brown
	color.RGBA{255, 102, 0, 255},   // Orange
	color.RGBA{170, 170, 170, 255}, // Grey 2
	color.RGBA{255, 153, 136, 255}, // Pink
	color.RGBA{17, 221, 0, 255},    // Green
	color.RGBA{255, 255, 0, 255},   // Yellow
	color.RGBA{68, 255, 153, 255},  // Aquamarine
	color.RGBA{255, 255, 255, 255}, // White
}
*/

func filterNTSCColor(in *image.RGBA, mask *image.Alpha, screenMode int) *image.RGBA {
	colorMap := ntscColorMap // or rgbColorMap
	colorMapLow := ntscColorMap
	if screenMode == ScreenModeNTSC {
		colorMapLow = attenuatedColorMap
	}

	b := in.Bounds()
	width := b.Dx()
	height := b.Dy()
	size := image.Rect(0, 0, width+4, height)
	out := image.NewRGBA(size)

	if width < 2*hiResWidth {
		panic(fmt.Sprintf("The image has width %v. We can't apply the NTSC filter.", width))
	}

	for y := 0; y < height; y++ {
		// We store the last four bits. We start with 0000
		v := 0
		for x := 0; x < width; x++ {
			cIn := in.At(x, y)
			r, _, _, _ := cIn.RGBA()

			pos := 1 << uint(x%4)
			if r != 0 {
				v |= pos
			} else {
				v &^= pos
			}

			var cOut color.Color
			if r != 0 {
				cOut = colorMap[v]
			} else {
				cOut = colorMapLow[v]
			}
			if mask != nil {
				// RGB mode7
				_, _, _, a := mask.At(x, y).RGBA()
				if a > 0 {
					cOut = cIn
				}
			}

			out.Set(x, y, cOut)
		}

		// We fade for the last three positions
		for x := width; x < width+4; x++ {
			v >>= 1
			cOut := colorMap[v]
			out.Set(x, y, cOut)
		}
	}
	return out
}
