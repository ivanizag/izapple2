package main

/*
#include "libretro.h"
#include "shim.h"
*/
import "C"

import (
	"unsafe"

	"github.com/ivanizag/izapple2/screen"
)

/*
The core renders every mode into a framebuffer of a fixed size, so the geometry
never has to change.

Only the screen modes without scan lines are used, so the image keeps its 192
lines instead of becoming four times taller. Drawing the gaps is left to the
shaders of the frontend, which do it better. The colour of the phosphor is
another matter and is a core option: no shader turns the colour picture into the
half width pixels of a green monitor.

With super hi-res, the Videx video and the RGB card left out, every remaining
mode is generated as 560x192, and the NTSC color filter widens it by four
pixels. Green skips that filter and stays at 560, so the columns it does not
cover are left black.
*/
const (
	frameWidth  = 564
	frameHeight = 192
)

type videoOutput struct {
	buffer     []uint32
	screenMode int
}

func newVideoOutput(screenMode int) *videoOutput {
	return &videoOutput{
		buffer:     make([]uint32, frameWidth*frameHeight),
		screenMode: screenMode,
	}
}

func (v *videoOutput) setScreenMode(screenMode int) {
	v.screenMode = screenMode
}

// render builds the frame and hands it to the frontend as XRGB8888
func (v *videoOutput) render(vs screen.VideoSource) {
	img := screen.Snapshot(vs, v.screenMode)
	if img == nil {
		return
	}

	// A mode wider or taller than the framebuffer should not happen with the
	// cards this core allows, but crop it instead of writing out of bounds
	bounds := img.Bounds()
	width := min(bounds.Dx(), frameWidth)
	height := min(bounds.Dy(), frameHeight)

	for y := range height {
		src := y * img.Stride
		dst := y * frameWidth
		for x := range width {
			p := src + 4*x
			v.buffer[dst+x] = uint32(img.Pix[p])<<16 |
				uint32(img.Pix[p+1])<<8 |
				uint32(img.Pix[p+2])
		}
		for x := width; x < frameWidth; x++ {
			v.buffer[dst+x] = 0
		}
	}
	for y := height; y < frameHeight; y++ {
		dst := y * frameWidth
		for x := range frameWidth {
			v.buffer[dst+x] = 0
		}
	}

	C.shim_video_refresh(
		unsafe.Pointer(&v.buffer[0]),
		C.uint(frameWidth),
		C.uint(frameHeight),
		C.size_t(frameWidth*4))
}
