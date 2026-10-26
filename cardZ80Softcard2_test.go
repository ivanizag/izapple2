package izapple2

import (
	"strings"
	"testing"
)

func TestCPM288Boot(t *testing.T) {
	at, err := makeApple2Tester("cpm288", nil)
	if err != nil {
		t.Fatal(err)
	}

	banner := "Softcard II CP/M"
	prompt := "A>"
	at.terminateCondition = buildTerminateConditionTexts([]string{banner, prompt}, testTextMode80, 10_000_000)

	at.run()

	text := at.getText(testTextMode80)
	if !strings.Contains(text, banner) {
		t.Errorf("Expected '%s', got '%s'", banner, text)
	}
	if !strings.Contains(text, prompt) {
		t.Errorf("Expected prompt '%s', got '%s'", prompt, text)
	}

}
