//go:build js

package main

import (
	"fmt"
	"image"

	"github.com/ivanizag/izapple2"
	a_screen "github.com/ivanizag/izapple2/screen"

	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	a        *izapple2.Apple2
	image    *ebiten.Image
	keyboard *wasmKeyboard
	speaker  *wasmAudio

	updates    uint64
	screenMode int

	// Reusable buffer for screen snapshots to avoid allocations
	lastSnapshotBounds image.Rectangle
}

const (
	virtualWidth  = 1128
	virtualHeight = 768
)

func (g *Game) Update() error {
	g.keyboard.update()
	g.speaker.update()

	if g.updates%3 == 0 && !g.a.IsPaused() { // 20 times per second
		vs := g.a.GetVideoSource()

		img := a_screen.Snapshot(vs, g.screenMode)

		if img != nil {
			// Check if the image dimensions changed (shouldn't happen often)
			if g.lastSnapshotBounds == (image.Rectangle{}) ||
				g.lastSnapshotBounds.Dx() != img.Bounds().Dx() ||
				g.lastSnapshotBounds.Dy() != img.Bounds().Dy() {

				// Dimensions changed or first run - create new resources
				g.lastSnapshotBounds = img.Bounds()
				g.image = ebiten.NewImage(img.Bounds().Dx(), img.Bounds().Dy())
				fmt.Printf("Screen size changed to %dx%d\n", img.Bounds().Dx(), img.Bounds().Dy())
			} else {
				// Reuse existing Ebiten image - just update pixels
				// This avoids creating a new WebGL texture every frame
				g.image.WritePixels(img.Pix)
			}

		}
	}

	g.updates++
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

var appInitialized bool
var gameInitialized bool

func main() {
	if appInitialized {
		fmt.Println("main() called again - ignoring to prevent duplicate Game instances")
		return
	}
	appInitialized = true

	a, err := izapple2.CreateConfiguredApple()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	if a != nil {
		ebitenRun(a)
	}
}

func ebitenRun(a *izapple2.Apple2) {
	if gameInitialized {
		fmt.Println("ebitenRun() called again - ignoring to prevent duplicate Game instances")
		return
	}
	gameInitialized = true

	game := &Game{
		a:          a,
		speaker:    newWasmAudio(a.GetClockMhz()),
		keyboard:   newWasmKeyBoard(a),
		screenMode: a_screen.ScreenModeNTSC,
	}

	// Set up providers
	for _, source := range a.GetAudioSources() {
		source.SetAudioSink(game.speaker.mixer.NewSource())
	}

	// Setup API exports for React
	setupAPI(a, game)

	go a.Run()
	if err := ebiten.RunGame(game); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
