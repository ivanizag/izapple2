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

func TestNibDecodeWithPrologAcrossTheEnd(t *testing.T) {
	data := make([]byte, bytesPerTrack)
	for i := range bytesPerTrack {
		data[i] = byte(i % 100)
	}
	nib := nibEncodeTrack(data, 255, 0, &dos33SectorsLogicalOrder)

	// Turn the track so that its first address prolog starts on the last
	// byte and goes on at the start, as it can be on a track being written
	start := findProlog(diskPrologByte3Address, nib, 0) - 3
	turned := append(append([]byte{}, nib[start+1:]...), nib[:start+1]...)

	data2, err := nibDecodeTrack(turned, &dos33SectorsLogicalOrder)
	if err != nil {
		t.Fatal(err)
	}
	for i := range bytesPerTrack {
		if data[i] != data2[i] {
			t.Fatalf("Mismatch in %v: %02x -> %02x", i, data[i], data2[i])
		}
	}
}
