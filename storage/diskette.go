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

// MakeDiskette returns a Diskette by detecting the format
func MakeDiskette(data []byte, filename string, writeable bool) (Diskette, error) {
	if isFileD13(data) {
		return nil, errors.New("files with .d13 format are not supported for 13 sectors disk, use .nib or .woz")
	}

	if isFileNib(data) {
		var d disketteNib
		d.nib = newFileNib(data)
		return &d, nil
	}

	if isFileDsk(data) {
		var d disketteNibWritable
		d.nib = newFileDsk(data, filename)
		d.nib.supportsWrite = d.nib.supportsWrite && writeable
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
