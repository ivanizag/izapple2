package apple2

import (
	"testing"
)

func TestNibBackAndForth(t *testing.T) {
	// Init data
	data := make([]byte, bytesPerTrack)
	for i := 0; i < bytesPerTrack; i++ {
		data[i] = byte(i % 100)
	}

	nib := nibEncodeTrack(data, 255, 0, &dos33SectorsLogicalOrder)
	data2, err := nibDecodeTrack(nib, &dos33SectorsLogicalOrder)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < bytesPerTrack; i++ {
		if data[i] != data2[i] {
			t.Errorf("Mismatch in %v: %02x -> %02x", i, data[i], data2[i])
		}
	}
}
