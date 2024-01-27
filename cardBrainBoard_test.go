package izapple2

import (
	"strings"
	"testing"
)

func buildBrainBoardTester(t *testing.T, conf string) *apple2Tester {
	overrides := newConfiguration()
	overrides.set(confS2, conf)
	overrides.set(confS3, "empty")
	overrides.set(confS4, "empty")
	overrides.set(confS5, "empty")
	overrides.set(confS6, "empty")
	overrides.set(confS7, "empty")

	at, err := makeApple2Tester("2plus", overrides)
	if err != nil {
		t.Fatal(err)
	}
	return at
}

func TestBrainBoardCardWozaniam(t *testing.T) {
	at := buildBrainBoardTester(t, "brainboard,switch=up")

	at.terminateCondition = func(a *Apple2) bool {
		return a.cpu.GetCycles() > 10_000_000
	}
	at.run()

	at.terminateCondition = buildTerminateConditionText(at, "_@_@_@_@_@_@_@_@_@_@_@_@_@_@_@_@_@_@_@_@", false, 100_000)

	text := at.getText()
	if !strings.Contains(text, "_@_@_@_@_@_@_@_@_@_@_@_@_@_@_@_@_@_@_@_@") {
		t.Errorf("Expected screen filled with _@_@', got '%s'", text)
	}
}

func TestBrainBoardCardIntegerBasic(t *testing.T) {
	at := buildBrainBoardTester(t, "brainboard,switch=down")

	at.terminateCondition = buildTerminateConditionText(at, "APPLE ][\n>", false, 1_000_000)
	at.run()

	text := at.getText()
	if !strings.Contains(text, "APPLE ][\n>") {
		t.Errorf("Expected APPLE ][' and '>', got '%s'", text)
	}
}
