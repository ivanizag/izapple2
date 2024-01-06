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

	at.terminateCondition = func(a *Apple2) bool {
		return a.cpu.GetCycles() > 10_000_000
	}
	at.run()

	text := at.getText80()
	if !strings.Contains(text, "HOW TO USE SWYFTCARD") {
		t.Errorf("Expected 'HOW TO USE SWYFTCARD', got '%s'", text)
	}

}
