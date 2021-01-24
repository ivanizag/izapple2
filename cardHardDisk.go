package izapple2

import (
	"fmt"
	"strconv"

	"github.com/ivanizag/izapple2/storage"
)

/*
To implement a hard drive we just have to support boot from #PR7 and the PRODOS expextations.

See:
	Beneath Prodos, section 6-6, 7-13 and 5-8. (http://www.apple-iigs.info/doc/fichiers/beneathprodos.pdf)
	Apple IIc Technical Reference, 2nd Edition. Chapter 8. https://ia800207.us.archive.org/19/items/AppleIIcTechnicalReference2ndEd/Apple%20IIc%20Technical%20Reference%202nd%20ed.pdf
	https://prodos8.com/docs/technote/21/
	https://prodos8.com/docs/technote/20/


*/

// CardHardDisk represents a SmartPort card
type CardHardDisk struct {
	cardBase

	filename string
	trace    bool

	disk      *storage.BlockDisk
	mliParams uint16
}

// NewCardHardDisk creates a new SmartPort card
func NewCardHardDisk() *CardHardDisk {
	var c CardHardDisk
	c.name = "Smartport Card"
	return &c
}

// GetInfo returns smartport info
func (c *CardHardDisk) GetInfo() map[string]string {
	info := make(map[string]string)
	info["filename"] = c.filename
	info["trace"] = strconv.FormatBool(c.trace)
	return info
}

// LoadImage loads a disk image
func (c *CardHardDisk) LoadImage(filename string) error {
	hd, err := storage.OpenBlockDisk(filename)
	if err != nil {
		return err
	}
	c.disk = hd
	c.filename = filename
	return nil
}

const (
	proDosDeviceCommandStatus = 0
	proDosDeviceCommandRead   = 1
	proDosDeviceCommandWrite  = 2
	proDosDeviceCommandFormat = 3
)

const (
	proDosDeviceNoError             = uint8(0)
	proDosDeviceErrorIO             = uint8(0x27)
	proDosDeviceErrorNoDevice       = uint8(0x28)
	proDosDeviceErrorWriteProtected = uint8(0x2b)
)

func (c *CardHardDisk) assign(a *Apple2, slot int) {
	c.loadRom(buildHardDiskRom(slot))

	c.addCardSoftSwitchR(0, func(*ioC0Page) uint8 {
		// Prodos entry point
		command := a.mmu.Peek(0x42)
		unit := a.mmu.Peek(0x43)
		address := uint16(a.mmu.Peek(0x44)) + uint16(a.mmu.Peek(0x45))<<8
		block := uint16(a.mmu.Peek(0x46)) + uint16(a.mmu.Peek(0x47))<<8
		if c.trace {
			fmt.Printf("[CardHardDisk] Prodos command %v on slot %v, unit $%x, block %v to $%x.\n", command, slot, unit, block, address)
		}

		switch command {
		case proDosDeviceCommandStatus:
			return c.status(unit, address)
		case proDosDeviceCommandRead:
			return c.readBlock(block, address)
		case proDosDeviceCommandWrite:
			return c.writeBlock(block, address)
		default:
			// Prodos device command not supported
			return proDosDeviceErrorIO
		}
	}, "HDCOMMAND")
	c.addCardSoftSwitchR(1, func(*ioC0Page) uint8 {
		// Blocks available, low byte
		return uint8(c.disk.GetSizeInBlocks())
	}, "HDBLOCKSLO")
	c.addCardSoftSwitchR(2, func(*ioC0Page) uint8 {
		// Blocks available, high byte
		return uint8(c.disk.GetSizeInBlocks() >> 8)
	}, "HDBLOCKHI")

	c.addCardSoftSwitchR(3, func(*ioC0Page) uint8 {
		// Smart port entry point
		command := c.a.mmu.Peek(c.mliParams + 1)
		paramsAddress := uint16(c.a.mmu.Peek(c.mliParams+2)) + uint16(c.a.mmu.Peek(c.mliParams+3))<<8
		unit := a.mmu.Peek(paramsAddress + 1)
		address := uint16(a.mmu.Peek(paramsAddress+2)) + uint16(a.mmu.Peek(paramsAddress+3))<<8
		block := uint16(a.mmu.Peek(paramsAddress+4)) + uint16(a.mmu.Peek(paramsAddress+5))<<8
		if c.trace {
			fmt.Printf("[CardHardDisk] Smart port command %v on slot %v, unit $%x, block %v to $%x.\n", command, slot, unit, block, address)
		}

		switch command {
		case proDosDeviceCommandStatus:
			return c.status(unit, address)
		case proDosDeviceCommandRead:
			return c.readBlock(block, address)
		case proDosDeviceCommandWrite:
			return c.writeBlock(block, address)
		default:
			// Smartport device command not supported
			return proDosDeviceErrorIO
		}
	}, "HDSMARTPORT")
	c.addCardSoftSwitchW(4, func(_ *ioC0Page, value uint8) {
		c.mliParams = (c.mliParams & 0xff00) + uint16(value)
		if c.trace {
			fmt.Printf("[CardHardDisk] Smart port LO: 0x%x.\n", c.mliParams)
		}
	}, "HDSMARTPORTLO")
	c.addCardSoftSwitchW(5, func(_ *ioC0Page, value uint8) {
		c.mliParams = (c.mliParams & 0x00ff) + (uint16(value) << 8)
		if c.trace {
			fmt.Printf("[CardHardDisk] Smart port HI: 0x%x.\n", c.mliParams)
		}
	}, "HDSMARTPORTHI")

	c.cardBase.assign(a, slot)
}

