package main

import (
	"bytes"
	"fmt"
	"image/color"

	"github.com/ivanizag/izapple2"
	"github.com/ivanizag/izapple2/frontend/shared"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/pkg/profile"
)

type Game struct {
	a           *izapple2.Apple2
	image       *ebiten.Image
	view        *shared.View
	keyboard    *ebitenKeyboard
	speaker     *ebitenAudio
	mouse       *shared.Mouse
	dropTargets *ebitenDropTargets
	fontSource  *text.GoTextFaceSource

	paused bool
	title  string

	updates uint64
	freq    float64
}

const (
	virtualWidth  = 1128
	virtualHeight = 768
	hudFontSize   = 50
)

var hudColor = color.RGBA{208, 241, 141, 255} // Yellow

func (g *Game) Update() error {
	g.keyboard.update()
	g.speaker.update()
	g.updateMouse()
	g.dropTargets.update()

	if g.paused != g.a.IsPaused() {
		if g.a.IsPaused() {
			ebiten.SetWindowTitle(g.title + " - PAUSED")
		} else {
			ebiten.SetWindowTitle(g.title)
		}
		g.paused = g.a.IsPaused()
	}

	if g.updates%3 == 0 && !g.a.IsPaused() { // 20 times per second
		img, viewTitle := g.view.Snapshot(g.a, g.dropTargets.targets,
			g.dropTargets.pointedDrive())
		if viewTitle != "" {
			ebiten.SetWindowTitle(viewTitle)
		}
		if img != nil {
			g.image = ebiten.NewImageFromImage(img)
		}
	}

	if g.updates%60 == 0 { // Once per second
		g.freq = g.a.GetCurrentFreqMHz()
	}

	g.updates++
	return nil
}

/*
updateMouse reads where the pointer is and whether its button is pressed.
Ebitengine has no mouse events, they are polled when the frame is updated.

The positions are on the virtual screen the game is laid out on, not on the
window, so the pointer lands on the same place of the picture whatever the size
of the window is.
*/
func (g *Game) updateMouse() {
	x, y := ebiten.CursorPosition()
	g.mouse.SetPosition(x, y, virtualWidth, virtualHeight)
	g.mouse.SetButton(ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft))
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

	if g.keyboard.showFreq {
		msg := fmt.Sprintf("%0.2f Hz, FPS %0.0f", g.freq, ebiten.ActualFPS())
		op := &text.DrawOptions{}
		op.GeoM.Translate(20, 20)
		op.ColorScale.ScaleWithColor(hudColor)
		text.Draw(screen, msg, &text.GoTextFace{
			Source: g.fontSource,
			Size:   hudFontSize,
		}, op)

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

	view := shared.NewView()
	game := &Game{
		a:           a,
		view:        view,
		keyboard:    newEbitenKeyBoard(a, view),
		speaker:     newEbitenAudio(a),
		mouse:       shared.NewMouse(),
		dropTargets: newEbitenDropTargets(a),
	}
	a.SetMouseProvider(game.mouse)

	var err error
	game.fontSource, err = text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		panic(err)
	}

	go a.Run()
	if err := ebiten.RunGame(game); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

/*
To test the WebAssembly version, run:
	go run github.com/hajimehoshi/wasmserve@latest .
*/
