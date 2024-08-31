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

	kp := newSDLKeyBoard(a)

	s := newSDLSpeaker()
	s.start()
	a.SetSpeakerProvider(s)

	j := newSDLJoysticks(true)
	a.SetJoysticksProvider(j)

	m := newSDLMouse()
	a.SetMouseProvider(m)

	go a.Run()

	var x int32
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
				x = t.X
			case *sdl.MouseButtonEvent:
				j.putMouseButtonEvent(t)
				m.putMouseButtonEvent(t)
			case *sdl.DropEvent:
				switch t.Type {
				case sdl.DROPFILE:
					w, _ := window.GetSize()
					drive := int(2 * x / w)
					fmt.Printf("Loading '%s' in drive %v\n", t.File, drive+1)
					a.SendLoadDisk(drive, t.File)
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
			var img *image.RGBA
			vs := a.GetVideoSource()
			if kp.showHelp {
				img = screen.SnapshotMessageGenerator(vs, helpMessage)
			} else if kp.showCharGen {
				cgPage, cgPages := a.GetCgPageInfo()
				img = screen.SnapshotCharacterGenerator(vs, kp.showAltText)
				window.SetTitle(fmt.Sprintf("%v character map, page %v/%v", a.Name, cgPage+1, cgPages))
			} else if kp.showPages {
				img = screen.SnapshotParts(vs, kp.screenMode)
				window.SetTitle(fmt.Sprintf("%v %v %vx%v", a.Name, screen.VideoModeName(vs), img.Rect.Dx()/2, img.Rect.Dy()/2))
			} else {
				img = screen.Snapshot(vs, kp.screenMode)
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

var helpMessage = `

          F1: Show/Hide help
     Ctrl-F2: Reset
          F4: Show/Hide CPU trace
          F5: Fast/Normal speed
     Ctrl-F5: Show speed
          F6: Next screen mode
          F7: Show/Hide pages
         F10: Next character set
    Ctrl-F10: Show/Hide character set
   Shift-F10: Show/Hide alternate text
         F12: Save screen snapshot
       Pause: Pause the emulation

  Left alt or option key: Open-Apple
 Right alt or option key: Closed-Apple

Drop a file on the left or right
side of the window to load a disk

 Run izapple2 -h for more options
   https://github.com/ivanizag/izapple2
`

///////////////////////////////////////
