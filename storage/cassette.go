package storage

import (
	"fmt"
	"math"
)

/*
Loader of cassette tape recordings.

The Apple II cassette input circuit is a comparator that detects the
zero crossings of the audio signal coming from the tape player. The
tape is converted to the list of CPU cycles, from the start of the
tape, at which the signal crosses zero.

The audio file is cleaned up on load: a high pass filter removes any
DC offset and the zero crossings are detected with hysteresis
relative to the signal envelope to ignore the noise of old
recordings.
*/

const (
	// Threshold for a transition relative to the signal envelope
	tapeHysteresisRatio = 0.1
	// Minimum threshold for a transition, relative to full scale, to ignore noise on silence
	tapeHysteresisFloor = 0.02
	// High pass filter cutoff in Hz to remove the DC offset
	tapeHighPassHz = 20.0
	// Time constant in seconds for the envelope follower
	tapeEnvelopeSeconds = 0.020
)

// MakeTape loads a WAV recording of a cassette tape and returns the
// zero crossings of the audio signal as CPU cycles from the start of
// the recording
func MakeTape(data []uint8, cyclesPerSecond float64) ([]uint64, error) {
	wav, err := parseWavFile(data)
	if err != nil {
		return nil, err
	}

	transitions := extractTransitions(wav, cyclesPerSecond)
	if len(transitions) == 0 {
		return nil, fmt.Errorf("no signal found in the tape recording")
	}
	return transitions, nil
}

func extractTransitions(wav *wavFile, cyclesPerSecond float64) []uint64 {
	channel := wav.bestChannel()
	frames := wav.frames()
	dt := 1.0 / float64(wav.sampleRate)
	cyclesPerFrame := cyclesPerSecond / float64(wav.sampleRate)

	// One pole high pass filter coefficient
	rc := 1.0 / (2 * math.Pi * tapeHighPassHz)
	alpha := rc / (rc + dt)
	// Envelope follower decay per sample
	decay := math.Exp(-dt / tapeEnvelopeSeconds)

	transitions := make([]uint64, 0, 1000)
	prevX := 0.0
	prevY := 0.0
	envelope := 0.0
	high := false
	lastCross := 0
	for frame := 0; frame < frames; frame++ {
		x := wav.sample(frame, channel)
		y := alpha * (prevY + x - prevX)
		if (y > 0) != (prevY > 0) {
			lastCross = frame
		}
		prevX = x
		prevY = y

		envelope = max(math.Abs(y), envelope*decay)
		threshold := max(envelope*tapeHysteresisRatio, tapeHysteresisFloor)
		if (!high && y > threshold) || (high && y < -threshold) {
			high = !high
			// The transition is timed at the last zero crossing, not at
			// the threshold crossing, as the comparator on a real Apple
			// II flips right at the zero crossing. The delay reaching
			// the threshold would shorten the next half-cycle
			transitions = append(transitions, uint64(float64(lastCross)*cyclesPerFrame))
		}
	}

	return transitions
}
