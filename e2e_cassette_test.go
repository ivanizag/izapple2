package izapple2

import (
	"fmt"
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

// TestCassetteChangeTape puts two tapes, one after the other, in a machine
// that started without one, and reads both with the Monitor
func TestCassetteChangeTape(t *testing.T) {
	tapes := make([][]uint8, 2)
	paths := make([]string, 2)
	for n := range tapes {
		tapes[n] = make([]uint8, 16)
		for i := range tapes[n] {
			tapes[n][i] = uint8(i*13 + 7 + n*50)
		}
		var recorder tapeRecorder
		recorder.record(tapes[n])
		paths[n] = filepath.Join(t.TempDir(), "tape.wav")
		err := os.WriteFile(paths[n], wavBytes8BitMono(recorder.samples), 0o644)
		if err != nil {
			t.Fatal(err)
		}
	}

	a, err := CreateAppleFromModel("2plus", map[string]string{"s6": "empty"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	keys := &strobedKeys{}
	a.SetKeyboardProvider(keys)
	a.Init()
	cyclesToText(t, a, "\n]", 5_000_000)
	typeLines(t, a, keys, "CALL -151")

	for n, address := range []uint16{0x0800, 0x0900} {
		if err := a.LoadTape(paths[n]); err != nil {
			t.Fatal(err)
		}
		end := address + uint16(len(tapes[n])) - 1
		typeLines(t, a, keys, fmt.Sprintf("%X.%XR", address, end))
		a.RunCycles(30_000_000)
		for i, value := range tapes[n] {
			if got := a.mmu.Peek(address + uint16(i)); got != value {
				t.Fatalf("tape %v: expected $%02x at $%04x, got $%02x", n+1, value, address+uint16(i), got)
			}
		}
	}
	if strings.Contains(a.ScreenText(), "ERR") {
		t.Errorf("the Monitor reported a checksum error:\n%v", a.ScreenText())
	}
}

func TestCassetteNoTapeOnTheIIc(t *testing.T) {
	a, err := CreateAppleFromModel("2c", map[string]string{"s6": "iicdisk"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.LoadTape("<internal>/dos33.dsk"); err == nil {
		t.Error("expected the IIc to refuse a tape")
	}
}
