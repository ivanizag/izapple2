package storage

/*
See:
	"Beneath Apple DOS" https://fabiensanglard.net/fd_proxy/prince_of_persia/Beneath%20Apple%20DOS.pdf
	https://github.com/TomHarte/CLK/wiki/Apple-GCR-disk-encoding
*/

type disketteNibWritable struct {
	nib      *fileNib
	position int
	powered  bool
	shifting bool // The next read is the latch shifting in a nibble

	// Needed to write back
	hasDirtyTrack bool
	dirtyTrack    int
}

func (d *disketteNibWritable) PowerOn(cycle uint64) {
	d.powered = true
}
func (d *disketteNibWritable) PowerOff(_ uint64) {
	d.powered = false
	d.commit()
}

func (d *disketteNibWritable) Read(quarterTrack int, cycle uint64) uint8 {
	track := d.nib.track[quarterTrack/4]
	if !d.powered {
		// The disk is still, the latch keeps what it has
		return track[d.position]
	}
	if d.shifting {
		d.shifting = false
		return latchShifting(track[d.position])
	}
	value := track[d.position]
	d.position = (d.position + 1) % nibBytesPerTrack
	d.shifting = true
	return value
}

func (d *disketteNibWritable) Write(quarterTrack int, value uint8, _ uint64) {
	track := quarterTrack / 4

	if d.hasDirtyTrack && track != d.dirtyTrack {
		d.commit()
	}

	d.shifting = false
	d.nib.track[track][d.position] = value
	if d.powered { // the position can't change if the drive is off
		d.position = (d.position + 1) % nibBytesPerTrack
	}
	d.hasDirtyTrack = true
	d.dirtyTrack = track
}

func (d *disketteNibWritable) commit() {
	if d.hasDirtyTrack {
		d.nib.saveTrack(d.dirtyTrack)
		d.hasDirtyTrack = false
	}
}

func (d *disketteNibWritable) Is13Sectors() bool {
	// It amy be 13 sectors but we don't know
	return false
}

/*
IsWriteProtected returns false. The writes always land in memory and the
software reads them back, they are saved when the image or an overlay can be
written. Refusing them would break the software that saves to a diskette that
could not be saved, an embedded, compressed or downloaded one.
*/
func (d *disketteNibWritable) IsWriteProtected() bool {
	return false
}
