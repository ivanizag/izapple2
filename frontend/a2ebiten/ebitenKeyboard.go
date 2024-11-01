package main

import (
	"fmt"

	"github.com/ivanizag/izapple2"
	"github.com/ivanizag/izapple2/screen"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type ebitenKeyboard struct {
	a          *izapple2.Apple2
	keyChannel *izapple2.KeyboardChannel

	showHelp    bool
	showPages   bool
	showCharGen bool
	showAltText bool
	showFreq    bool
	screenMode  int

	debug bool
}

func newEbitenKeyBoard(a *izapple2.Apple2) *ebitenKeyboard {
	var k ebitenKeyboard
	k.a = a
	k.keyChannel = izapple2.NewKeyboardChannel(a)

	k.screenMode = screen.ScreenModeNTSC
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
	k.keyChannel.PutText(text)
}

func (k *ebitenKeyboard) putKey(key ebiten.Key) {
	/*
		See "Apple II reference manual", page 5

		To get keys as understood by the Apple2 hardware run:
		10 A=PEEK(49152)
		20 PRINT A, A - 128
		30 GOTO 10
	*/

	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl)
	shift := ebiten.IsKeyPressed(ebiten.KeyShift)

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
	case ebiten.KeyF1:
		k.showHelp = !k.showHelp
	case ebiten.KeyF2:
		if ctrl {
			k.a.SendCommand(izapple2.CommandReset)
		}
	case ebiten.KeyF4:
		k.a.SendCommand(izapple2.CommandToggleCPUTrace)
	case ebiten.KeyF5:
		if ctrl {
			k.showFreq = !k.showFreq
		} else {
			k.a.SendCommand(izapple2.CommandToggleSpeed)
		}
	case ebiten.KeyF6:
		k.screenMode = screen.NextScreenMode(k.screenMode)
	case ebiten.KeyF7:
		k.showPages = !k.showPages
	case ebiten.KeyF9:
		k.a.SendCommand(izapple2.CommandDumpDebugInfo)
	case ebiten.KeyF10:
		if ctrl {
			k.showCharGen = !k.showCharGen
		} else if shift {
			k.showAltText = !k.showAltText
		} else {
			k.a.SendCommand(izapple2.CommandNextCharGenPage)
		}
	case ebiten.KeyF12:
		fallthrough
	case ebiten.KeyPrintScreen:
		if ctrl {
			screen.AddScenario(k.a.GetVideoSource(), "../../screen/test_resources/")
		} else {
			err := screen.SaveSnapshot(k.a.GetVideoSource(), screen.ScreenModeNTSC, "snapshot.png")
			if err != nil {
				fmt.Printf("Error saving snapshoot: %v.\n.", err)
			} else {
				fmt.Println("Saving snapshot 'snapshot.png'")
			}
		}
	case ebiten.KeyPause:
		k.a.SendCommand(izapple2.CommandPauseUnpause)
	}

	// Missing values 91 to 95. Usually control for [\]^_
	// On the Base64A it's control for \]./

	if result != 0 {
		k.keyChannel.PutChar(result)
	}
}
