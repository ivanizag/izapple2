//go:build !js

package main

import (
	"github.com/ivanizag/izapple2"
	"github.com/ivanizag/izapple2/frontend/shared"
	"github.com/veandco/go-sdl2/sdl"
)

// sdlKeyboard turns the key events of SDL2 into the keys the frontends share
type sdlKeyboard struct {
	keyboard *shared.Keyboard
}

func newSDLKeyBoard(a *izapple2.Apple2, view *shared.View) *sdlKeyboard {
	var k sdlKeyboard
	k.keyboard = shared.NewKeyboard(a, view)
	return &k
}

func (k *sdlKeyboard) putText(text string) {
	k.keyboard.PutText(text)
}

func (k *sdlKeyboard) putKey(keyEvent *sdl.KeyboardEvent) {
	if keyEvent.Type != sdl.KEYDOWN {
		// Process only key pushes
		return
	}

	key := keyEvent.Keysym
	ctrl := key.Mod&sdl.KMOD_CTRL != 0
	shift := key.Mod&sdl.KMOD_SHIFT != 0
	command := key.Mod&sdl.KMOD_GUI != 0

	if ctrl && k.keyboard.PutCtrlLetter(rune(key.Sym)) {
		return
	}

	// Pasting reads the clipboard of SDL, it is not a shared command
	if (key.Sym == sdl.K_INSERT && shift) || (key.Sym == sdl.K_v && command) {
		text, _ := sdl.GetClipboardText()
		k.keyboard.Paste(text)
		return
	}

	k.keyboard.PutKey(sdlKey(key.Sym), ctrl, shift)
}

// sdlKey translates a key of SDL2
func sdlKey(sym sdl.Keycode) shared.Key {
	switch sym {
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
