package izapple2

import (
	"testing"
)

func TestVidexCharText(t *testing.T) {
	for _, c := range []struct {
		char     uint8
		expected string
	}{
		{'U', "U"},
		{'U' | 0x80, "U"}, // The second character set, inverse or highlighted
		{0x0f, " "},       // A graphic glyph of the Ultraterm
		{0x8f, " "},
	} {
		if got := videxCharText(c.char); got != c.expected {
			t.Errorf("expected %q for $%02x, got %q", c.expected, c.char, got)
		}
	}
}

func TestUltratermScreenText(t *testing.T) {
	// The demonstration of the Ultraterm shows its text in 80 columns
	a, err := CreateAppleFromModel("ultraterm", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	a.Init()
	cyclesToText(t, a, "Ultra-Versatility, Ultra-Quality", 60_000_000)
}
