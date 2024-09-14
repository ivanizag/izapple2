package izapple2

import (
	"fmt"
	"strconv"

	"github.com/ivanizag/izapple2/storage"
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

// CardDisk2 is a DiskII interface card
type CardDisk2 struct {
	cardBase
	sectors13 bool

	selected int  // q5, Only 0 and 1 supported
	power    bool // q4
	drive    [2]cardDisk2Drive
	fastMode bool

	dataLatch uint8
	q6        bool
	q7        bool

	trackTracer trackTracer
}

type drive interface {
	insertDiskette(path string) error
}

type cardDisk2Drive struct {
	name      string
	diskette  storage.Diskette
	phases    uint8 // q3, q2, q1 and q0 with q0 on the LSB. Magnets that are active on the stepper motor
	trackStep int   // Stepmotor for tracks position. 4 steps per track
}

func newCardDisk2Builder() *cardBuilder {
	return &cardBuilder{
		name:        "Disk II",
		description: "Disk II interface card",
		defaultParams: &[]paramSpec{
			{"disk1", "Diskette image for drive 1", ""},
			{"disk2", "Diskette image for drive 2", ""},
			{"tracktracer", "Trace how the disk head moves between tracks", "false"},
			{"fast", "Enable CPU burst when accessing the disk", "true"},
			{"sectors13", "Use 13 sectors per track ROM", "false"},
		},
		buildFunc: func(params map[string]string) (Card, error) {
			var c CardDisk2
			c.sectors13 = paramsGetBool(params, "sectors13")

			disk1 := paramsGetPath(params, "disk1")
			if disk1 != "" {
				err := c.drive[0].insertDiskette(disk1)
				if err != nil {
					return nil, err
				}
				if c.drive[0].diskette.Is13Sectors() && !c.sectors13 {
					// Auto configure for 13 sectors per track
					c.sectors13 = true
				}
			}
			disk2 := paramsGetPath(params, "disk2")
			if disk2 != "" {
				err := c.drive[1].insertDiskette(disk2)
				if err != nil {
					return nil, err
				}
			}

			P5RomFile := "<internal>/Apple Disk II 16 Sector Interface Card ROM P5 - 341-0027.bin"
			if c.sectors13 {
				P5RomFile = "<internal>/Apple Disk II 13 Sector Interface Card ROM P5 - 341-0009.bin"
			}

			err := c.loadRomFromResource(P5RomFile, cardRomSimple)
			if err != nil {
				return nil, err
			}

			trackTracer := paramsGetBool(params, "tracktracer")
			if trackTracer {
				c.trackTracer = makeTrackTracerLogger()
			}
			c.fastMode = paramsGetBool(params, "fast")
			return &c, nil
		},
	}
}

// GetInfo returns smartPort info
func (c *CardDisk2) GetInfo() map[string]string {
	info := make(map[string]string)
	if c.sectors13 {
		info["rom"] = "13 sector"
	} else {
		info["rom"] = "16 sector"
	}
	info["power"] = strconv.FormatBool(c.power)

	info["D1 name"] = c.drive[0].name
	info["D1 track"] = strconv.FormatFloat(float64(c.drive[0].trackStep)/4, 'f', 2, 64)

	info["D2 name"] = c.drive[1].name
	info["D2 track"] = strconv.FormatFloat(float64(c.drive[1].trackStep)/4, 'f', 2, 64)
	return info
}

func (c *CardDisk2) reset() {
	// UtA2e 9-12, all switches forced to off
	drive := &c.drive[c.selected]
	drive.phases = 0      // q0, q1, q2, q3
	c.softSwitchQ4(false) // q4
	c.selected = 0        // q5
	c.q6 = false
	c.q7 = false
}

func (c *CardDisk2) setTrackTracer(tt trackTracer) {
	c.trackTracer = tt
}

func (c *CardDisk2) assign(a *Apple2, slot int) {
	a.registerRemovableMediaDrive(&c.drive[0])
	a.registerRemovableMediaDrive(&c.drive[1])

	// Q1, Q2, Q3 and Q4 phase control soft switches,
	for i := uint8(0); i < 4; i++ {
		phase := i
		c.addCardSoftSwitchRW(phase<<1, func() uint8 {
			// Update magnets and position
			drive := &c.drive[c.selected]
			drive.phases &^= (1 << phase)
			drive.trackStep = moveDriveStepper(drive.phases, drive.trackStep)

			if c.trackTracer != nil {
				c.trackTracer.traceTrack(drive.trackStep, c.slot, c.selected)
			}

			return c.dataLatch // All even addresses return the last dataLatch
		}, fmt.Sprintf("PHASE%vOFF", phase))

		c.addCardSoftSwitchRW((phase<<1)+1, func() uint8 { // Update magnets and position
			drive := &c.drive[c.selected]
			drive.phases |= (1 << phase)
			drive.trackStep = moveDriveStepper(drive.phases, drive.trackStep)

			if c.trackTracer != nil {
				c.trackTracer.traceTrack(drive.trackStep, slot, c.selected)
			}

			return 0
		}, fmt.Sprintf("PHASE%vON", phase))
	}

	// Q4, power switch
	c.addCardSoftSwitchRW(0x8, func() uint8 {
		c.softSwitchQ4(false)
		return c.dataLatch
	}, "Q4DRIVEOFF")
	c.addCardSoftSwitchRW(0x9, func() uint8 {
		c.softSwitchQ4(true)
		return 0
	}, "Q4DRIVEON")

	// Q5, drive selecion
	c.addCardSoftSwitchRW(0xA, func() uint8 {
		c.softSwitchQ5(0)
		return c.dataLatch
	}, "Q5SELECT1")
	c.addCardSoftSwitchRW(0xB, func() uint8 {
		c.softSwitchQ5(1)
		return 0
	}, "Q5SELECT2")

	// Q6, Q7
	for i := uint8(0xC); i <= 0xF; i++ {
		iCopy := i
		c.addCardSoftSwitchR(iCopy, func() uint8 {
			return c.softSwitchQ6Q7(iCopy, 0)
		}, "Q6Q7")
		c.addCardSoftSwitchW(iCopy, func(value uint8) {
			c.softSwitchQ6Q7(iCopy, value)
		}, "Q6Q7")
	}

	c.cardBase.assign(a, slot)
}

func (c *CardDisk2) softSwitchQ4(value bool) {
	if !value && c.power {
		// Turn off
		c.power = false
		if c.fastMode {
			c.a.ReleaseFastMode()
		}
		drive := &c.drive[c.selected]
		if drive.diskette != nil {
			drive.diskette.PowerOff(c.a.GetCycles())
		}
	} else if value && !c.power {
		// Turn on
		c.power = true
		if c.fastMode {
			c.a.RequestFastMode()
		}
		drive := &c.drive[c.selected]
		if drive.diskette != nil {
			drive.diskette.PowerOn(c.a.GetCycles())
		}
	}
}

func (c *CardDisk2) softSwitchQ5(selected int) {
	if c.power && c.selected != selected {
		// Selected changed with power on, power goes to the other disk
		if c.drive[c.selected].diskette != nil {
			c.drive[c.selected].diskette.PowerOff(c.a.GetCycles())
		}
		if c.drive[selected].diskette != nil {
			c.drive[selected].diskette.PowerOn(c.a.GetCycles())
		}
	}

	c.selected = selected
}

// Q6: shift/load
// Q7: read/write

func (c *CardDisk2) softSwitchQ6Q7(index uint8, in uint8) uint8 {
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

func (c *CardDisk2) processQ6Q7(in uint8) {
	d := &c.drive[c.selected]
	if d.diskette == nil {
		return
	}
	if !c.q6 { // shift
		if !c.q7 { // Q6L-Q7L: Read
			c.dataLatch = d.diskette.Read(d.trackStep, c.a.GetCycles())
		} else { // Q6L-Q7H: Write the dataLatch value to disk. Shift data out
			d.diskette.Write(d.trackStep, c.dataLatch, c.a.GetCycles())
		}
	} else { // load
		if !c.q7 { // Q6H-Q7L: Sense write protect / prewrite state
			// Bit 7 of the control status register means write protected
			c.dataLatch = 0 // Never write protected
		} else { // Q6H-Q7H: Load data into the controller
			c.dataLatch = in
		}
	}

	/*
		if c.dataLatch >= 0x80 {
			fmt.Printf("Datalach: 0x%.2x in cycle %v\n", c.dataLatch, c.a.GetCycles())
		}
	*/
}

func (d *cardDisk2Drive) insertDiskette(name string) error {
	diskette, err := LoadDiskette(name)
	if err != nil {
		return err
	}

	d.name = name
	d.diskette = diskette
	return nil
}
