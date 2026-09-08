package izapple2

import (
	"strings"
	"testing"
)

const forthBanner = "79-FORTH  2.2"

func TestForthRomCardBoots(t *testing.T) {
	at, err := makeApple2Tester("forth", nil)
	if err != nil {
		t.Fatal(err)
	}

	at.terminateCondition = buildTerminateConditionText(forthBanner, testTextMode40, 1_000_000)
	at.run()

	text := at.getText(testTextMode40)
	if !strings.Contains(text, forthBanner) {
		t.Errorf("Expected '%s', got '%s'", forthBanner, text)
	}
}

// TestForthRomCardRuns types an expression on the Forth interpreter in ROM
func TestForthRomCardRuns(t *testing.T) {
	at, err := makeApple2Tester("forth", nil)
	if err != nil {
		t.Fatal(err)
	}

	var kb testKeyboard
	at.a.SetKeyboardProvider(&kb)
	kb.putLine("2 3 + . ")

	const expected = "2 3 + .  5 OK"
	at.terminateCondition = buildTerminateConditionText(expected, testTextMode40, 10_000_000)
	at.run()

	text := at.getText(testTextMode40)
	if !strings.Contains(text, expected) {
		t.Errorf("Expected '%s', got '%s'", expected, text)
	}
}

// TestForthRomCardOnAnotherSlot verifies that the card replaces the Basic ROM
// from any slot, and that it wins over the language card of the Apple ][+
func TestForthRomCardOnAnotherSlot(t *testing.T) {
	overrides := newConfiguration()
	overrides.set(confS2, "forthrom")
	overrides.set(confS6, "empty")

	at, err := makeApple2Tester("2plus", overrides)
	if err != nil {
		t.Fatal(err)
	}

	at.terminateCondition = buildTerminateConditionText(forthBanner, testTextMode40, 1_000_000)
	at.run()

	text := at.getText(testTextMode40)
	if !strings.Contains(text, forthBanner) {
		t.Errorf("Expected '%s', got '%s'", forthBanner, text)
	}
}
