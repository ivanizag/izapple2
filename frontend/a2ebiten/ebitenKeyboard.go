package main

import (
	"fmt"

	"github.com/ivanizag/izapple2"
	"github.com/ivanizag/izapple2/frontend/shared"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// ebitenKeyboard turns the key events of Ebitengine into the keys the
// frontends share
type ebitenKeyboard struct {
	keyboard *shared.Keyboard

	showFreq bool // The speed on the HUD, only this frontend has one

	debug bool
}

func newEbitenKeyBoard(a *izapple2.Apple2, view *shared.View) *ebitenKeyboard {
	var k ebitenKeyboard
	k.keyboard = shared.NewKeyboard(a, view)
	return &k
}

func (k *ebitenKeyboard) update() {
	runes := ebiten.AppendInputChars(nil)
	if len(runes) > 0 {
		if k.debug {
			fmt.Println("Runes: ", string(runes))
		}
		k.putText(string(runes))
	}

	keys := inpututil.AppendJustPressedKeys(nil)
	for _, key := range keys {
		if k.debug {
			s := key.String()
			fmt.Println("Key pressed: ", s)
		}
		k.putKey(key)
	}
}

func (k *ebitenKeyboard) putText(text string) {
	k.keyboard.PutText(text)
}

func (k *ebitenKeyboard) putKey(key ebiten.Key) {
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl)
	shift := ebiten.IsKeyPressed(ebiten.KeyShift)

	if ctrl && key >= ebiten.KeyA && key <= ebiten.KeyZ {
		k.keyboard.PutCtrlLetter('a' + rune(key-ebiten.KeyA))
		return
	}

	if ctrl && key == ebiten.KeyF5 {
		// Only this frontend draws the speed, the others report it
		k.showFreq = !k.showFreq
		return
	}

	k.keyboard.PutKey(ebitenKey(key), ctrl, shift)
}

// ebitenKey translates a key of Ebitengine
func ebitenKey(key ebiten.Key) shared.Key {
	switch key {
	case ebiten.KeyEscape:
		return shared.KeyEscape
	case ebiten.KeyBackspace:
		return shared.KeyBackspace
	case ebiten.KeyEnter, ebiten.KeyNumpadEnter:
		return shared.KeyReturn
	case ebiten.KeyTab:
		return shared.KeyTab
	case ebiten.KeyDelete:
		return shared.KeyDelete
	case ebiten.KeyLeft:
		return shared.KeyLeft
	case ebiten.KeyRight:
		return shared.KeyRight
	case ebiten.KeyUp:
		return shared.KeyUp
	case ebiten.KeyDown:
		return shared.KeyDown

	case ebiten.KeyF1:
		return shared.KeyF1
	case ebiten.KeyF2:
		return shared.KeyF2
	case ebiten.KeyF3:
		return shared.KeyF3
	case ebiten.KeyF4:
		return shared.KeyF4
	case ebiten.KeyF5:
		return shared.KeyF5
	case ebiten.KeyF6:
		return shared.KeyF6
	case ebiten.KeyF7:
		return shared.KeyF7
	case ebiten.KeyF8:
		return shared.KeyF8
	case ebiten.KeyF9:
		return shared.KeyF9
	case ebiten.KeyF10:
		return shared.KeyF10
	case ebiten.KeyF12:
		return shared.KeyF12
	case ebiten.KeyPause:
		return shared.KeyPause
	case ebiten.KeyPrintScreen:
		return shared.KeyPrintScreen
	}

	return shared.KeyNone
}
