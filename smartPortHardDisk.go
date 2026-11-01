package izapple2

import (
	"fmt"

	"github.com/ivanizag/izapple2/storage"
)

/*
To implement a smartPort hard drive we have to support the smartPort commands.

See:
	Beneath Prodos, section 6-6, 7-13 and 5-8. (http://www.apple-iigs.info/doc/fichiers/beneathprodos.pdf)
	Apple IIc Technical Reference, 2nd Edition. Chapter 8. https://ia800207.us.archive.org/19/items/AppleIIcTechnicalReference2ndEd/Apple%20IIc%20Technical%20Reference%202nd%20ed.pdf

*/

// SmartPortHardDisk represents a hard disk
type SmartPortHardDisk struct {
	filename string
	trace    bool
	disk     storage.BlockDisk
}

// NewSmartPortHardDisk creates a new hard disk with the smartPort interface.
// An empty saveDirectory writes the changes back to the image.
func NewSmartPortHardDisk(filename string, saveDirectory string) (*SmartPortHardDisk, error) {
	var d SmartPortHardDisk
	d.filename = filename

	hd, err := LoadBlockDisk(filename, saveDirectory)
	if err != nil {
		return nil, err
	}
	d.disk = hd

	return &d, nil
}

func (d *SmartPortHardDisk) exec(call *smartPortCall) (uint8, []uint8) {
	var result uint8
	var data []uint8

	switch call.command {
	case smartPortCommandStatus:
		if call.statusCode() == smartPortStatusCodeDevice {
			result, data = d.deviceStatus()
		} else {
			result, data = d.status()
		}

	case smartPortCommandReadBlock:
		block := call.param24(4)
		result, data = d.readBlock(block)

	case smartPortCommandWriteBlock:
		block := call.param24(4)
		result = d.writeBlock(block, call.data)

	default:
		// Prodos device command not supported
		result = smartPortErrorIO
	}

	if d.trace {
		fmt.Printf("[SmartPortHardDisk] Command %v, return %s \n",
			call, smartPortErrorMessage(result))
	}

	return result, data
}

func (d *SmartPortHardDisk) readBlock(block uint32) (uint8, []uint8) {
	if d.trace {
		fmt.Printf("[SmartPortHardDisk] Read block %v.\n", block)
	}

	data, err := d.disk.Read(block)
	if err != nil {
		return smartPortErrorIO, nil
	}
	return smartPortNoError, data
}

func (d *SmartPortHardDisk) writeBlock(block uint32, data []uint8) uint8 {
	if d.trace {
		fmt.Printf("[SmartPortHardDisk] Write block %v.\n", block)
	}

	if d.disk.IsReadOnly() {
		return smartPortErrorWriteProtected
	}
	if len(data) != int(storage.ProDosBlockSize) {
		return smartPortErrorIO
	}

	err := d.disk.Write(block, data)
	if err != nil {
		return smartPortErrorIO
	}

	return smartPortNoError
}

/*
deviceStatus is the status list of the STATUS with code 0: the status byte
and the size in blocks. The SmartPort card answers it with the status of
the host for all the units, it comes from the SmartPort bus.
*/
func (d *SmartPortHardDisk) deviceStatus() (uint8, []uint8) {
	status := smartPortStatusCodeTypeBlock | smartPortStatusCodeTypeRead |
		smartPortStatusCodeTypeOnline | smartPortStatusCodeTypeFormat
	if d.disk.IsReadOnly() {
		status |= smartPortStatusCodeTypeProtected
	} else {
		status |= smartPortStatusCodeTypeWrite
	}
	blocks := d.disk.GetSizeInBlocks()
	return smartPortNoError, []uint8{status, uint8(blocks), uint8(blocks >> 8), uint8(blocks >> 16)}
}

func (d *SmartPortHardDisk) status() (uint8, []uint8) {
	if d.trace {
		fmt.Printf("[SmartPortHardDisk] Status.\n")
	}

	// See http://www.1000bit.it/support/manuali/apple/technotes/smpt/tn.smpt.2.html
	return smartPortNoError, []uint8{
		0x01, // One device
		0xff, // No interrupt
		0x00,
		0x00, // Unknown manufacturer
		0x01,
		0x00, // Version 1.0 final
		0x00,
		0x00, // Reserved
	}
}
