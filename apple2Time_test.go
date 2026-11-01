package izapple2

import (
	"hash/fnv"
	"math"
	"slices"
	"testing"
	"time"
)

// cyclesFor returns the cycles, rounded up, that take d at the nominal clock
func cyclesFor(d time.Duration) uint64 {
	clockMhz := CPUClockMhz
	return uint64(math.Ceil(float64(d.Microseconds()) * clockMhz))
}

func TestParseStartTime(t *testing.T) {
	expected := time.Date(1985, 3, 15, 10, 30, 0, 0, time.UTC)
	for _, value := range []string{"1985-03-15T10:30:00", "1985-03-15T10:30", "1985-03-15T10:30:00Z"} {
		got, err := parseStartTime(value)
		if err != nil {
			t.Fatalf("%s: %v", value, err)
		}
		if !got.Equal(expected) {
			t.Errorf("%s: expected %v, got %v", value, expected, got)
		}
	}

	got, err := parseStartTime("1985-03-15T10:30:00+01:00")
	if err != nil {
		t.Fatal(err)
	}
	if got.Hour() != 10 || !got.Equal(expected.Add(-time.Hour)) {
		t.Errorf("the zone given is not kept, got %v", got)
	}

	_, err = parseStartTime("15/03/1985")
	if err == nil {
		t.Error("an invalid start time is accepted")
	}
}

func TestNowWithoutStartTime(t *testing.T) {
	var a Apple2
	if err := a.setStartTime("none"); err != nil {
		t.Fatal(err)
	}
	if a.HasStartTime() || a.random != nil {
		t.Fatal("'none' sets a start time")
	}
	a.cycles = 1_000_000_000
	if time.Since(a.Now()).Abs() > time.Minute {
		t.Errorf("without a start time Now is not the host time, got %v", a.Now())
	}
}

func TestNowWithStartTime(t *testing.T) {
	var a Apple2
	if err := a.setStartTime("1985-03-15T10:30:00"); err != nil {
		t.Fatal(err)
	}

	if a.Now() != a.startTime {
		t.Errorf("expected the start time on cycle 0, got %v", a.Now())
	}

	// A minute of cycles at the nominal clock
	a.cycles = cyclesFor(60 * time.Second)
	expected := time.Date(1985, 3, 15, 10, 31, 0, 0, time.UTC)
	if diff := a.Now().Sub(expected).Abs(); diff > time.Microsecond {
		t.Errorf("expected %v, got %v", expected, a.Now())
	}

	// The speed of the emulation doesn't change the time of the machine
	a.cycleDurationNs = 0
	if diff := a.Now().Sub(expected).Abs(); diff > time.Microsecond {
		t.Errorf("with full speed, expected %v, got %v", expected, a.Now())
	}
}

func TestBlinkWithStartTime(t *testing.T) {
	var a Apple2
	if err := a.setStartTime("1985-03-15"); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		frame    uint64
		flashed  bool
		fastShow bool
	}{
		{0, false, false},
		{7, false, false},
		{8, false, true},
		{15, false, true},
		{16, true, false},
		{24, true, true},
		{31, true, true},
		{32, false, false},
	}
	for _, c := range cases {
		a.cycles = c.frame*screenDrawCycles + 1
		if a.IsFlashedFrame() != c.flashed {
			t.Errorf("frame %v: expected flashed %v", c.frame, c.flashed)
		}
		if a.isBlinkOn(16, time.Second/2) != c.fastShow {
			t.Errorf("frame %v: expected fast blink %v", c.frame, c.fastShow)
		}
	}
}

func TestRandomSeededByStartTime(t *testing.T) {
	sequence := func(startTime string) []int {
		var a Apple2
		if err := a.setStartTime(startTime); err != nil {
			t.Fatal(err)
		}
		values := make([]int, 20)
		for i := range values {
			values[i] = a.random.Intn(1000)
		}
		return values
	}

	first := sequence("1985-03-15T10:30:00")
	if !slices.Equal(first, sequence("1985-03-15T10:30:00")) {
		t.Error("the same start time gives different random values")
	}
	if slices.Equal(first, sequence("1985-03-15T10:30:01")) {
		t.Error("a different start time gives the same random values")
	}
}

func TestNoSlotClockWithStartTime(t *testing.T) {
	var a Apple2
	if err := a.setStartTime("1985-03-15T10:30:00"); err != nil {
		t.Fatal(err)
	}
	// Five seconds and a half later
	a.cycles = cyclesFor(5500 * time.Millisecond)

	nsc := newNoSlotClockDS1216(&a, nil)
	nsc.loadTime()

	// 85-03-15, a Friday, the 6th day, 10:30:05.50
	expected := uint64(0x8503_1506_1030_0550)
	if nsc.timeCapture != expected {
		t.Errorf("expected %x, got %x", expected, nsc.timeCapture)
	}
}

func TestThunderClockWithStartTime(t *testing.T) {
	overrides := newConfiguration()
	overrides.set(confStartTime, "1985-03-15T10:30:00")
	overrides.set(confS4, "thunderclock")
	at, err := makeApple2Tester("2plus", overrides)
	if err != nil {
		t.Fatal(err)
	}
	card, ok := at.a.cards[4].(*CardThunderClockPlus)
	if !ok {
		t.Fatalf("not a ThunderClock Plus card, %T", at.a.cards[4])
	}
	if card.upd1990.Now == nil || !card.upd1990.Now().Equal(at.a.startTime) {
		t.Error("the ThunderClock Plus doesn't use the time of the machine")
	}
}

// runWozFor runs a machine booting a WOZ diskette for a number of cycles and
// returns a hash of its main RAM
func runWozFor(t *testing.T, sequencer bool, file string, startTime string, cycles uint64) uint64 {
	overrides := newConfiguration()
	card := "diskii"
	if sequencer {
		card = "diskiiseq"
	}
	overrides.set(confS6, card+",disk1=\"woz_test_images/"+file+"\"")
	overrides.set(confStartTime, startTime)
	at, err := makeApple2Tester("2enh", overrides)
	if err != nil {
		t.Fatal(err)
	}

	at.terminateCondition = func(a *Apple2) bool {
		return a.GetCycles() >= cycles
	}
	at.run()

	h := fnv.New64()
	for address := range 0xc000 {
		h.Write([]byte{at.a.mmu.Peek(uint16(address))})
	}
	return h.Sum64()
}

// Congo Bongo reads the random bits where its diskette has no flux: with the
// time of the host, the memory after its boot is different on every run.
// With a start time it is the same on every run, and different with another
// start time.
func TestWozReproducibleWithStartTime(t *testing.T) {
	const file = "Congo Bongo.woz"
	const cycles = 8_000_000

	for _, sequencer := range []bool{false, true} {
		name := "diskii"
		if sequencer {
			name = "diskiiseq"
		}
		t.Run(name, func(t *testing.T) {
			first := runWozFor(t, sequencer, file, "1985-03-15T10:30:00", cycles)
			again := runWozFor(t, sequencer, file, "1985-03-15T10:30:00", cycles)
			other := runWozFor(t, sequencer, file, "1990-01-01", cycles)
			if first != again {
				t.Error("the memory changes from run to run with the same start time")
			}
			if first == other {
				t.Error("the memory is the same with a different start time")
			}
		})
	}
}
