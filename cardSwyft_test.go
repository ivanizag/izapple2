package izapple2

import (
	"strings"
	"testing"
)

func TestSwyftTutorial(t *testing.T) {
	at := makeApple2Tester("2e")
	at.a.AddSwyftCard()
	err := at.a.AddDisk2(6, "<internal>/SwyftWare_-_SwyftCard_Tutorial.woz", "", nil)
	if err != nil {
		panic(err)
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
