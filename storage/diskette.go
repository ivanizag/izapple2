package storage

import (
	"errors"
)

// Diskette represents a diskette and it's RW mechanism
type Diskette interface {
	PowerOn(cycle uint64)
	PowerOff(cycle uint64)
	Read(quarterTrack int, cycle uint64) uint8
	Write(quarterTrack int, value uint8, cycle uint64)
	Is13Sectors() bool
}

// IsDiskette returns true if the files looks like a 5 1/4 diskette
func IsDiskette(data []byte) bool {
	return isFileNib(data) || isFileDsk(data) || isFileWoz(data)
}

/*
MakeDiskette returns a Diskette by detecting the format.

With an overlayFilename, the changes the software writes are kept in that file
and the image loaded is never modified. It also makes writable the diskettes
that could not be written otherwise, the ones loaded from a compressed file, an
URL or the embedded resources. An empty overlayFilename writes the changes back
to the image, which only works for the DSK and PO files on disk.
*/
func MakeDiskette(data []byte, filename string, writeable bool, overlayFilename string) (Diskette, error) {
	if isFileD13(data) {
		return nil, errors.New("files with .d13 format are not supported for 13 sectors disk, use .nib or .woz")
	}

	if isFileNib(data) {
		var d disketteNib
		d.nib = newFileNib(data)
		return &d, nil
	}

	if isFileDsk(data) {
		var overlay *overlay
		if overlayFilename != "" {
			// Only now is it worth identifying the image, which means reading
			// all of it. The checksum is of the image as it was loaded, before
			// anything saved before is put back on top of it.
			var err error
			overlay, err = openOverlay(overlayFilename, bytesPerTrack, numberOfTracks,
				checksumOfBytes(data))
			if err != nil {
				return nil, err
			}
			if err := overlay.applyTo(data); err != nil {
				return nil, err
			}
		}

		var d disketteNibWritable
		d.nib = newFileDsk(data, filename)
		d.nib.overlay = overlay
		d.nib.supportsWrite = overlay != nil || (d.nib.supportsWrite && writeable)
		return &d, nil
	}

	if isFileWoz(data) {
		f, err := NewFileWoz(data)
		if err != nil {
			return nil, err
		}

		return newDisquetteWoz(f)
	}

	return nil, errors.New("diskette format not supported")
}
