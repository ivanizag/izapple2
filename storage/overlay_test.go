package storage

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

const testBlocks = 16

// makeTestImage writes a raw ProDOS image with recognizable content
func makeTestImage(t *testing.T, path string) []uint8 {
	t.Helper()

	data := make([]uint8, testBlocks*int(ProDosBlockSize))
	for i := range data {
		data[i] = uint8(i)
	}
	if err := os.WriteFile(path, data, 0666); err != nil {
		t.Fatal(err)
	}
	return data
}

func openTestDisk(t *testing.T, path string, overlayPath string) BlockDisk {
	t.Helper()

	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	base, err := NewBlockDiskFile(file, false)
	if err != nil {
		t.Fatal(err)
	}
	disk, err := NewBlockDiskOverlay(base, overlayPath)
	if err != nil {
		t.Fatal(err)
	}
	return disk
}

func TestBlockOverlayKeepsTheImageUntouched(t *testing.T) {
	dir := t.TempDir()
	image := filepath.Join(dir, "disk.hdv")
	overlayPath := filepath.Join(dir, "disk.hdv.ovl")
	original := makeTestImage(t, image)

	written := bytes.Repeat([]uint8{0xaa}, int(ProDosBlockSize))
	disk := openTestDisk(t, image, overlayPath)
	if err := disk.Write(3, written); err != nil {
		t.Fatal(err)
	}

	// The block written comes back from the overlay, the others from the image
	got, err := disk.Read(3)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, written) {
		t.Error("the block written did not come back")
	}
	got, err = disk.Read(4)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, original[4*int(ProDosBlockSize):5*int(ProDosBlockSize)]) {
		t.Error("an untouched block did not come from the image")
	}

	// The image on disk is exactly as it was
	after, err := os.ReadFile(image)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, original) {
		t.Error("the image was modified, it should have been left alone")
	}
}

func TestBlockOverlaySurvivesReopening(t *testing.T) {
	dir := t.TempDir()
	image := filepath.Join(dir, "disk.hdv")
	overlayPath := filepath.Join(dir, "disk.hdv.ovl")
	makeTestImage(t, image)

	written := bytes.Repeat([]uint8{0x5a}, int(ProDosBlockSize))
	disk := openTestDisk(t, image, overlayPath)
	if err := disk.Write(7, written); err != nil {
		t.Fatal(err)
	}

	reopened := openTestDisk(t, image, overlayPath)
	got, err := reopened.Read(7)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, written) {
		t.Error("what was saved did not survive reopening the disk")
	}
}

func TestOverlayIsRefusedForAnotherImage(t *testing.T) {
	dir := t.TempDir()
	image := filepath.Join(dir, "disk.hdv")
	overlayPath := filepath.Join(dir, "disk.hdv.ovl")
	makeTestImage(t, image)

	disk := openTestDisk(t, image, overlayPath)
	if err := disk.Write(2, bytes.Repeat([]uint8{1}, int(ProDosBlockSize))); err != nil {
		t.Fatal(err)
	}

	// Another disk with the same name, the overlay does not belong to it
	changed := makeTestImage(t, image)
	changed[0] ^= 0xff
	if err := os.WriteFile(image, changed, 0666); err != nil {
		t.Fatal(err)
	}

	file, err := os.OpenFile(image, os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	base, err := NewBlockDiskFile(file, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewBlockDiskOverlay(base, overlayPath); err == nil {
		t.Error("expected the overlay of another image to be refused")
	}
}

func TestOverlayIsRefusedForAnotherShape(t *testing.T) {
	dir := t.TempDir()
	overlayPath := filepath.Join(dir, "shared.ovl")

	o, err := openOverlay(overlayPath, int(ProDosBlockSize), testBlocks, 1234)
	if err != nil {
		t.Fatal(err)
	}
	if err := o.write(1, bytes.Repeat([]uint8{9}, int(ProDosBlockSize))); err != nil {
		t.Fatal(err)
	}

	// Same content, but taken as a diskette of tracks instead of blocks
	if _, err := openOverlay(overlayPath, bytesPerTrack, numberOfTracks, 1234); err == nil {
		t.Error("expected the overlay of a differently shaped disk to be refused")
	}
}

func TestOverlayStoresOnlyWhatWasWritten(t *testing.T) {
	dir := t.TempDir()
	overlayPath := filepath.Join(dir, "sparse.ovl")

	// A full 32 Mb hard disk, with a single block saved
	o, err := openOverlay(overlayPath, int(ProDosBlockSize), int(proDosMaxBlocks), 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := o.write(10, bytes.Repeat([]uint8{7}, int(ProDosBlockSize))); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(overlayPath)
	if err != nil {
		t.Fatal(err)
	}
	// The header, the bitmap of 8 Kb and the eleven blocks up to the one
	// written, far from the 32 Mb of the disk
	if info.Size() > 64*1024 {
		t.Errorf("the overlay of one block grew to %v bytes", info.Size())
	}
}
