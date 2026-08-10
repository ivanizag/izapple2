//go:build js

package main

import (
	"fmt"

	"github.com/ivanizag/izapple2"
	"github.com/ivanizag/izapple2/frontend/shared"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type wasmKeyboard struct {
	a          *izapple2.Apple2
	keyChannel *izapple2.KeyboardChannel

	debug bool
}

func newWasmKeyBoard(a *izapple2.Apple2) *wasmKeyboard {
	var k wasmKeyboard
	k.a = a
	k.keyChannel = izapple2.NewKeyboardChannel(a)
	return &k
}

func (k *wasmKeyboard) update() {
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

func (k *wasmKeyboard) putText(text string) {
	k.keyChannel.PutText(text)
}

func (k *wasmKeyboard) putChar(ch uint8) {
	k.keyChannel.PutChar(ch)
}

func (k *wasmKeyboard) putKey(key ebiten.Key) {
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl)

	if ctrl && key >= ebiten.KeyA && key <= ebiten.KeyZ {
		if char, ok := shared.CharForCtrlLetter('a' + rune(key-ebiten.KeyA)); ok {
			fmt.Println("Control Key: ", key.String())
			k.keyChannel.PutChar(char)
		}
		return
	}

	if char, ok := shared.CharForKey(wasmKey(key), ctrl); ok {
		k.keyChannel.PutChar(char)
		return
	}

	// Control of the emulator. The page has its own controls for the rest.
	switch key {
	case ebiten.KeyF2:
		if ctrl {
			k.a.SendCommand(izapple2.CommandReset)
		}
	case ebiten.KeyF9:
		k.a.SendCommand(izapple2.CommandDumpDebugInfo)
	case ebiten.KeyF12, ebiten.KeyPause:
		k.a.SendCommand(izapple2.CommandPauseUnpause)
	}
}

// wasmKey translates a key of Ebitengine
func wasmKey(key ebiten.Key) shared.Key {
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
	case ebiten.KeyF3:
		return shared.KeyF3
	}

	return shared.KeyNone
}
