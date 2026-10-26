package izapple2

import (
	"strings"
	"testing"
)

func TestCPMApplicardBoot(t *testing.T) {
	at, err := makeApple2Tester("cpmapplicard", nil)
	if err != nil {
		t.Fatal(err)
	}

	banner := "PCPI Ver 2.0"
	prompt := "A>"
	at.terminateCondition = buildTerminateConditionTexts([]string{banner, prompt}, testTextMode80, 30_000_000)

	at.run()

	text := at.getText(testTextMode80)
	if !strings.Contains(text, banner) {
		t.Errorf("Expected '%s', got '%s'", banner, text)
	}
	if !strings.Contains(text, prompt) {
		t.Errorf("Expected prompt '%s', got '%s'", prompt, text)
	}

}