func (c *CardHardDisk) readBlock(block uint16, dest uint16) uint8 {
	if c.trace {
		fmt.Printf("[CardHardDisk] Read block %v into $%x.\n", block, dest)
	}

	data, err := c.disk.Read(uint32(block))
	if err != nil {
		return proDosDeviceErrorIO
	}
	// Byte by byte transfer to memory using the full Poke code path
	for i := uint16(0); i < uint16(len(data)); i++ {
		c.a.mmu.Poke(dest+i, data[i])
	}

	return proDosDeviceNoError
}

func (c *CardHardDisk) writeBlock(block uint16, source uint16) uint8 {
	if c.trace {
		fmt.Printf("[CardHardDisk] Write block %v from $%x.\n", block, source)
	}

	if c.disk.IsReadOnly() {
		return proDosDeviceErrorWriteProtected
	}

	// Byte by byte transfer from memory using the full Peek code path
	buf := make([]uint8, storage.ProDosBlockSize)
	for i := uint16(0); i < uint16(len(buf)); i++ {
		buf[i] = c.a.mmu.Peek(source + i)
	}

	err := c.disk.Write(uint32(block), buf)
	if err != nil {
		return proDosDeviceErrorIO
	}

	return proDosDeviceNoError
}

func (c *CardHardDisk) status(unit uint8, dest uint16) uint8 {
	if c.trace {
		fmt.Printf("[CardHardDisk] Status for %v into $%x.\n", unit, dest)
	}

	// See http://www.1000bit.it/support/manuali/apple/technotes/smpt/tn.smpt.2.html
	c.a.mmu.Poke(dest+0, 0x02) // One device
	c.a.mmu.Poke(dest+1, 0xff) // No interrupt
	c.a.mmu.Poke(dest+2, 0x00)
	c.a.mmu.Poke(dest+3, 0x00) // Unknown manufacturer
	c.a.mmu.Poke(dest+4, 0x01)
	c.a.mmu.Poke(dest+5, 0x00) // Version 1.0 final
	c.a.mmu.Poke(dest+6, 0x00)
	c.a.mmu.Poke(dest+7, 0x00) // Reserved

	return proDosDeviceNoError
}

