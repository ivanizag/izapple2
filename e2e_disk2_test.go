package izapple2

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

/*
A DOS 3.3 boot on an Apple II takes a few seconds, five million cycles to the
banner. DOS turns the drive off after each access and, at the start of the
next, reads the data latch to tell whether the disk still turns, before it
waits for the motor to come up to speed. If it can't tell, each sector costs
that wait, and the boot takes more than thirty million cycles.
*/
const dos33BootCycles = 10_000_000

// cyclesToText runs a machine until a text is on the screen and returns the
// cycles it took, or fails after a limit
func cyclesToText(t *testing.T, a *Apple2, text string, limit uint64) uint64 {
	t.Helper()
	for !strings.Contains(a.ScreenText(), text) {
		if a.GetCycles() > limit {
			t.Fatalf("%q was not on the screen after %v cycles:\n%v", text, limit, a.ScreenText())
		}
		a.RunCycles(100_000)
	}
	return a.GetCycles()
}

// writableCopy is a copy of an embedded disk image in a file of the test
func writableCopy(t *testing.T, resource string) string {
	t.Helper()
	data, _, err := LoadResource(resource)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), filepath.Base(resource))
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestDisk2BootsDOS33InFewSeconds(t *testing.T) {
	for name, disk := range map[string]string{
		"read only": "<internal>/dos33.dsk",
		"writable":  writableCopy(t, "<internal>/dos33.dsk"),
	} {
		t.Run(name, func(t *testing.T) {
			a, err := CreateAppleFromModel("2plus", map[string]string{"s6": "diskii,disk1=\"" + disk + "\""}, nil)
			if err != nil {
				t.Fatal(err)
			}
			a.Init()
			cycles := cyclesToText(t, a, "DOS VERSION 3.3", 100_000_000)
			t.Logf("DOS 3.3 started in %v cycles", cycles)
			if cycles > dos33BootCycles {
				t.Errorf("DOS 3.3 took %v cycles to start, more than %v", cycles, dos33BootCycles)
			}
		})
	}
}

// strobedKeys is a keyboard that hands a key only once the program has
// read the one before, as the latch of the keyboard does
type strobedKeys struct {
	pending []uint8
}

func (k *strobedKeys) GetKey(strobed bool) (uint8, bool) {
	if !strobed || len(k.pending) == 0 {
		return 0, false
	}
	key := k.pending[0]
	k.pending = k.pending[1:]
	return key, true
}

func TestDisk2InitializesADisketteThatStarts(t *testing.T) {
	blank := filepath.Join(t.TempDir(), "blank.dsk")
	if err := os.WriteFile(blank, make([]byte, 143_360), 0o644); err != nil {
		t.Fatal(err)
	}

	a, err := CreateAppleFromModel("2plus", map[string]string{
		"s6": "diskii,disk1=<internal>/dos33.dsk,disk2=\"" + blank + "\"",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	keys := &strobedKeys{}
	a.SetKeyboardProvider(keys)
	a.Init()
	cyclesToText(t, a, "DOS VERSION 3.3", 100_000_000)

	for _, line := range []string{"NEW", `10 PRINT "INITIALIZED HERE"`, "INIT HELLO,D2"} {
		keys.pending = append(keys.pending, []uint8(line+"\r")...)
	}
	start := a.GetCycles()
	for len(keys.pending) > 0 {
		if a.GetCycles()-start > 100_000_000 {
			t.Fatalf("the keys were not read:\n%v", a.ScreenText())
		}
		a.RunCycles(100_000)
	}
	a.RunCycles(100_000_000) // INIT formats the whole diskette

	again, err := CreateAppleFromModel("2plus", map[string]string{"s6": "diskii,disk1=\"" + blank + "\""}, nil)
	if err != nil {
		t.Fatal(err)
	}
	again.Init()
	cyclesToText(t, again, "INITIALIZED HERE", 100_000_000)
}
