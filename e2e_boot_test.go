package izapple2

import (
	"strings"
	"testing"
)

func testBoots(t *testing.T, model string, disk string, cycles uint64, banner string, prompt string, col80 bool) {
	overrides := newConfiguration()
	if disk != "" {
		overrides.set(confS6, "diskii,disk1=\""+disk+"\"")
	} else {
		overrides.set(confS6, "empty")
	}

	at, err := makeApple2Tester(model, overrides)
	if err != nil {
		t.Fatal(err)
	}
	at.terminateCondition = buildTerminateConditionTexts(at, []string{banner, prompt}, col80, cycles)
	at.run()

	var text string
	if col80 {
		text = at.getText80()
	} else {
		text = at.getText()
	}
	if !strings.Contains(text, banner) {
		t.Errorf("Expected '%s', got '%s'", banner, text)
	}
	if !strings.Contains(text, prompt) {
		t.Errorf("Expected prompt '%s', got '%s'", prompt, text)
	}

}

func TestPlusBoots(t *testing.T) {
	testBoots(t, "2plus", "", 200_000, "APPLE ][", "\n]", false)
}

func Test2EBoots(t *testing.T) {
	testBoots(t, "2e", "", 200_000, "Apple ][", "\n]", false)
}

func Test2EnhancedBoots(t *testing.T) {
	testBoots(t, "2enh", "", 200_000, "Apple //e", "\n]", false)
}

func TestBase64Boots(t *testing.T) {
	testBoots(t, "base64a", "", 1_000_000, "BASE 64A", "\n]", false)
}

func TestPlusDOS33Boots(t *testing.T) {
	testBoots(t, "2plus", "<internal>/dos33.dsk", 100_000_000, "DOS VERSION 3.3", "\n]", false)
}

func TestProdDOSBoots(t *testing.T) {
	testBoots(t, "2enh", "<internal>/ProDOS_2_4_3.po", 100_000_000, "BITSY  BYE", "NEW VOL", false)
}

func TestCPM65Boots(t *testing.T) {
	testBoots(t, "2enh", "<internal>/cpm65.po", 5_000_000, "CP/M-65 for the Apple II", "\nA>", true)
}
