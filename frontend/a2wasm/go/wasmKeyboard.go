package main

import (
	"fmt"

	"github.com/ivanizag/izapple2"

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
	/*
		See "Apple II reference manual", page 5

		To get keys as understood by the Apple2 hardware run:
		10 A=PEEK(49152)
		20 PRINT A, A - 128
		30 GOTO 10
	*/

	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl)
	//#shift := ebiten.IsKeyPressed(ebiten.KeyShift)

	if ctrl {
		if key >= ebiten.KeyA && key <= ebiten.KeyZ {
			fmt.Println("Control Key: ", key.String())
			k.keyChannel.PutChar(uint8(key-ebiten.KeyA) - 97 + 1)
			return
		}
	}

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
		if ctrl {
			result = 31 // Base64A
		} else {
			result = 8
		}
	case ebiten.KeyRight:
		result = 21

	// Apple //e
	case ebiten.KeyUp:
		result = 11 // 31 in the Base64A
	case ebiten.KeyDown:
		result = 10
	case ebiten.KeyTab:
		result = 9
	case ebiten.KeyDelete:
		result = 127 // 24 in the Base64A

	// Base64A clone particularities
	case ebiten.KeyF3:
		result = 127 // Base64A

	// Control of the emulator
	case ebiten.KeyF2:
		if ctrl {
			k.a.SendCommand(izapple2.CommandReset)
		}
	case ebiten.KeyF9:
		k.a.SendCommand(izapple2.CommandDumpDebugInfo)
	case ebiten.KeyF12:
		fallthrough
	case ebiten.KeyPause:
		k.a.SendCommand(izapple2.CommandPauseUnpause)
	}

	// Missing values 91 to 95. Usually control for [\]^_
	// On the Base64A it's control for \]./

	if result != 0 {
		k.keyChannel.PutChar(result)
	}
}
