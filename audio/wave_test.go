package audio

import (
	"math"
	"testing"
)

// The standard NTSC Apple II clock
const testClockMhz = 14.318 / 14

// testToggler pushes speaker-like level toggles to a source
type testToggler struct {
	source *Source
	high   bool
}

const testSpeakerLevel = 0.25

func (tg *testToggler) click(cycle uint64) {
	tg.high = !tg.high
	if tg.high {
		tg.source.PushLevel(cycle, testSpeakerLevel)
	} else {
		tg.source.PushLevel(cycle, -testSpeakerLevel)
	}
}

func TestSilenceBeforeFirstEvent(t *testing.T) {
	m := NewMixer(testClockMhz)
	m.NewSource()
	buf := make([]float32, 1024)
	buf[10] = 42.0 // Ensure the buffer is overwritten
	m.ReadSamples(buf)
	for i, v := range buf {
		if v != 0 {
			t.Fatalf("expected silence before the first event, got %v at sample %v", v, i)
		}
	}
}

// waveDriver simulates the emulation and the audio device running in
// parallel: on each step it queues the clicks of a square wave for the
// next chunk of emulated time and reads the equivalent audio samples.
type waveDriver struct {
	m         *Mixer
	togglers  []*testToggler
	nextClick uint64
	cycle     uint64
}

const driverChunkSamples = 480 // 10 ms

func newWaveDriver(startCycle uint64, togglers int) *waveDriver {
	d := &waveDriver{
		m:         NewMixer(testClockMhz),
		nextClick: startCycle,
		cycle:     startCycle,
	}
	for range togglers {
		d.togglers = append(d.togglers, &testToggler{source: d.m.NewSource()})
	}
	return d
}

// step advances 10 ms, toggling every halfPeriodCycles, 0 for silence
func (d *waveDriver) step(halfPeriodCycles uint64) []float32 {
	chunkCycles := float64(driverChunkSamples) * d.m.cyclesPerSample
	d.cycle += uint64(chunkCycles)
	if halfPeriodCycles != 0 {
		for d.nextClick < d.cycle {
			for _, tg := range d.togglers {
				tg.click(d.nextClick)
			}
			d.nextClick += halfPeriodCycles
		}
	} else {
		d.nextClick = d.cycle
	}
	buf := make([]float32, driverChunkSamples)
	d.m.ReadSamples(buf)
	return buf
}

func (d *waveDriver) run(halfPeriodCycles uint64, steps int) []float32 {
	out := make([]float32, 0, steps*driverChunkSamples)
	for range steps {
		out = append(out, d.step(halfPeriodCycles)...)
	}
	return out
}

func TestSquareWaveFrequency(t *testing.T) {
	// A square wave toggling every 512 cycles, ~977 Hz, for 400 ms
	d := newWaveDriver(1_000_000, 1)
	out := d.run(512, 40)

	// Check the average period on the second half, once in steady state
	steady := out[len(out)/2:]
	risings := make([]int, 0, 200)
	for i := 1; i < len(steady); i++ {
		if steady[i-1] <= 0 && steady[i] > 0 {
			risings = append(risings, i)
		}
	}
	if len(risings) < 50 {
		t.Fatalf("expected a sustained square wave, got %v rising edges", len(risings))
	}

	gotPeriod := float64(risings[len(risings)-1]-risings[0]) / float64(len(risings)-1)
	wantPeriod := 2 * 512 / d.m.cyclesPerSample // ~48 samples
	if math.Abs(gotPeriod-wantPeriod) > 0.1 {
		t.Errorf("expected a period of %.2f samples, got %.2f", wantPeriod, gotPeriod)
	}
}

func TestAmplitudeAndDCRemoval(t *testing.T) {
	d := newWaveDriver(1_000_000, 1)
	out := d.run(512, 40)

	sum := 0.0
	for i, v := range out {
		if v > 1.0 || v < -1.0 {
			t.Fatalf("sample %v out of range: %v", i, v)
		}
		sum += float64(v)
	}
	mean := sum / float64(len(out))
	if math.Abs(mean) > 0.01 {
		t.Errorf("expected the DC filter to remove the offset, got a mean of %v", mean)
	}
}

func TestSilenceDecaysAfterSound(t *testing.T) {
	d := newWaveDriver(1_000_000, 1)
	d.run(512, 20)

	// A second of silence
	out := d.run(0, 100)

	tail := out[len(out)-100:]
	for i, v := range tail {
		if math.Abs(float64(v)) > 0.001 {
			t.Fatalf("expected silence to decay to zero, got %v at tail sample %v", v, i)
		}
	}
}

func TestSoundResumesAfterPause(t *testing.T) {
	d := newWaveDriver(1_000_000, 1)
	d.run(512, 20)
	d.run(0, 100) // A second of silence

	// The sound must resume promptly, not delayed by the pause length
	out := d.run(512, 6)
	found := false
	for _, v := range out {
		if math.Abs(float64(v)) > 0.05 {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected the sound to resume within 60 ms after a pause")
	}
}

func TestResyncAfterFastForward(t *testing.T) {
	d := newWaveDriver(1_000_000, 1)
	d.run(512, 20)

	// The emulation fast-forwards ten seconds ahead of the audio clock
	d.cycle += 10_000_000
	d.nextClick = d.cycle

	out := d.run(512, 6)
	found := false
	for _, v := range out {
		if math.Abs(float64(v)) > 0.05 {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected the sound to resume promptly after a fast forward")
	}
}
