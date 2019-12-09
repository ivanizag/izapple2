package apple2

import (
	"errors"
	"os"
)

/*
See:
	"Beneath Apple DOS" https://fabiensanglard.net/fd_proxy/prince_of_persia/Beneath%20Apple%20DOS.pdf
	https://github.com/TomHarte/CLK/wiki/Apple-GCR-disk-encoding
*/

const (
	numberOfTracks   = 35
	numberOfSectors  = 16
	bytesPerSector   = 256
	bytesPerTrack    = numberOfSectors * bytesPerSector
	nibBytesPerTrack = 6656
	nibImageSize     = numberOfTracks * nibBytesPerTrack
	dskImageSize     = numberOfTracks * numberOfSectors * bytesPerSector
	defaultVolumeTag = 254
	cyclesPerBit     = 4
)

type diskette16sector struct {
	track     [numberOfTracks][]byte
	timeBased bool
	// Not time based implementation
	position int // For not time based implemenation
	// Time based implementation, expermiental
	cycleOn uint64 // Cycle when the disk was last turned on
}

func (d *diskette16sector) powerOn(cycle uint64) {
	d.cycleOn = cycle
}
func (d *diskette16sector) powerOff(_ uint64) {
	// Not needed
}

func (d *diskette16sector) getBitPositionInTrack(cycle uint64) int {
	// Calculate how long the disk has been spinning. We move one bit every 4 cycles.
	// In this implementation we don't take into account hot long the motor takes to be at full speed.
	cycles := cycle - d.cycleOn
	position := cycles / cyclesPerBit
	return int(position % (8 * nibBytesPerTrack)) // Ignore full turns
}

func (d *diskette16sector) read(quarterTrack int, cycle uint64) uint8 {
	track := d.track[quarterTrack/stepsPerTrack]
	if d.timeBased {
		bitPosition := d.getBitPositionInTrack(cycle)
		bytePosition := bitPosition / 8
		shift := uint(bitPosition % 8)
		if shift == 1 {
			// We continue having the previous data for a little longer
			shift = 0
		}
		value := track[bytePosition]
		value >>= shift
		//fmt.Printf("%v, %v, %v, %x\n", bitPosition, shift, bytePosition, uint8(data))
		return value
	}
	value := track[d.position]
	d.position = (d.position + 1) % nibBytesPerTrack
	//fmt.Printf("%v, %v, %v, %x\n", 0, 0, d.position, uint8(value))
	return value
}

func (d *diskette16sector) write(quarterTrack int, value uint8, _ uint64) {
	if d.timeBased {
		panic("Write not implmented on time based disk implementation")
	}
	track := quarterTrack / stepsPerTrack
	d.track[track][d.position] = value
	d.position = (d.position + 1) % nibBytesPerTrack
}

func loadDisquette(filename string) (*diskette16sector, error) {
	var d diskette16sector

	// Experimental
	d.timeBased = true

	data, err := loadResource(filename)
	if err != nil {
		return nil, err
	}
	size := len(data)

	if size == nibImageSize {
		// Load file already in nib format
		for i := 0; i < numberOfTracks; i++ {
			d.track[i] = data[nibBytesPerTrack*i : nibBytesPerTrack*(i+1)]
		}
	} else if size == dskImageSize {
		// Convert to nib
		for i := 0; i < numberOfTracks; i++ {
			trackData := data[i*bytesPerTrack : (i+1)*bytesPerTrack]
			d.track[i] = nibEncodeTrack(trackData, defaultVolumeTag, byte(i))
		}
	} else {
		return nil, errors.New("Invalid disk size")
	}

	return &d, nil
}

func (d *diskette16sector) saveNib(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, v := range d.track {
		_, err := f.Write(v)
		if err != nil {
			return err
		}
	}

	return nil
}

var dos33SectorsLogicOrder = [16]int{
	0x0, 0x7, 0xE, 0x6, 0xD, 0x5, 0xC, 0x4,
	0xB, 0x3, 0xA, 0x2, 0x9, 0x1, 0x8, 0xF,
}

