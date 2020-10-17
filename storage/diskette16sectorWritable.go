package storage

/*
See:
	"Beneath Apple DOS" https://fabiensanglard.net/fd_proxy/prince_of_persia/Beneath%20Apple%20DOS.pdf
	https://github.com/TomHarte/CLK/wiki/Apple-GCR-disk-encoding
*/

type diskette16sectorWritable struct {
	nib      *fileNib
	position int

	// Needed to write back
	hasDirtyTrack bool
	dirtyTrack    int
}

func (d *diskette16sectorWritable) PowerOn(cycle uint64) {
	// Not used
}
func (d *diskette16sectorWritable) PowerOff(_ uint64) {
	d.commit()
}

func (d *diskette16sectorWritable) Read(quarterTrack int, cycle uint64) uint8 {
	track := d.nib.track[quarterTrack/4]
	value := track[d.position]
	d.position = (d.position + 1) % nibBytesPerTrack
	return value
}

func (d *diskette16sectorWritable) Write(quarterTrack int, value uint8, _ uint64) {
	track := quarterTrack / 4

	if d.hasDirtyTrack && track != d.dirtyTrack {
		d.commit()
	}

	d.nib.track[track][d.position] = value
	d.position = (d.position + 1) % nibBytesPerTrack

	d.hasDirtyTrack = true
	d.dirtyTrack = track
}

func (d *diskette16sectorWritable) commit() {
	if d.hasDirtyTrack {
		d.nib.saveTrack(d.dirtyTrack)
		d.hasDirtyTrack = false
	}
}
