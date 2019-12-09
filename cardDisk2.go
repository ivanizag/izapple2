package apple2

import (
	"encoding/binary"
	"fmt"
	"io"
)

/*
https://applesaucefdc.com/woz/reference2/

Good explanation of the softswitches and the phases:
http://yesterbits.com/media/pubs/AppleOrchard/articles/disk-ii-part-1-1983-apr.pdf

"IMW Floppy Disk I/O Controller info" (https://www.brutaldeluxe.fr/documentation/iwm/apple2_IWM_INFO_19840510.pdf)

"Understanfing the Apple II, chapter 9"

35 tracks, 16 sectors, 256 bytes
NIB: 35 tracks 6656 bytes, 232960 bytes

*/
const maxHalfTrack = 68

type cardDisk2 struct {
	cardBase
	selected int // q5, Only 0 and 1 supported
	drive    [2]cardDisk2Drive

	dataLatch uint8
	q6        bool
	q7        bool
}

type cardDisk2Drive struct {
	diskette   diskette
	power      bool  // q4
	phases     uint8 // q3, q2, q1 and q0 with q0 on the LSB. Magnets that are active on the stepper motor
	tracksStep int   // Stepmotor for tracks position. 4 steps per track
}

type diskette interface {
	powerOn(cycle uint64)
	powerOff(cycle uint64)
	read(quarterTrack int, cycle uint64) uint8
	write(quarterTrack int, value uint8, cycle uint64)
}

const (
	diskBitCycle           = 4   // There is a dataLatch bit transferred every 4 cycles
	diskLatchReadCycles    = 7   // Loaded data is available for a little more than 7ns
	diskWriteByteCycle     = 32  // Load data to write every 32 cycles
	diskWriteSelfSyncCycle = 40  // Save $FF every 40 cycles. Self sync is 10 bits: 1111 1111 00
	diskMotorStartMs       = 150 // Time with the disk spinning to get full speed

)

func (c *cardDisk2) assign(a *Apple2, slot int) {
	// Q1, Q2, Q3 and Q4 phase control soft switches,
	for i := uint8(0); i < 4; i++ {
		phase := i
		c.addCardSoftSwitchR(phase<<1, func(_ *ioC0Page) uint8 {
			// Update magnets and position
			drive := &c.drive[c.selected]
			drive.phases &^= (1 << phase)
			drive.tracksStep = moveStep(drive.phases, drive.tracksStep)

			return c.dataLatch // All even addresses return the last dataLatch
		}, fmt.Sprintf("PHASE%vOFF", phase))

		c.addCardSoftSwitchR((phase<<1)+1, func(_ *ioC0Page) uint8 {
			// Update magnets and position
			drive := &c.drive[c.selected]
			drive.phases |= (1 << phase)
			drive.tracksStep = moveStep(drive.phases, drive.tracksStep)

			return 0
		}, fmt.Sprintf("PHASE%vOFF", phase))
	}

	// Q4, power switch
	c.addCardSoftSwitchR(0x8, func(_ *ioC0Page) uint8 {
		drive := c.drive[c.selected]
		if !drive.power {
			drive.power = false
			c.a.releaseFastMode()
			if drive.diskette != nil {
				drive.diskette.powerOff(c.a.cpu.GetCycles())
			}
		}
		return c.dataLatch
	}, "Q4DRIVEOFF")
	c.addCardSoftSwitchR(0x9, func(_ *ioC0Page) uint8 {
		drive := c.drive[c.selected]
		if !drive.power {
			drive.power = true
			c.a.requestFastMode()
			if drive.diskette != nil {
				drive.diskette.powerOn(c.a.cpu.GetCycles())
			}
		}
		return 0
	}, "")

	// Q5, drive selecion
	c.addCardSoftSwitchR(0xA, func(_ *ioC0Page) uint8 {
		c.selected = 0
		return c.dataLatch
	}, "Q5SELECT1")
	c.addCardSoftSwitchR(0xB, func(_ *ioC0Page) uint8 {
		c.selected = 1
		return 0
	}, "Q5SELECT2")

	// Q6, Q7
	for i := uint8(0xC); i <= 0xF; i++ {
		iCopy := i
		c.addCardSoftSwitchR(iCopy, func(_ *ioC0Page) uint8 {
			return c.softSwitchQ6Q7(iCopy, 0)
		}, "Q6Q7")
		c.addCardSoftSwitchW(iCopy, func(_ *ioC0Page, value uint8) {
			c.softSwitchQ6Q7(iCopy, value)
		}, "Q6Q7")
	}

	c.cardBase.assign(a, slot)
}

// Q6: shift/load
// Q7: read/write

