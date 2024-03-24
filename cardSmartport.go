package izapple2

import (
	"fmt"
	"strconv"
)

/*
To implement a hard drive we just have to support boot from #PR7 and the PRODOS expectations.

See:
	Beneath Prodos, section 6-6, 7-13 and 5-8. (http://www.apple-iigs.info/doc/fichiers/beneathprodos.pdf)
	Apple IIc Technical Reference, 2nd Edition. Chapter 8. https://ia800207.us.archive.org/19/items/AppleIIcTechnicalReference2ndEd/Apple%20IIc%20Technical%20Reference%202nd%20ed.pdf
	https://prodos8.com/docs/technote/21/
	https://prodos8.com/docs/technote/20/


*/

// CardSmartPort represents a SmartPort card
type CardSmartPort struct {
	cardBase
	devices        []smartPortDevice
	hardDiskBlocks uint32

	mliParams uint16
	trace     bool
}

func newCardSmartPortStorageBuilder() *cardBuilder {
	return &cardBuilder{
		name:        "SmartPort",
		description: "SmartPort interface card",
		defaultParams: &[]paramSpec{
			{"image1", "Disk image for unit 1", ""},
			{"image2", "Disk image for unit 2", ""},
			{"image3", "Disk image for unit 3", ""},
			{"image4", "Disk image for unit 4", ""},
			{"image5", "Disk image for unit 5", ""},
			{"image6", "Disk image for unit 6", ""},
			{"image7", "Disk image for unit 7", ""},
			{"image8", "Disk image for unit 8", ""},
			{"tracesp", "Trace SmartPort calls", "false"},
			{"tracehd", "Trace image accesses", "false"},
		},
		buildFunc: func(params map[string]string) (Card, error) {
			var c CardSmartPort
			c.trace = paramsGetBool(params, "tracesp")
			traceHD := paramsGetBool(params, "tracehd")
			for i := 1; i <= 8; i++ {
				image := paramsGetPath(params, "image"+strconv.Itoa(i))
				if image != "" {
					err := c.LoadImage(image, traceHD)
					if err != nil {
						return nil, err
					}
				}
			}
			return &c, nil
		},
	}
}

func newCardSmartPortFujinetBuilder() *cardBuilder {
	return &cardBuilder{
		name:        "Fujinet",
		description: "SmartPort interface card hosting the Fujinet",
		defaultParams: &[]paramSpec{
			{"tracesp", "Trace SmartPort calls", "false"},
			{"tracenet", "Trace on the network device", "false"},
			{"traceclock", "Trace on the clock device", "false"},
		},
		buildFunc: func(params map[string]string) (Card, error) {
			var c CardSmartPort
			c.trace = paramsGetBool(params, "tracesp")

			net := NewSmartPortFujinetNetwork(&c)
			net.trace = paramsGetBool(params, "tracenet")
			c.AddDevice(net)

			clock := NewSmartPortFujinetClock(&c)
			clock.trace = paramsGetBool(params, "traceclock")
			c.AddDevice(clock)

			return &c, nil
		},
	}
}

// GetInfo returns smartPort info
func (c *CardSmartPort) GetInfo() map[string]string {
	info := make(map[string]string)
	info["trace"] = strconv.FormatBool(c.trace)
	return info
}

// LoadImage loads a disk image
func (c *CardSmartPort) LoadImage(filename string, trace bool) error {
	device, err := NewSmartPortHardDisk(c, filename)
	if err == nil {
		device.trace = trace
		c.devices = append(c.devices, device)
		c.hardDiskBlocks = device.disk.GetSizeInBlocks() // Needed for the PRODOS status
	}
	return err
}

// LoadImage loads a disk image
func (c *CardSmartPort) AddDevice(device smartPortDevice) {
	c.devices = append(c.devices, device)
	c.hardDiskBlocks = 0 // Needed for the PRODOS status
}

