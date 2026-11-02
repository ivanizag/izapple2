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
		testCardDetectedInternal(t, "2enh", "memexp", "s2", 55_000_000, "2   Apple II Memory Expansion Card (SP)")
	})

	t.Run("test Mouse card", func(t *testing.T) {
		testCardDetectedInternal(t, "2enh", "mouse", "s2", 55_000_000, "2   Apple II Mouse Card")
	})

	t.Run("test Parallel printer card", func(t *testing.T) {
		testCardDetectedInternal(t, "2enh", "parallel", "s2", 55_000_000, "2   Apple Parallel Interface Card")
	})

	t.Run("test ThunderClock Plus card", func(t *testing.T) {
		testCardDetectedInternal(t, "2enh", "thunderclock", "s2", 55_000_000, "2   ThunderClock Plus Card")
	})

	t.Run("test Z80 Softcard card", func(t *testing.T) {
		testCardDetectedInternal(t, "2enh", "z80softcard", "s2", 55_000_000, "2   Z80 Card")
	})

	t.Run("test VidHD card", func(t *testing.T) {
		testCardDetectedInternal(t, "2enh", "vidhd", "s2", 55_000_000, "2   No Firmware Card Detected")
	})

	t.Run("test Saturn card", func(t *testing.T) {
		testCardDetectedInternal(t, "2plus", "saturn", "s0", 55_000_000, "SATURN 128K CARD IN SLOT 0")
	})

	t.Run("test Videx Videoterm card", func(t *testing.T) {
		testCardDetectedInternal(t, "2plus", "videx", "s3", 55_000_000, "3   Videx 80 Column Text Display Card")
	})

	t.Run("test Videx Ultraterm card", func(t *testing.T) {
		testCardDetectedInternal(t, "2plus", "videxultraterm", "s3", 55_000_000, "3   ? Unknown 80-Column Display Card")
	})

	t.Run("test Dan 2 SD card", func(t *testing.T) {
		testCardDetectedInternal(t, "2enh", "dan2sd", "s2", 55_000_000, "2   DAN II Card")
	})

	t.Run("test ProDOS ROM Drive card", func(t *testing.T) {
		testCardDetectedInternal(t, "2enh", "prodosromdrive", "s2", 55_000_000, "2   ProDOS ROM Drive Card")
	})

	t.Run("test RAMWorks aux card", func(t *testing.T) {
		testCardDetectedInternal(t, "2enh", "4096", "ramworks", 55_000_000, "RAMWorks 4096K Card in Aux Slot")
	})

	t.Run("test Mockingboard card", func(t *testing.T) {
		testCardDetectedInternal(t, "2enh", "mockingboard", "s4", 55_000_000, "4   MockingBoard Card")
	})

	t.Run("test Liron card", func(t *testing.T) {
		testCardDetectedInternal(t, "2enh", "liron", "s2", 55_000_000, "2   Apple II LIRON Card (SP)")
	})

	t.Run("test Super Serial Card", func(t *testing.T) {
		testCardDetectedInternal(t, "2enh", "ssc", "s2", 55_000_000, "2   Super Serial Card")
	})

	t.Run("test Disk II card", func(t *testing.T) {
		testCardDetectedInternal(t, "2enh", "diskii", "s5", 55_000_000, "5   5.25 Disk Interface Card")
	})

	t.Run("test Disk II card with the sequencer", func(t *testing.T) {
		testCardDetectedInternal(t, "2enh", "diskiiseq", "s5", 55_000_000, "5   5.25 Disk Interface Card")
	})

	// Swyftcard not compatible with Card Cat
	// Brainboard, Brainboard II, Forth ROM and MultiROM replace the main ROM
	// Unknown cards: smartport, prodosblock, fujinet, profile, fastchip,
	// applicard, z80softcard2, softswitchlogger and prodosromcard3
}

/*
TestCardsDetectedOnTheIIc runs Card Cat on the Apple IIc models. It tells the
version of the ROM and finds what the IIc has built in on the ports its
firmware uses. The lines are cut at 40 columns, only their start is checked.
*/
func TestCardsDetectedOnTheIIc(t *testing.T) {
	for _, c := range []struct {
		model    string
		expected []string
	}{
		{"2c_rom255", []string{
			"Apple IIc (Original - 16K ROM)",
			"1   Super Serial Card",
			"2   Super Serial Card",
			"4   Apple II Mouse Card",
			"6   5.25 Disk Interface Card",
		}},
		{"2c_rom0", []string{
			"Apple IIc (32K ROM V0)",
			"1   Super Serial Card",
			"2   Super Serial Card",
			"4   Apple II Mouse Card",
			"5   Apple II LIRON Card (SP)", // The SmartPort of the external port
			"6   5.25 Disk Interface Card",
			"7   ? Unknown Network or Bus Interface", // The AppleTalk firmware
		}},
		{"2c_rom3", []string{
			"Apple IIc (32K ROM V3)",
			"1   Super Serial Card",
			"2   Super Serial Card",
			"4   Apple II Memory Expansion 1024K",
			"5   Apple II LIRON Card (SP)",
			"6   5.25 Disk Interface Card",
			"7   Apple II Mouse Card",
		}},
		{"2c", []string{
			"Apple IIc (32K ROM V4)",
			"1   Super Serial Card",
			"2   Super Serial Card",
			"4   Apple II Memory Expansion 1024K",
			"5   Apple II LIRON Card (SP)",
			"6   5.25 Disk Interface Card",
			"7   Apple II Mouse Card",
		}},
	} {
		t.Run(c.model, func(t *testing.T) {
			overrides := newConfiguration()
			overrides.set(confS6, "iicdisk,disk1=\"<internal>/Card Cat 1.7.dsk\"")
			at, err := makeApple2Tester(c.model, overrides)
			if err != nil {
				t.Fatal(err)
			}
			last := c.expected[len(c.expected)-1]
			at.terminateCondition = buildTerminateConditionText(last, testTextMode80, 55_000_000)
			at.run()

			text := at.getTextBest()
			for _, expected := range c.expected {
				if !strings.Contains(text, expected) {
					t.Errorf("expected '%s', got '%s'", expected, text)
				}
			}
		})
	}
}
