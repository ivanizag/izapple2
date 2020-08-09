package apple2

import (
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

func filterNTSCColor(in *image.RGBA, mask *image.Alpha) *image.RGBA {
	colorMap := ntscColorMap // or rgbColorMap

	b := in.Bounds()
	size := image.Rect(0, 0, b.Dx()+3, b.Dy())
	out := image.NewRGBA(size)

	for y := b.Min.Y; y < b.Max.Y; y++ {
		// We store the last four bits. We start with 0000
		v := 0
		for x := b.Min.X; x < b.Dx(); x++ {
			cIn := in.At(x, y)
			r, _, _, _ := cIn.RGBA()

			pos := 1 << uint(x%4)
			if r != 0 {
				v |= pos
			} else {
				v &^= pos
			}

			cOut := colorMap[v]
			if mask != nil {
				// RGM mode7
				_, _, _, a := mask.At(x, y).RGBA()
				if a > 0 {
					cOut = cIn
				}
			}

			out.Set(x, y, cOut)
		}

		// We fade for the last three positions
		for x := b.Dx(); x < b.Max.X; x++ {
			v >>= 1
			cOut := colorMap[v]
			out.Set(x, y, cOut)
		}
	}
	return out
}
