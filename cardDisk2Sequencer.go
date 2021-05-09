package izapple2

import (
	"fmt"

	"github.com/ivanizag/izapple2/component"
	"github.com/ivanizag/izapple2/storage"
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

// CardDisk2Sequencer is a DiskII interface card
type CardDisk2Sequencer struct {
	cardBase

	p6ROM      []uint8
	latch      component.SN74LS259 // 8-bit latch
	register   uint8               // 8-bit shift/storage register SN74LS323
	sequence   uint8               // 4 bits stored in an hex flip-flop SN74LS174
	motorDelay uint8               // NE556 timer, used to delay motor off
	drive      [2]cardDisk2SequencerDrive

	lastPulse       bool
	lastPulseCycles uint8 // There is a new pulse every 4ms, that's 8 cycles of 2Mhz

	lastCycle uint64 // 2 Mhz cycles
}

const disk2MotorOffDelay = uint8(20) // 2 Mhz cycles, TODO: how long?
const disk2PulseCcyles = uint8(8)    // 8 cycles = 4ms * 2Mhz

// NewCardDisk2Sequencer creates a new CardDisk2Sequencer
func NewCardDisk2Sequencer() *CardDisk2Sequencer {
	var c CardDisk2Sequencer
	c.name = "Disk II"
	c.loadRomFromResource("<internal>/DISK2.rom")

	data, _, err := storage.LoadResource("<internal>/DISK2P6.rom")
	if err != nil {
		// The resource should be internal and never fail
		panic(err)
	}
	c.p6ROM = data

	return &c
}

func (c *CardDisk2Sequencer) dumpState() {
	fmt.Printf("Q5 %v, Q4 %v, delay %v\n", c.latch.Q(5), c.latch.Q(4), c.motorDelay)
}

// GetInfo returns card info
func (c *CardDisk2Sequencer) GetInfo() map[string]string {
	info := make(map[string]string)
	info["rom"] = "16 sector"
	// TODO: add drives info
	return info
}

func (c *CardDisk2Sequencer) reset() {
	// UtA2e 9-12, all switches forced to off
	c.latch.Reset()
}

func (c *CardDisk2Sequencer) assign(a *Apple2, slot int) {
	c.addCardSoftSwitches(func(_ *ioC0Page, address uint8, data uint8, write bool) uint8 {
		/*
			Slot card pins to SN74LS259 mapping:
				slot_address[3,2,1] => latch_address[2,1,0]
				slot_address[0] => latch_data
				slot_dev_selct =>  latch_write_enable // It will be true
		*/
		c.latch.Write(address>>1, (address&1) != 0, true)

		// Advance the Disk2 state machine since the last call to softswitches
		c.catchUp(data)
		//c.dumpState()
		/*
			Slot card pins to SN74LS259 mapping:
				slot_address[0] => latch_oe2_n
		*/
		register_output_enable_neg := (address & 1) != 0
		if !register_output_enable_neg {
			//if c.register >= 0x80 && address == 0xc {
			//	fmt.Printf("Byte %x\n", c.register)
			//}
			return c.register
		} else {
			return 33
		}
	}, "DISK2SEQ")

	c.cardBase.assign(a, slot)
}

func (c *CardDisk2Sequencer) catchUp(data uint8) {
	currentCycle := c.a.cpu.GetCycles() << 1 // Disk2 cycles are x2 cpu cycle

	motorOn := c.step(data, true)
	//if motorOn && c.lastCycle == 0 {
	if motorOn && currentCycle > c.lastCycle+100000 { // With 10000, cross track snc not working
		// The motor was off, now on. We start the count. We do at least a couple 2 Mhz cycles
		c.lastCycle = currentCycle - 2
	}
	c.lastCycle++

	for motorOn && c.lastCycle <= currentCycle {
		motorOn = c.step(data, false)
		c.lastCycle++
	}

	if !motorOn {
		c.lastCycle = 0 // No tracking done
	}
}

func (c *CardDisk2Sequencer) step(data uint8, firstStep bool) bool {
	/*
		Q4 and Q6 set on the sofswitches is stored on the
		latch.
	*/
	q5 := c.latch.Q(5) // Drive selection
	q4 := c.latch.Q(4) // Motor on (before delay)

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
		q0 := c.latch.Q(0)
		q1 := c.latch.Q(1)
		q2 := c.latch.Q(2)
		q3 := c.latch.Q(3)
		c.drive[0].moveHead(q0, q1, q2, q3)
		c.drive[1].moveHead(q0, q1, q2, q3)
	}

	/*
		The reading from the drive is converted to a pulse detecting
		changes using Q3 and Q4 of the flip flop, combined with
		the last quarter of the 74LS132 NAND.
		The woz format provides the pulse directly and we won't emulate
		this detection.
	*/
	pulse := false
	c.lastPulseCycles++
	if c.lastPulseCycles == disk2PulseCcyles {
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
	seqBits := component.ByteToPins(c.sequence)

	high := c.register >= 0x80
	romAddress := component.PinsToByte([8]bool{
		seqBits[1], // seq1
		high,
		c.latch.Q(6),
		c.latch.Q(7),
		!pulse,
		seqBits[0], // seq0
		seqBits[2], // seq2
		seqBits[3], //seq3
	})

	//fmt.Printf("For Q6(%v) Q7(%v) H(%v) P(%v) Seq(%x) => ",
	//	c.latch.Q(6), c.latch.Q(6), high, pulse, c.sequence)

	romData := c.p6ROM[romAddress]
	inst := romData & 0xf
	next := romData >> 4
	//fmt.Printf("cmd(%x) seq(%x) ", inst, next)

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
			c.register = c.register >> 1
			if wProt {
				c.register |= 0x80
			}
		case 3:
			// Load
			c.register = data
		}
	}
	c.sequence = next

	//fmt.Printf("reg %02x\n", c.register)
	return true
}
