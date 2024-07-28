package izapple2

import (
	"strings"
	"testing"
)

func testA2AuditInternal(t *testing.T, model string, removeLangCard bool, cycles uint64, messages []string) {
	overrides := newConfiguration()
	if removeLangCard {
		overrides.set(confS0, "empty")
	}
	overrides.set(confS1, "empty")
	overrides.set(confS2, "empty")
	overrides.set(confS3, "empty")
	overrides.set(confS4, "empty")
	overrides.set(confS5, "empty")
	overrides.set(confS6, "diskii,disk1=\"<internal>/audit.dsk\"")
	overrides.set(confS7, "empty")

	at, err := makeApple2Tester(model, overrides)
	if err != nil {
		t.Fatal(err)
	}
	at.terminateCondition = buildTerminateConditionTexts(messages, testTextMode40, cycles)
	at.run()

	text := at.getText(testTextMode40)
	for _, message := range messages {
		if !strings.Contains(text, message) {
			t.Errorf("Expected '%s', got '%s'", message, text)
		}
	}
}

func TestA2Audit(t *testing.T) {

	t.Run("test a2audit on Apple IIe enhanced", func(t *testing.T) {
		testA2AuditInternal(t, "2enh", false, 4_000_000, []string{
			"MEMORY:128K",
			"APPLE IIE (ENHANCED)",
			"LANGUAGE CARD TESTS SUCCEEDED",
			"AUXMEM TESTS SUCCEEDED",
			"SOFTSWITCH TESTS SUCCEEDED",
		})
	})

	t.Run("test a2audit on Apple IIe", func(t *testing.T) {
		testA2AuditInternal(t, "2e", false, 4_000_000, []string{
			"MEMORY:128K",
			"APPLE IIE",
			"LANGUAGE CARD TESTS SUCCEEDED",
			"AUXMEM TESTS SUCCEEDED",
			"SOFTSWITCH TESTS SUCCEEDED",
		})
	})

	t.Run("test a2audit on Apple II plus", func(t *testing.T) {
		testA2AuditInternal(t, "2plus", false, 4_000_000, []string{
			"MEMORY:64K",
			"APPLE II PLUS",
			"LANGUAGE CARD TESTS SUCCEEDED",
			"64K OR LESS:SKIPPING AUXMEM TEST",
			"NOT IIE OR IIC:SKIPPING SOFTSWITCH TEST",
		})
	})

	t.Run("test a2audit on Apple II plus without lang card", func(t *testing.T) {
		testA2AuditInternal(t, "2plus", true, 4_000_000, []string{
			"MEMORY:48K",
			"APPLE II PLUS",
			"48K:SKIPPING LANGUAGE CARD TEST",
			"64K OR LESS:SKIPPING AUXMEM TEST",
			"NOT IIE OR IIC:SKIPPING SOFTSWITCH TEST",
		})
	})

	/*
	   	t.Run("test Mouse card", func(t *testing.T) {
	   		testCardDetectedInternal(t, "2enh", "mouse", 50_000_000, "2   38-18-01-20  Apple II Mouse Card")
	   	})

	   	t.Run("test Parallel printer card", func(t *testing.T) {
	   		testCardDetectedInternal(t, "2enh", "parallel", 50_000_000, "2   48-48-58-FF  Apple Parallel Interface Card")
	   	})

	   // Saturn not detected
	   // VidHD not detected
	   // Swyftcard not compatible with Card Cat
	   // Pending to try Saturn, 80col with 2plus.

	   	t.Run("test ThunderClock Plus card", func(t *testing.T) {
	   		testCardDetectedInternal(t, "2enh", "thunderclock", 50_000_000, "2   FF-05-18-B8  ThunderClock Plus Card")
	   	})
	*/
}
