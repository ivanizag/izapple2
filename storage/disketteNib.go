package storage

/*
See:
	"Beneath Apple DOS" https://fabiensanglard.net/fd_proxy/prince_of_persia/Beneath%20Apple%20DOS.pdf
	https://github.com/TomHarte/CLK/wiki/Apple-GCR-disk-encoding
*/

type disketteNib struct {
	nib      *fileNib
	position int
}

func (d *disketteNib) PowerOn(cycle uint64) {
	// Not used
}
func (d *disketteNib) PowerOff(_ uint64) {
	// Not used
}

func (d *disketteNib) Read(quarterTrack int, cycle uint64) uint8 {
	track := d.nib.track[quarterTrack/4]
	value := track[d.position]
	d.position = (d.position + 1) % nibBytesPerTrack
	// fmt.Printf("%v, %v, %v, %x\n", 0, 0, d.position, uint8(value))
	return value
}

func (d *disketteNib) Write(quarterTrack int, value uint8, _ uint64) {
	track := quarterTrack / 4
	d.nib.track[track][d.position] = value
	d.position = (d.position + 1) % nibBytesPerTrack
}

func (d *disketteNib) Is13Sectors() bool {
	// It amy be 13 sectors but we don't know
	return false
}
