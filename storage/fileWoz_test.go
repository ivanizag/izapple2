package storage

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"os"
	"path/filepath"
	"testing"
)

const testWozImage = "../woz_test_images/DOS 3.3 System Master.woz"

// testWozCopy writes a copy of a WOZ image in a file of the test, with the
// write protect tab on or off
func testWozCopy(t *testing.T, writeProtected bool) string {
	t.Helper()
	data, err := os.ReadFile(testWozImage)
	if err != nil {
		t.Fatal(err)
	}
	data[22] = 0 // The WriteProtected field of INFO
	if writeProtected {
		data[22] = 1
	}
	path := filepath.Join(t.TempDir(), "disk.woz")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func testWozDiskette(t *testing.T, path string, writeable bool, overlayFilename string) *disketteWoz {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	d, err := MakeDiskette(data, path, writeable, overlayFilename)
	if err != nil {
		t.Fatal(err)
	}
	return d.(*disketteWoz)
}

func testTrackData() []byte {
	data := make([]byte, bytesPerTrack)
	for i := range data {
		data[i] = byte(i % 251)
	}
	return data
}

/*
testWriteTrack formats the track 0 with the data, as DOS would: one Write per
nibble, 32 cycles apart and 40 for the sync nibbles. The gaps are shorter than
the ones of the nib files for the track to fit in a turn, and the writing starts
on the first bit of the track and goes around it once.
*/
func testWriteTrack(t *testing.T, d *disketteWoz, data []byte) {
	t.Helper()
	bitCount := d.data.tracks[d.data.trackMap[0]].bitCount
	d.position = bitCount - 1
	d.positionMax = bitCount

	nibbles := nibEncodeTrack(data, defaultVolumeTag, 0, &dos33SectorsLogicalOrder)
	cycle := uint64(0)
	bits := uint32(0)
	syncs := 0
	for _, nibble := range nibbles {
		if nibble == 0xff {
			syncs++
			if syncs > 12 {
				continue
			}
		} else {
			syncs = 0
		}
		d.Write(0, nibble, cycle)
		if nibble == 0xff {
			cycle += 40
			bits += 10
		} else {
			cycle += 32
			bits += 8
		}
	}
	if bits > bitCount {
		t.Fatalf("the track of %v bits does not fit the %v written", bitCount, bits)
	}
	for bits < bitCount {
		d.Write(0, 0xff, cycle)
		cycle += 40
		bits += 10
	}
	d.Read(0, cycle) // Back to reading, the last nibble is done
}

func testReadTrack(t *testing.T, f *FileWoz) []byte {
	t.Helper()
	data, err := nibDecodeTrack(f.DumpTrackAsNib(0), &dos33SectorsLogicalOrder)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestWozWriteReadsBack(t *testing.T) {
	d := testWozDiskette(t, testWozCopy(t, false), false, "")
	written := testTrackData()
	testWriteTrack(t, d, written)

	if !bytes.Equal(testReadTrack(t, d.data), written) {
		t.Error("the track written did not read back")
	}
}

func TestWozWriteIsSavedToTheFile(t *testing.T) {
	path := testWozCopy(t, false)
	d := testWozDiskette(t, path, true, "")
	written := testTrackData()
	testWriteTrack(t, d, written)
	d.PowerOff(0)

	saved, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	crc := binary.LittleEndian.Uint32(saved[wozCrcOffset:])
	if crc != crc32.ChecksumIEEE(saved[wozFirstChunkPos:]) {
		t.Error("the CRC of the file saved is not valid")
	}

	reloaded := testWozDiskette(t, path, true, "")
	if !bytes.Equal(testReadTrack(t, reloaded.data), written) {
		t.Error("the track written was not saved to the file")
	}
}

func TestWozWriteProtected(t *testing.T) {
	path := testWozCopy(t, true)
	d := testWozDiskette(t, path, true, "")
	if !d.IsWriteProtected() {
		t.Fatal("the diskette should be write protected")
	}

	before := testReadTrack(t, d.data)
	testWriteTrack(t, d, testTrackData())
	d.PowerOff(0)
	if !bytes.Equal(testReadTrack(t, d.data), before) {
		t.Error("a write protected diskette was written")
	}
}

func TestWozOverlayKeepsTheImageUntouched(t *testing.T) {
	path := testWozCopy(t, true)
	overlayPath := filepath.Join(t.TempDir(), "disk.woz.ovl")
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	// The overlay makes writable even a write protected image
	d := testWozDiskette(t, path, false, overlayPath)
	if d.IsWriteProtected() {
		t.Fatal("the diskette with an overlay should not be write protected")
	}
	written := testTrackData()
	testWriteTrack(t, d, written)
	d.PowerOff(0)

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, original) {
		t.Error("the image was modified, it should have been left alone")
	}

	reopened := testWozDiskette(t, path, false, overlayPath)
	if !bytes.Equal(testReadTrack(t, reopened.data), written) {
		t.Error("what was saved did not survive reopening the diskette")
	}

	// Another image with the same name, the overlay does not belong to it
	changed := append([]byte{}, original...)
	changed[len(changed)-1] ^= 0xff
	if _, err := MakeDiskette(changed, path, false, overlayPath); err == nil {
		t.Error("expected the overlay of another image to be refused")
	}
}
