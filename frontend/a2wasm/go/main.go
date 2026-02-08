package main

import (
	"bytes"
	"fmt"

	"github.com/ivanizag/izapple2"
	a_screen "github.com/ivanizag/izapple2/screen"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type Game struct {
	a          *izapple2.Apple2
	image      *ebiten.Image
	speaker    *wasmSpeaker
	fontSource *text.GoTextFaceSource

	updates    uint64
	screenMode int
	keyChannel *izapple2.KeyboardChannel
}

const (
	virtualWidth  = 1128
	virtualHeight = 768
)

func (g *Game) Update() error {
	// Handle keyboard input from Ebiten (AppendInputChars for printable chars)
	runes := ebiten.AppendInputChars(nil)
	if len(runes) > 0 {
		g.keyChannel.PutText(string(runes))
	}

	// Handle special keys (Enter, arrows, etc.)
	keys := inpututil.AppendJustPressedKeys(nil)
	for _, key := range keys {
		g.putKey(key)
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

// putKey handles special keys from Ebiten
func (g *Game) putKey(key ebiten.Key) {
	result := uint8(0)

	switch key {
	case ebiten.KeyEscape:
		result = 27
	case ebiten.KeyBackspace:
		result = 8
	case ebiten.KeyEnter:
		result = 13
	case ebiten.KeyNumpadEnter:
		result = 13
	case ebiten.KeyLeft:
		result = 8
	case ebiten.KeyRight:
		result = 21
	case ebiten.KeyUp:
		result = 11
	case ebiten.KeyDown:
		result = 10
	case ebiten.KeyTab:
		result = 9
	case ebiten.KeyDelete:
		result = 127
	}

	if result != 0 {
		g.keyChannel.PutChar(result)
	}
}

var globalGame *Game
var initialized bool

func main() {
	if initialized {
		fmt.Println("main() called again - ignoring to prevent duplicate Game instances")
		return
	}
	initialized = true

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
	if globalGame != nil {
		fmt.Println("ebitenRun() called again - ignoring to prevent duplicate Game instances")
		return
	}

	game := &Game{
		a:          a,
		speaker:    newWasmSpeaker(),
		keyChannel: izapple2.NewKeyboardChannel(a), // This already sets itself as the keyboard provider
		screenMode: a_screen.ScreenModeNTSC,
	}

	// Set up providers
	a.SetSpeakerProvider(game.speaker)

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
