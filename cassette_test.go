package izapple2

import (
	"encoding/binary"
	"testing"
)

/*
Tape recording generator for the tests, using the Monitor ROM
encoding: a 770 Hz header tone, a short sync cycle, the data bytes
MSB first with 2 kHz cycles for "0" bits and 1 kHz cycles for "1"
bits, and a final checksum byte.
*/

const testTapeSampleRate = 44100

type tapeRecorder struct {
	samples []float64
	high    bool
	carry   float64 // Fractional sample carried to the next half cycle
}

func (r *tapeRecorder) halfCycle(microseconds float64) {
	exact := microseconds*1e-6*testTapeSampleRate + r.carry
	count := int(exact)
	r.carry = exact - float64(count)

	level := 0.6
	if !r.high {
		level = -0.6
	}
	for range count {
		r.samples = append(r.samples, level)
	}
	r.high = !r.high
}

func (r *tapeRecorder) tone(hz float64, seconds float64) {
	halfCycles := int(seconds * hz * 2)
	for range halfCycles {
		r.halfCycle(500_000 / hz)
	}
}

func (r *tapeRecorder) bit(value bool) {
	duration := 250.0 // A "0" is a full 2 kHz cycle
	if value {
		duration = 500.0 // A "1" is a full 1 kHz cycle
	}
	r.halfCycle(duration)
	r.halfCycle(duration)
}

func (r *tapeRecorder) byte(value uint8) {
	for i := 7; i >= 0; i-- {
		r.bit(value&(1<<i) != 0)
	}
}

func (r *tapeRecorder) record(data []uint8) {
	r.tone(770, 4.0) // Header, 10 seconds on real tapes
	r.halfCycle(200) // Sync
	r.halfCycle(250)
	checksum := uint8(0xff)
	for _, value := range data {
		r.byte(value)
		checksum ^= value
	}
	r.byte(checksum)
	r.tone(770, 0.1) // Trailing tone
}

func (r *tapeRecorder) halfCycles(data []uint8) int {
	// One zero crossing per half cycle recorded
	headers := int(4.0*770*2) + int(0.1*770*2)
	sync := 2
	bits := (len(data) + 1) * 8 * 2
	return headers + sync + bits
}

func buildWavFile(channels int, bitsPerSample int, sampleRate int, data []uint8) []uint8 {
	blockAlign := channels * bitsPerSample / 8
	wav := make([]uint8, 0, 44+len(data))
	wav = append(wav, "RIFF"...)
	wav = binary.LittleEndian.AppendUint32(wav, uint32(36+len(data)))
	wav = append(wav, "WAVE"...)
	wav = append(wav, "fmt "...)
	wav = binary.LittleEndian.AppendUint32(wav, 16)
	wav = binary.LittleEndian.AppendUint16(wav, 1) // PCM format
	wav = binary.LittleEndian.AppendUint16(wav, uint16(channels))
	wav = binary.LittleEndian.AppendUint32(wav, uint32(sampleRate))
	wav = binary.LittleEndian.AppendUint32(wav, uint32(sampleRate*blockAlign))
	wav = binary.LittleEndian.AppendUint16(wav, uint16(blockAlign))
	wav = binary.LittleEndian.AppendUint16(wav, uint16(bitsPerSample))
	wav = append(wav, "data"...)
	wav = binary.LittleEndian.AppendUint32(wav, uint32(len(data)))
	wav = append(wav, data...)
	return wav
}

func wavBytes8BitMono(samples []float64) []uint8 {
	data := make([]uint8, len(samples))
	for i, sample := range samples {
		data[i] = uint8(128 + sample*127)
	}
	return buildWavFile(1, 8, testTapeSampleRate, data)
}

// wavBytes16BitStereo puts the signal on one channel, leaving the other silent
func wavBytes16BitStereo(samples []float64, channel int) []uint8 {
	data := make([]uint8, len(samples)*4)
	for i, sample := range samples {
		value := int16(sample * 32000)
		binary.LittleEndian.PutUint16(data[i*4+channel*2:], uint16(value))
	}
	return buildWavFile(2, 16, testTapeSampleRate, data)
}

func testTapeTransitions(t *testing.T, wav []uint8, expected int) {
	var a Apple2
	c, err := newCassette(&a, wav)
	if err != nil {
		t.Fatal(err)
	}

	tolerance := 2
	if len(c.transitions) < expected-tolerance || len(c.transitions) > expected+tolerance {
		t.Errorf("Expected %v transitions, got %v", expected, len(c.transitions))
	}
}

func TestCassetteTransitions8BitMono(t *testing.T) {
	var recorder tapeRecorder
	data := []uint8{0x00, 0xff, 0xa5, 0x5a}
	recorder.record(data)
	testTapeTransitions(t, wavBytes8BitMono(recorder.samples), recorder.halfCycles(data))
}

func TestCassetteTransitions16BitStereo(t *testing.T) {
	var recorder tapeRecorder
	data := []uint8{0x00, 0xff, 0xa5, 0x5a}
	recorder.record(data)
	testTapeTransitions(t, wavBytes16BitStereo(recorder.samples, 1), recorder.halfCycles(data))
}

func TestCassetteInvalidFile(t *testing.T) {
	var a Apple2
	_, err := newCassette(&a, []uint8{0x01, 0x02, 0x03, 0x04})
	if err == nil {
		t.Error("Expected an error loading a file that is not a WAV")
	}
}
