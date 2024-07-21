package storage

type disketteNibTimed struct {
	nib     *fileNib
	cycleOn uint64 // Cycle when the disk was last turned on
}

func (d *disketteNibTimed) PowerOn(cycle uint64) {
	d.cycleOn = cycle
}
func (d *disketteNibTimed) PowerOff(_ uint64) {
	// Not needed
}

func (d *disketteNibTimed) getBitPositionInTrack(cycle uint64) int {
	// Calculate how long the disk has been spinning. We move one bit every 4 cycles.
	// In this implementation we don't take into account how long the motor takes to be at full speed.
	cycles := cycle - d.cycleOn
	position := cycles / cyclesPerBit
	return int(position % (8 * nibBytesPerTrack)) // Ignore full turns
}

func (d *disketteNibTimed) Read(quarterTrack int, cycle uint64) uint8 {
	track := d.nib.track[quarterTrack/4]
	bitPosition := d.getBitPositionInTrack(cycle)
	bytePosition := bitPosition / 8
	shift := uint(bitPosition % 8)
	if shift == 1 {
		// We continue having the unshifted byte for a little longer (4 cycles)
		shift = 0
	}
	value := track[bytePosition]
	value >>= shift
	//fmt.Printf("%v, %v, %v, %x\n", bitPosition, shift, bytePosition, uint8(value))
	return value
}

func (d *disketteNibTimed) Write(quarterTrack int, value uint8, _ uint64) {
	panic("Write not implemented on time based disk implementation")
}

func (d *disketteNibTimed) Is13Sectors() bool {
	// It amy be 13 sectors but we don't know
	return false
}
