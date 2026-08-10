//go:build !js

package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/ivanizag/izapple2"
	"github.com/ivanizag/izapple2/frontend/shared"

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

	view := shared.NewView()
	kp := newSDL3Keyboard(a, view)

	s := newSDL3Audio(a)
	s.start()
	defer s.close()

	j := newSDL3Joysticks(!a.UsesMouse())
	a.SetJoysticksProvider(j.paddles)

	m := shared.NewMouse()
	a.SetMouseProvider(m)

	d := newSDL3DropTargets(a, window)

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
			case sdl.EVENT_WINDOW_FOCUS_LOST:
				j.paddles.ReleaseAppleKeys()
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
				m.SetPosition(int(e.X), int(e.Y), int(w), int(h))
				d.dragEnded()
			case sdl.EVENT_MOUSE_BUTTON_DOWN, sdl.EVENT_MOUSE_BUTTON_UP:
				e := event.MouseButtonEvent()
				j.putMouseButtonEvent(e)
				if sdl3MouseButton(e.Button) == shared.MouseButtonLeft {
					m.SetButton(e.Down)
				}
			case sdl.EVENT_DROP_BEGIN:
				// Unlike SDL2, SDL3 reports the file being dragged over the
				// window, so the drop targets can be shown while it moves.
				d.dragStarted()
			case sdl.EVENT_DROP_POSITION:
				d.dragMoved(event.DropEvent().X)
			case sdl.EVENT_DROP_COMPLETE:
				d.dragEnded()
			case sdl.EVENT_DROP_FILE:
				e := event.DropEvent()
				d.dragEnded()
				drive := d.dropped(e.X)
				if drive >= 0 {
					fmt.Printf("Loading '%s' in drive %v\n", e.Data, drive+1)
					a.SendLoadDisk(drive, e.Data)
				} else {
					fmt.Printf("There are no drives to load '%s' on\n", e.Data)
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

///////////////////////////////////////
