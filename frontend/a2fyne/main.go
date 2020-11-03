package main

import (
	"fmt"
	"image"
	"strings"
	"time"

	"github.com/ivanizag/izapple2"
	"github.com/ivanizag/izapple2/screen"

	"github.com/pkg/profile"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/driver/desktop"
	"fyne.io/fyne/layout"
)

type state struct {
	a   *izapple2.Apple2
	app fyne.App
	win fyne.Window

	devices *panelDevices

	showPages  bool
	screenMode int
}

func (s *state) DefaultTitle() string {
	title := "iz-" + s.a.Name
	if s.a.IsForceCaps() {
		title = strings.ToUpper(title)
	}
	return title
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
	s.screenMode = screen.ScreenModeNTSC

	s.app = app.New()
	s.app.SetIcon(resourceApple2Png)
	s.win = s.app.NewWindow(s.DefaultTitle())
	s.win.SetIcon(resourceApple2Png)

	s.devices = newPanelDevices(s)
	s.devices.w.Hide()
	toolbar := buildToolbar(s)
	display := canvas.NewImageFromImage(nil)
	display.ScaleMode = canvas.ImageScalePixels // Looks worst but loads less.
	display.SetMinSize(fyne.NewSize(280*2*2, 192*2*2))

	container := fyne.NewContainerWithLayout(
		layout.NewBorderLayout(nil, toolbar, nil, s.devices.w),
		display, toolbar, s.devices.w,
	)
	s.win.SetContent(container)
	s.win.SetPadded(false)

	registerKeyboardEvents(s)
	j := newJoysticks(s)
	j.start()
	s.a.SetJoysticksProvider(j)

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
						img = screen.SnapshotParts(s.a, s.screenMode)
						s.win.SetTitle(fmt.Sprintf("%v %v %vx%v", s.a.Name, screen.VideoModeName(s.a), img.Rect.Dx()/2, img.Rect.Dy()/2))
					} else {
						img = screen.Snapshot(s.a, s.screenMode)
					}
					display.Image = img
					canvas.Refresh(display)
				}
			}
		}
	}()

	s.win.SetOnClosed(func() {
		done <- true
	})

	s.win.Show()
	s.app.Run()
}

func registerKeyboardEvents(s *state) {
	kp := newKeyboard(s)
	canvas := s.win.Canvas()

	// Events
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
