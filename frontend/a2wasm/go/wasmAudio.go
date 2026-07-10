//go:build js

package main

import (
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"

	a2audio "github.com/ivanizag/izapple2/audio"
)

// wasmAudio sends the mixed audio of the machine to the ebiten audio
// player. The audio sources of the machine attach to the mixer.
type wasmAudio struct {
	mixer *a2audio.Mixer

	audioContext *audio.Context
	audioPlayer  *audio.Player
	samples      []float32
}

func newWasmAudio(clockMhz float64) *wasmAudio {
	return &wasmAudio{
		mixer: a2audio.NewMixer(clockMhz),
	}
}

// Read is io.Reader's Read, it fills the buffer with audio samples
func (s *wasmAudio) Read(buf []byte) (n int, err error) {
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

func (s *wasmAudio) update() error {
	if s.audioContext == nil {
		s.audioContext = audio.NewContext(a2audio.SampleRate)
	}
	if s.audioPlayer == nil {
		var err error
		s.audioPlayer, err = s.audioContext.NewPlayerF32(s)
		if err != nil {
			return err
		}
		// Increased buffer size for web browsers (150ms vs 100ms native)
		s.audioPlayer.SetBufferSize(time.Duration(150) * time.Millisecond)
		s.audioPlayer.Play()
	}
	return nil
}
