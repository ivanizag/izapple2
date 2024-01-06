package izapple2

import (
	"strings"
	"testing"
)

func testBoots(t *testing.T, model string, disk string, cycles uint64, banner string, prompt string) {
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
	at.terminateCondition = func(a *Apple2) bool {
		return a.cpu.GetCycles() > cycles
	}
	at.run()

	text := at.getText()
	if !strings.Contains(text, banner) {
		t.Errorf("Expected '%s', got '%s'", banner, text)
	}
	if !strings.Contains(text, prompt) {
		t.Errorf("Expected prompt '%s', got '%s'", prompt, text)
	}

}

func TestPlusBoots(t *testing.T) {
	testBoots(t, "2plus", "", 200_000, "APPLE ][", "\n]")
}

func Test2EBoots(t *testing.T) {
	testBoots(t, "2e", "", 200_000, "Apple ][", "\n]")
}

func Test2EnhancedBoots(t *testing.T) {
	testBoots(t, "2enh", "", 200_000, "Apple //e", "\n]")
}

func TestBase64Boots(t *testing.T) {
	testBoots(t, "base64a", "", 1_000_000, "BASE 64A", "\n]")
}

func TestPlusDOS33Boots(t *testing.T) {
	testBoots(t, "2plus", "<internal>/dos33.dsk", 100_000_000, "DOS VERSION 3.3", "\n]")
}
