package storage

import (
	"encoding/binary"
	"errors"
	"io"
)

/*
Valid for ProDos disks in 2MG format.

See:
	https://apple2.org.za/gswv/a2zine/Docs/DiskImage_2MG_Info.txt
*/

const (
	file2mgPreamble     = uint32(1196247346) // "2IMG"
	file2mgFormatProdos = 1
	file2mgVersion      = 1
)

type file2mgHeader struct {
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

func parse2mg(bd *BlockDisk) error {
	fileInfo, err := bd.file.Stat()
	if err != nil {
		return err
	}

	var header file2mgHeader
	minHeaderSize := binary.Size(&header)
	if fileInfo.Size() < int64(minHeaderSize) {
		return errors.New("invalid 2MG file")
	}

	err = readHeader(bd.file, &header)
	if err != nil {
		return err
	}

	bd.blocks = header.Blocks
	bd.dataOffset = header.OffsetData

	if fileInfo.Size() < int64(bd.dataOffset+bd.blocks*ProDosBlockSize) {
		return errors.New("the 2MG file is too small")
	}

	return nil
}

func readHeader(buf io.Reader, header *file2mgHeader) error {
	err := binary.Read(buf, binary.LittleEndian, header)
	if err != nil {
		return err
	}

	if header.Preamble != file2mgPreamble {
		return errors.New("the 2mg file must start with '2IMG'")
	}

	if header.Format != file2mgFormatProdos {
		return errors.New("only prodos disks are supported")
	}

	if header.Version != file2mgVersion {
		return errors.New("version of 2MG image not supported")
	}

	return nil
}
