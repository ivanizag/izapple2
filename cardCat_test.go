package izapple2

import (
	"strings"
	"testing"
)

func testCardDetectedInternal(t *testing.T, model string, card string, slot string, cycles uint64, banner string) {
	overrides := newConfiguration()
	overrides.set(confS2, "empty")
	overrides.set(confS3, "empty")
	overrides.set(confS4, "empty")
	overrides.set(confS5, "empty")
	overrides.set(confS7, "empty")
	overrides.set(confRamworks, "none")
	overrides.set(slot, card)

	overrides.set(confS6, "diskii,disk1=\"<internal>/Card Cat 1.7.dsk\"")

	at, err := makeApple2Tester(model, overrides)
	if err != nil {
		t.Fatal(err)
	}
	at.terminateCondition = buildTerminateConditionText(banner, testTextMode80, cycles)
	at.run()

	text := at.getTextBest()
	if !strings.Contains(text, banner) {
		t.Errorf("Expected '%s', got '%s'", banner, text)
	}
}

func TestCardsDetected(t *testing.T) {

	t.Run("test Memory Expansion card", func(t *testing.T) {
		testCardDetectedInternal(t, "2enh", "memexp", "s2", 50_000_000, "2   Apple II Memory Expansion Card (SP)")
	})

	t.Run("test Mouse card", func(t *testing.T) {
		testCardDetectedInternal(t, "2enh", "mouse", "s2", 50_000_000, "2   Apple II Mouse Card")
	})

	t.Run("test Parallel printer card", func(t *testing.T) {
		testCardDetectedInternal(t, "2enh", "parallel", "s2", 50_000_000, "2   Apple Parallel Interface Card")
	})

	t.Run("test ThunderClock Plus card", func(t *testing.T) {
		testCardDetectedInternal(t, "2enh", "thunderclock", "s2", 50_000_000, "2   ThunderClock Plus Card")
	})

	t.Run("test Z80 Softcard card", func(t *testing.T) {
		testCardDetectedInternal(t, "2enh", "z80softcard", "s2", 50_000_000, "2   Z80 Card")
	})

	t.Run("test VidHD card", func(t *testing.T) {
		testCardDetectedInternal(t, "2enh", "vidhd", "s2", 50_000_000, "2   No Firmware Card Detected")
	})

	t.Run("test Saturn card", func(t *testing.T) {
		testCardDetectedInternal(t, "2plus", "saturn", "s0", 50_000_000, "SATURN 128K CARD IN SLOT 0")
	})

	t.Run("test Videx Videoterm card", func(t *testing.T) {
		testCardDetectedInternal(t, "2plus", "videx", "s3", 50_000_000, "3   Videx 80 Column Text Display Card")
	})

	t.Run("test Videx Ultraterm card", func(t *testing.T) {
		testCardDetectedInternal(t, "2plus", "videxultraterm", "s3", 50_000_000, "3   ? Unknown 80-Column Display Card")
	})

	t.Run("test Dan 2 SD card", func(t *testing.T) {
		testCardDetectedInternal(t, "2enh", "dan2sd", "s2", 50_000_000, "2   DAN II Card")
	})

	t.Run("test ProDOS ROM Drive card", func(t *testing.T) {
		testCardDetectedInternal(t, "2enh", "prodosromdrive", "s2", 50_000_000, "2   ProDOS ROM Drive Card")
	})

	t.Run("test RAMWorks aux card", func(t *testing.T) {
		testCardDetectedInternal(t, "2enh", "4096", "ramworks", 50_000_000, "RAMWorks 4096K Card in Aux Slot")
	})

	// Swyftcard not compatible with Card Cat
	// Unknonw cards: prodosromcard3
}
