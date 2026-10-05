package storage

/*
See:
	"Beneath Apple DOS" https://fabiensanglard.net/fd_proxy/prince_of_persia/Beneath%20Apple%20DOS.pdf
	https://github.com/TomHarte/CLK/wiki/Apple-GCR-disk-encoding
*/

type disketteNib struct {
	nib      *fileNib
	position int
	shifting bool // The next read is the latch shifting in a nibble
}

/*
latchShifting is what the data latch of the controller holds while a nibble is
shifted in: its first bits, with bit 7 still clear. A disk read a nibble at a
time gives every other read as this, as the latch of a turning disk is never
the same nibble twice in a row. DOS relies on it to tell a disk that turns
from one that does not, reading the latch eight times: in a gap of 0xff sync
nibbles, a nibble each read would look still, and DOS would wait each time for
the motor to come up to speed. Programs reading nibbles skip the values with
bit 7 clear.
*/
func latchShifting(next uint8) uint8 {
	return next >> 1
}

func (d *disketteNib) PowerOn(cycle uint64) {
	// Not used
}
func (d *disketteNib) PowerOff(_ uint64) {
	// Not used
}

func (d *disketteNib) Read(quarterTrack int, cycle uint64) uint8 {
	track := d.nib.track[quarterTrack/4]
	if d.shifting {
		d.shifting = false
		return latchShifting(track[d.position])
	}
	value := track[d.position]
	d.position = (d.position + 1) % nibBytesPerTrack
	d.shifting = true
	return value
}

func (d *disketteNib) Write(quarterTrack int, value uint8, _ uint64) {
	track := quarterTrack / 4
	d.shifting = false
	d.nib.track[track][d.position] = value
	d.position = (d.position + 1) % nibBytesPerTrack
}

func (d *disketteNib) Is13Sectors() bool {
	// It may be 13 sectors but we don't know
	return false
}
