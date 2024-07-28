package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/ivanizag/izapple2"
	"github.com/ivanizag/izapple2/screen"
)

type keyboard struct {
	s          *state
	keyChannel *izapple2.KeyboardChannel

	controlLeft  bool
	controlRight bool
}

func newKeyboard(s *state) *keyboard {
	var k keyboard
	k.s = s
	k.keyChannel = izapple2.NewKeyboardChannel(s.a)
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

	if press && ctrl {
		// F keys with ctrl do not generate events in putKey()
		switch keyEvent.Name {
		case fyne.KeyF1:
			k.s.a.SendCommand(izapple2.CommandReset)
		case fyne.KeyF12:
			screen.AddScenario(k.s.a.GetVideoSource(), "../../screen/test_resources/")
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

	// Keys with control are not generating events in putKey()
	//ctrl := k.controlLeft || k.controlRight

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
	case fyne.KeyTab: // The Tab is not reaching here
		result = 9
	case fyne.KeyDelete:
		result = 127 // 24 in the Base64A

	// Base64A clone particularities
	case fyne.KeyF2:
		result = 127 // Base64A

	// Control of the emulator
	case fyne.KeyF1:
		/*if ctrl {
			k.s.a.SendCommand(izapple2.CommandReset)
		}*/
	case fyne.KeyF5:
		k.s.a.SendCommand(izapple2.CommandShowSpeed)
	case fyne.KeyF6:
		if k.s.screenMode != screen.ScreenModeGreen {
			k.s.screenMode = screen.ScreenModeGreen
		} else {
			k.s.screenMode = screen.ScreenModeNTSC
		}
	case fyne.KeyF7:
		k.s.showPages = !k.s.showPages
	case fyne.KeyF9:
		k.s.a.SendCommand(izapple2.CommandDumpDebugInfo)
	case fyne.KeyF10:
		k.s.a.SendCommand(izapple2.CommandNextCharGenPage)
	case fyne.KeyF11:
		k.s.a.SendCommand(izapple2.CommandToggleCPUTrace)
	case fyne.KeyF12:
		//case fyne.KeyPrintScreen:
		err := screen.SaveSnapshot(k.s.a.GetVideoSource(), k.s.screenMode, "snapshot.png")
		if err != nil {
			fmt.Printf("Error saving snapshoot: %v.\n.", err)
		} else {
			fmt.Println("Saving snapshot")
		}
		//case fyne.KeyPause:
		//	k.s.a.SendCommand(izapple2.CommandPauseUnpause)
	}

	if result != 0 {
		k.keyChannel.PutChar(result)
	}
}
