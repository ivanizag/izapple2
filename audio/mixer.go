/*
Package audio synthesizes the sound of the Apple II.

The emulation core reports the sound activity as events with the CPU
cycle when they happen: speaker clicks and output level changes of the
sound cards. This package reconstructs a mixed 48 kHz float32 mono
stream from those events. It is shared by the frontends, that only need
to send the samples to their audio device.

The synthesis works as follows:

  - The Mixer owns a render clock, measured in CPU cycles, that advances
    exactly one sample period per sample generated, whether there is
    sound or not.
  - Each Source holds the timestamped level changes of one sound
    generator. For every output sample, the level of each source is
    averaged over the sample window of ~21.31 CPU cycles, with the
    changes applied at their exact fractional position (box filter).
    This avoids most of the aliasing of naive point sampling.
  - The render clock lags the newest received event by a target latency,
    building a cushion against jitter between the emulation and audio
    clocks. If the events get too far ahead (fast forward) or fall
    behind (pause, hiccup) the render clock resynchronizes.
  - The sources are summed and a DC-blocking high-pass filter models the
    AC coupling of a real speaker: constant levels decay to silence.
*/
package audio

const (
	// SampleRate is the sample rate of the generated audio stream
	SampleRate = 48000

	// The render clock stays this far behind the newest event, as a
	// cushion against jitter of the emulation clock
	targetLatencyMs = 30

	// Resynchronize if the newest event is further ahead than this
	maxLatencyMs = 200

	// Pole of the DC-blocking filter, ~30 Hz cutoff at 48 kHz
	dcFilterPole = 0.996

	eventChannelSize = 10_000
)

type levelEvent struct {
	cycle uint64
	level float32
}

// Source receives the output level changes of one sound generator. It
// implements the izapple2.AudioSink interface.
type Source struct {
	events   chan levelEvent
	pending  []levelEvent
	consumed int
	level    float64 // Current output level of the generator
}

// PushLevel reports that the generator output changed to level at the
// given CPU cycle. It can be called from any goroutine.
func (s *Source) PushLevel(cycle uint64, level float32) {
	select {
	case s.events <- levelEvent{cycle, level}:
	default:
		// The channel is full, the event is lost
	}
}

// renderWindow returns the average level over the sample window,
// applying the pending events that fall inside it
func (s *Source) renderWindow(start, end float64) float64 {
	acc := 0.0
	pos := start
	for s.consumed < len(s.pending) {
		event := s.pending[s.consumed]
		cycle := float64(event.cycle)
		if cycle >= end {
			break
		}
		if cycle > pos {
			acc += s.level * (cycle - pos)
			pos = cycle
		}
		s.level = float64(event.level)
		s.consumed++
	}
	acc += s.level * (end - pos)
	return acc / (end - start)
}

// Mixer generates the audio stream of the machine, summing the streams
// of its sources.
type Mixer struct {
	sources []*Source

	// Conversions for the configured CPU clock
	cyclesPerSample     float64 // ~21.31 cycles at the standard clock
	targetLatencyCycles float64
	maxLatencyCycles    float64

	synced      bool    // False until the first event arrives
	renderCycle float64 // Render clock, the CPU cycle of the next sample

	// DC-blocking filter state
	dcPrevIn  float64
	dcPrevOut float64
}

// NewMixer creates a Mixer for a CPU running at clockMhz
func NewMixer(clockMhz float64) *Mixer {
	var m Mixer
	m.cyclesPerSample = 1000 * clockMhz * 1000 / SampleRate
	m.targetLatencyCycles = 1000 * clockMhz * targetLatencyMs
	m.maxLatencyCycles = 1000 * clockMhz * maxLatencyMs
	return &m
}

// NewSource adds a sound source to the mix. All the sources must be
// created before the audio device starts calling ReadSamples.
func (m *Mixer) NewSource() *Source {
	s := &Source{
		events:  make(chan levelEvent, eventChannelSize),
		pending: make([]levelEvent, 0, eventChannelSize),
	}
	m.sources = append(m.sources, s)
	return s
}

// ReadSamples fills buf with the next samples of the audio stream. Mono,
// values in [-1.0, 1.0]. It must always be called from the same goroutine
// or audio callback thread.
func (m *Mixer) ReadSamples(buf []float32) {
	// Take the queued events of each source
	someEvents := false
	var oldest, newest float64
	for _, s := range m.sources {
		for done := false; !done; {
			select {
			case event := <-s.events:
				s.pending = append(s.pending, event)
			default:
				done = true
			}
		}
		if len(s.pending) > 0 {
			first := float64(s.pending[0].cycle)
			last := float64(s.pending[len(s.pending)-1].cycle)
			if !someEvents || first < oldest {
				oldest = first
			}
			if !someEvents || last > newest {
				newest = last
			}
			someEvents = true
		}
	}

	// Synchronize the render clock with the event timestamps
	if someEvents {
		if !m.synced || oldest < m.renderCycle || newest > m.renderCycle+m.maxLatencyCycles {
			m.renderCycle = newest - m.targetLatencyCycles
			m.synced = true
		}
	}

	if !m.synced {
		// No events received yet, keep silent
		for i := range buf {
			buf[i] = 0
		}
		return
	}

	for i := range buf {
		windowEnd := m.renderCycle + m.cyclesPerSample
		sample := 0.0
		for _, s := range m.sources {
			sample += s.renderWindow(m.renderCycle, windowEnd)
		}

		// DC-blocking filter
		out := sample - m.dcPrevIn + dcFilterPole*m.dcPrevOut
		m.dcPrevIn = sample
		m.dcPrevOut = out

		buf[i] = float32(out)
		m.renderCycle = windowEnd
	}

	// Keep the events not rendered yet for the next call
	for _, s := range m.sources {
		s.pending = s.pending[:copy(s.pending, s.pending[s.consumed:])]
		s.consumed = 0
	}
}
