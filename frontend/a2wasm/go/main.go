package main

import (
	"bytes"
	"fmt"

	"github.com/ivanizag/izapple2"
	a_screen "github.com/ivanizag/izapple2/screen"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type Game struct {
	a          *izapple2.Apple2
	image      *ebiten.Image
	speaker    *wasmSpeaker
	fontSource *text.GoTextFaceSource

	updates     uint64
	screenMode  int
	keyChannel  *izapple2.KeyboardChannel
}

const (
	virtualWidth  = 1128
	virtualHeight = 768
)

func (g *Game) Update() error {
	// Handle keyboard input from Ebiten (AppendInputChars)
	runes := ebiten.AppendInputChars(nil)
	if len(runes) > 0 {
		text := string(runes)
		fmt.Printf("Ebiten keyboard: '%s'\n", text)
		g.keyChannel.PutText(text)
	}

	g.speaker.update()

	if g.updates%3 == 0 && !g.a.IsPaused() { // 20 times per second
		vs := g.a.GetVideoSource()
		img := a_screen.Snapshot(vs, g.screenMode)
		if img != nil {
			g.image = ebiten.NewImageFromImage(img)
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

var globalGame *Game

func main() {
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
	// Note: Window functions don't apply in WASM - canvas size is controlled by CSS
	// ebiten.SetWindowSize(virtualWidth/2, virtualHeight/2)
	// ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	// ebiten.SetWindowTitle("iz-" + a.Name)

	fmt.Printf("ebitenRun: Apple2 instance at %p\n", a)

	game := &Game{
		a:          a,
		speaker:    newWasmSpeaker(),
		keyChannel: izapple2.NewKeyboardChannel(a), // This already sets itself as the keyboard provider
		screenMode: a_screen.ScreenModeNTSC,
	}

	fmt.Printf("ebitenRun: keyChannel at %p\n", game.keyChannel)

	// Set up providers
	a.SetSpeakerProvider(game.speaker)
	// Don't set keyboard provider - NewKeyboardChannel already did it

	var err error
	game.fontSource, err = text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		panic(err)
	}

	globalGame = game

	// Setup API exports for React
	setupAPI(a, game)

	go a.Run()
	if err := ebiten.RunGame(game); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
