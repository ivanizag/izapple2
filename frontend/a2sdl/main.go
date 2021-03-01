package main

import (
	"fmt"
	"image"
	"unsafe"

	"github.com/ivanizag/izapple2"
	"github.com/ivanizag/izapple2/screen"

	"github.com/pkg/profile"
	"github.com/veandco/go-sdl2/sdl"
)

func main() {
	a := izapple2.MainApple()
	if a != nil {
		if a.IsProfiling() {
			// See the log with:
			//    go tool pprof --pdf ~/go/bin/izapple2sdl /tmp/profile329536248/cpu.pprof > profile.pdf
			defer profile.Start().Stop()
		}

		sdlRun(a)
	}
}

func sdlRun(a *izapple2.Apple2) {

	window, renderer, err := sdl.CreateWindowAndRenderer(4*40*7+8, 4*24*8,
		sdl.WINDOW_SHOWN)
	if err != nil {
		panic("Failed to create window")
	}
	window.SetResizable(true)

	defer window.Destroy()
	defer renderer.Destroy()
	window.SetTitle("iz-" + a.Name)

	kp := newSDLKeyBoard(a)

	s := newSDLSpeaker()
	s.start()
	a.SetSpeakerProvider(s)

	j := newSDLJoysticks(true)
	a.SetJoysticksProvider(j)

	m := newSDLMouse()
	a.SetMouseProvider(m)

	go a.Run()

	paused := false
	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				a.SendCommand(izapple2.CommandKill)
				running = false
			case *sdl.KeyboardEvent:
				kp.putKey(t)
				j.putKey(t)
			case *sdl.TextInputEvent:
				kp.putText(t.GetText())
			case *sdl.JoyAxisEvent:
				j.putAxisEvent(t)
			case *sdl.JoyButtonEvent:
				j.putButtonEvent(t)
			case *sdl.MouseMotionEvent:
				w, h := window.GetSize()
				j.putMouseMotionEvent(t, w, h)
				m.putMouseMotionEvent(t, w, h)
			case *sdl.MouseButtonEvent:
				j.putMouseButtonEvent(t)
				m.putMouseButtonEvent(t)
			}
		}

		if paused != a.IsPaused() {
			if a.IsPaused() {
				window.SetTitle("iz-" + a.Name + " - PAUSED!")
			} else {
				window.SetTitle("iz-" + a.Name)
			}
			paused = a.IsPaused()
		}

		if !a.IsPaused() {
			var img *image.RGBA
			if kp.showCharGen {
				img = screen.SnapshotCharacterGenerator(a, kp.showAltText)
				window.SetTitle(fmt.Sprintf("%v character map", a.Name))
			} else if kp.showPages {
				img = screen.SnapshotParts(a, screen.ScreenModeNTSC)
				window.SetTitle(fmt.Sprintf("%v %v %vx%v", a.Name, screen.VideoModeName(a), img.Rect.Dx()/2, img.Rect.Dy()/2))
			} else {
				img = screen.Snapshot(a, screen.ScreenModeNTSC)
			}
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
				renderer.Copy(texture, nil, nil)
				renderer.Present()
				surface.Free()
				texture.Destroy()
			}
		}
		sdl.Delay(1000 / 30)
	}

}
