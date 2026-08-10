//go:build !js

package main

import (
	"github.com/ivanizag/izapple2"
	"github.com/ivanizag/izapple2/frontend/shared"

	"github.com/Zyko0/go-sdl3/sdl"
)

// sdl3Keyboard turns the key events of SDL3 into the keys the frontends share
type sdl3Keyboard struct {
	keyboard *shared.Keyboard
}

func newSDL3Keyboard(a *izapple2.Apple2, view *shared.View) *sdl3Keyboard {
	var k sdl3Keyboard
	k.keyboard = shared.NewKeyboard(a, view)
	return &k
}

func (k *sdl3Keyboard) putText(text string) {
	k.keyboard.PutText(text)
}

func (k *sdl3Keyboard) putKey(keyEvent *sdl.KeyboardEvent) {
	if !keyEvent.Down {
		// Process only key pushes
		return
	}

	ctrl := keyEvent.Mod&sdl.KMOD_CTRL != 0
	shift := keyEvent.Mod&sdl.KMOD_SHIFT != 0
	command := keyEvent.Mod&sdl.KMOD_GUI != 0

	if ctrl && k.keyboard.PutCtrlLetter(rune(keyEvent.Key)) {
		return
	}

	// Pasting reads the clipboard of SDL, it is not a shared command
	if (keyEvent.Key == sdl.K_INSERT && shift) || (keyEvent.Key == 'v' && command) {
		text, _ := sdl.GetClipboardText()
		k.keyboard.Paste(text)
		return
	}

	k.keyboard.PutKey(sdl3Key(keyEvent.Key), ctrl, shift)
}

// sdl3Key translates a key of SDL3
func sdl3Key(key sdl.Keycode) shared.Key {
	switch key {
	case sdl.K_ESCAPE:
		return shared.KeyEscape
	case sdl.K_BACKSPACE:
		return shared.KeyBackspace
	case sdl.K_RETURN, sdl.K_RETURN2:
		return shared.KeyReturn
	case sdl.K_TAB:
		return shared.KeyTab
	case sdl.K_DELETE:
		return shared.KeyDelete
	case sdl.K_LEFT:
		return shared.KeyLeft
	case sdl.K_RIGHT:
		return shared.KeyRight
	case sdl.K_UP:
		return shared.KeyUp
	case sdl.K_DOWN:
		return shared.KeyDown

	case sdl.K_F1:
		return shared.KeyF1
	case sdl.K_F2:
		return shared.KeyF2
	case sdl.K_F3:
		return shared.KeyF3
	case sdl.K_F4:
		return shared.KeyF4
	case sdl.K_F5:
		return shared.KeyF5
	case sdl.K_F6:
		return shared.KeyF6
	case sdl.K_F7:
		return shared.KeyF7
	case sdl.K_F8:
		return shared.KeyF8
	case sdl.K_F9:
		return shared.KeyF9
	case sdl.K_F10:
		return shared.KeyF10
	case sdl.K_F12:
		return shared.KeyF12
	case sdl.K_PAUSE:
		return shared.KeyPause
	case sdl.K_PRINTSCREEN:
		return shared.KeyPrintScreen
	}

	return shared.KeyNone
}