var sixAndTwoTranslateTable = [0x40]byte{
	0x96, 0x97, 0x9a, 0x9b, 0x9d, 0x9e, 0x9f, 0xa6,
	0xa7, 0xab, 0xac, 0xad, 0xae, 0xaf, 0xb2, 0xb3,
	0xb4, 0xb5, 0xb6, 0xb7, 0xb9, 0xba, 0xbb, 0xbc,
	0xbd, 0xbe, 0xbf, 0xcb, 0xcd, 0xce, 0xcf, 0xd3,
	0xd6, 0xd7, 0xd9, 0xda, 0xdb, 0xdc, 0xdd, 0xde,
	0xdf, 0xe5, 0xe6, 0xe7, 0xe9, 0xea, 0xeb, 0xec,
	0xed, 0xee, 0xef, 0xf2, 0xf3, 0xf4, 0xf5, 0xf6,
	0xf7, 0xf9, 0xfa, 0xfb, 0xfc, 0xfd, 0xfe, 0xff,
}

const (
	gap1Len             = 48
	gap2Len             = 5
	primaryBufferSize   = bytesPerSector
	secondaryBufferSize = bytesPerSector/3 + 1
)

func oddEvenEncodeByte(b byte) []byte {
	/*
		A byte is encoded in two bytes to make sure the bytes start with 1 and
		does not have two consecutive zeros.
		   Data byte: D7-D6-D5-D4-D3-D2-D1-D0
		   resutl[0]:  1-D7- 1-D5- 1-D3-1 -D1
		   resutl[1]:  1-D6- 1-D4- 1-D2-1 -D0
	*/
	e := make([]byte, 2)
	e[0] = ((b >> 1) & 0x55) | 0xaa
	e[1] = (b & 0x55) | 0xaa
	return e
}

func nibEncodeTrack(data []byte, volume byte, track byte) []byte {
	b := make([]byte, 0, nibBytesPerTrack) // Buffer slice with enough capacity
	// Initialize gaps to be copied for each sector
	gap1 := make([]byte, gap1Len)
	for i := range gap1 {
		gap1[i] = 0xff
	}
	gap2 := make([]byte, gap2Len)
	for i := range gap2 {
		gap2[i] = 0xff
	}
	for physicalSector := byte(0); physicalSector < numberOfSectors; physicalSector++ {
		/* On the DSK file the sectors are in DOS3.3 logical order
		but on the physical encoded track as well as in the nib
		files they are in phisical order.
		*/
		logicalSector := dos33SectorsLogicOrder[physicalSector]
		sectorData := data[logicalSector*bytesPerSector : (logicalSector+1)*bytesPerSector]

		//  6and2 prenibbilizing.
		primaryBuffer := make([]byte, primaryBufferSize)
		secondaryBuffer := make([]byte, secondaryBufferSize)
		for i, v := range sectorData {
			// Primary buffer is easy: the 6 MSB
			primaryBuffer[i] = v >> 2
			// Secondary buffer: the 2 LSB reversed, shifted and in their place
			shift := uint((i / secondaryBufferSize) * 2)
			bit0 := ((v & 0x01) << 1) << shift
			bit1 := ((v & 0x02) >> 1) << shift
			position := i % secondaryBufferSize
			secondaryBuffer[position] |= bit0 | bit1
		}

		// Render sector
		// Address field
		b = append(b, gap1...)
		b = append(b, 0xd5, 0xaa, 0x96)                                  // Address prolog
		b = append(b, oddEvenEncodeByte(volume)...)                      // 4-4 encoded volume
		b = append(b, oddEvenEncodeByte(track)...)                       // 4-4 encoded track
		b = append(b, oddEvenEncodeByte(physicalSector)...)              // 4-4 encoded sector
		b = append(b, oddEvenEncodeByte(volume^track^physicalSector)...) // Checksum
		b = append(b, 0xde, 0xaa, 0xeb)                                  // Epilog
		// Data field
		b = append(b, gap2...)
		b = append(b, 0xd5, 0xaa, 0xad) // Data prolog
		prevV := byte(0)
		for _, v := range secondaryBuffer {
			b = append(b, sixAndTwoTranslateTable[v^prevV])
			prevV = v
		}
		for _, v := range primaryBuffer {
			b = append(b, sixAndTwoTranslateTable[v^prevV])
			prevV = v
		}
		b = append(b, sixAndTwoTranslateTable[prevV]) // Checksum
		b = append(b, 0xde, 0xaa, 0xeb)               // Data epilog
	}

	return b
}
