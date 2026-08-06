package storage

import (
	"errors"
	"hash/crc64"
)

/*
blockDiskOverlay sends the blocks written to an overlay, leaving the image
untouched, and reads back from it the ones already saved. The blocks are taken
one at a time as ProDOS asks for them, so a 32 Mb hard disk costs nothing to
open however few blocks a game saves.
*/
type blockDiskOverlay struct {
	base    BlockDisk
	overlay *overlay
}

// NewBlockDiskOverlay wraps a block device so that the writes go to the
// overlay kept in overlayFilename instead of to the image. It also makes
// writable a device that was not, as the image is only read. An empty
// overlayFilename returns the device untouched.
func NewBlockDiskOverlay(base BlockDisk, overlayFilename string) (BlockDisk, error) {
	if overlayFilename == "" {
		return base, nil
	}

	checksum, err := checksumOfBlockDisk(base)
	if err != nil {
		return nil, err
	}

	o, err := openOverlay(overlayFilename, int(ProDosBlockSize),
		int(base.GetSizeInBlocks()), checksum)
	if err != nil {
		return nil, err
	}
	return &blockDiskOverlay{base: base, overlay: o}, nil
}

// checksumOfBlockDisk identifies the image an overlay was made from. It reads
// the whole device, which for a 32 Mb hard disk is the price of opening it
// once with the overlays in use.
func checksumOfBlockDisk(base BlockDisk) (uint64, error) {
	hash := crc64.New(overlayChecksumTable)
	for block := range base.GetSizeInBlocks() {
		data, err := base.Read(block)
		if err != nil {
			return 0, err
		}
		hash.Write(data)
	}
	return hash.Sum64(), nil
}

func (bd *blockDiskOverlay) GetSizeInBlocks() uint32 {
	return bd.base.GetSizeInBlocks()
}

func (bd *blockDiskOverlay) IsReadOnly() bool {
	return false
}

func (bd *blockDiskOverlay) Read(block uint32) ([]uint8, error) {
	if !bd.overlay.has(int(block)) {
		return bd.base.Read(block)
	}

	buf := make([]uint8, ProDosBlockSize)
	if err := bd.overlay.read(int(block), buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func (bd *blockDiskOverlay) Write(block uint32, data []uint8) error {
	if block >= bd.base.GetSizeInBlocks() {
		return errors.New("disk block number is too big")
	}
	return bd.overlay.write(int(block), data)
}
