package apple2

type diskette16sectorTimed struct {
	nib     *fileNib
	cycleOn uint64 // Cycle when the disk was last turned on
}

func (d *diskette16sectorTimed) powerOn(cycle uint64) {
	d.cycleOn = cycle
}
func (d *diskette16sectorTimed) powerOff(_ uint64) {
	// Not needed
}

func (d *diskette16sectorTimed) getBitPositionInTrack(cycle uint64) int {
	// Calculate how long the disk has been spinning. We move one bit every 4 cycles.
	// In this implementation we don't take into account how long the motor takes to be at full speed.
	cycles := cycle - d.cycleOn
	position := cycles / cyclesPerBit
	return int(position % (8 * nibBytesPerTrack)) // Ignore full turns
}

func (d *diskette16sectorTimed) read(quarterTrack int, cycle uint64) uint8 {
	track := d.nib.track[quarterTrack/stepsPerTrack]
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

func (d *diskette16sectorTimed) write(quarterTrack int, value uint8, _ uint64) {
	panic("Write not implemented on time based disk implementation")
}

func loadDisquetteTimed(filename string) (*diskette16sectorTimed, error) {
	var d diskette16sectorTimed

	f, err := loadNibOrDsk(filename)
	if err != nil {
		return nil, err
	}
	d.nib = f

	return &d, nil
}
