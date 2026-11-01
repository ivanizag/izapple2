package izapple2

import (
	"fmt"
	"math/rand"
	"time"
)

/*
The machine reads the time of the host by default: the clock cards give the
date of the host, and the cursor and the flashing characters blink with the
host clock, so they look right when the emulation runs faster than the real
machine.

With a start time, the machine runs on a time of its own, the start time
plus the cycles run. The clocks give that time, the blinks follow the frames
of the emulated machine and the random sources are seeded from the start
time. The same steps make the same results on every run.
*/

// framesPerFlash is the period of the flashing characters and of the slow
// cursor of the 6845 in frames, about half a second
const framesPerFlash = 32

// startTimeLayouts are the formats accepted for the start time. Without a
// zone, the time is taken as UTC, so that it doesn't depend on the host
var startTimeLayouts = []string{
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02T15:04",
	"2006-01-02",
}

func parseStartTime(value string) (time.Time, error) {
	for _, layout := range startTimeLayouts {
		t, err := time.Parse(layout, value)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid start time '%s', use a format like 1985-03-15T10:30:00", value)
}

// setStartTime makes the machine use its own time from startTime. An empty
// value keeps the time of the host.
func (a *Apple2) setStartTime(value string) error {
	if value == "" || value == "none" {
		return nil
	}
	t, err := parseStartTime(value)
	if err != nil {
		return err
	}
	a.startTime = t
	a.random = rand.New(rand.NewSource(t.UnixNano()))
	return nil
}

// HasStartTime returns true when the machine runs on its own time instead
// of the time of the host
func (a *Apple2) HasStartTime() bool {
	return !a.startTime.IsZero()
}

// Now returns the time of the machine: the start time plus the time of the
// cycles run at the nominal clock, or the time of the host without a start
// time. The speed of the emulation doesn't change it.
func (a *Apple2) Now() time.Time {
	if !a.HasStartTime() {
		return time.Now()
	}
	elapsed := time.Duration(float64(a.cycles) * 1000.0 / CPUClockMhz)
	return a.startTime.Add(elapsed)
}

// isBlinkOn tells if something that blinks is in the second half of its
// period: counted in frames of the machine with a start time, with the
// host clock otherwise.
func (a *Apple2) isBlinkOn(periodFrames uint64, hostPeriod time.Duration) bool {
	if a.HasStartTime() {
		frame := a.cycles / screenDrawCycles
		return frame%periodFrames >= periodFrames/2
	}
	return time.Now().UnixNano()%int64(hostPeriod) >= int64(hostPeriod)/2
}

// IsFlashedFrame tells if the flashing characters are shown inverse now
func (a *Apple2) IsFlashedFrame() bool {
	return a.isBlinkOn(framesPerFlash, time.Second)
}
