package izapple2

import (
	"github.com/ivanizag/izapple2/component"
)

// TODO: fast mode

/*
See:
	"Understanding the Apple II, chapter 9"
	Beneath Apple ProDOS, Appendix
	Phases: http://yesterbits.com/media/pubs/AppleOrchard/articles/disk-ii-part-1-1983-apr.pdf
	IMW Floppy Disk I/O Controller info: (https://www.brutaldeluxe.fr/documentation/iwm/apple2_IWM_INFO_19840510.pdf)
	Woz https://applesaucefdc.com/woz/reference2/
	Schematic: https://mirrors.apple2.org.za/ftp.apple.asimov.net/documentation/hardware/schematics/APPLE_DiskII_SCH.pdf


*/

// CardDisk2Sequencer is a DiskII interface card with the Woz state machine
type CardDisk2Sequencer struct {
	cardBase

	sectors13 bool
	p6ROM     []uint8

	q          [8]bool // 8-bit latch SN74LS259
	register   uint8   // 8-bit shift/storage register SN74LS323
	sequence   uint8   // 4 bits stored in an hex flip-flop SN74LS174
	motorDelay uint64  // NE556 timer, used to delay motor off
	drive      [2]cardDisk2SequencerDrive

	lastWriteValue  bool  // We write transitions to the WOZ file. We store the last value to send a pulse on change.
	lastPulseCycles uint8 // There is a new pulse every 4ms, that's 8 cycles of 2Mhz

	lastCycle uint64 // 2 Mhz cycles

	trackTracer trackTracer
}

// Shared methods between both versions on the Disk II card
type cardDisk2Shared interface {
	setTrackTracer(tt trackTracer)
}

const (
	disk2MotorOffDelay = uint64(2 * 1000 * 1000) // 2 Mhz cycles. Total 1 second.
	disk2PulseCyles    = uint8(8)                // 8 cycles = 4ms * 2Mhz

	/*
	   We skip register calculations for long periods with the motor
	   on but not reading bytes. It's an optimizations, 10000 is too
	   short for cross track sync copy protections.
	*/
	disk2CyclestoLoseSsync = 100000
)

func newCardDisk2SequencerBuilder() *cardBuilder {
	return &cardBuilder{
		name:        "Disk II Sequencer",
		description: "Disk II interface card emulating the Woz state machine",
		defaultParams: &[]paramSpec{
			{"disk1", "Diskette image for drive 1", ""},
			{"disk2", "Diskette image for drive 2", ""},
			{"tracktracer", "Trace how the disk head moves between tracks", "false"},
		},
		buildFunc: func(params map[string]string) (Card, error) {
			var c CardDisk2Sequencer

			disk1 := paramsGetString(params, "disk1")
			if disk1 != "" {
				err := c.drive[0].insertDiskette(disk1)
				if err != nil {
					return nil, err
				}
				c.sectors13 = c.drive[0].data.Info.BootSectorFormat == 2 // Woz 13 sector disk
			}

			disk2 := paramsGetString(params, "disk2")
			if disk2 != "" {
				err := c.drive[1].insertDiskette(disk2)
				if err != nil {
					return nil, err
				}
			}

			P5RomFile := "<internal>/Apple Disk II 16 Sector Interface Card ROM P5 - 341-0027.bin"
			P6RomFile := "<internal>/Apple Disk II 16 Sector Interface Card ROM P6 - 341-0028.bin"
			if c.sectors13 {
				P5RomFile = "<internal>/Apple Disk II 13 Sector Interface Card ROM P5 - 341-0009.bin"
				// Buggy sequencer not need for 13 sectors disks to work
				// P6RomFile = "<internal>/Apple Disk II 13 Sector Interface Card ROM P6 - 341-0010.bin"
			}

			err := c.loadRomFromResource(P5RomFile, cardRomSimple)
			if err != nil {
				return nil, err
			}

			data, _, err := LoadResource(P6RomFile)
			if err != nil {
				return nil, err
			}
			c.p6ROM = data

			trackTracer := paramsGetBool(params, "tracktracer")
			if trackTracer {
				c.trackTracer = makeTrackTracerLogger()
			}
			return &c, nil
		},
	}
}

// GetInfo returns card info
func (c *CardDisk2Sequencer) GetInfo() map[string]string {
	info := make(map[string]string)
	if c.sectors13 {
		info["rom"] = "13 sector"
	} else {
		info["rom"] = "16 sector"
	}
	// TODO: add drives info
	return info
}

func (c *CardDisk2Sequencer) reset() {
	// UtA2e 9-12, all switches forced to off
	c.q = [8]bool{}
}

func (c *CardDisk2Sequencer) setTrackTracer(tt trackTracer) {
	c.trackTracer = tt
}

func (c *CardDisk2Sequencer) assign(a *Apple2, slot int) {
	a.registerRemovableMediaDrive(&c.drive[0])
	a.registerRemovableMediaDrive(&c.drive[1])

	c.addCardSoftSwitches(func(address uint8, data uint8, _ bool) uint8 {
		/*
			Slot card pins to SN74LS259 latch mapping:
				slot_address[3,2,1] => latch_address[2,1,0]
				slot_address[0] => latch_data
				slot_dev_selct =>  latch_write_enable ;It will be true
		*/
		c.q[address>>1] = (address & 1) != 0

		// Advance the Disk2 state machine since the last call to softswitches
		c.catchUp(data)

		/*
			Slot card pins to SN74LS259 mapping:
				slot_address[0] => latch_oe2_n
		*/
		registerOutputEnableNeg := (address & 1) != 0
		if !registerOutputEnableNeg {
			return c.register
		} else {
			return 33 // Floating
		}
	}, "DISK2SEQ")

	c.cardBase.assign(a, slot)
}

