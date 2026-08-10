//go:build !js

package main

import (
	"fmt"
	"strconv"
	"unsafe"

	"github.com/ivanizag/izapple2"
	"github.com/ivanizag/izapple2/audio"
	"github.com/ivanizag/izapple2/frontend/shared"

	"github.com/Zyko0/go-sdl3/sdl"
)

// Samples per SDL audio buffer, ~21 ms
const bufferSamples = 1024

// sdl3Audio sends the mixed audio of the machine to the SDL audio device.
// The audio sources of the machine attach to the mixer.
type sdl3Audio struct {
	mixer  *audio.Mixer
	stream *sdl.AudioStream
	buf    []float32
}

func newSDL3Audio(a *izapple2.Apple2) *sdl3Audio {
	return &sdl3Audio{
		mixer: shared.NewMixer(a),
		buf:   make([]float32, bufferSamples),
	}
}

// feed is called on the SDL audio thread to get more sound buffer data.
// Unlike SDL2, SDL3 asks for the samples instead of handing over a buffer
// to fill, so the pointer to the sdl3Audio can be captured in the closure.
func (s *sdl3Audio) feed(stream *sdl.AudioStream, additionalAmount int32, _ int32) {
	if additionalAmount <= 0 {
		return
	}

	samples := int(additionalAmount) / 4
	if samples > cap(s.buf) {
		s.buf = make([]float32, samples)
	}
	buf := s.buf[:samples]
	s.mixer.ReadSamples(buf)

	// Adapt the slice of float32 samples to the byte buffer SDL expects
	bytes := unsafe.Slice((*byte)(unsafe.Pointer(unsafe.SliceData(buf))), samples*4)
	err := stream.PutData(bytes)
	if err != nil {
		fmt.Printf("Error sending audio samples: %v.\n", err)
	}
}

func (s *sdl3Audio) start() {
	// SDL3 has no buffer size in the audio spec, it is a hint now
	sdl.SetHint(sdl.HINT_AUDIO_DEVICE_SAMPLE_FRAMES, strconv.Itoa(bufferSamples))

	spec := &sdl.AudioSpec{
		Format:   sdl.AUDIO_F32,
		Channels: 1,
		Freq:     audio.SampleRate,
	}

	s.stream = sdl.AUDIO_DEVICE_DEFAULT_PLAYBACK.OpenAudioDeviceStream(spec,
		sdl.NewAudioStreamCallback(s.feed))
	if s.stream == nil {
		fmt.Print("Error opening the SDL audio channel.\n")
		return
	}

	err := s.stream.ResumeDevice()
	if err != nil {
		fmt.Printf("Error starting SDL audio: %v.\n", err)
	}
}

func (s *sdl3Audio) close() {
	if s.stream != nil {
		s.stream.Destroy()
		s.stream = nil
	}
}
