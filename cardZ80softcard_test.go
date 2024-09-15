package izapple2

import (
	"strings"
	"testing"
)

func TestCPMBoot(t *testing.T) {
	at, err := makeApple2Tester("cpm", nil)
	if err != nil {
		t.Fatal(err)
	}

	banner := "APPLE ][ CP/M"
	prompt := "A>"
	at.terminateCondition = buildTerminateConditionTexts([]string{banner, prompt}, testTextMode40, 10_000_000)

	at.run()

	text := at.getText(testTextMode40)
	if !strings.Contains(text, banner) {
		t.Errorf("Expected '%s', got '%s'", banner, text)
	}
	if !strings.Contains(text, prompt) {
		t.Errorf("Expected prompt '%s', got '%s'", prompt, text)
	}

}
