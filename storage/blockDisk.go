package storage

import (
	"errors"
	"fmt"
	"os"
)

/*
Valid for ProDOS disks with 512 bytes blocks. Can be diskettes or hard disks
*/

const (
	// ProDosBlockSize is the size of the blocks on the ProDOS devices
	ProDosBlockSize = uint32(512)
	proDosMaxBlocks = uint32(65536)
)

// BlockDisk is any block device with 512 bytes blocks
type BlockDisk struct {
	file       *os.File
	readOnly   bool
	dataOffset uint32
	blocks     uint32
}

// OpenBlockDisk creates a new block device links to a file
func OpenBlockDisk(filename string) (*BlockDisk, error) {
	var bd BlockDisk

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

	if fileInfo.Size() > int64(ProDosBlockSize*proDosMaxBlocks) {
		return nil, fmt.Errorf("file is too big OR %s", err2mg.Error())
	}

	size := uint32(fileInfo.Size())
	if size%ProDosBlockSize != 0 {
		return nil, fmt.Errorf("file size os invalid OR %s", err2mg.Error())
	}

	// It's a valid raw file
	bd.blocks = size / ProDosBlockSize
	bd.dataOffset = 0
	return &bd, nil
}

// GetSizeInBlocks returns the number of blocks of the device
func (bd *BlockDisk) GetSizeInBlocks() uint32 {
	return bd.blocks
}

// IsReadOnly returns true if the device is read only
func (bd *BlockDisk) IsReadOnly() bool {
	return bd.readOnly
}

func (bd *BlockDisk) Read(block uint32) ([]uint8, error) {
	if block >= bd.blocks {
		return nil, errors.New("disk block number is too big")
	}

	buf := make([]uint8, ProDosBlockSize)

	offset := int64(bd.dataOffset + block*ProDosBlockSize)
	_, err := bd.file.ReadAt(buf, offset)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func (bd *BlockDisk) Write(block uint32, data []uint8) error {
	if bd.readOnly {
		return errors.New("can't write in a readonly disk")
	}
	if block >= bd.blocks {
		return errors.New("disk block number is too big")
	}

	offset := int64(bd.dataOffset + block*ProDosBlockSize)
	_, err := bd.file.WriteAt(data, offset)
	if err != nil {
		return err
	}

	return nil
}
