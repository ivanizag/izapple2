package storage

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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
type BlockDisk interface {
	GetSizeInBlocks() uint32
	IsReadOnly() bool
	Read(block uint32) ([]uint8, error)
	Write(block uint32, data []uint8) error
}

type blockDiskBase struct {
	file       *os.File
	readOnly   bool
	dataOffset uint32
	blocks     uint32
}

type blockDiskFile struct {
	blockDiskBase
	file *os.File
}

type blockDiskMemory struct {
	blockDiskBase
	data []uint8
}

// OpenBlockDisk creates a new block device linked to a file
func NewBlockDiskFile(file *os.File, readOnly bool) (BlockDisk, error) {
	var bd blockDiskFile
	bd.file = file
	bd.readOnly = readOnly

	fileInfo, err := bd.file.Stat()
	if err != nil {
		return nil, err
	}

	size := uint32(fileInfo.Size())
	bd.blocks, bd.dataOffset, err = getBlockAndOffset(bd.file, size)
	if err != nil {
		return nil, err
	}
	return &bd, nil
}

func NewBlockDiskMemory(data []uint8) (BlockDisk, error) {
	var bd blockDiskMemory
	bd.data = data
	bd.readOnly = true

	var err error
	bd.blocks, bd.dataOffset, err = getBlockAndOffset(bytes.NewReader(data), uint32(len(data)))
	if err != nil {
		return nil, err
	}
	return &bd, nil
}

func getBlockAndOffset(reader io.Reader, size uint32) (uint32, uint32, error) {
	header, err := parse2mg(reader, size)
	if err == nil {
		// It's a 2mg file
		return header.Blocks, header.OffsetData, nil
	}

	// Let's try to load as raw ProDOS Blocks
	if size > ProDosBlockSize*proDosMaxBlocks {
		return 0, 0, fmt.Errorf("file is too big OR %s", err.Error())
	}

	if size%ProDosBlockSize != 0 {
		return 0, 0, fmt.Errorf("file size os invalid OR %s", err.Error())
	}

	// It's a valid raw file
	return size / ProDosBlockSize, 0, nil
}

// GetSizeInBlocks returns the number of blocks of the device
func (bd *blockDiskBase) GetSizeInBlocks() uint32 {
	return bd.blocks
}

// IsReadOnly returns true if the device is read only
func (bd *blockDiskBase) IsReadOnly() bool {
	return bd.readOnly
}

func (bd *blockDiskFile) Read(block uint32) ([]uint8, error) {
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

func (bd *blockDiskMemory) Read(block uint32) ([]uint8, error) {
	if block >= bd.blocks {
		return nil, errors.New("disk block number is too big")
	}

	offset := bd.dataOffset + block*ProDosBlockSize
	return bd.data[offset : offset+ProDosBlockSize], nil
}

func (bd *blockDiskFile) Write(block uint32, data []uint8) error {
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

func (bd *blockDiskMemory) Write(block uint32, data []uint8) error {
	return errors.New("can't write in a readonly disk")
}
