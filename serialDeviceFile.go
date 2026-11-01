package izapple2

import (
	"os"
)

// serialDeviceFile appends what is sent to a file, as it is
type serialDeviceFile struct {
	file *os.File
}

func newSerialDeviceFile(filename string) (*serialDeviceFile, error) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &serialDeviceFile{file}, nil
}

func (d *serialDeviceFile) Send(value uint8) {
	d.file.Write([]byte{value})
}

func (d *serialDeviceFile) Receive() (uint8, bool) {
	return 0, false
}

func (d *serialDeviceFile) isReady() bool {
	return true
}
