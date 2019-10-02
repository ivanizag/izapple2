package apple2

/*
To implement a hard drive we just have to support boot from #PR7 and the PRODOS expextations.
See Beneath Prodos, section 6-6, 7-13 and 5-8. (http://www.apple-iigs.info/doc/fichiers/beneathprodos.pdf)
*/

type cardHardDisk struct {
	cardBase
	disk *hardDisk
}

func buildHardDiskRom(slot int) []uint8 {
	data := make([]uint8, 256)
	ssBase := 0x80 + uint8(slot<<4)

	copy(data, []uint8{
		// Preamble bytes to comply with the expectation in $Cn01, 3, 5 and 7
		0xa9, 0x20, // LDA #$20
		0xa9, 0x00, // LDA #$20
		0xa9, 0x03, // LDA #$20
		0xa9, 0x3c, // LDA #$3c

		// Boot code: SS will load block 0 in address $0800. The jump there.
		// Note: on execution the first block expects $42 to $47 to have
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
		0xa2, uint8(slot << 4), // LDX $s7 ;Slot on hign nibble of X
		0x4c, 0x01, 0x08, // JMP $801
	})

	copy(data[0x80:], []uint8{
		0xad, ssBase + 0, 0xc0, // LDA $C0n0 ; Softswitch 0, execute command. Error code in reg A.
		0xae, ssBase + 1, 0xc0, // LDX $C0n1 ; Softswitch 2, LO(Blocks), STATUS needs that in reg X.
		0xac, ssBase + 2, 0xc0, // LDY $C0n2 ; Softswitch 3, HI(Blocks). STATUS needs that in reg Y.
		0x18, // CLC ; Clear carry for no errors.
		0x60, // RTS
	})

	data[0xfc] = 0
	data[0xfd] = 0
	data[0xfe] = 3    // Status and Read. No write, no format. Single volume
	data[0xff] = 0x80 // Driver entry point

	return data
}

const (
	proDosDeviceCommandStatus = 0
	proDosDeviceCommandRead   = 1
	proDosDeviceCommandWrite  = 2
	proDosDeviceCommandFormat = 3
)

const (
	proDosDeviceNoError             = 0
	proDosDeviceErrorIO             = 0x27
	proDosDeviceErrorNoDevice       = 0x28
	proDosDeviceErrorWriteProtected = 0x2b
)

func (c *cardHardDisk) assign(a *Apple2, slot int) {
	c.ssr[0] = func(*ioC0Page) uint8 {
		// Prodos entry point
		command := a.mmu.Peek(0x42)
		//unit := a.mmu.Peek(0x43)
		dest := uint16(a.mmu.Peek(0x44)) + uint16(a.mmu.Peek(0x45))<<8
		block := uint16(a.mmu.Peek(0x46)) + uint16(a.mmu.Peek(0x47))<<8
		//fmt.Printf("[CardHardDisk] Command %v on unit $%x, block %v to $%x.\n", command, unit, block, dest)

		switch command {
		case proDosDeviceCommandStatus:
			return proDosDeviceNoError
		case proDosDeviceCommandRead:
			c.readBlock(block, dest)
			return proDosDeviceNoError
		default:
			panic("Prodos device command not supported.")
		}
	}
	c.ssr[1] = func(*ioC0Page) uint8 {
		// Blocks available, low byte
		return uint8(c.disk.header.Blocks)
	}
	c.ssr[2] = func(*ioC0Page) uint8 {
		// Blocks available, high byte
		return uint8(c.disk.header.Blocks >> 8)
	}

	c.cardBase.assign(a, slot)
}

func (c *cardHardDisk) readBlock(block uint16, dest uint16) {
	//fmt.Printf("[CardHardDisk] Read block %v into $%x.\n", block, dest)

	data := c.disk.read(uint32(block))
	// Byte by byte transfer to memory using the full Poke code path
	for i := uint16(0); i < uint16(proDosBlockSize); i++ {
		c.a.mmu.Poke(dest+i, data[i])
	}

}

func (c *cardHardDisk) addDisk(disk *hardDisk) {
	c.disk = disk
}
