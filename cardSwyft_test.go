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
