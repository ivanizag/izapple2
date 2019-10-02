package apple2

import (
	"bytes"
	"encoding/binary"
)

/*
Valid for ProDos hard disks in 2MG format. Read only.

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
	data   []uint8
	header hardDisk2mgHeader
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

func (hd *hardDisk) read(block uint32) []uint8 {
	if block >= hd.header.Blocks {
		return nil
	}
	offset := hd.header.OffsetData + block*proDosBlockSize
	return hd.data[offset : offset+proDosBlockSize]
}

func loadHardDisk2mg(filename string) *hardDisk {
	var hd hardDisk

	hd.data = loadResource(filename)

	size := len(hd.data)
	if size < binary.Size(&hd.header) {
		panic("2mg file is too short")
	}

	buf := bytes.NewReader(hd.data)
	err := binary.Read(buf, binary.LittleEndian, &hd.header)
	if err != nil {
		panic(err)
	}

	if hd.header.Preamble != hardDisk2mgPreamble {
		panic("2mg file must start with '2IMG'")
	}

	if hd.header.Format != hardDisk2mgFormatProdos {
		panic("Only prodos hard disks are supported")
	}

	if hd.header.Version != hardDisk2mgVersion {
		panic("Version of 2MG image not supported")
	}

	if size < int(hd.header.OffsetData+hd.header.Blocks*proDosBlockSize) {
		panic("Thr 2MG file is too small")
	}

	return &hd
}
