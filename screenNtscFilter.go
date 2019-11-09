package apple2

import (
	"image"
	"image/color"
)

func getNTSCColorMap() []color.Color {
	// RGB values from https://mrob.com/pub/xapple2/colors.html
	black := color.RGBA{0, 0, 0, 255} /*COLOR=0 */ /*  0   0    0*/

	dkBlue := color.RGBA{96, 78, 189, 255} /*COLOR=2 */ /*  0  60   25*/
	red := color.RGBA{227, 30, 96, 255}    /*COLOR=1 */ /* 90  60   25*/
	brown := color.RGBA{96, 114, 3, 255}   /*COLOR=8 */ /*180  60   25*/
	dkGreen := color.RGBA{0, 163, 96, 255} /*COLOR=4 */ /*270  60   25*/

	grey := color.RGBA{156, 156, 156, 255}  /*COLOR=10*/ /*  0   0   50*/
	purple := color.RGBA{255, 68, 253, 255} /*HCOLOR=2*/ /* 45 100   50*/
	orange := color.RGBA{255, 106, 60, 255} /*HCOLOR=5*/ /*135 100   50*/
	green := color.RGBA{20, 245, 60, 255}   /*HCOLOR=1*/ /*225 100   50*/
	blue := color.RGBA{20, 207, 253, 255}   /*HCOLOR=6*/ /*315 100   50*/

	//purple := color.RGBA{255, 68, 253, 255} /*COLOR=3 */ /* 45 100   50*/
	//orange := color.RGBA{255, 106, 60, 255} /*COLOR=9 */ /*135 100   50*/
	//green := color.RGBA{20, 245, 60, 255} /*COLOR=12*/ /*225 100   50*/
	//blue := color.RGBA{20, 207, 253, 255} /*COLOR=6 */ /*315 100   50*/

	ltBlue := color.RGBA{208, 195, 255, 255} /*COLOR=7 */ /*  0  60   75*/
	pink := color.RGBA{255, 160, 208, 255}   /*COLOR=11*/ /* 90  60   75*/
	yellow := color.RGBA{208, 221, 141, 255} /*COLOR=13*/ /*180  60   75*/
	aqua := color.RGBA{114, 255, 208, 255}   /*COLOR=14*/ /*270  60   75*/

	white := color.RGBA{255, 255, 255, 255} /*COLOR=15*/ /*  0   0  100*/

	colorMap := []color.Color{
		/* 0000 */ black,
		/* 0001 */ brown,
		/* 0010 */ dkGreen,
		/* 0011 */ green,
		/* 0100 */ dkBlue,
		/* 0101 */ grey,
		/* 0110 */ blue,
		/* 0111 */ aqua,
		/* 1000 */ red,
		/* 1001 */ orange,
		/* 1010 */ grey,
		/* 1011 */ yellow,
		/* 1100 */ purple,
		/* 1101 */ pink,
		/* 1110 */ ltBlue,
		/* 1111 */ white,
	}

	return colorMap
}

func filterNTSCColor(blacker bool, in *image.RGBA) *image.RGBA {
	colorMap := getNTSCColorMap()

	b := in.Bounds()
	size := image.Rect(0, 0, b.Dx()+3, b.Dy())
	out := image.NewRGBA(size)

	for y := b.Min.Y; y < b.Max.Y; y++ {
		// We store the last four bits. We start with 0000
		v := 0
		for x := b.Min.X; x < b.Dx(); x++ {
			cIn := in.At(x, y)
			r, _, _, _ := cIn.RGBA()

			pos := 1 << (3 - uint(x%4))
			var cOut color.Color
			if r != 0 {
				v |= pos
				cOut = colorMap[v]
			} else {
				v &^= pos
				if blacker {
					// If there is no luminance, let's have black anyway
					cOut = colorMap[0]
				} else {
					cOut = colorMap[v]
				}
			}
			out.Set(x, y, cOut)
		}

		// We fade for the last three positions
		for x := b.Dx(); x < b.Max.X; x++ {
			v = (v << 1) & 0xF
			cOut := colorMap[v]
			out.Set(x, y, cOut)
		}
	}
	return out
}