func (c *cardDisk2) softSwitchQ6Q7(index uint8, in uint8) uint8 {
	switch index {
	case 0xC: // Q6L
		c.q6 = false
	case 0xD: // Q6H
		c.q6 = true
	case 0xE: // Q7L
		c.q7 = false
	case 0xF: // Q7H
		c.q7 = true
	}

	c.processQ6Q7(in)
	if index&1 == 0 {
		// All even addresses return the last dataLatch
		return c.dataLatch
	}
	return 0
}

func (c *cardDisk2) processQ6Q7(in uint8) {
	d := &c.drive[c.selected]
	if d.diskette == nil {
		return
	}
	if !c.q6 { // shift
		if !c.q7 { // Q6L-Q7L: Read
			c.dataLatch = d.diskette.read(d.tracksStep, c.a.cpu.GetCycles())
		} else { // Q6L-Q7H: Write the dataLatch value to disk. Shift data out
			d.diskette.write(d.tracksStep, c.dataLatch, c.a.cpu.GetCycles())
		}
	} else { // load
		if !c.q7 { // Q6H-Q7L: Sense write protect / prewrite state
			// Bit 7 of the control status register means write protected
			c.dataLatch = 0 // Never write protected
		} else { // Q6H-Q7H: Load data into the controller
			c.dataLatch = in
		}
	}
}

/*
Stepper motor to position the track.

There are a number of group of four magnets. The stepper motor can be thought as a long
line of groups of magnets, each group on the same configuration. We call phase each of those
magnets. The cog is attracted to the enabled magnets, and can stay aligned to a magnet or
between two.

Phases (magents):                       3   2   1   0   3   2   1   0   3   2   1   0
Cog direction (step withn a group):   7 6 5 4 3 2 1 0 7 6 5 4 3 2 1 0 7 6 5 4 3 2 1 0

We will consider that the cog would go to the prefferred postion if there is one. Independenly
of the previous position. The previous position is only used to know if it goes up or down
a full group.
*/

const (
	undefinedPosition = -1
	maxStep           = 68 * 2 // What is the maximum quarter tracks a DiskII can go?
	stepsPerGroup     = 8
	stepsPerTrack     = 4
)

var cogPositions = []int{
	undefinedPosition, // 0000, phases active
	0,                 // 0001
	2,                 // 0010
	1,                 // 0011
	4,                 // 0100
	undefinedPosition, // 0101
	3,                 // 0110
	2,                 // 0111
	6,                 // 1000
	7,                 // 1001
	undefinedPosition, // 1010
	0,                 // 1011
	5,                 // 1100
	6,                 // 1101
	4,                 // 1110
	undefinedPosition, // 1111
}

func moveStep(phases uint8, prevStep int) int {

	//fmt.Printf("magnets: 0x%x\n", phases)

	cogPosition := cogPositions[phases]
	if cogPosition == undefinedPosition {
		// Don't move if magnets don't push on a defined direction.
		return prevStep
	}

	prevPosition := prevStep % stepsPerGroup // Direction, step in the current group of magnets.
	delta := cogPosition - prevPosition
	if delta < 0 {
		delta = delta + stepsPerGroup
	}

	var nextStep int
	if delta < 4 {
		// Steps up
		nextStep = prevStep + delta
		if nextStep > maxStep {
			nextStep = maxStep
		}
	} else if delta == 4 {
		// Don't move if magnets push on the oposite direction
		nextStep = prevStep
	} else { // delta > 4
		// Steps down
		nextStep = prevStep + delta - stepsPerGroup
		if nextStep < 0 {
			nextStep = 0
		}
	}
	return nextStep
}

func (d *cardDisk2Drive) insertDiskette(dt diskette) {
	d.diskette = dt
}

func (c *cardDisk2) save(w io.Writer) error {
	err := binary.Write(w, binary.BigEndian, c.selected)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, c.dataLatch)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, c.q6)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, c.q7)
	if err != nil {
		return err
	}
	err = c.drive[0].save(w)
	if err != nil {
		return err
	}
	err = c.drive[1].save(w)
	if err != nil {
		return err
	}
	return c.cardBase.save(w)
}

func (c *cardDisk2) load(r io.Reader) error {
	err := binary.Read(r, binary.BigEndian, &c.selected)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.BigEndian, &c.dataLatch)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.BigEndian, &c.q6)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.BigEndian, &c.q7)
	if err != nil {
		return err
	}
	err = c.drive[0].load(r)
	if err != nil {
		return err
	}
	err = c.drive[1].load(r)
	if err != nil {
		return err
	}
	return c.cardBase.load(r)
}

func (d *cardDisk2Drive) save(w io.Writer) error {
	err := binary.Write(w, binary.BigEndian, d.power)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, d.phases)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, d.tracksStep)
	if err != nil {
		return err
	}
	return nil
}

func (d *cardDisk2Drive) load(r io.Reader) error {
	err := binary.Read(r, binary.BigEndian, &d.power)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.BigEndian, &d.phases)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.BigEndian, &d.tracksStep)
	if err != nil {
		return err
	}
	return nil
}
