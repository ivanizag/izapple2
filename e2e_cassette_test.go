package izapple2

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testKeyboard sends queued keys respecting the strobe, so bursts of
// keys are not dropped as with KeyboardChannel
type testKeyboard struct {
	keys []uint8
}

func (k *testKeyboard) GetKey(strobed bool) (uint8, bool) {
	if strobed && len(k.keys) > 0 {
		key := k.keys[0]
		k.keys = k.keys[1:]
		return key, true
	}
	return 0, false
}

func (k *testKeyboard) putLine(text string) {
	k.keys = append(k.keys, []uint8(text)...)
	k.keys = append(k.keys, 13)
}

// TestCassetteMonitorLoad reads a synthesized tape into memory using
// the Monitor ROM tape read routine
func TestCassetteMonitorLoad(t *testing.T) {
	data := make([]uint8, 16)
	for i := range data {
		data[i] = uint8(i*13 + 7)
	}
	var recorder tapeRecorder
	recorder.record(data)

	path := filepath.Join(t.TempDir(), "test.wav")
	err := os.WriteFile(path, wavBytes8BitMono(recorder.samples), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	overrides := newConfiguration()
	overrides.set(confS6, "empty")
	overrides.set(confTape, path)

	at, err := makeApple2Tester("2plus", overrides)
	if err != nil {
		t.Fatal(err)
	}

	var kb testKeyboard
	at.a.SetKeyboardProvider(&kb)
	const address = uint16(0x0800)
	dataLoaded := func(a *Apple2) bool {
		for i, value := range data {
			if a.mmu.Peek(address+uint16(i)) != value {
				return false
			}
		}
		return true
	}

	stage := 0
	lastCheck := uint64(0)
	at.terminateCondition = func(a *Apple2) bool {
		cycles := a.GetCycles()
		if cycles > 100_000_000 {
			return true
		}
		if cycles-lastCheck < textCheckInterval {
			return false
		}
		lastCheck = cycles

		switch stage {
		case 0: // Wait for the BASIC prompt and enter the Monitor
			if strings.Contains(at.getText(testTextMode40), "\n]") {
				kb.putLine("CALL -151")
				stage = 1
			}
		case 1: // Wait for the Monitor prompt and read the tape
			if strings.Contains(at.getText(testTextMode40), "\n*") {
				kb.putLine("800.80FR")
				stage = 2
			}
		case 2:
			return dataLoaded(a)
		}
		return false
	}
	at.run()

	if !dataLoaded(at.a) {
		t.Errorf("The tape data was not loaded at $%04x", address)
	}
	if strings.Contains(at.getText(testTextMode40), "ERR") {
		t.Errorf("The Monitor reported a checksum error:\n%v", at.getText(testTextMode40))
	}
}
