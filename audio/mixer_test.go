package audio

import (
	"testing"
)

func waveAmplitude(buf []float32) float32 {
	peak := float32(0)
	for _, v := range buf {
		if v > peak {
			peak = v
		}
	}
	return peak
}

func TestMixerSumsSources(t *testing.T) {
	// The same square wave on one speaker and on two in-phase speakers
	single := newWaveDriver(1_000_000, 1)
	double := newWaveDriver(1_000_000, 2)

	singlePeak := waveAmplitude(single.run(512, 20))
	doublePeak := waveAmplitude(double.run(512, 20))

	ratio := doublePeak / singlePeak
	if ratio < 1.8 || ratio > 2.2 {
		t.Errorf("expected two in-phase sources to double the amplitude, got a ratio of %v", ratio)
	}
}

func TestMixerSilentSourceDoesNotBlock(t *testing.T) {
	// An extra source with no events must not prevent the others from
	// playing
	d := newWaveDriver(1_000_000, 1)
	d.m.NewSource()

	out := d.run(512, 20)
	if waveAmplitude(out) < 0.1 {
		t.Error("expected the speaker to play with a silent source attached")
	}
}

func TestMixerSourceHoldsLevel(t *testing.T) {
	// A source that reports one level change holds it, and the DC filter
	// slowly brings the output back to zero
	m := NewMixer(testClockMhz)
	s := m.NewSource()

	s.PushLevel(1_000_000, 0.5)
	buf := make([]float32, SampleRate) // One second
	m.ReadSamples(buf)

	if waveAmplitude(buf) < 0.4 {
		t.Error("expected the level step to reach the output")
	}
	tail := buf[len(buf)-100:]
	for _, v := range tail {
		if v > 0.01 || v < -0.01 {
			t.Fatalf("expected the DC filter to decay a constant level, got %v", v)
		}
	}
}
