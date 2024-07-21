package izapple2

import (
	"strings"
	"testing"
)

func testBoots(t *testing.T, model string, disk string, overrides *configuration, cycles uint64, banner string, prompt string, col80 bool) {
	if overrides == nil {
		overrides = newConfiguration()
	}

	if disk != "" {
		if overrides.has(confS6) {
			t.Fatal("Do not set custom slot 6 configuration and custom disk")
		}
		overrides.set(confS6, "diskii,disk1=\""+disk+"\"")
	} else if !overrides.has(confS6) {
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
	testBoots(t, "2plus", "", nil, 200_000, "APPLE ][", "\n]", false)
}

func Test2EBoots(t *testing.T) {
	testBoots(t, "2e", "", nil, 200_000, "Apple ][", "\n]", false)
}

func Test2EnhancedBoots(t *testing.T) {
	testBoots(t, "2enh", "", nil, 200_000, "Apple //e", "\n]", false)
}

func TestBase64Boots(t *testing.T) {
	testBoots(t, "base64a", "", nil, 1_000_000, "BASE 64A", "\n]", false)
}

func TestPlusDOS32Boots(t *testing.T) {
	overrides := newConfiguration()
	overrides.set(confS0, "multirom,bank=7,basic=0")
	overrides.set(confS6, "diskii,sectors13,disk1=<internal>/dos32.nib")
	testBoots(t, "2plus", "", overrides, 100_000_000, "MASTER DISKETTE VERSION 3.2 STANDARD", "\n>", false)
}

func TestPlusDOS33Boots(t *testing.T) {
	testBoots(t, "2plus", "<internal>/dos33.dsk", nil, 100_000_000, "DOS VERSION 3.3", "\n]", false)
}

func TestProdDOSBoots(t *testing.T) {
	testBoots(t, "2enh", "<internal>/ProDOS_2_4_3.po", nil, 100_000_000, "BITSY  BYE", "NEW VOL", false)
}

func TestCPM65Boots(t *testing.T) {
	testBoots(t, "2enh", "<internal>/cpm65.po", nil, 5_000_000, "CP/M-65 for the Apple II", "\nA>", true)
}
