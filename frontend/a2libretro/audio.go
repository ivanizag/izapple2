package main

/*
#include "libretro.h"
#include "shim.h"
*/
import "C"

import (
	"github.com/ivanizag/izapple2"
	a2audio "github.com/ivanizag/izapple2/audio"
	"github.com/ivanizag/izapple2/frontend/shared"
)

const sampleRate = a2audio.SampleRate

/*
The mixer needs no special handling here. It keeps its render clock a fixed
30 ms behind the newest sound event, and the core pushes one frame of events
and reads one frame of samples on every call, so the two stay in step and that
lag is just constant latency.
*/
type audioOutput struct {
	mixer *a2audio.Mixer

	samplesPerFrame float64
	// Samples owed to the frontend, as the samples per frame are not a whole
	// number
	pending float64

	mono   []float32
	stereo []C.int16_t
}

func newAudioOutput(a *izapple2.Apple2) *audioOutput {
	samplesPerFrame := sampleRate * cyclesPerFrame / (a.GetClockMhz() * 1_000_000)

	// Room for a frame and a bit, the pending samples never add up to a full
	// extra sample
	size := int(samplesPerFrame) + 2

	return &audioOutput{
		mixer:           shared.NewMixer(a),
		samplesPerFrame: samplesPerFrame,
		mono:            make([]float32, size),
		stereo:          make([]C.int16_t, 2*size),
	}
}

// render sends the samples of the frame to the frontend as interleaved stereo
func (s *audioOutput) render() {
	s.pending += s.samplesPerFrame
	samples := int(s.pending)
	s.pending -= float64(samples)
	if samples == 0 {
		return
	}

	s.mixer.ReadSamples(s.mono[:samples])
	for i, v := range s.mono[:samples] {
		sample := toInt16(v)
		s.stereo[2*i] = sample
		s.stereo[2*i+1] = sample
	}

	C.shim_audio_sample_batch(&s.stereo[0], C.size_t(samples))
}

func toInt16(v float32) C.int16_t {
	scaled := int32(v * 32767)
	if scaled > 32767 {
		scaled = 32767
	} else if scaled < -32768 {
		scaled = -32768
	}
	return C.int16_t(scaled)
}
