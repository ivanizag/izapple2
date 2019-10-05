package apple2

import (
	"encoding/binary"
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
	diskette     *diskette16sector
	currentPhase int
	power        bool // q4
	halfTrack    int
	position     int
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
	for i := 0; i < 4; i++ {
		phase := i
		c.ssr[phase<<1] = func(_ *ioC0Page) uint8 {
			return c.dataLatch // All even addresses return the last dataLatch
		}
		c.ssr[(phase<<1)+1] = func(_ *ioC0Page) uint8 {
			// Move the head up or down depending on the previous phase.
			drive := &c.drive[c.selected]
			delta := (phase - drive.currentPhase + 4) % 4
			switch delta {
			case 1: // Up
				drive.halfTrack++
			case 2: // Illegal, let's say up
				drive.halfTrack++
			case 3: // Down
				drive.halfTrack--
			case 0: // No chamge
			}

			// Don't go over the limits
			if drive.halfTrack > maxHalfTrack {
				drive.halfTrack = maxHalfTrack
			} else if drive.halfTrack < 0 {
				drive.halfTrack = 0
			}

			drive.currentPhase = phase
			//fmt.Printf("DISKII: Current halftrack is %v\n", drive.halfTrack)
			return 0
		}
	}

	// Q4, power switch
	c.ssr[0x8] = func(_ *ioC0Page) uint8 {
		if c.drive[c.selected].power {
			c.drive[c.selected].power = false
			c.a.releaseFastMode()
		}
		return c.dataLatch
	}
	c.ssr[0x9] = func(_ *ioC0Page) uint8 {
		if !c.drive[c.selected].power {
			c.drive[c.selected].power = true
			c.a.requestFastMode()
		}
		return 0
	}

	// Q5, drive selecion
	c.ssr[0xA] = func(_ *ioC0Page) uint8 {
		c.selected = 0
		return c.dataLatch
	}
	c.ssr[0xB] = func(_ *ioC0Page) uint8 {
		c.selected = 1
		return 0
	}

	// Q6, Q7
	for i := 0xC; i <= 0xF; i++ {
		iCopy := i
		c.ssr[iCopy] = func(_ *ioC0Page) uint8 {
			return c.softSwitchQ6Q7(iCopy, 0)
		}
		c.ssw[iCopy] = func(_ *ioC0Page, value uint8) {
			c.softSwitchQ6Q7(iCopy, value)
		}
	}

	c.cardBase.assign(a, slot)
}

func (c *cardDisk2) softSwitchQ6Q7(index int, in uint8) uint8 {
	switch index {
	case 0xC: // Q6L
		c.q6 = false
	case 0xD: // Q6H
		c.q6 = true
	case 0xE: // Q/L
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
	if !c.q6 {
		if !c.q7 { // Q6L-Q7L: Read
			track := d.halfTrack / 2
			c.dataLatch, d.position = d.diskette.read(track, d.position)
		} else { // Q6L-Q7H: Write the dataLatch value to disk. Shift data out
			track := d.halfTrack / 2
			d.position = d.diskette.write(track, d.position, c.dataLatch)
		}
	} else {
		if !c.q7 { // Q6H-Q7L: Sense write protect / prewrite state
			// Bit 7 of the control status register means write protected
			c.dataLatch = 0 // Never write protected
		} else { // Q6H-Q7H: Load data into the controller
			c.dataLatch = in
		}
	}
}

func (d *cardDisk2Drive) insertDiskette(dt *diskette16sector) {
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
	err := binary.Write(w, binary.BigEndian, d.currentPhase)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, d.power)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, d.halfTrack)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, d.position)
	if err != nil {
		return err
	}
	return nil
}

func (d *cardDisk2Drive) load(r io.Reader) error {
	err := binary.Read(r, binary.BigEndian, &d.currentPhase)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.BigEndian, &d.power)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.BigEndian, &d.halfTrack)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.BigEndian, &d.position)
	if err != nil {
		return err
	}
	return nil
}
