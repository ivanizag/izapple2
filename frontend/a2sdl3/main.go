//go:build !js

package main

import (
	"fmt"
	"image"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/ivanizag/izapple2"
	"github.com/ivanizag/izapple2/screen"

	"github.com/Zyko0/go-sdl3/bin/binsdl"
	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/pkg/profile"
)

func main() {
	a, err := izapple2.CreateConfiguredApple()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	if a != nil {
		if a.IsProfiling() {
			// See the log with:
			//    go tool pprof --pdf ~/go/bin/izapple2sdl3 /tmp/profile329536248/cpu.pprof > profile.pdf
			defer profile.Start().Stop()
		}

		sdlRun(a)
	}
}

func sdlRun(a *izapple2.Apple2) {
	// SDL3 is embedded in the binary. It is extracted to a temporary
	// directory, loaded, and removed when this function returns.
	defer binsdl.Load().Unload()

	err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_AUDIO | sdl.INIT_JOYSTICK)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize SDL: %v", err))
	}
	defer sdl.Quit()

	title := "iz-" + a.Name + " (F1 for help)"
	window, renderer, err := sdl.CreateWindowAndRenderer(title, 4*40*7+8, 4*24*8,
		sdl.WINDOW_RESIZABLE)
	if err != nil {
		panic("Failed to create window")
	}
	defer window.Destroy()
	defer renderer.Destroy()

	// SDL3 replaces the SDL2 hint SDL_HINT_RENDER_SCALE_QUALITY with a
	// per renderer default scale mode.
	renderer.SetDefaultTextureScaleMode(sdl.SCALEMODE_LINEAR)

	// Unlike SDL2, SDL3 does not send text input events unless asked to.
	err = window.StartTextInput()
	if err != nil {
		fmt.Printf("Error starting text input: %v.\n", err)
	}

	kp := newSDL3Keyboard(a)

	s := newSDL3Audio(a.GetClockMhz())
	for _, source := range a.GetAudioSources() {
		source.SetAudioSink(s.mixer.NewSource())
	}
	s.start()
	defer s.close()

	j := newSDL3Joysticks(!a.UsesMouse())
	a.SetJoysticksProvider(j)

	m := newSDL3Mouse()
	a.SetMouseProvider(m)

	// go-sdl3 asks for SIGINT and SIGTERM to remove the folder it extracted
	// SDL3 to, but it does not end the process afterwards. Asking for them too
	// puts them back to work: a Ctrl-C or a kill now quits the same way as
	// closing the window, releasing everything on the way out.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)
	go func() {
		<-signals
		err := sdl.PushEvent(&sdl.Event{Type: sdl.EVENT_QUIT})
		if err != nil {
			fmt.Printf("Error requesting the quit: %v.\n", err)
		}
	}()

	go a.Run()

	paused := false
	running := true
	for running {
		var event sdl.Event
		for sdl.PollEvent(&event) {
			switch event.Type {
			case sdl.EVENT_QUIT:
				a.SendCommand(izapple2.CommandKill)
				running = false
			case sdl.EVENT_KEY_DOWN, sdl.EVENT_KEY_UP:
				e := event.KeyboardEvent()
				kp.putKey(e)
				j.putKey(e)
			case sdl.EVENT_TEXT_INPUT:
				kp.putText(event.TextInputEvent().Text)
			case sdl.EVENT_JOYSTICK_AXIS_MOTION:
				j.putAxisEvent(event.JoyAxisEvent())
			case sdl.EVENT_JOYSTICK_BUTTON_DOWN, sdl.EVENT_JOYSTICK_BUTTON_UP:
				j.putButtonEvent(event.JoyButtonEvent())
			case sdl.EVENT_MOUSE_MOTION:
				w, h, _ := window.Size()
				e := event.MouseMotionEvent()
				j.putMouseMotionEvent(e, w, h)
				m.putMouseMotionEvent(e, w, h)
			case sdl.EVENT_MOUSE_BUTTON_DOWN, sdl.EVENT_MOUSE_BUTTON_UP:
				e := event.MouseButtonEvent()
				j.putMouseButtonEvent(e)
				m.putMouseButtonEvent(e)
			case sdl.EVENT_DROP_FILE:
				// Unlike SDL2, SDL3 reports where the file was dropped.
				e := event.DropEvent()
				w, _, _ := window.Size()
				drive := int(2 * int32(e.X) / w)
				fmt.Printf("Loading '%s' in drive %v\n", e.Data, drive+1)
				a.SendLoadDisk(drive, e.Data)
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
				// image.RGBA stores the bytes as R, G, B, A. That is
				// PIXELFORMAT_RGBA32, an alias resolved for the host endianness.
				surface, err := sdl.CreateSurfaceFrom(
					img.Bounds().Dx(), img.Bounds().Dy(),
					sdl.PIXELFORMAT_RGBA32,
					img.Pix, 4*img.Bounds().Dx())
				if err != nil {
					panic(err)
				}

				texture, err := renderer.CreateTextureFromSurface(surface)
				if err != nil {
					panic(err)
				}
				// The surface points at img.Pix without owning it.
				runtime.KeepAlive(img)

				renderer.Clear()
				renderer.RenderTexture(texture, nil, nil)
				renderer.Present()
				surface.Destroy()
				texture.Destroy()
			}
		}
		sdl.Delay(1000 / 30)
	}
}

var helpMessage = `
          F1: Show/Hide help
     Ctrl-F2: Reset
      F1, F2: Reset
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
