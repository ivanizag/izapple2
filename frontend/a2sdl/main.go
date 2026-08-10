//go:build !js

package main

import (
	"fmt"
	"unsafe"

	"github.com/ivanizag/izapple2"
	"github.com/ivanizag/izapple2/frontend/shared"

	"github.com/pkg/profile"
	"github.com/veandco/go-sdl2/sdl"
)

func main() {
	a, err := izapple2.CreateConfiguredApple()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
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

	title := "iz-" + a.Name + " (F1 for help)"
	window.SetTitle(title)

	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "best")

	// We only care about DROPFILE. Besides, sdl2-compat (the SDL2 API on top of
	// SDL3, used by recent Linux distributions) leaves the file field of the
	// DROPBEGIN and DROPCOMPLETE events with the window id instead of the NULL
	// the SDL2 docs promise. go-sdl2 then reads a string from that bogus
	// pointer and the emulator crashes when a disk is dropped. See
	// https://github.com/ivanizag/izapple2/issues/39
	sdl.EventState(sdl.DROPBEGIN, sdl.DISABLE)
	sdl.EventState(sdl.DROPCOMPLETE, sdl.DISABLE)

	view := shared.NewView()
	kp := newSDLKeyBoard(a, view)

	s := newSDLAudio(a)
	s.start()

	j := newSDLJoysticks(!a.UsesMouse())
	a.SetJoysticksProvider(j.paddles)

	m := shared.NewMouse()
	a.SetMouseProvider(m)

	d := newSDLDropTargets(a, window)

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
			case *sdl.WindowEvent:
				if t.Event == sdl.WINDOWEVENT_FOCUS_LOST {
					j.paddles.ReleaseAppleKeys()
				}
			case *sdl.TextInputEvent:
				kp.putText(t.GetText())
			case *sdl.JoyAxisEvent:
				j.putAxisEvent(t)
			case *sdl.JoyButtonEvent:
				j.putButtonEvent(t)
			case *sdl.MouseMotionEvent:
				w, h := window.GetSize()
				j.putMouseMotionEvent(t, w, h)
				m.SetPosition(int(t.X), int(t.Y), int(w), int(h))
			case *sdl.MouseButtonEvent:
				j.putMouseButtonEvent(t)
				if sdlMouseButton(t.Button) == shared.MouseButtonLeft {
					m.SetButton(t.State == sdl.PRESSED)
				}
			case *sdl.DropEvent:
				switch t.Type {
				case sdl.DROPFILE:
					drive := d.dropped()
					if drive >= 0 {
						fmt.Printf("Loading '%s' in drive %v\n", t.File, drive+1)
						a.SendLoadDisk(drive, t.File)
					} else {
						fmt.Printf("There are no drives to load '%s' on\n", t.File)
					}
				}
			}
		}

		if paused != a.IsPaused() {
			if a.IsPaused() {
				window.SetTitle(title + " - PAUSED!")
			} else {
				window.SetTitle(title)
			}
			paused = a.IsPaused()
		}

		if !a.IsPaused() {
			img, viewTitle := view.Snapshot(a, d.targets, d.pointedDrive())
			if viewTitle != "" {
				window.SetTitle(viewTitle)
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

///////////////////////////////////////
