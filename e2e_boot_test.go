package izapple2

import (
	"strings"
	"testing"
)

func testBoots(t *testing.T, model string, disk string, overrides *configuration, cycles uint64, banner string, prompt string, textMode testTextModeFunc) {
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
	at.terminateCondition = buildTerminateConditionTexts([]string{banner, prompt}, textMode, cycles)
	at.run()

	text := at.getText(textMode)
	if !strings.Contains(text, banner) {
		t.Errorf("Expected '%s', got '%s'", banner, text)
	}
	if !strings.Contains(text, prompt) {
		t.Errorf("Expected prompt '%s', got '%s'", prompt, text)
	}

}

func TestPlusBoots(t *testing.T) {
	testBoots(t, "2plus", "", nil, 200_000, "APPLE ][", "\n]", testTextMode40)
}

func Test2EBoots(t *testing.T) {
	testBoots(t, "2e", "", nil, 200_000, "Apple ][", "\n]", testTextMode40)
}

func Test2EnhancedBoots(t *testing.T) {
	testBoots(t, "2enh", "", nil, 200_000, "Apple //e", "\n]", testTextMode40)
}

func TestBase64Boots(t *testing.T) {
	testBoots(t, "base64a", "", nil, 1_000_000, "BASE 64A", "\n]", testTextMode40)
}

func TestBasis108Boots(t *testing.T) {
	testBoots(t, "basis108", "", nil, 1_000_000, "B a s i s   1 0 8", "\n]", testTextMode80AltOrder)
}

func TestPlusDOS32Boots(t *testing.T) {
	overrides := newConfiguration()
	overrides.set(confS0, "multirom,bank=7,basic=0")
	overrides.set(confS6, "diskii,sectors13,disk1=<internal>/dos32.nib")
	testBoots(t, "2plus", "", overrides, 100_000_000, "MASTER DISKETTE VERSION 3.2 STANDARD", "\n>", testTextMode40)
}

func TestPlusDOS33Boots(t *testing.T) {
	testBoots(t, "2plus", "<internal>/dos33.dsk", nil, 100_000_000, "DOS VERSION 3.3", "\n]", testTextMode40)
}

func TestProdDOSBoots(t *testing.T) {
	testBoots(t, "2enh", "<internal>/ProDOS_2_4_3.po", nil, 100_000_000, "BITSY  BYE", "NEW VOL", testTextMode40)
}

func TestCPM65Boots(t *testing.T) {
	testBoots(t, "2enh", "<internal>/cpm65.po", nil, 5_000_000, "CP/M-65 for the Apple II", "\nA>", testTextMode80)
}
