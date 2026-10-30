package izapple2

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// wozWriteCards are the Disk II cards that write to the WOZ diskettes
var wozWriteCards = []string{"diskii", "diskiiseq"}

// wozCopy is a copy of the DOS 3.3 System Master in a file of the test, with
// the write protect tab on or off
func wozCopy(t *testing.T, writeProtected bool) string {
	t.Helper()
	data, err := os.ReadFile("woz_test_images/DOS 3.3 System Master.woz")
	if err != nil {
		t.Fatal(err)
	}
	data[22] = 0 // The WriteProtected field of INFO
	if writeProtected {
		data[22] = 1
	}
	path := filepath.Join(t.TempDir(), "master.woz")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// bootWoz starts DOS 3.3 from the WOZ with a Disk II card
func bootWoz(t *testing.T, card string, woz string, saveDir string) (*Apple2, *strobedKeys) {
	t.Helper()
	params := card + ",disk1=\"" + woz + "\""
	if saveDir != "" {
		params += ",savedir=\"" + saveDir + "\""
	}
	a, err := CreateAppleFromModel("2plus", map[string]string{"s6": params}, nil)
	if err != nil {
		t.Fatal(err)
	}
	keys := &strobedKeys{}
	a.SetKeyboardProvider(keys)
	a.Init()
	cyclesToText(t, a, "COPYRIGHT APPLE", 100_000_000)
	return a, keys
}

// typeLines types the lines and waits for the program to read them all
func typeLines(t *testing.T, a *Apple2, keys *strobedKeys, lines ...string) {
	t.Helper()
	for _, line := range lines {
		keys.pending = append(keys.pending, []uint8(line+"\r")...)
	}
	start := a.GetCycles()
	for len(keys.pending) > 0 {
		if a.GetCycles()-start > 100_000_000 {
			t.Fatalf("the keys were not read:\n%v", a.ScreenText())
		}
		a.RunCycles(100_000)
	}
}

// saveProgram saves a program to the diskette and waits for the motor to
// stop, when the track written is saved
func saveProgram(t *testing.T, a *Apple2, keys *strobedKeys) {
	t.Helper()
	typeLines(t, a, keys, "NEW", `10 PRINT "SAVED ON THE WOZ"`, "SAVE WOZTEST")
	a.RunCycles(20_000_000)
}

func runProgram(t *testing.T, card string, woz string, saveDir string) {
	t.Helper()
	a, keys := bootWoz(t, card, woz, saveDir)
	typeLines(t, a, keys, "RUN WOZTEST")
	cyclesToText(t, a, "SAVED ON THE WOZ", a.GetCycles()+20_000_000)
}

func TestWozSavesAProgram(t *testing.T) {
	for _, card := range wozWriteCards {
		t.Run(card, func(t *testing.T) {
			woz := wozCopy(t, false)
			a, keys := bootWoz(t, card, woz, "")
			saveProgram(t, a, keys)

			runProgram(t, card, woz, "")
		})
	}
}

func TestWozWriteProtectedRefusesToSave(t *testing.T) {
	for _, card := range wozWriteCards {
		t.Run(card, func(t *testing.T) {
			woz := wozCopy(t, true)
			original, err := os.ReadFile(woz)
			if err != nil {
				t.Fatal(err)
			}

			a, keys := bootWoz(t, card, woz, "")
			typeLines(t, a, keys, "SAVE WOZTEST")
			cyclesToText(t, a, "WRITE PROTECTED", a.GetCycles()+20_000_000)

			after, err := os.ReadFile(woz)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(after, original) {
				t.Error("the write protected image was modified")
			}
		})
	}
}

func TestWozSavesToTheSaveDirectory(t *testing.T) {
	for _, card := range wozWriteCards {
		t.Run(card, func(t *testing.T) {
			// Even a write protected image is written with a save directory
			woz := wozCopy(t, true)
			saveDir := t.TempDir()
			original, err := os.ReadFile(woz)
			if err != nil {
				t.Fatal(err)
			}

			a, keys := bootWoz(t, card, woz, saveDir)
			saveProgram(t, a, keys)

			after, err := os.ReadFile(woz)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(after, original) {
				t.Error("the image was modified, it should have been left alone")
			}
			if _, err := os.Stat(filepath.Join(saveDir, "master.woz.ovl")); err != nil {
				t.Errorf("no overlay in the save directory: %v", err)
			}

			runProgram(t, card, woz, saveDir)
		})
	}
}