func buildHardDiskRom(slot int) []uint8 {
	data := make([]uint8, 256)
	ssBase := 0x80 + uint8(slot<<4)

	copy(data, []uint8{
		// Preamble bytes to comply with the expectation in $Cn01, 3, 5 and 7
		0xa9, 0x20, // LDA #$20
		0xa9, 0x00, // LDA #$00
		0xa9, 0x03, // LDA #$03
		0xa9, 0x00, // LDA #$00

		// Boot code: SS will load block 0 in address $0800. The jump there.
		// Note: after execution the first block expects $42 to $47 to have
		// valid values to read block 0. At least Total Replay expects that.
		0xa9, 0x01, // LDA·#$01
		0x85, 0x42, // STA $42 ; Command READ(1)
		0xa9, 0x00, // LDA·#$00
		0x85, 0x43, // STA $43 ; Unit 0
		0x85, 0x44, // STA $44 ; Dest LO($0800)
		0x85, 0x46, // STA $46 ; Block LO(0)
		0x85, 0x47, // STA $47 ; Block HI(0)
		0xa9, 0x08, // LDA·#$08
		0x85, 0x45, // STA $45 ; Dest HI($0800)

		0xad, ssBase, 0xc0, // LDA $C0n1 ;Call to softswitch 0.
		0xa2, uint8(slot << 4), // LDX $s7 ; Slot on hign nibble of X
		0x4c, 0x01, 0x08, // JMP $801 ; Jump to loaded boot sector
	})

	if slot == 7 {
		// It should be 0 for SmartPort, but with 0 it's not bootable with the II+ ROM
		// See http://www.1000bit.it/support/manuali/apple/technotes/udsk/tn.udsk.2.html
		data[0x07] = 0x3c
	}

	// Entrypoints and Smartport body
	copy(data[0x40:], []uint8{
		0x4c, 0x80, 0xc0 + uint8(slot), // JMP $cs80 ; Prodos Entrypoint

		// 3 bytes later, smartport entrypoint. Uses the ProDos MLI calling convention
		0x68,                   // PLA
		0x8d, ssBase + 4, 0xc0, // STA $c0n4 ; Softswitch 4, store LO(cmdBlock)
		0xa8,                   // TAY ; We will need it later
		0x68,                   // PLA
		0x8d, ssBase + 5, 0xc0, // STA $c0n5 ; Softswitch 5, store HI(cmdBlock)
		0x48,       // PHA
		0x98,       // TYA
		0x18,       // CLC
		0x69, 0x03, // ADC #$03 ; Fix return address past the cmdblock
		0x48,                   // PHA
		0xad, ssBase + 3, 0xc0, // LDA $C0n3 ; Softswitch 3, execute command. Error code in reg A.
		0x18,       // CLC ; Clear carry for no errors.
		0xF0, 0x01, // BEQ $01 ; Skips the SEC if reg A is zero
		0x38, // SEC ; Set carry on errors
		0x60, // RTS
	})

	// Prodos entrypoint body
	copy(data[0x80:], []uint8{
		0xad, ssBase + 0, 0xc0, // LDA $C0n0 ; Softswitch 0, execute command. Error code in reg A.
		0x48,                   // PHA
		0xae, ssBase + 1, 0xc0, // LDX $C0n1 ; Softswitch 1, LO(Blocks), STATUS needs that in reg X.
		0xac, ssBase + 2, 0xc0, // LDY $C0n2 ; Softswitch 2, HI(Blocks). STATUS needs that in reg Y.
		0x18,       // CLC ; Clear carry for no errors.
		0x68,       // PLA ; Sets Z if no error
		0xF0, 0x01, // BEQ $01 ; Skips the SEC if reg A is zero
		0x38, // SEC ; Set carry on errors
		0x60, // RTS
	})

	data[0xfc] = 0
	data[0xfd] = 0
	data[0xfe] = 3    // Status and Read. No write, no format. Single volume
	data[0xff] = 0x40 // Driver entry point

	return data
}