func (c *CardDisk2Sequencer) catchUp(data uint8) {
	currentCycle := c.a.GetCycles() << 1 // Disk2 cycles are x2 cpu cycle

	motorOn := c.step(data, true)
	if motorOn && currentCycle > c.lastCycle+disk2CyclestoLoseSsync {
		// We have lost sync. We start the count.
		// We do at least a couple 2 Mhz cycles
		c.lastCycle = currentCycle - 2
	}
	c.lastCycle++

	for motorOn && c.lastCycle <= currentCycle {
		motorOn = c.step(data, false)
		c.lastCycle++
	}

	if !motorOn {
		c.lastCycle = 0 // Sync lost
	}
}

func (c *CardDisk2Sequencer) step(data uint8, firstStep bool) bool {
	/*
		Q4 and Q6 set on the sofswitches is stored on the
		latch.
	*/
	q5 := c.q[5] // Drive selection
	q4 := c.q[4] // Motor on (before delay)

	/*
		Motor On comes from the latched q4 via the 556 to
		provide a delay. The delay is reset while q4 is on.
	*/
	if q4 {
		c.motorDelay = disk2MotorOffDelay
	}
	motorOn := c.motorDelay > 0

	/*
		The pins for the cable drives ENBL1 and ENBL2 are
		connected to q5 and motor using half of the 74LS132
		NAND to combine them.
	*/
	c.drive[0].enable(!q5 && motorOn)
	c.drive[1].enable(q5 && motorOn)

	/*
		Motor on AND the 2 Mhz clock (Q3 pin 37 of the slot)
		are connected to the clok pulse of the shift register
		if off, the sequences does not advance. The and uses
		another quarter of the 74LS132 NAND.
	*/
	if !motorOn {
		c.sequence = 0
		return false
	}
	c.motorDelay--

	/*
		Head movements. We assume it's instantaneous on Q0-Q3 change. We
		will place it on the first step.
		Q0 to Q3 are connected directly to the drives.
	*/
	if firstStep {
		q0 := c.q[0]
		q1 := c.q[1]
		q2 := c.q[2]
		q3 := c.q[3]
		c.drive[0].moveHead(q0, q1, q2, q3, c.trackTracer, c.slot, 0)
		c.drive[1].moveHead(q0, q1, q2, q3, c.trackTracer, c.slot, 1)
	}

	/*
		The reading from the drive is converted to a pulse detecting
		changes using Q3 and Q4 of the flip flop, combined with
		the last quarter of the 74LS132 NAND.∫
		The woz format provides the pulse directly and we won't emulate
		this detection.
	*/
	pulse := false
	c.lastPulseCycles++
	if c.lastPulseCycles == disk2PulseCyles {
		// Read
		pulse = c.drive[0].readPulse() ||
			c.drive[1].readPulse()
		c.lastPulseCycles = 0
	}

	/*
		The write protected signal comes directly from any of the
		drives being enabled (motor on) and write protected.
	*/
	wProt := (c.drive[0].enabled && c.drive[0].writeProtected) ||
		(c.drive[1].enabled && c.drive[1].writeProtected)

	/*
		The next instruction for the sequencer is retrieved from
		the ROM P6 using the address:
			A0, A5, A6, A7 <= sequence from 74LS174
			A1 =< high, MSB of register (pin Q7)
			A2 <= Q6 from 9334
			A3 <= Q7 from 9334
			A4 <= pulse transition
	*/
	high := c.register >= 0x80
	seqBits := component.ByteToPins(c.sequence)
	romAddress := component.PinsToByte([8]bool{
		seqBits[1], // seq1
		high,
		c.q[6],
		c.q[7],
		!pulse,
		seqBits[0], // seq0
		seqBits[2], // seq2
		seqBits[3], // seq3
	})

	romData := c.p6ROM[romAddress]
	inst := romData & 0xf
	next := romData >> 4

	/*
		The pins for the register shifter update are:
			SR(CLR) <- ROM D3
			S1      <- ROM D0
			S0      <- ROM D1
			DS0(SR) <- WPROT pin of the selected drive
			DS7(SL) <- ROM D2
			IO[7.0] <-> D[0-7] slot data bus (the order is reversed)

	*/
	if inst < 8 {
		c.register = 0 // Bit 4 clear to reset
	} else {
		switch inst & 0x3 { // Bit 0 and 1 are the operation
		case 0:
			// Nothing
		case 1:
			// Shift left bringing bit 1
			c.register = (c.register << 1) | ((inst >> 2) & 1)
		case 2:
			// Shift right bringing wProt
			c.register >>= 1
			if wProt {
				c.register |= 0x80
			}
		case 3:
			// Load
			c.register = data
		}

		if c.q[7] && (inst&0x3) != 0 {
			currentWriteValue := next >= 0x8
			writePulse := currentWriteValue != c.lastWriteValue
			c.drive[0].writePulse(writePulse)
			c.drive[1].writePulse(writePulse)
			c.lastWriteValue = currentWriteValue

		}
	}

	// fmt.Printf("[D2SEQ] Step. seq:%x inst:%x next:%x reg:%02x\n",
	//	c.sequence, inst, next, c.register)

	c.sequence = next
	return true
}
