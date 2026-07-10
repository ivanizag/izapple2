package izapple2

/*
Audio interface between the sound generators and the frontends.

The built-in speaker and the cards that generate sound implement
AudioSource. The frontend attaches an AudioSink to each source and
renders the reported level changes, typically with the audio package
Mixer.
*/

// AudioSink receives the output level changes of a sound source with the
// CPU cycle when they happen. Implemented by the frontends.
type AudioSink interface {
	PushLevel(cycle uint64, level float32)
}

// AudioSource is implemented by the cards that generate sound.
type AudioSource interface {
	GetAudioSourceName() string
	SetAudioSink(sink AudioSink)
}

// GetAudioSources returns the sound generators of the machine: the
// built-in speaker and the sound cards
func (a *Apple2) GetAudioSources() []AudioSource {
	sources := []AudioSource{&a.io.speaker}
	for _, card := range a.cards {
		if source, ok := card.(AudioSource); ok {
			sources = append(sources, source)
		}
	}
	return sources
}

// Position of the speaker cone, giving half of the full scale peak to peak
const speakerLevel = 0.25

// speakerAudioSource is the AudioSource of the built-in speaker. Each
// access to $C030 toggles the position of the speaker cone.
type speakerAudioSource struct {
	sink AudioSink
	high bool // Position of the speaker cone
}

func (s *speakerAudioSource) GetAudioSourceName() string {
	return "speaker"
}

func (s *speakerAudioSource) SetAudioSink(sink AudioSink) {
	s.sink = sink
}

func (s *speakerAudioSource) click(cycle uint64) {
	if s.sink == nil {
		return
	}
	s.high = !s.high
	if s.high {
		s.sink.PushLevel(cycle, speakerLevel)
	} else {
		s.sink.PushLevel(cycle, -speakerLevel)
	}
}
