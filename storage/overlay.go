package storage

import (
	"encoding/binary"
	"fmt"
	"hash/crc64"
	"os"
)

/*
An overlay keeps the changes made to a disk apart from the image it was loaded
from, so that saving a game does not modify the file the user has.

Only the parts the software writes are stored, indexed by a bitmap in the
header. Nothing is written for the rest, so the file has holes and takes on
disk as much as what was actually saved and not the size of the disk. That
matters for the hard disks, where an image can be 32 Mb and a saved game a
handful of blocks.

The unit is whatever the disk writes at a time: a track of 4096 bytes for the
diskettes, a block of 512 bytes for the ProDOS devices.

The header holds a checksum of the image the overlay was made from. An overlay
belongs to one image, and applying the one of another disk would corrupt it, so
a mismatch is refused instead of guessed.
*/

const (
	// overlayMagic is 24 bytes, the version is the last character
	overlayMagic = "IZAPPLE2 DISK OVERLAY 2\n"

	// overlayHeaderSize leaves room after the magic for the shape of the disk
	// and the checksum, and keeps the bitmap aligned
	overlayHeaderSize = 64
)

var overlayChecksumTable = crc64.MakeTable(crc64.ECMA)

type overlay struct {
	filename string
	unitSize int
	units    int
	checksum uint64

	// present has a bit set for each unit stored in the overlay
	present []byte
}

/*
openOverlay returns the overlay kept in filename for a disk of the given shape
and content. The file is not created until something is written to it, and an
overlay of another disk is refused.

The name says open because the overlay is read here, but nothing is left open:
each read and write opens the file and closes it. The callers only ask for an
overlay when there is a filename for it, and identifying a disk costs a read of
the whole image, so nothing of this happens without one.
*/
func openOverlay(filename string, unitSize int, units int, checksum uint64) (*overlay, error) {
	o := &overlay{
		filename: filename,
		unitSize: unitSize,
		units:    units,
		checksum: checksum,
		present:  make([]byte, (units+7)/8),
	}

	file, err := os.Open(filename)
	if os.IsNotExist(err) {
		// Nothing saved yet
		return o, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	header := make([]byte, overlayHeaderSize)
	if _, err := file.ReadAt(header, 0); err != nil {
		return nil, fmt.Errorf("the overlay %s is damaged: %w", filename, err)
	}
	if string(header[:len(overlayMagic)]) != overlayMagic {
		return nil, fmt.Errorf("%s is not a disk overlay", filename)
	}

	storedUnitSize := binary.LittleEndian.Uint32(header[24:])
	storedUnits := binary.LittleEndian.Uint32(header[28:])
	storedChecksum := binary.LittleEndian.Uint64(header[32:])

	if int(storedUnitSize) != unitSize || int(storedUnits) != units {
		return nil, fmt.Errorf(
			"the overlay %s was made for a disk of a different shape, delete it to start over",
			filename)
	}
	if storedChecksum != checksum {
		return nil, fmt.Errorf(
			"the overlay %s belongs to another disk image, delete it to start over",
			filename)
	}

	if _, err := file.ReadAt(o.present, overlayBitmapOffset); err != nil {
		return nil, fmt.Errorf("the overlay %s is damaged: %w", filename, err)
	}

	return o, nil
}

// has returns whether the overlay holds the unit
func (o *overlay) has(unit int) bool {
	return o.present[unit/8]&(1<<uint(unit%8)) != 0
}

// isEmpty returns whether nothing has been stored in the overlay yet, in which
// case there is not even a file
func (o *overlay) isEmpty() bool {
	for _, b := range o.present {
		if b != 0 {
			return false
		}
	}
	return true
}

// read fills the buffer with a unit stored in the overlay
func (o *overlay) read(unit int, buffer []uint8) error {
	file, err := os.Open(o.filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return o.readUnit(file, unit, buffer)
}

func (o *overlay) readUnit(file *os.File, unit int, buffer []uint8) error {
	if _, err := file.ReadAt(buffer, o.unitOffset(unit)); err != nil {
		return fmt.Errorf("can't read the unit %v of the overlay %s: %w",
			unit, o.filename, err)
	}
	return nil
}

// write stores a unit, creating the overlay the first time
func (o *overlay) write(unit int, data []uint8) error {
	file, err := o.openForWriting()
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteAt(data, o.unitOffset(unit)); err != nil {
		return err
	}

	// Only the byte of the bitmap that changed goes back to the file
	o.present[unit/8] |= 1 << uint(unit%8)
	_, err = file.WriteAt(o.present[unit/8:unit/8+1], overlayBitmapOffset+int64(unit/8))
	return err
}

// applyTo replaces in the image the units kept in the overlay. It is for the
// disks read whole into memory, the diskettes.
func (o *overlay) applyTo(data []uint8) error {
	if o.isEmpty() {
		return nil
	}

	file, err := os.Open(o.filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for unit := range o.units {
		if !o.has(unit) {
			continue
		}
		destination := data[unit*o.unitSize : (unit+1)*o.unitSize]
		if err := o.readUnit(file, unit, destination); err != nil {
			return err
		}
	}
	return nil
}

/*
openForWriting returns the overlay file for the caller to write to and close,
writing the header the first time. The header identifies the disk the overlay
belongs to, so openOverlay can refuse it for another one.
*/
func (o *overlay) openForWriting() (*os.File, error) {
	_, err := os.Stat(o.filename)
	isNew := os.IsNotExist(err)

	file, err := os.OpenFile(o.filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	if isNew {
		header := make([]byte, overlayHeaderSize)
		copy(header, overlayMagic)
		binary.LittleEndian.PutUint32(header[24:], uint32(o.unitSize))
		binary.LittleEndian.PutUint32(header[28:], uint32(o.units))
		binary.LittleEndian.PutUint64(header[32:], o.checksum)
		if _, err := file.WriteAt(header, 0); err != nil {
			file.Close()
			return nil, err
		}
		if _, err := file.WriteAt(o.present, overlayBitmapOffset); err != nil {
			file.Close()
			return nil, err
		}
	}

	return file, nil
}

const overlayBitmapOffset = int64(overlayHeaderSize)

// unitOffset is where a unit is stored, after the header and the bitmap
func (o *overlay) unitOffset(unit int) int64 {
	bitmapSize := int64(len(o.present))
	// Keep the units aligned, the bitmap is 5 bytes for a diskette and 8 Kb
	// for a full hard disk
	bitmapSize = (bitmapSize + overlayHeaderSize - 1) / overlayHeaderSize * overlayHeaderSize
	return overlayBitmapOffset + bitmapSize + int64(unit*o.unitSize)
}

// checksumOfBytes identifies the content an overlay was made from
func checksumOfBytes(data []uint8) uint64 {
	return crc64.Checksum(data, overlayChecksumTable)
}
