package izapple2

import (
	"errors"
	"math/rand"
)

/*
See:
	https://applesaucefdc.com/woz/

Emulation status for the disk used on the reference:
	- How to begin
		- DOS 3.3: Works
		- * DOS 3.2: Not working, 13 sector disks can't boot
	- Next choices
		- Bouncing Kamungas: Works
		- *** Commando: Not working
		- Planetfall: Working
		- Rescue Raiders: Working
		- *** Sammy Lightfoot: Not working
		- Stargate: Working
	- Cross track sync
		- *** Blazing Paddles: Not working
		- *** Take 1: Not working
		- *** Hard Hat Mack: Not working
	- Half tracks
		- The Bilestoad: Working
	- Even more bit fiddling
		- Dino Eggs: Working
		- Crisis Mountain: Working
		- Miner 2049er II: Working
	- When bits aren't really bits
		- The Print Shop Companion: Working
	- What is the lifespan of the data latch?
		- *** First Math Adventures - Understanding Word Problems
	- Reading Offset Data Streams
		- *** Wings of Fury: Not working
		- Stickybear Town Builder: Working
	- Optimal bit timing of WOZ 2,0
		- * Border Zone: Unknown, there is no UI to swap disks

*/

type disketteWoz struct {
	data    *fileWoz
	cycleOn uint64 // Cycle when the disk was last turned on
	turning bool

	latch     uint8
	position  uint32
	cycle     uint64
	trackSize uint32

	mc3470Buffer uint8 // Four bit buffer to detect weak bits and to add latency

	visibleLatch          uint8
	visibleLatchCountDown int8 // The visible latch stores a valid latch reading for 2 bit timings
}

func newDisquetteWoz(f *fileWoz) (*disketteWoz, error) {
	// Discard not supported features
	if f.info.DiskType != 1 {
		return nil, errors.New("Only 5.25 disks are supported")
	}
	if f.info.BootSectorFormat == 2 { // Info not available in WOZ 1.0
		return nil, errors.New("Woz 13 sector disks are not supported")
	}

	var d disketteWoz
	d.data = f
	return &d, nil
}

func (d *disketteWoz) powerOn(cycle uint64) {
	d.turning = true
	d.cycleOn = cycle
}

func (d *disketteWoz) powerOff(_ uint64) {
	d.turning = false
}

func (d *disketteWoz) read(quarterTrack int, cycle uint64) uint8 {
	// Count cycles to know how many bits have been read
	cycles := cycle - d.cycle
	deltaBits := cycles / cyclesPerBit // TODO: Use Woz optimal bit timing

	// Process bits from woz
	// TODO: avoid processing too many bits if delta is big
	for i := uint64(0); i < deltaBits; i++ {
		// Get next bit taking into account the MC3470 latency and weak bits
		d.position++
		fluxBit := d.data.getBit(d.position, quarterTrack)
		d.mc3470Buffer = (d.mc3470Buffer<<1 + fluxBit) & 0x0f
		bit := (d.mc3470Buffer >> 1) & 0x1 // Use the previous to last bit to add latency
		if d.mc3470Buffer == 0 && rand.Intn(100) < 3 {
			// Four consecutive zeros.It'a a fake bit.
			// Output a random value. 70% zero, 30% one
			bit = 1
		}

		d.latch = (d.latch << 1) + bit
		if d.latch >= 0x80 {
			// Valid byte, store value a bit longer and clear the internal latch
			//fmt.Printf("Valid 0x%.2x\n", d.latch)
			d.visibleLatch = d.latch
			d.visibleLatchCountDown = 1
			d.latch = 0
		} else if d.visibleLatchCountDown > 0 {
			// Continue showing the valid byte
			d.visibleLatchCountDown--
		} else {
			// The valid byte is lost, show the internal latch
			d.visibleLatch = d.latch
		}
	}

	//fmt.Printf("Visible: 0x%.2x, latch: 0x%.2x, bits: %v, cycles: %v\n", d.visibleLatch, d.latch, deltaBits, cycle-d.cycle)

	// Update the internal last cycle without losing the remainder not processed
	d.cycle += deltaBits * cyclesPerBit

	return d.visibleLatch
}

func (d *disketteWoz) write(quarterTrack int, value uint8, _ uint64) {
	panic("Write not implemented on woz disk implementation")
}
