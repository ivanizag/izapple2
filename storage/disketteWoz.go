package storage

import (
	"errors"
	"math/rand"
)

/*
See:
	https://applesaucefdc.com/woz/
*/

type disketteWoz struct {
	data    *FileWoz
	cycleOn uint64 // Cycle when the disk was last turned on
	turning bool

	latch       uint8
	position    uint32
	positionMax uint32 // As tracks may have different lengths position is related of positionMax of the las track
	cycle       uint64

	mc3470Buffer uint8 // Four bit buffer to detect weak bits and to add latency

	visibleLatch          uint8
	visibleLatchCountDown int8 // The visible latch stores a valid latch reading for 2 bit timings
}

func newDisquetteWoz(f *FileWoz) (*disketteWoz, error) {
	// Discard not supported features
	if f.Info.DiskType != 1 {
		return nil, errors.New("Only 5.25 disks are supported")
	}

	var d disketteWoz
	d.data = f
	return &d, nil
}

func (d *disketteWoz) PowerOn(cycle uint64) {
	d.turning = true
	d.cycleOn = cycle
}

func (d *disketteWoz) PowerOff(_ uint64) {
	d.turning = false
}

func (d *disketteWoz) Read(quarterTrack int, cycle uint64) uint8 {
	// Count cycles to know how many bits have been read
	cycles := cycle - d.cycle
	deltaBits := cycles / cyclesPerBit // TODO: Use Woz optimal bit timing

	// Process bits from woz
	// TODO: avoid processing too many bits if delta is big
	for i := uint64(0); i < deltaBits; i++ {
		// Get next bit taking into account the MC3470 latency and weak bits
		var fluxBit bool
		fluxBit, d.position, d.positionMax = d.data.GetNextBitAndPosition(d.position, d.positionMax, quarterTrack)
		d.mc3470Buffer = (d.mc3470Buffer << 1) & 0x0f
		if fluxBit {
			d.mc3470Buffer++
		}
		bit := (d.mc3470Buffer >> 1) & 0x1 // Use the previous to last bit to add latency
		if d.mc3470Buffer == 0 && rand.Intn(100) < 3 {
			// Four consecutive zeros. It'a a fake bit.
			// Output a random value. 70% zero, 30% one
			bit = 1
		}

		d.latch = (d.latch << 1) + bit
		if d.latch >= 0x80 {
			// Valid byte, store value a bit longer and clear the internal latch
			// fmt.Printf("Valid 0x%.2x\n", d.latch)
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

	// fmt.Printf("Visible: 0x%.2x, latch: 0x%.2x, bits: %v, cycles: %v\n", d.visibleLatch, d.latch, deltaBits, cycle-d.cycle)

	// Update the internal last cycle without losing the remainder not processed
	d.cycle += deltaBits * cyclesPerBit

	return d.visibleLatch
}

func (d *disketteWoz) Write(quarterTrack int, value uint8, _ uint64) {
	panic("Write not implemented on woz disk implementation")
}

func (d *disketteWoz) Is13Sectors() bool {
	return d.data.version == 2 && d.data.Info.BootSectorFormat == 2
}
