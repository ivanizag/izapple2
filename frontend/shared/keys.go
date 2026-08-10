package shared

/*
Every toolkit has its own constants for the keys of the keyboard. The frontends
translate theirs to the keys here, and this decides what the Apple II keyboard
would produce for them, so that all the frontends type the same.

See "Apple II reference manual", page 5.

To get the keys as understood by the Apple II hardware run:

	10 A=PEEK(49152)
	20 PRINT A, A - 128
	30 GOTO 10
*/

// Key is a key of the keyboard, whatever the toolkit of the frontend calls it
type Key int

const (
	KeyNone Key = iota // Not a key the frontends care about

	KeyEscape
	KeyBackspace
	KeyReturn
	KeyTab
	KeyDelete
	KeyInsert
	KeyLeft
	KeyRight
	KeyUp
	KeyDown

	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF12
	KeyPause
	KeyPrintScreen
)

// CharForKey returns the character the Apple II keyboard produces for a key,
// and whether it produces one at all
func CharForKey(key Key, ctrl bool) (uint8, bool) {
	switch key {
	case KeyEscape:
		return 27, true
	case KeyBackspace:
		return 8, true
	case KeyReturn:
		return 13, true
	case KeyLeft:
		if ctrl {
			return 31, true // Base64A
		}
		return 8, true
	case KeyRight:
		return 21, true

	// Apple //e
	case KeyUp:
		return 11, true // 31 in the Base64A
	case KeyDown:
		return 10, true
	case KeyTab:
		return 9, true
	case KeyDelete:
		return 127, true // 24 in the Base64A

	// Base64A clone particularities
	case KeyF3:
		return 127, true // Base64A
	}

	// Missing values 91 to 95. Usually control for [\]^_
	// On the Base64A it's control for \]./
	return 0, false
}

// CharForCtrlLetter returns the character a letter pressed with the control
// key produces, and whether it was a letter at all
func CharForCtrlLetter(letter rune) (uint8, bool) {
	if letter >= 'A' && letter <= 'Z' {
		letter += 'a' - 'A'
	}
	if letter < 'a' || letter > 'z' {
		return 0, false
	}

	return uint8(letter-'a') + 1, true
}
