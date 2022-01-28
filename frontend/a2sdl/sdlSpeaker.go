package main

/*
typedef unsigned char Uint8;
void SpeakerCallback(void *userdata, Uint8 *stream, int len);
*/
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/ivanizag/izapple2"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	samplingHz = 48000
	bufferSize = 1000
	// bufferSize/samplingHz will be the max delay of the sound
	sampleDurationCycles = 1000000 * izapple2.CPUClockMhz / samplingHz
	// each sample on the sound stream is 21.31 cpu cycles approx
	maxOutOfSyncMs = 2000
	decayLevel     = 128
)

type sdlSpeaker struct {
	clickChannel  chan uint64
	pendingClicks []uint64
	lastCycle     uint64
	lastState     bool
	lastLevel     C.Uint8
}

/*
I have not found a way to encode the pointer to sdlSpeaker on the userdata of
the call to SpeakerCallback(). I use a global as workaround...
*/
var theSDLSpeaker *sdlSpeaker

func newSDLSpeaker() *sdlSpeaker {
	var s sdlSpeaker
	s.clickChannel = make(chan uint64, bufferSize)
	s.pendingClicks = make([]uint64, 0, bufferSize)
	s.lastLevel = decayLevel // Mid position to avoid starting clicks.
	return &s
}

// Click receives a speaker click. The argument is the CPU cycle when it is generated
func (s *sdlSpeaker) Click(cycle uint64) {
	select {
	case s.clickChannel <- cycle:
		// Sent
	default:
		// The channel is full, the click is lost.
	}
}

func stateToLevel(state bool) C.Uint8 {
	if state {
		return 255
	}
	return 0
}

// SpeakerCallback is called to get more sound buffer data
//export SpeakerCallback
func SpeakerCallback(userdata unsafe.Pointer, stream *C.Uint8, length C.int) {
	s := theSDLSpeaker
	if s == nil {
		return
	}

	// Adapt C buffer
	buf := unsafe.Slice(stream, length)

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
		}
	}

	// Build wave
	var i, p int
	level := s.lastLevel
	for p = 0; p < len(s.pendingClicks); p++ {
		cycle := s.pendingClicks[p]
		if cycle < s.lastCycle {
			// Too old, ignore
			continue
		}

		// Fill with samples
		samplesNeeded := int(float64(cycle-s.lastCycle) / sampleDurationCycles)
		if samplesNeeded+i > bufferSize {
			samplesNeeded = bufferSize - i
		}
		for j := 0; j < samplesNeeded; j++ {
			buf[i] = level
			i++
		}

		// Update state
		s.lastCycle = cycle
		s.lastState = !s.lastState
		level = stateToLevel(s.lastState)

		if i == bufferSize {
			// Buffer is complete
			break
		}
	}

	// If the buffer is empty lets decay the signal
	if i == 0 {
		for level != decayLevel && i < bufferSize {
			if i%100 == 0 {
				if level > decayLevel {
					level--
				} else {
					level++
				}
			}
			buf[i] = level
			i++
		}
	}

	// Complete the buffer if needed
	for b := i; b < bufferSize; b++ {
		buf[b] = level
	}
	s.lastLevel = level

	// Remove processed clicks, store the rest for later
	remainingClicks := len(s.pendingClicks) - p
	for r := 0; r < remainingClicks; r++ {
		s.pendingClicks[r] = s.pendingClicks[p+r]
	}
	s.pendingClicks = s.pendingClicks[0:remainingClicks]
}

func (s *sdlSpeaker) start() {
	err := sdl.Init(sdl.INIT_AUDIO)
	if err != nil {
		fmt.Printf("Error starting SDL audio: %v.\n", err)
		return
	}

	spec := &sdl.AudioSpec{
		Freq:     samplingHz,
		Format:   sdl.AUDIO_U8,
		Channels: 1,
		Samples:  bufferSize,
		Callback: sdl.AudioCallback(C.SpeakerCallback),
	}

	if err := sdl.OpenAudio(spec, nil); err != nil {
		fmt.Printf("Error opening the SDL audio channel: %v.\n", err)
		return
	}
	sdl.PauseAudio(false)
	theSDLSpeaker = s
}

func (s *sdlSpeaker) close() {
	sdl.CloseAudio()
	sdl.Quit()
}
