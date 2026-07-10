package main

import (
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"

	a2audio "github.com/ivanizag/izapple2/audio"
)

// ebitenAudio sends the mixed audio of the machine to the ebiten audio
// player. The audio sources of the machine attach to the mixer.
type ebitenAudio struct {
	mixer *a2audio.Mixer

	audioContext *audio.Context
	audioPlayer  *audio.Player
	samples      []float32
}

func newEbitenAudio(clockMhz float64) *ebitenAudio {
	return &ebitenAudio{
		mixer: a2audio.NewMixer(clockMhz),
	}
}

// Read is io.Reader's Read, it fills the buffer with audio samples
func (s *ebitenAudio) Read(buf []byte) (n int, err error) {
	const bytesPerSample = 8 // Two float32, one for each channel
	samples := len(buf) / bytesPerSample

	if cap(s.samples) < samples {
		s.samples = make([]float32, samples)
	}
	s.samples = s.samples[:samples]
	s.mixer.ReadSamples(s.samples)

	for i, v := range s.samples {
		putFloat32InBuffer(buf, i, v)
	}
	return samples * bytesPerSample, nil
}

func putFloat32InBuffer(buf []byte, i int, f float32) {
	v := math.Float32bits(f)
	buf[i*8] = byte(v)
	buf[i*8+1] = byte(v >> 8)
	buf[i*8+2] = byte(v >> 16)
	buf[i*8+3] = byte(v >> 24)
	buf[i*8+4] = byte(v)
	buf[i*8+5] = byte(v >> 8)
	buf[i*8+6] = byte(v >> 16)
	buf[i*8+7] = byte(v >> 24)
}

func (s *ebitenAudio) update() error {
	if s.audioContext == nil {
		s.audioContext = audio.NewContext(a2audio.SampleRate)
	}
	if s.audioPlayer == nil {
		var err error
		s.audioPlayer, err = s.audioContext.NewPlayerF32(s)
		if err != nil {
			return err
		}
		s.audioPlayer.SetBufferSize(time.Duration(100) * time.Millisecond)
		s.audioPlayer.Play()
	}
	return nil
}
