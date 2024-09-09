package izapple2

import (
	"slices"
	"unicode"
)

// KeyboardProvider provides a keyboard implementation
type KeyboardProvider interface {
	GetKey(strobe bool) (key uint8, ok bool)
}

// KeyboardChannel is a possible implemetation of a Keyboard provider
type KeyboardChannel struct {
	keyChannel chan uint8
	a          *Apple2
}

// NewKeyboardChannel returns an instance of KeyboardChannel
func NewKeyboardChannel(a *Apple2) *KeyboardChannel {
	var k KeyboardChannel
	k.keyChannel = make(chan uint8, 100)
	k.a = a
	a.SetKeyboardProvider(&k)
	return &k
}

// PutText sends texts to the emulator as succesive chars
func (k *KeyboardChannel) PutText(text string) {
	for _, ch := range text {
		k.PutRune(ch)
	}
}

var macOptionChars = []rune("ı˝•£‰ ⁄‘’≈œæ€®†¥øπå∫∂ƒ™¶§∑©√ßµ„…≤≥çñŒÆ€‡∏ﬂ¯ˇ˘‹›◊˙˚")
var macOptionSubs = []rune("!\"·$%&/()=qwertyopasdfghjkxcvbm,.<>cnQWETPGJKLZXVNM")

// PutRune sends a rune to the emulator if it is valid printable ASCII
func (k *KeyboardChannel) PutRune(ch rune) {

	// Some substitutions useful for Macs that transform chars with the option key
	pos := slices.Index(macOptionChars, ch)
	if pos >= 0 {
		ch = rune(macOptionSubs[pos])
	}

	// We will use computed text only for printable ASCII chars
	if ch >= ' ' && ch <= '~' {
		if k.a.IsForceCaps() && ch >= 'a' && ch <= 'z' {
			ch = unicode.ToUpper(ch)
		}
		k.PutChar(uint8(ch))
	}
}

// PutChar sends a character to the emulator
func (k *KeyboardChannel) PutChar(ch uint8) {
	k.keyChannel <- ch
}

// GetKey returns a pressed key if available
func (k *KeyboardChannel) GetKey(_ bool) (key uint8, ok bool) {
	select {
	case key = <-k.keyChannel:
		ok = true
	default:
		ok = false
	}
	return
}
