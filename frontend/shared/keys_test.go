package shared

import "testing"

func TestCharForKey(t *testing.T) {
	cases := []struct {
		key      Key
		ctrl     bool
		expected uint8
		is       bool
	}{
		{KeyEscape, false, 27, true},
		{KeyBackspace, false, 8, true},
		{KeyReturn, false, 13, true},
		{KeyTab, false, 9, true},
		{KeyDelete, false, 127, true},
		{KeyLeft, false, 8, true},
		{KeyLeft, true, 31, true}, // Base64A
		{KeyRight, false, 21, true},
		{KeyUp, false, 11, true},
		{KeyDown, false, 10, true},
		{KeyF3, false, 127, true}, // Base64A
		{KeyF1, false, 0, false},  // A command of the emulator, not a character
		{KeyF12, true, 0, false},
		{KeyNone, false, 0, false},
	}

	for _, c := range cases {
		actual, is := CharForKey(c.key, c.ctrl)
		if actual != c.expected || is != c.is {
			t.Errorf("CharForKey(%v, %v) is (%v, %v), expected (%v, %v)",
				c.key, c.ctrl, actual, is, c.expected, c.is)
		}
	}
}

func TestCharForCtrlLetter(t *testing.T) {
	letters := map[rune]uint8{
		'a': 1, 'm': 13, 'z': 26,
		'A': 1, 'M': 13, 'Z': 26, // The upper case letters are the same keys
	}

	for letter, expected := range letters {
		actual, is := CharForCtrlLetter(letter)
		if actual != expected || !is {
			t.Errorf("CharForCtrlLetter(%q) is (%v, %v), expected (%v, true)",
				letter, actual, is, expected)
		}
	}

	for _, letter := range []rune{'0', ' ', '[', '@', 'ñ'} {
		if _, is := CharForCtrlLetter(letter); is {
			t.Errorf("CharForCtrlLetter(%q) should not be a letter", letter)
		}
	}
}
