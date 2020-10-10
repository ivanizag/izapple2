package main

import (
	"fmt"

	"fyne.io/fyne"
	"fyne.io/fyne/driver/desktop"
	"github.com/ivanizag/izapple2"
)

type keyboard struct {
	a          *izapple2.Apple2
	keyChannel *izapple2.KeyboardChannel

	controlLeft  bool
	controlRight bool
	showPages    bool
}

func newKeyboard(a *izapple2.Apple2) *keyboard {
	var k keyboard
	k.a = a
	k.keyChannel = izapple2.NewKeyboardChannel(a)
	return &k
}

func (k *keyboard) putRune(ch rune) {
	k.keyChannel.PutRune(ch)
}

// PutChar sends a character to the emulator
func (k *keyboard) PutChar(ch uint8) {
	k.keyChannel.PutChar(ch)
}

func (k *keyboard) putKeyAction(keyEvent *fyne.KeyEvent, press bool) {

	ctrl := k.controlLeft || k.controlRight
	if press && ctrl && len(keyEvent.Name) == 1 {
		// Hacky. We relay on the letter values to be a single uppercase char.
		ch := keyEvent.Name[0]
		if ch >= 'A' && ch <= 'Z' {
			k.keyChannel.PutChar(uint8(ch) - 65 + 1)
			return
		}
	}

	switch keyEvent.Name {
	case desktop.KeyControlLeft:
		k.controlLeft = press
	case desktop.KeyControlRight:
		k.controlRight = press
	}
}

func (k *keyboard) putKey(keyEvent *fyne.KeyEvent) {
	/*
		See "Apple II reference manual", page 5

		To get keys as understood by the Apple2 hardware run:
		10 A=PEEK(49152)
		20 PRINT A, A - 128
		30 GOTO 10
	*/
	ctrl := k.controlLeft || k.controlRight
	result := uint8(0)
	switch keyEvent.Name {
	case fyne.KeyEscape:
		result = 27
	case fyne.KeyBackspace:
		result = 8
	case fyne.KeyReturn:
		result = 13
	case fyne.KeyEnter:
		result = 13
	case fyne.KeyLeft:
		result = 8
	case fyne.KeyRight:
		result = 21

	// Apple //e
	case fyne.KeyUp:
		result = 11 // 31 in the Base64A
	case fyne.KeyDown:
		result = 10
	case fyne.KeyTab:
		result = 9
	case fyne.KeyDelete:
		result = 127 // 24 in the Base64A

	// Base64A clone particularities
	case fyne.KeyF2:
		result = 127 // Base64A

	// Control of the emulator
	case fyne.KeyF1:
		if ctrl {
			k.a.SendCommand(izapple2.CommandReset)
		}
	case fyne.KeyF5:
		if ctrl {
			k.a.SendCommand(izapple2.CommandShowSpeed)
		} else {
			k.a.SendCommand(izapple2.CommandToggleSpeed)
		}
	case fyne.KeyF6:
		k.a.SendCommand(izapple2.CommandToggleColor)
	case fyne.KeyF7:
		k.showPages = !k.showPages
	case fyne.KeyF9:
		k.a.SendCommand(izapple2.CommandDumpDebugInfo)
	case fyne.KeyF10:
		k.a.SendCommand(izapple2.CommandNextCharGenPage)
	case fyne.KeyF11:
		k.a.SendCommand(izapple2.CommandToggleCPUTrace)
	case fyne.KeyF12:
	case fyne.KeyPrintScreen:
		err := izapple2.SaveSnapshot(k.a, "snapshot.png")
		if err != nil {
			fmt.Printf("Error saving snapshoot: %v.\n.", err)
		} else {
			fmt.Println("Saving snapshot")
		}
	case fyne.KeyPause:
		k.a.SendCommand(izapple2.CommandPauseUnpauseEmulator)
	}

	// Missing values 91 to 95. Usually control for [\]^_
	// On the Base64A it's control for \]./

	if result != 0 {
		k.keyChannel.PutChar(result)
	}
}
