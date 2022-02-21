package izapple2

import (
	"strings"
	"testing"
)

func TestPlusBoots(t *testing.T) {
	at := makeApple2Tester("2plus")
	at.terminateCondition = func(a *Apple2) bool {
		return a.cpu.GetCycles() > 200_000
	}
	at.run()

	text := at.getText()
	if !strings.Contains(text, "APPLE ][") {
		t.Errorf("Expected 'APPLE ][', got '%s'", text)
	}
	if !strings.Contains(text, "\n]") {
		t.Errorf("Expected ] prompt, got '%s'", text)
	}
}

func Test2EBoots(t *testing.T) {
	at := makeApple2Tester("2e")
	at.terminateCondition = func(a *Apple2) bool {
		return a.cpu.GetCycles() > 200_000
	}
	at.run()

	text := at.getText()
	if !strings.Contains(text, "Apple ][") {
		t.Errorf("Expected 'Apple ][', got '%s'", text)
	}
	if !strings.Contains(text, "\n]") {
		t.Errorf("Expected ] prompt, got '%s'", text)
	}
}

func Test2EnhancedBoots(t *testing.T) {
	at := makeApple2Tester("2enh")
	at.terminateCondition = func(a *Apple2) bool {
		return a.cpu.GetCycles() > 200_000
	}
	at.run()

	text := at.getText()
	if !strings.Contains(text, "Apple //e") {
		t.Errorf("Expected 'Apple //e', got '%s'", text)
	}
	if !strings.Contains(text, "\n]") {
		t.Errorf("Expected ] prompt, got '%s'", text)
	}
}

func TestBase64Boots(t *testing.T) {
	at := makeApple2Tester("base64a")
	at.terminateCondition = func(a *Apple2) bool {
		return a.cpu.GetCycles() > 1_000_000
	}
	at.run()

	text := at.getText()
	if !strings.Contains(text, "BASE 64A") {
		t.Errorf("Expected 'BASE 64A', got '%s'", text)
	}
	if !strings.Contains(text, "\n]") {
		t.Errorf("Expected ] prompt, got '%s'", text)
	}
}

func TestPlusDOS33Boots(t *testing.T) {
	at := makeApple2Tester("2plus")

	err := at.a.AddDisk2(6, "<internal>/dos33.dsk", "")
	if err != nil {
		panic(err)
	}

	at.terminateCondition = func(a *Apple2) bool {
		return a.cpu.GetCycles() > 100_000_000
	}
	at.run()

	text := at.getText()
	if !strings.Contains(text, "DOS VERSION 3.3") {
		t.Errorf("Expected 'APPLE ][', got '%s'", text)
	}
	if !strings.Contains(text, "\n]") {
		t.Errorf("Expected ] prompt, got '%s'", text)
	}
}
