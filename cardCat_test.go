package izapple2

import (
	"strings"
	"testing"
)

func testCardDetectedInternal(t *testing.T, model string, card string, cycles uint64, banner string) {
	overrides := newConfiguration()
	overrides.set(confS3, "empty")
	overrides.set(confS4, "empty")
	overrides.set(confS5, "empty")
	overrides.set(confS7, "empty")

	overrides.set(confS2, card)
	overrides.set(confS6, "diskii,disk1=\"<internal>/Card Cat 1.0b9.dsk\"")

	at, err := makeApple2Tester(model, overrides)
	if err != nil {
		t.Fatal(err)
	}
	at.terminateCondition = buildTerminateConditionText(at, banner, true, cycles)
	at.run()

	text := at.getText80()
	if !strings.Contains(text, banner) {
		t.Errorf("Expected '%s', got '%s'", banner, text)
	}
}

func TestCardsDetected(t *testing.T) {

	t.Run("test Memory Expansion card", func(t *testing.T) {
		testCardDetectedInternal(t, "2enh", "memexp", 50_000_000, "2   03-00-05-D0  Apple II Memory Expansion Card (SP)")
	})

	t.Run("test Mouse card", func(t *testing.T) {
		testCardDetectedInternal(t, "2enh", "mouse", 50_000_000, "2   38-18-01-20  Apple II Mouse Card")
	})

	t.Run("test Parallel printer card", func(t *testing.T) {
		testCardDetectedInternal(t, "2enh", "parallel", 50_000_000, "2   48-48-58-FF  Apple Parallel Interface Card")
	})

	t.Run("test ThunderClock Plus card", func(t *testing.T) {
		testCardDetectedInternal(t, "2enh", "thunderclock", 50_000_000, "2   FF-05-18-B8  ThunderClock Plus Card")
	})

	// Saturn not detected
	// VidHD not detected
	// Swyftcard not compatible with Card Cat
	// Pending to try Saturn, 80col with 2plus but fails with an illegal opcode

}
