package main

import (
	"flag"
	"unsafe"

	"github.com/ivanizag/apple2"
	"github.com/pkg/profile"
	"github.com/veandco/go-sdl2/sdl"
)

func main() {
	prof := flag.Bool(
		"profile",
		false,
		"generate profile trace to analyse with pprof")
	flag.Parse()
	if *prof {
		// See the log with:
		//    go tool pprof --pdf ~/go/bin/apple2sdl /tmp/profile329536248/cpu.pprof > profile.pdf
		defer profile.Start().Stop()
	}

	a := apple2.MainApple()
	if a != nil {
		SDLRun(a)
	}
}

// SDLRun starts the Apple2 emulator on SDL
func SDLRun(a *apple2.Apple2) {

	window, renderer, err := sdl.CreateWindowAndRenderer(4*40*7, 4*24*8,
		sdl.WINDOW_SHOWN)
	if err != nil {
		panic("Failed to create window")
	}
	window.SetResizable(true)

	defer window.Destroy()
	defer renderer.Destroy()
	window.SetTitle(a.Name)

	kp := newSDLKeyBoard(a)
	a.SetKeyboardProvider(kp)

	s := newSDLSpeaker()
	s.start()
	a.SetSpeakerProvider(s)

	j := newSDLJoysticks()
	a.SetJoysticksProvider(j)

	go a.Run()

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				a.SendCommand(apple2.CommandKill)
				running = false
			case *sdl.KeyboardEvent:
				kp.putKey(t)
				j.putKey(t)
			case *sdl.TextInputEvent:
				kp.putText(t)
			case *sdl.JoyAxisEvent:
				j.putAxisEvent(t)
			case *sdl.JoyButtonEvent:
				j.putButtonEvent(t)
			}
		}

		img := apple2.Snapshot(a)
		if img != nil {
			surface, err := sdl.CreateRGBSurfaceFrom(unsafe.Pointer(&img.Pix[0]),
				int32(img.Bounds().Dx()), int32(img.Bounds().Dy()),
				32, 4*img.Bounds().Dx(),
				0x0000ff, 0x0000ff00, 0x00ff0000, 0xff000000)
			// Valid for little endian. Should we reverse for big endian?
			// 0xff000000, 0x00ff0000, 0x0000ff00, 0x000000ff)

			if err != nil {
				panic(err)
			}

			texture, err := renderer.CreateTextureFromSurface(surface)
			if err != nil {
				panic(err)
			}

			renderer.Clear()
			w, h := window.GetSize()
			renderer.Copy(texture, nil, &sdl.Rect{X: 0, Y: 0, W: w, H: h})
			renderer.Present()

			surface.Free()
			texture.Destroy()
		}
		sdl.Delay(1000 / 30)
	}

}
