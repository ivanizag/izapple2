package main

import (
	"fmt"
	"image"

	"github.com/ivanizag/izapple2"
	a_screen "github.com/ivanizag/izapple2/screen"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pkg/profile"
)

type Game struct {
	a        *izapple2.Apple2
	image    *ebiten.Image
	keyboard *ebitenKeyboard
	speaker  *ebitenSpeaker

	paused bool
	title  string
}

const (
	virtualWidth  = 1128
	virtualHeight = 768
)

func (g *Game) Update() error {
	g.keyboard.update()
	g.speaker.update()

	if g.paused != g.a.IsPaused() {
		if g.a.IsPaused() {
			ebiten.SetWindowTitle(g.title + " - PAUSED")
		} else {
			ebiten.SetWindowTitle(g.title)
		}
		g.paused = g.a.IsPaused()
	}

	if !g.a.IsPaused() {
		var img *image.RGBA
		vs := g.a.GetVideoSource()
		if g.keyboard.showHelp {
			img = a_screen.SnapshotMessageGenerator(vs, helpMessage)
		} else if g.keyboard.showCharGen {
			cgPage, cgPages := g.a.GetCgPageInfo()
			img = a_screen.SnapshotCharacterGenerator(vs, g.keyboard.showAltText)
			ebiten.SetWindowTitle(fmt.Sprintf("%v character map, page %v/%v", g.a.Name, cgPage+1, cgPages))
		} else if g.keyboard.showPages {
			img = a_screen.SnapshotParts(vs, g.keyboard.screenMode)
			ebiten.SetWindowTitle(fmt.Sprintf("%v %v %vx%v", g.a.Name, a_screen.VideoModeName(vs), img.Rect.Dx()/2, img.Rect.Dy()/2))
		} else {
			img = a_screen.Snapshot(vs, g.keyboard.screenMode)
		}
		if img != nil {
			g.image = ebiten.NewImageFromImage(img)
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.image != nil {
		op := &ebiten.DrawImageOptions{}
		size := g.image.Bounds().Size()
		scaleX := virtualWidth / float64(size.X)
		scaleY := virtualHeight / float64(size.Y)
		op.GeoM.Scale(scaleX, scaleY)

		screen.DrawImage(g.image, op)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return virtualWidth, virtualHeight
}

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

		ebitenRun(a)
	}
}

func ebitenRun(a *izapple2.Apple2) {
	ebiten.SetWindowSize(virtualWidth/2, virtualHeight/2)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	title := "iz-" + a.Name + " (F1 for help)"
	ebiten.SetWindowTitle(title)

	go a.Run()

	game := &Game{
		a:        a,
		keyboard: newEbitenKeyBoard(a),
		speaker:  newEbitenSpeaker(),
	}

	a.SetSpeakerProvider(game.speaker)

	if err := ebiten.RunGame(game); err != nil {
		fmt.Printf("Error: %v\n", err)
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
