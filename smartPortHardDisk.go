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
	host     *CardSmartPort // For DMA
	filename string
	trace    bool
	disk     storage.BlockDisk
}

// NewSmartPortHardDisk creates a new hard disk with the smartPort interface
func NewSmartPortHardDisk(host *CardSmartPort, filename string) (*SmartPortHardDisk, error) {
	var d SmartPortHardDisk
	d.host = host
	d.filename = filename

	hd, err := LoadBlockDisk(filename)
	if err != nil {
		return nil, err
	}
	d.disk = hd

	return &d, nil
}

func (d *SmartPortHardDisk) exec(call *smartPortCall) uint8 {
	var result uint8

	switch call.command {
	case smartPortCommandStatus:
		address := call.param16(2)
		result = d.status(address)

	case smartPortCommandReadBlock:
		address := call.param16(2)
		block := call.param24(4)
		result = d.readBlock(block, address)

	case smartPortCommandWriteBlock:
		address := call.param16(2)
		block := call.param24(4)
		result = d.writeBlock(block, address)

	default:
		// Prodos device command not supported
		result = smartPortErrorIO
	}

	if d.trace {
		fmt.Printf("[SmartPortHardDisk] Command %v, return %s \n",
			call, smartPortErrorMessage(result))
	}

	return result
}

func (d *SmartPortHardDisk) readBlock(block uint32, dest uint16) uint8 {
	if d.trace {
		fmt.Printf("[SmartPortHardDisk] Read block %v into $%x.\n", block, dest)
	}

	data, err := d.disk.Read(block)
	if err != nil {
		return smartPortErrorIO
	}

	// Byte by byte transfer to memory using the full Poke code path
	for i := uint16(0); i < uint16(len(data)); i++ {
		d.host.a.mmu.Poke(dest+i, data[i])
	}

	return smartPortNoError
}

func (d *SmartPortHardDisk) writeBlock(block uint32, source uint16) uint8 {
	if d.trace {
		fmt.Printf("[SmartPortHardDisk] Write block %v from $%x.\n", block, source)
	}

	if d.disk.IsReadOnly() {
		return smartPortErrorWriteProtected
	}

	// Byte by byte transfer from memory using the full Peek code path
	buf := make([]uint8, storage.ProDosBlockSize)
	for i := uint16(0); i < uint16(len(buf)); i++ {
		buf[i] = d.host.a.mmu.Peek(source + i)
	}

	err := d.disk.Write(block, buf)
	if err != nil {
		return smartPortErrorIO
	}

	return smartPortNoError
}

func (d *SmartPortHardDisk) status(dest uint16) uint8 {
	if d.trace {
		fmt.Printf("[SmartPortHardDisk] Status into $%x.\n", dest)
	}

	// See http://www.1000bit.it/support/manuali/apple/technotes/smpt/tn.smpt.2.html
	d.host.a.mmu.Poke(dest+0, 0x01) // One device
	d.host.a.mmu.Poke(dest+1, 0xff) // No interrupt
	d.host.a.mmu.Poke(dest+2, 0x00)
	d.host.a.mmu.Poke(dest+3, 0x00) // Unknown manufacturer
	d.host.a.mmu.Poke(dest+4, 0x01)
	d.host.a.mmu.Poke(dest+5, 0x00) // Version 1.0 final
	d.host.a.mmu.Poke(dest+6, 0x00)
	d.host.a.mmu.Poke(dest+7, 0x00) // Reserved

	return smartPortNoError
}
