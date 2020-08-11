package apple2

import (
	"errors"
	"fmt"
	"os"
)

/*
Valid for ProDos disks with 512 bytes blocks. Can be diskettes or hard disks
*/

const (
	proDosBlockSize = uint32(512)
	proDosMaxBlocks = uint32(65536)
)

type blockDisk struct {
	file       *os.File
	readOnly   bool
	dataOffset uint32
	blocks     uint32
}

func openBlockDisk(filename string) (*blockDisk, error) {
	var bd blockDisk

	bd.readOnly = false
	file, err := os.OpenFile(filename, os.O_RDWR, 0)
	if os.IsPermission(err) {
		// Retry in read-only mode
		bd.readOnly = true
		file, err = os.OpenFile(filename, os.O_RDONLY, 0)
	}
	if err != nil {
		return nil, err
	}
	bd.file = file

	err2mg := parse2mg(&bd)
	if err2mg == nil {
		// It's a 2mg file, ready to use
		return &bd, nil
	}

	// Let's try to load as raw ProDOS Blocks
	fileInfo, err := bd.file.Stat()
	if err != nil {
		return nil, err
	}

	if fileInfo.Size() > int64(proDosBlockSize*proDosMaxBlocks) {
		return nil, fmt.Errorf("File is too big OR %s", err2mg.Error())
	}

	size := uint32(fileInfo.Size())
	if size%proDosBlockSize != 0 {
		return nil, fmt.Errorf("File size os invalid OR %s", err2mg.Error())
	}

	// It's a valid raw file
	bd.blocks = size / proDosBlockSize
	bd.dataOffset = 0
	return &bd, nil
}

func (bd *blockDisk) read(block uint32) ([]uint8, error) {
	if block >= bd.blocks {
		return nil, errors.New("disk block number is too big")
	}

	buf := make([]uint8, proDosBlockSize)

	offset := int64(bd.dataOffset + block*proDosBlockSize)
	_, err := bd.file.ReadAt(buf, offset)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func (bd *blockDisk) write(block uint32, data []uint8) error {
	if bd.readOnly {
		return errors.New("can't write in a readonly disk")
	}
	if block >= bd.blocks {
		return errors.New("disk block number is too big")
	}

	offset := int64(bd.dataOffset + block*proDosBlockSize)
	_, err := bd.file.WriteAt(data, offset)
	if err != nil {
		return err
	}

	return nil
}
