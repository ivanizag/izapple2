package main

import (
	"time"

	"github.com/ivanizag/izapple2"

	"github.com/pkg/profile"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
)

func main() {
	a := izapple2.MainApple()
	if a != nil {
		if a.IsProfiling() {
			// See the log with:
			//    go tool pprof --pdf ~/go/bin/izapple2sdl /tmp/profile329536248/cpu.pprof > profile.pdf
			defer profile.Start().Stop()
		}

		FyneRun(a)
	}
}

// FyneRun starts the Apple2 emulator on fyne.io
func FyneRun(a *izapple2.Apple2) {
	app := app.New()
	// app.SetIcon(xxx)
	window := app.NewWindow("iz-" + a.Name)
	// window.SetIcon(xxx)

	screen := canvas.NewImageFromImage(nil)
	top := widget.NewLabel("Top")
	bottom := widget.NewLabel("Bottom")
	right := widget.NewLabel("Right")

	container := fyne.NewContainerWithLayout(
		layout.NewBorderLayout(top, bottom, nil, right),
		screen, top, bottom, right,
	)
	window.SetContent(container)

	go a.Run()
	ticker := time.NewTicker(60 * time.Millisecond)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if !a.IsPaused() {
					img := a.Snapshot()
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