func (c *CardSmartPort) assign(a *Apple2, slot int) {
	c.loadRom(buildHardDiskRom(slot), cardRomSimple)

	c.addCardSoftSwitchR(0, func() uint8 {
		// Prodos entry point
		command := a.mmu.Peek(0x42)
		unit := a.mmu.Peek(0x43) & 0x0f

		// Generate Smarport compatible params
		var call *smartPortCall
		if command == smartPortCommandStatus {
			call = newSmartPortCallSynthetic(c, command, []uint8{
				3, // 3 args
				unit,
				a.mmu.Peek(0x44), a.mmu.Peek(0x45), // data address
				0,
			})
		} else if command == smartPortCommandReadBlock || command == smartPortCommandWriteBlock {
			call = newSmartPortCallSynthetic(c, command, []uint8{
				3, // 3args
				unit,
				a.mmu.Peek(0x44), a.mmu.Peek(0x45), // data address
				a.mmu.Peek(0x46), a.mmu.Peek(0x47), 0, // block number
			})
		} else {
			return smartPortBadCommand
		}

		return c.exec(call)
	}, "SMARTPORTPRODOSCOMMAND")

	c.addCardSoftSwitchR(1, func() uint8 {
		// Blocks available, low byte
		return uint8(c.hardDiskBlocks)
	}, "HDBLOCKSLO")
	c.addCardSoftSwitchR(2, func() uint8 {
		// Blocks available, high byte
		return uint8(c.hardDiskBlocks >> 8)
	}, "HDBLOCKHI")

	c.addCardSoftSwitchR(3, func() uint8 {
		// Smart port entry point
		command := c.a.mmu.Peek(c.mliParams + 1)
		paramsAddress := uint16(c.a.mmu.Peek(c.mliParams+2)) + uint16(c.a.mmu.Peek(c.mliParams+3))<<8

		call := newSmartPortCall(c, command, paramsAddress)
		return c.exec(call)
	}, "SMARTPORTEXEC")

	c.addCardSoftSwitchW(4, func(value uint8) {
		c.mliParams = (c.mliParams & 0xff00) + uint16(value)
	}, "HDSMARTPORTLO")
	c.addCardSoftSwitchW(5, func(value uint8) {
		c.mliParams = (c.mliParams & 0x00ff) + (uint16(value) << 8)
	}, "HDSMARTPORTHI")

	c.cardBase.assign(a, slot)
}

func (c *CardSmartPort) exec(call *smartPortCall) uint8 {
	var result uint8
	unit := int(call.unit())

	if call.command == smartPortCommandStatus &&
		// Call to the host
		call.statusCode() == smartPortStatusCodeDevice {

		result = c.hostStatus(call)
	} else if unit > len(c.devices) {
		result = smartPortErrorNoDevice
	} else {
		if unit == 0 {
			unit = 1 // For unit 0(host) use the first device
		}

		if unit > len(c.devices) {
			result = smartPortErrorNoDevice
		} else {
			result = c.devices[unit-1].exec(call)
		}
	}

	if c.trace {
		fmt.Printf("[CardSmartPort] Command %v on slot %v => result %s.\n",
			call, c.slot, smartPortErrorMessage(result))
	}
	return result
}

func (c *CardSmartPort) hostStatus(call *smartPortCall) uint8 {
	dest := call.param16(2)
	if c.trace {
		fmt.Printf("[CardSmartPort] Host status into $%x.\n", dest)
	}

	// See http://www.1000bit.it/support/manuali/apple/technotes/smpt/tn.smpt.2.html
	c.a.mmu.Poke(dest+0, uint8(len(c.devices)))
	c.a.mmu.Poke(dest+1, 0xff) // No interrupt
	c.a.mmu.Poke(dest+2, 0x00)
	c.a.mmu.Poke(dest+3, 0x00) // Unknown manufacturer
	c.a.mmu.Poke(dest+4, 0x01)
	c.a.mmu.Poke(dest+5, 0x00) // Version 1.0 final
	c.a.mmu.Poke(dest+6, 0x00)
	c.a.mmu.Poke(dest+7, 0x00) // Reserved

	return smartPortNoError
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
		0xd0, 0x36, // BNE bootcode, there is no space for a jmp
	})

	if slot == 7 {
		// It should be 0 for SmartPort, but with 0 it's not bootable with the II+ ROM
		// See http://www.1000bit.it/support/manuali/apple/technotes/udsk/tn.udsk.2.html
		data[0x07] = 0x3c
	}

	copy(data[0x0a:], []uint8{
		// Entrypoints and SmartPort body it has to be in $Cx0a
		0x4c, 0x80, 0xc0 + uint8(slot), // JMP $cs80 ; Prodos Entrypoint

		// 3 bytes later, smartPort entrypoint. Uses the ProDos MLI calling convention
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

	copy(data[0x40:], []uint8{
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
	data[0xff] = 0x0a // Driver entry point  // Must be $0a

	return data
}
