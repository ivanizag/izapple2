package izapple2

import (
	"strings"
	"testing"
)

func TestSwyftTutorial(t *testing.T) {
	at, err := makeApple2Tester("swyft", nil)
	if err != nil {
		t.Fatal(err)
	}

	at.terminateCondition = buildTerminateConditionText("HOW TO USE SWYFTCARD", testTextMode80, 10_000_000)

	at.run()

	text := at.getText(testTextMode80)
	if !strings.Contains(text, "HOW TO USE SWYFTCARD") {
		t.Errorf("Expected 'HOW TO USE SWYFTCARD', got '%s'", text)
	}

}

// TestSwyftCalc runs Calc, Use Front G, as page 61 of the tutorial does:
// what was typed since the last leap goes to Applesoft, that runs from the
// language card RAM where the card has copied and patched the ROM, and the
// answer comes back to the Text
func TestSwyftCalc(t *testing.T) {
	at, err := makeApple2Tester("swyft", nil)
	if err != nil {
		t.Fatal(err)
	}

	var kb testKeyboard
	at.a.SetKeyboardProvider(&kb)

	// Typed when the tutorial shows, a key every tenth of a second as the
	// card loses the keys typed faster. Control-G is Calc.
	const expected = "? 5.6 + 38.6"
	toType := []uint8("? 5.6 + 3\x07")
	tutorial := buildTerminateConditionText("HOW TO USE SWYFTCARD", testTextMode80, 20_000_000)
	answer := buildTerminateConditionText(expected, testTextMode80, 40_000_000)
	shown := false
	nextKey := uint64(0)
	at.terminateCondition = func(a *Apple2) bool {
		if !shown {
			shown = tutorial(a)
			nextKey = a.GetCycles() + 1_000_000
			return a.GetCycles() > 20_000_000
		}
		if len(toType) > 0 && a.GetCycles() > nextKey {
			kb.keys = append(kb.keys, toType[0])
			toType = toType[1:]
			nextKey = a.GetCycles() + 100_000
		}
		return answer(a)
	}
	at.run()

	text := at.getText(testTextMode80)
	if !strings.Contains(text, expected) {
		t.Errorf("Expected '%s', got '%s'", expected, text)
	}
}
