package main

import (
	"fmt"
	"image"
	"time"

	"github.com/ivanizag/izapple2"

	"github.com/pkg/profile"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/driver/desktop"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

type state struct {
	a *izapple2.Apple2

	showPages bool
}

func main() {
	var s state
	s.a = izapple2.MainApple()
	if s.a != nil {
		if s.a.IsProfiling() {
			// See the log with:
			//    go tool pprof --pdf ~/go/bin/izapple2sdl /tmp/profile329536248/cpu.pprof > profile.pdf
			defer profile.Start().Stop()
		}

		fyneRun(&s)
	}
}

func fyneRun(s *state) {
	app := app.New()
	// app.SetIcon(xxx)
	window := app.NewWindow("iz-" + s.a.Name)
	// window.SetIcon(xxx)

	bottom := widget.NewToolbar(
		widget.NewToolbarAction(
			theme.NewThemedResource(resourceRestartSvg, nil), func() {
				s.a.SendCommand(izapple2.CommandReset)
			}),
		widget.NewToolbarAction(
			theme.NewThemedResource(resourcePauseSvg, nil), func() {
				s.a.SendCommand(izapple2.CommandPauseUnpauseEmulator)
			}),
		widget.NewToolbarAction(
			theme.NewThemedResource(resourceFastForwardSvg, nil), func() {
				s.a.SendCommand(izapple2.CommandToggleSpeed)
			}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(
			theme.NewThemedResource(resourcePaletteSvg, nil), func() {
				s.a.SendCommand(izapple2.CommandToggleColor)
			}),
		widget.NewToolbarAction(
			theme.NewThemedResource(resourceLayersTripleSvg, nil), func() {
				s.showPages = !s.showPages
				if !s.showPages {
					window.SetTitle("iz-" + s.a.Name)
				}
			}),
		widget.NewToolbarAction(
			theme.NewThemedResource(resourceCameraSvg, nil), func() {
				err := izapple2.SaveSnapshot(s.a, "snapshot.png")
				if err != nil {
					app.SendNotification(fyne.NewNotification(window.Title(),
						fmt.Sprintf("Error saving snapshoot: %v.\n.", err)))
				} else {
					app.SendNotification(fyne.NewNotification(window.Title(),
						"Saving snapshot on 'snapshot.png'"))
				}
			}),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.ViewFullScreenIcon(), func() {
			window.SetFullScreen(!window.FullScreen())
		}),
	)

	screen := canvas.NewImageFromImage(nil)
	screen.SetMinSize(fyne.NewSize(380, 192))
	container := fyne.NewContainerWithLayout(
		layout.NewBorderLayout(nil, bottom, nil, nil),
		screen, bottom,
	)
	window.SetContent(container)
	window.SetPadded(false)

	registerKeyboardEvents(s, window.Canvas())

	go s.a.Run()

	ticker := time.NewTicker(60 * time.Millisecond)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if !s.a.IsPaused() {
					var img *image.RGBA
					if s.showPages {
						img = s.a.SnapshotParts()
						window.SetTitle(fmt.Sprintf("%v %v %vx%v", s.a.Name, s.a.VideoModeName(), img.Rect.Dx()/2, img.Rect.Dy()/2))
					} else {
						img = s.a.Snapshot()
					}
					screen.Image = img
					canvas.Refresh(screen)
				}
			}
		}
	}()

	window.SetOnClosed(func() {
		done <- true
	})

	window.Show()
	app.Run()

}

func registerKeyboardEvents(s *state, canvas fyne.Canvas) {
	kp := newKeyboard(s)

	// Koyboard events
	canvas.SetOnTypedKey(func(ke *fyne.KeyEvent) {
		//fmt.Printf("Event: %v\n", ke.Name)
		kp.putKey(ke)
	})
	canvas.SetOnTypedRune(func(ch rune) {
		//fmt.Printf("Rune: %v\n", ch)
		kp.putRune(ch)
	})
	if deskCanvas, ok := canvas.(desktop.Canvas); ok {
		deskCanvas.SetOnKeyDown(func(ke *fyne.KeyEvent) {
			kp.putKeyAction(ke, true)
			//fmt.Printf("Event down: %v\n", ke.Name)
		})
		deskCanvas.SetOnKeyUp(func(ke *fyne.KeyEvent) {
			kp.putKeyAction(ke, false)
			//fmt.Printf("Event up: %v\n", ke.Name)
		})
	}
}
