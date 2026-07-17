package storage

import (
	"encoding/binary"
	"fmt"
	"math"
)

/*
Minimal WAV file decoder to load cassette tape recordings.

Supports PCM integer samples of 8, 16, 24 and 32 bits and 32 bit
IEEE float samples. For multichannel files, the channel with the
most energy is used.

See: http://soundfile.sapp.org/doc/WaveFormat/
*/

const (
	wavFormatPCM        = 1
	wavFormatIEEEFloat  = 3
	wavFormatExtensible = 0xfffe
)

type wavFile struct {
	audioFormat   uint16
	channels      int
	sampleRate    int
	bitsPerSample int
	blockAlign    int
	data          []uint8 // Raw sample frames
}

func parseWavFile(data []uint8) (*wavFile, error) {
	if len(data) < 12 || string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return nil, fmt.Errorf("not a WAV file")
	}

	var w wavFile
	hasFormat := false

	// Iterate the RIFF chunks
	pos := 12
	for pos+8 <= len(data) {
		chunkID := string(data[pos : pos+4])
		chunkSize := int(binary.LittleEndian.Uint32(data[pos+4 : pos+8]))
		chunkStart := pos + 8
		chunkEnd := chunkStart + chunkSize
		if chunkEnd > len(data) {
			chunkEnd = len(data) // Tolerate truncated files
		}
		chunk := data[chunkStart:chunkEnd]

		switch chunkID {
		case "fmt ":
			if len(chunk) < 16 {
				return nil, fmt.Errorf("invalid WAV format chunk")
			}
			w.audioFormat = binary.LittleEndian.Uint16(chunk[0:2])
			w.channels = int(binary.LittleEndian.Uint16(chunk[2:4]))
			w.sampleRate = int(binary.LittleEndian.Uint32(chunk[4:8]))
			w.blockAlign = int(binary.LittleEndian.Uint16(chunk[12:14]))
			w.bitsPerSample = int(binary.LittleEndian.Uint16(chunk[14:16]))
			if w.audioFormat == wavFormatExtensible && len(chunk) >= 26 {
				// The actual format is in the first bytes of the subformat GUID
				w.audioFormat = binary.LittleEndian.Uint16(chunk[24:26])
			}
			hasFormat = true
		case "data":
			w.data = chunk
		}

		// Chunks are padded to an even size
		pos = chunkStart + chunkSize + (chunkSize & 1)
	}

	if !hasFormat || w.data == nil {
		return nil, fmt.Errorf("incomplete WAV file")
	}
	if w.channels < 1 || w.sampleRate < 1 || w.blockAlign < 1 {
		return nil, fmt.Errorf("invalid WAV file")
	}
	if w.audioFormat == wavFormatPCM &&
		(w.bitsPerSample == 8 || w.bitsPerSample == 16 || w.bitsPerSample == 24 || w.bitsPerSample == 32) {
		// Supported
	} else if w.audioFormat == wavFormatIEEEFloat && w.bitsPerSample == 32 {
		// Supported
	} else {
		return nil, fmt.Errorf("unsupported WAV format %v with %v bits per sample",
			w.audioFormat, w.bitsPerSample)
	}

	return &w, nil
}

func (w *wavFile) frames() int {
	return len(w.data) / w.blockAlign
}

// sample returns a sample scaled to [-1, 1]
func (w *wavFile) sample(frame int, channel int) float64 {
	offset := frame*w.blockAlign + channel*(w.bitsPerSample/8)
	switch w.bitsPerSample {
	case 8:
		// 8 bit WAV samples are unsigned
		return (float64(w.data[offset]) - 128) / 128
	case 16:
		v := int16(binary.LittleEndian.Uint16(w.data[offset : offset+2]))
		return float64(v) / 0x8000
	case 24:
		v := int32(w.data[offset]) | int32(w.data[offset+1])<<8 | int32(w.data[offset+2])<<16
		v = v << 8 >> 8 // Sign extend
		return float64(v) / 0x80_0000
	case 32:
		bits := binary.LittleEndian.Uint32(w.data[offset : offset+4])
		if w.audioFormat == wavFormatIEEEFloat {
			return float64(math.Float32frombits(bits))
		}
		return float64(int32(bits)) / 0x8000_0000
	}
	return 0
}

// bestChannel returns the channel with the most energy
func (w *wavFile) bestChannel() int {
	if w.channels == 1 {
		return 0
	}

	frames := w.frames()
	energy := make([]float64, w.channels)
	for frame := 0; frame < frames; frame++ {
		for channel := 0; channel < w.channels; channel++ {
			v := w.sample(frame, channel)
			energy[channel] += v * v
		}
	}

	best := 0
	for channel := 1; channel < w.channels; channel++ {
		if energy[channel] > energy[best] {
			best = channel
		}
	}
	return best
}
