package apple2

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
)

/*
Valid for ProDos hard disks in 2MG format.

See:
	https://apple2.org.za/gswv/a2zine/Docs/DiskImage_2MG_Info.txt
*/

const (
	proDosBlockSize         = uint32(512)
	hardDisk2mgPreamble     = uint32(1196247346) // "2IMG"
	hardDisk2mgFormatProdos = 1
	hardDisk2mgVersion      = 1
)

type hardDisk struct {
	file     *os.File
	readOnly bool
	header   hardDisk2mgHeader
}

type hardDisk2mgHeader struct {
	Preamble      uint32
	Creator       uint32
	HeaderSize    uint16
	Version       uint16
	Format        uint32
	Flags         uint32
	Blocks        uint32
	OffsetData    uint32
	LengthData    uint32
	OffsetComment uint32
	LengthComment uint32
	OffsetCreator uint32
	LengthCreator uint32
}

func (hd *hardDisk) read(block uint32) ([]uint8, error) {
	if block >= hd.header.Blocks {
		return nil, errors.New("disk block number is too big")
	}

	buf := make([]uint8, proDosBlockSize)

	offset := int64(hd.header.OffsetData + block*proDosBlockSize)
	_, err := hd.file.ReadAt(buf, offset)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func (hd *hardDisk) write(block uint32, data []uint8) error {
	if hd.readOnly {
		return errors.New("can't write in a readonly disk")
	}
	if block >= hd.header.Blocks {
		return errors.New("disk block number is too big")
	}

	offset := int64(hd.header.OffsetData + block*proDosBlockSize)
	_, err := hd.file.WriteAt(data, offset)
	if err != nil {
		return err
	}

	return nil
}

func openHardDisk2mg(filename string) (*hardDisk, error) {
	var hd hardDisk

	hd.readOnly = false
	file, err := os.OpenFile(filename, os.O_RDWR, 0)
	if os.IsPermission(err) {
		// Retry in read-only mode
		hd.readOnly = true
		file, err = os.OpenFile(filename, os.O_RDONLY, 0)
	}
	if err != nil {
		return nil, err
	}
	hd.file = file

	fileInfo, err := hd.file.Stat()
	if err != nil {
		return nil, err
	}

	minHeaderSize := binary.Size(&hd.header)
	if fileInfo.Size() < int64(minHeaderSize) {
		return nil, errors.New("Invalid 2MG file")
	}

	err = readHeader(hd.file, &hd.header)
	if err != nil {
		return nil, err
	}

	if fileInfo.Size() < int64(hd.header.OffsetData+hd.header.Blocks*proDosBlockSize) {
		return nil, errors.New("Thr 2MG file is too small")
	}

	return &hd, nil
}

func readHeader(buf io.Reader, header *hardDisk2mgHeader) error {
	err := binary.Read(buf, binary.LittleEndian, header)
	if err != nil {
		return err
	}

	if header.Preamble != hardDisk2mgPreamble {
		return errors.New("2mg file must start with '2IMG'")
	}

	if header.Format != hardDisk2mgFormatProdos {
		return errors.New("Only prodos hard disks are supported")
	}

	if header.Version != hardDisk2mgVersion {
		return errors.New("Version of 2MG image not supported")
	}

	return nil
}
