/*
Package audio synthesizes the sound of the Apple II speaker.

The Apple II speaker is a 1-bit device: each access to $C030 toggles the
speaker cone position. The emulator core reports those accesses as clicks
with the CPU cycle when they happen. This package reconstructs a 48 kHz
float32 audio stream from that stream of click timestamps. It is shared by
the frontends, that only need to send the samples to their audio device.

The synthesis works as follows:

  - A render clock, measured in CPU cycles, advances exactly one sample
    period per sample generated, whether there is sound or not.
  - Each output sample is the average of the speaker position over its
    window of ~21.31 CPU cycles, with the toggles placed at their exact
    fractional position (box filter). This avoids most of the aliasing of
    naive point sampling.
  - The render clock lags the newest received click by a target latency,
    building a cushion against jitter between the emulation and audio
    clocks. If the clicks get too far ahead (fast forward) or fall behind
    (pause, hiccup) the render clock resynchronizes.
  - A DC-blocking high-pass filter models the AC coupling of a real
    speaker: constant levels decay to silence and no artificial decay or
    resting level is needed.
*/
package audio

const (
	// SampleRate is the sample rate of the generated audio stream
	SampleRate = 48000

	// The render clock stays this far behind the newest click, as a
	// cushion against jitter of the emulation clock
	targetLatencyMs = 30

	// Resynchronize if the newest click is further ahead than this
	maxLatencyMs = 200

	// Speaker cone positions, giving half of the full scale peak to peak
	speakerLevel = 0.25

	// Pole of the DC-blocking filter, ~30 Hz cutoff at 48 kHz
	dcFilterPole = 0.996

	clickChannelSize = 10_000
)

// Speaker generates the audio stream of the Apple II speaker. It
// implements the izapple2.SpeakerProvider interface.
type Speaker struct {
	clickChannel  chan uint64
	pendingClicks []uint64

	// Conversions for the configured CPU clock
	cyclesPerSample     float64 // ~21.31 cycles at the standard clock
	targetLatencyCycles float64
	maxLatencyCycles    float64

	synced      bool    // False until the first click arrives
	renderCycle float64 // Render clock, the CPU cycle of the next sample
	level       float64 // Current speaker position

	// DC-blocking filter state
	dcPrevIn  float64
	dcPrevOut float64
}

// NewSpeaker creates a Speaker for a CPU running at clockMhz
func NewSpeaker(clockMhz float64) *Speaker {
	var s Speaker
	s.clickChannel = make(chan uint64, clickChannelSize)
	s.pendingClicks = make([]uint64, 0, clickChannelSize)
	s.cyclesPerSample = 1000 * clockMhz * 1000 / SampleRate
	s.targetLatencyCycles = 1000 * clockMhz * targetLatencyMs
	s.maxLatencyCycles = 1000 * clockMhz * maxLatencyMs
	s.level = speakerLevel
	return &s
}

// Click receives a speaker click. The argument is the CPU cycle when it is
// generated. It can be called from any goroutine.
func (s *Speaker) Click(cycle uint64) {
	select {
	case s.clickChannel <- cycle:
	default:
		// The channel is full, the click is lost
	}
}

// ReadSamples fills buf with the next samples of the audio stream. Mono,
// values in [-1.0, 1.0]. It must always be called from the same goroutine
// or audio callback thread.
func (s *Speaker) ReadSamples(buf []float32) {
	// Take the queued clicks
	for done := false; !done; {
		select {
		case cycle := <-s.clickChannel:
			s.pendingClicks = append(s.pendingClicks, cycle)
		default:
			done = true
		}
	}

	// Synchronize the render clock with the click timestamps
	if len(s.pendingClicks) > 0 {
		oldest := float64(s.pendingClicks[0])
		newest := float64(s.pendingClicks[len(s.pendingClicks)-1])
		if !s.synced || oldest < s.renderCycle || newest > s.renderCycle+s.maxLatencyCycles {
			s.renderCycle = newest - s.targetLatencyCycles
			s.synced = true
		}
	}

	if !s.synced {
		// No clicks received yet, keep silent
		for i := range buf {
			buf[i] = 0
		}
		return
	}

	consumed := 0
	for i := range buf {
		windowEnd := s.renderCycle + s.cyclesPerSample

		// Average the speaker position over the sample window, toggling
		// at the exact click positions
		acc := 0.0
		pos := s.renderCycle
		for consumed < len(s.pendingClicks) {
			click := float64(s.pendingClicks[consumed])
			if click >= windowEnd {
				break
			}
			if click > pos {
				acc += s.level * (click - pos)
				pos = click
			}
			s.level = -s.level
			consumed++
		}
		acc += s.level * (windowEnd - pos)
		sample := acc / s.cyclesPerSample

		// DC-blocking filter
		out := sample - s.dcPrevIn + dcFilterPole*s.dcPrevOut
		s.dcPrevIn = sample
		s.dcPrevOut = out

		buf[i] = float32(out)
		s.renderCycle = windowEnd
	}

	// Keep the clicks not rendered yet for the next call
	s.pendingClicks = s.pendingClicks[:copy(s.pendingClicks, s.pendingClicks[consumed:])]
}
