package main

import (
	"fmt"

	"github.com/ivanizag/izapple2"
	"github.com/ivanizag/izapple2/screen"
	"github.com/veandco/go-sdl2/sdl"
)

type sdlKeyboard struct {
	a          *izapple2.Apple2
	keyChannel *izapple2.KeyboardChannel

	showHelp    bool
	showPages   bool
	showCharGen bool
	showAltText bool
	screenMode  int
}

func newSDLKeyBoard(a *izapple2.Apple2) *sdlKeyboard {
	var k sdlKeyboard
	k.a = a
	k.keyChannel = izapple2.NewKeyboardChannel(a)

	k.screenMode = screen.ScreenModeNTSC
	return &k
}

func (k *sdlKeyboard) putText(text string) {
	k.keyChannel.PutText(text)
}

func (k *sdlKeyboard) putKey(keyEvent *sdl.KeyboardEvent) {
	/*
		See "Apple II reference manual", page 5

		To get keys as understood by the Apple2 hardware run:
		10 A=PEEK(49152)
		20 PRINT A, A - 128
		30 GOTO 10
	*/

	if keyEvent.Type != sdl.KEYDOWN {
		// Process only key pushes
		return
	}

	key := keyEvent.Keysym
	ctrl := key.Mod&sdl.KMOD_CTRL != 0
	shift := key.Mod&sdl.KMOD_SHIFT != 0

	if ctrl {
		if key.Sym >= 'a' && key.Sym <= 'z' {
			k.keyChannel.PutChar(uint8(key.Sym) - 97 + 1)
			return
		}
	}

	result := uint8(0)

	switch key.Sym {
	case sdl.K_ESCAPE:
		result = 27
	case sdl.K_BACKSPACE:
		result = 8
	case sdl.K_RETURN:
		result = 13
	case sdl.K_RETURN2:
		result = 13
	case sdl.K_LEFT:
		if ctrl {
			result = 31 // Base64A
		} else {
			result = 8
		}
	case sdl.K_RIGHT:
		result = 21

	// Apple //e
	case sdl.K_UP:
		result = 11 // 31 in the Base64A
	case sdl.K_DOWN:
		result = 10
	case sdl.K_TAB:
		result = 9
	case sdl.K_DELETE:
		result = 127 // 24 in the Base64A

	// Base64A clone particularities
	case sdl.K_F3:
		result = 127 // Base64A

	// Control of the emulator
	case sdl.K_F1:
		k.showHelp = !k.showHelp
	case sdl.K_F2:
		if ctrl {
			k.a.SendCommand(izapple2.CommandReset)
		}
	case sdl.K_F4:
		k.a.SendCommand(izapple2.CommandToggleCPUTrace)
	case sdl.K_F5:
		if ctrl {
			k.a.SendCommand(izapple2.CommandShowSpeed)
		} else {
			k.a.SendCommand(izapple2.CommandToggleSpeed)
		}
	case sdl.K_F6:
		k.screenMode = screen.NextScreenMode(k.screenMode)
	case sdl.K_F7:
		k.showPages = !k.showPages
	case sdl.K_F9:
		k.a.SendCommand(izapple2.CommandDumpDebugInfo)
	case sdl.K_F10:
		if ctrl {
			k.showCharGen = !k.showCharGen
		} else if shift {
			k.showAltText = !k.showAltText
		} else {
			k.a.SendCommand(izapple2.CommandNextCharGenPage)
		}
	case sdl.K_F12:
		fallthrough
	case sdl.K_PRINTSCREEN:
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
	case sdl.K_PAUSE:
		k.a.SendCommand(izapple2.CommandPauseUnpause)
	}

	// Missing values 91 to 95. Usually control for [\]^_
	// On the Base64A it's control for \]./

	if result != 0 {
		k.keyChannel.PutChar(result)
	}
}
