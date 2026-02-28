package storage

import (
	"testing"
)

func TestNibBackAndForth(t *testing.T) {
	// Init data
	data := make([]byte, bytesPerTrack)
	for i := range bytesPerTrack {
		data[i] = byte(i % 100)
	}

	nib := nibEncodeTrack(data, 255, 0, &dos33SectorsLogicalOrder)
	data2, err := nibDecodeTrack(nib, &dos33SectorsLogicalOrder)
	if err != nil {
		t.Error(err)
	}

	for i := range bytesPerTrack {
		if data[i] != data2[i] {
			t.Errorf("Mismatch in %v: %02x -> %02x", i, data[i], data2[i])
		}
	}
}
