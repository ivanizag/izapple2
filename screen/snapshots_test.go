package screen

import "testing"

func TestScreenModeByName(t *testing.T) {
	tests := []struct {
		name       string
		screenMode int
	}{
		{"green", ScreenModeGreen},
		{"color", ScreenModeColor},
		{"greenscanlines", ScreenModeGreenScanlines},
		{"colorscanlines", ScreenModeColorScanlines},
		{" Green ", ScreenModeGreen},
	}
	for _, test := range tests {
		screenMode, err := ScreenModeByName(test.name)
		if err != nil {
			t.Error(err)
		} else if screenMode != test.screenMode {
			t.Errorf("screen '%s' should be mode %v, got %v", test.name, test.screenMode, screenMode)
		}
	}

	_, err := ScreenModeByName("amber")
	if err == nil {
		t.Error("screen 'amber' should not be supported")
	}
}
