package apple2

import (
	"errors"
	"strings"
)

type diskette interface {
	powerOn(cycle uint64)
	powerOff(cycle uint64)
	read(quarterTrack int, cycle uint64) uint8
	write(quarterTrack int, value uint8, cycle uint64)
}

func loadDisquette(filename string) (diskette, error) {
	data, err := loadResource(filename)
	if err != nil {
		return nil, err
	}

	if isFileNib(data) {
		var d diskette16sector
		d.nib = newFileNib(data)
		return &d, nil
	}

	if isFileDsk(data) {
		isPO := strings.HasSuffix(strings.ToLower(filename), "po")
		var d diskette16sector
		d.nib = newFileDsk(data, isPO)
		return &d, nil
	}

	if isFileWoz(data) {
		f, err := newFileWoz(data)
		if err != nil {
			return nil, err
		}

		return newDisquetteWoz(f)
	}

	return nil, errors.New("Diskette format not supported")
}
