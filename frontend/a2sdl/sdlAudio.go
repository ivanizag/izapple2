package main

/*
typedef unsigned char Uint8;
void SpeakerCallback(void *userdata, Uint8 *stream, int len);
*/
import "C"
import (
	"fmt"
	"sync/atomic"
	"unsafe"

	"github.com/ivanizag/izapple2/audio"
	"github.com/veandco/go-sdl2/sdl"
)

// Samples per SDL audio buffer, ~21 ms
const bufferSamples = 1024

// sdlAudio sends the mixed audio of the machine to the SDL audio device.
// The audio sources of the machine attach to the mixer.
type sdlAudio struct {
	mixer *audio.Mixer
}

/*
I have not found a way to encode the pointer to sdlAudio on the userdata of
the call to SpeakerCallback(). I use a global as workaround... It is atomic
because the callback runs on the SDL audio thread.
*/
var theSDLAudio atomic.Pointer[sdlAudio]

func newSDLAudio(clockMhz float64) *sdlAudio {
	return &sdlAudio{
		mixer: audio.NewMixer(clockMhz),
	}
}

// SpeakerCallback is called to get more sound buffer data
//
//export SpeakerCallback
func SpeakerCallback(userdata unsafe.Pointer, stream *C.Uint8, length C.int) {
	// Adapt the C buffer to a slice of float32 samples
	buf := unsafe.Slice((*float32)(unsafe.Pointer(stream)), int(length)/4)

	s := theSDLAudio.Load()
	if s == nil {
		// SDL does not guarantee the buffer to be initialized
		for i := range buf {
			buf[i] = 0
		}
		return
	}

	s.mixer.ReadSamples(buf)
}

func (s *sdlAudio) start() {
	err := sdl.Init(sdl.INIT_AUDIO)
	if err != nil {
		fmt.Printf("Error starting SDL audio: %v.\n", err)
		return
	}

	spec := &sdl.AudioSpec{
		Freq:     audio.SampleRate,
		Format:   sdl.AUDIO_F32SYS,
		Channels: 1,
		Samples:  bufferSamples,
		Callback: sdl.AudioCallback(C.SpeakerCallback),
	}

	if err := sdl.OpenAudio(spec, nil); err != nil {
		fmt.Printf("Error opening the SDL audio channel: %v.\n", err)
		return
	}
	theSDLAudio.Store(s)
	sdl.PauseAudio(false)
}

func (s *sdlAudio) close() {
	sdl.CloseAudio()
	sdl.Quit()
}
