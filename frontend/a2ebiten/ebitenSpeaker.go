package main

import (
	"fmt"
	"math"
	"time"

	"github.com/ivanizag/izapple2"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const (
	samplingHz = 48000
	//bufferSize = 1000
	// bufferSize/samplingHz will be the max delay of the sound
	sampleDurationCycles = 1000000 * izapple2.CPUClockMhz / samplingHz
	// each sample on the sound stream is 21.31 cpu cycles approx
	maxOutOfSyncMs = 2000
	decayLevel     = 0.0
)

type ebitenSpeaker struct {
	audioContext *audio.Context
	audioPlayer  *audio.Player

	clickChannel  chan uint64
	pendingClicks []uint64
	lastCycle     uint64
	lastState     bool
	lastLevel     float32
}

func newEbitenSpeaker() *ebitenSpeaker {
	var s ebitenSpeaker
	s.clickChannel = make(chan uint64, 1000)
	s.pendingClicks = make([]uint64, 0, 1000)
	s.lastLevel = decayLevel // Mid position to avoid starting clicks.
	return &s
}

// Click receives a speaker click. The argument is the CPU cycle when it is generated
func (s *ebitenSpeaker) Click(cycle uint64) {
	select {
	case s.clickChannel <- cycle:
		// Sent
	default:
		fmt.Printf("Speaker click dropped in channel.\n")
		// The channel is full, the click is lost.
	}
}

func stateToLevel(state bool) float32 {
	if state {
		return 1.0
	}
	return -1.0
}

// Read is io.Reader's Read
func (s *ebitenSpeaker) Read(buf []byte) (n int, err error) {
	//Read queued clicks
	done := false
	for !done {
		select {
		case cycle := <-s.clickChannel:
			s.pendingClicks = append(s.pendingClicks, cycle)
		default:
			done = true
		}
	}

	// Verify that we are not too long behind
	var maxOutOfSyncCyclesFloat = 1000 * izapple2.CPUClockMhz * maxOutOfSyncMs
	var maxOutOfSyncCycles = uint64(maxOutOfSyncCyclesFloat)
	for _, pc := range s.pendingClicks {
		if pc-s.lastCycle > maxOutOfSyncCycles {
			// Fast forward
			s.lastCycle = pc
			fmt.Printf("Speaker fast forward.\n")
		}
	}

	// Build wave
	const bytesPerSample = 8 // Two floats32, 4 bytes each, one for each channel
	samples := len(buf) / bytesPerSample
	//fmt.Printf("smples: %v\n", smples)

	if len(s.pendingClicks) > 0 {
		fmt.Printf("pendingClicks: %v\n", len(s.pendingClicks))
	}
	var i, r int
	level := s.lastLevel
	for p := 0; p < len(s.pendingClicks); p++ {
		cycle := s.pendingClicks[p]
		if cycle < s.lastCycle {
			// Too old, ignore
			continue
		}

		// Fill with samples
		level = stateToLevel(s.lastState)
		samplesNeeded := int(float64(cycle-s.lastCycle) / sampleDurationCycles)
		if samplesNeeded+i > samples {
			// Partial fill, to be completed on the next callback
			samplesNeeded = samples - i
			s.lastCycle = cycle - uint64(float64(samplesNeeded)*sampleDurationCycles)
		} else {
			s.lastCycle = cycle
			s.lastState = !s.lastState
			r++ // Remove this pending click
		}

		for j := 0; j < samplesNeeded; j++ {
			putFloat32InBuffer(buf, i, level)
			i += 1
		}

		if i == samples {
			// Buffer is complete
			break
		}
	}

	// If the buffer is empty lets stop the signal
	if i == 0 && level != 0.0 {
		level = 0.0
		fmt.Printf("Speaker buffer empty, to zero.\n")
	}

	// Complete the buffer if needed
	for b := i; b < samples; b++ {
		putFloat32InBuffer(buf, b, level)
	}
	s.lastLevel = level

	// Remove processed clicks, store the rest for later
	s.pendingClicks = s.pendingClicks[r:]

	return len(buf), nil
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

func (s *ebitenSpeaker) update() error {
	if s.audioContext == nil {
		s.audioContext = audio.NewContext(samplingHz)
	}
	if s.audioPlayer == nil {
		var err error
		s.audioPlayer, err = s.audioContext.NewPlayerF32(s)
		if err != nil {
			return err
		}
		//s.audioPlayer.SetVolume(1.0)
		s.audioPlayer.SetBufferSize(time.Duration(100) * time.Millisecond)
		s.audioPlayer.Play()
	}
	return nil
}
