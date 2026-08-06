package izapple2

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// writeZip builds a zip file with the given entries, in order
func writeZip(t *testing.T, path string, entries map[string][]uint8, order []string) {
	t.Helper()

	var buffer bytes.Buffer
	w := zip.NewWriter(&buffer)
	for _, name := range order {
		f, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write(entries[name]); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, buffer.Bytes(), 0666); err != nil {
		t.Fatal(err)
	}
}

// A diskette is recognised by its size, a DSK image is 35 tracks of 16 sectors
const testDskSize = 35 * 16 * 256

func TestZipWithADisketteAndSomethingElse(t *testing.T) {
	diskette := make([]uint8, testDskSize)
	for i := range diskette {
		diskette[i] = uint8(i)
	}
	readme := bytes.Repeat([]uint8{'x'}, testDskSize*2) // Bigger than the diskette

	path := filepath.Join(t.TempDir(), "game.zip")
	writeZip(t, path,
		map[string][]uint8{"readme.txt": readme, "game.dsk": diskette},
		[]string{"readme.txt", "game.dsk"})

	data, _, err := LoadResource(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, diskette) {
		t.Error("expected the diskette and not the biggest file of the zip")
	}
}

func TestZipWithAHardDisk(t *testing.T) {
	// Nothing recognises a hard disk by its contents, so the biggest file of
	// the zip is the one to take
	hardDisk := make([]uint8, 64*512)
	for i := range hardDisk {
		hardDisk[i] = uint8(i / 512)
	}
	readme := bytes.Repeat([]uint8{'x'}, 100)

	path := filepath.Join(t.TempDir(), "collection.zip")
	writeZip(t, path,
		map[string][]uint8{"readme.txt": readme, "collection.hdv": hardDisk},
		[]string{"readme.txt", "collection.hdv"})

	data, _, err := LoadResource(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, hardDisk) {
		t.Errorf("expected the hard disk image, got %v bytes", len(data))
	}
}

func TestZipIsNotWriteable(t *testing.T) {
	diskette := make([]uint8, testDskSize)
	path := filepath.Join(t.TempDir(), "game.zip")
	writeZip(t, path, map[string][]uint8{"game.dsk": diskette}, []string{"game.dsk"})

	_, writeable, err := LoadResource(path)
	if err != nil {
		t.Fatal(err)
	}
	if writeable {
		t.Error("a disk inside a zip cannot be written back to the image")
	}
}
