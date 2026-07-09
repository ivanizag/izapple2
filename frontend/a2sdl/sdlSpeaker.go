package main

/*
typedef unsigned char Uint8;
void SpeakerCallback(void *userdata, Uint8 *stream, int len);
*/
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/ivanizag/izapple2/audio"
	"github.com/veandco/go-sdl2/sdl"
)

// Samples per SDL audio buffer, ~21 ms
const bufferSamples = 1024

// sdlSpeaker sends the audio from the shared synthesizer to the SDL audio
// device. It implements izapple2.SpeakerProvider.
type sdlSpeaker struct {
	speaker *audio.Speaker
}

/*
I have not found a way to encode the pointer to sdlSpeaker on the userdata of
the call to SpeakerCallback(). I use a global as workaround...
*/
var theSDLSpeaker *sdlSpeaker

func newSDLSpeaker() *sdlSpeaker {
	return &sdlSpeaker{speaker: audio.NewSpeaker()}
}

// Click receives a speaker click. The argument is the CPU cycle when it is generated
func (s *sdlSpeaker) Click(cycle uint64) {
	s.speaker.Click(cycle)
}

// SpeakerCallback is called to get more sound buffer data
//
//export SpeakerCallback
func SpeakerCallback(userdata unsafe.Pointer, stream *C.Uint8, length C.int) {
	s := theSDLSpeaker
	if s == nil {
		return
	}

	// Adapt the C buffer to a slice of float32 samples
	buf := unsafe.Slice((*float32)(unsafe.Pointer(stream)), int(length)/4)
	s.speaker.ReadSamples(buf)
}

func (s *sdlSpeaker) start() {
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
	sdl.PauseAudio(false)
	theSDLSpeaker = s
}

func (s *sdlSpeaker) close() {
	sdl.CloseAudio()
	sdl.Quit()
}
