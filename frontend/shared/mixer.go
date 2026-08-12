package shared

import (
	"github.com/ivanizag/izapple2"
	"github.com/ivanizag/izapple2/audio"
)

// NewFrontMixer creates the mixer of a machine with every sound source of it
// attached, ready for the frontend to send the samples to its audio device
func NewFrontMixer(a *izapple2.Apple2) *audio.Mixer {
	mixer := audio.NewMixer(a.GetClockMhz())
	for _, source := range a.GetAudioSources() {
		source.SetAudioSink(mixer.NewSource())
	}

	return mixer
}
