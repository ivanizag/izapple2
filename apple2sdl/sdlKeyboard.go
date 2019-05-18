package apple2sdl

import (
	"go6502/apple2"
	"unicode/utf8"

	"github.com/veandco/go-sdl2/sdl"
)

type sdlKeyboard struct {
	keyChannel chan uint8
	a          *apple2.Apple2
}

func newSDLKeyBoard(a *apple2.Apple2) *sdlKeyboard {
	var k sdlKeyboard
	k.keyChannel = make(chan uint8, 100)
	k.a = a
	return &k
}

func (k *sdlKeyboard) putText(textEvent *sdl.TextInputEvent) {
	text := textEvent.GetText()

	for _, ch := range text {
		// We will use computed text only for printable ASCII chars
		if ch < ' ' || ch > '~' {
			continue
		}

		buf := make([]uint8, 1)
		utf8.EncodeRune(buf, ch)

		k.putChar(buf[0])
	}
}

func (k *sdlKeyboard) putKey(keyEvent *sdl.KeyboardEvent) {
	/*
		See "Apple II reference manual", page 5

		To get keys as understood by the Apple2 hardware run:
		10 A=PEEK(49152)
		20 PRINT A, A - 128
		30 GOTO 10

		Missing Reset button
	*/
	if keyEvent.Type != sdl.KEYDOWN {
		// Process only key pushes
		return
	}

	key := keyEvent.Keysym
	ctrl := key.Mod&sdl.KMOD_CTRL != 0

	if ctrl {
		if key.Sym >= 'a' && key.Sym <= 'z' {
			k.putChar(uint8(key.Sym) - 97 + 1)
			return
		}
	}

	result := uint8(0)

	switch key.Sym {
	case sdl.K_ESCAPE:
		result = 27
	case sdl.K_BACKSPACE:
		result = 24
	case sdl.K_RETURN:
		result = 13
	case sdl.K_RETURN2:
		result = 13
	case sdl.K_LEFT:
		if ctrl {
			result = 31 // Base64A
		}
		result = 8
	case sdl.K_RIGHT:
		result = 21

	// Base64A clone particularities
	case sdl.K_F2:
		result = 127
	case sdl.K_UP:
		result = 31
	case sdl.K_DOWN:
		result = 10

	// Control of the emulator
	case sdl.K_F5:
		k.a.SendCommand(apple2.CommandToggleSpeed)
	case sdl.K_F6:
		k.a.SendCommand(apple2.CommandToggleColor)
	case sdl.K_F7:
		k.a.SendCommand(apple2.CommandSaveState)
	case sdl.K_F8:
		k.a.SendCommand(apple2.CommandLoadState)
	case sdl.K_F9:
		k.a.SendCommand(apple2.CommandDumpDebugInfo)
	}

	// Missing values 91 to 95. Usually control for [\]^_
	// On the Base64A it's control for \]./

	if result != 0 {
		k.putChar(result)
	}
}

func (k *sdlKeyboard) putChar(ch uint8) {
	k.keyChannel <- ch
}

func (k *sdlKeyboard) GetKey(_ bool) (key uint8, ok bool) {
	select {
	case key = <-k.keyChannel:
		ok = true
	default:
		ok = false
	}
	return
}
