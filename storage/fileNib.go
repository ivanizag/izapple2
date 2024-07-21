package storage

import (
	"errors"
	"fmt"
	"os"
	"strings"
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
	d13ImageSize     = numberOfTracks * 13 * bytesPerSector
	defaultVolumeTag = 254
	cyclesPerBit     = 4
)

type fileNib struct {
	track [numberOfTracks][]byte

	// Needed to write back
	supportsWrite bool
	filename      string
	logicalOrder  *[16]int
}

func isFileNib(data []uint8) bool {
	return len(data) == nibImageSize
}

func newFileNib(data []uint8) *fileNib {
	var f fileNib

	for i := 0; i < numberOfTracks; i++ {
		f.track[i] = data[nibBytesPerTrack*i : nibBytesPerTrack*(i+1)]
	}

	return &f
}

func isFileDsk(data []uint8) bool {
	return len(data) == dskImageSize
}

func isFileD13(data []uint8) bool {
	return len(data) == d13ImageSize
}

func newFileDsk(data []uint8, filename string) *fileNib {
	var f fileNib

	isPO := strings.HasSuffix(strings.ToLower(filename), "po")
	f.logicalOrder = &dos33SectorsLogicalOrder
	if isPO {
		f.logicalOrder = &prodosSectorsLogicalOrder
	}

	f.filename = filename
	f.supportsWrite = true

	for i := 0; i < numberOfTracks; i++ {
		trackData := data[i*bytesPerTrack : (i+1)*bytesPerTrack]
		f.track[i] = nibEncodeTrack(trackData, defaultVolumeTag, byte(i), f.logicalOrder)
	}

	return &f
}

func (f *fileNib) saveTrack(track int) {
	if f.supportsWrite {
		file, err := os.OpenFile(f.filename, os.O_RDWR, 0)
		if err != nil {
			// We can't open the file for writing"
			f.supportsWrite = false
			fmt.Printf("Data can't be written for %v\n", f.filename)
		}

		data, err := nibDecodeTrack(f.track[track], f.logicalOrder)
		if err != nil {
			f.supportsWrite = false
			fmt.Printf("Data written can't be decoded from nibbles\n")
		}

		offset := int64(track * bytesPerTrack)
		_, err = file.WriteAt(data, offset)
		if err != nil {
			f.supportsWrite = false
			fmt.Printf("Data can't be written\n")
		}
	}
}

func (f *fileNib) saveNib(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, v := range f.track {
		_, err := file.Write(v)
		if err != nil {
			return err
		}
	}

	return nil
}

// See Beneath Apple DOS, figure 3.24
var dos33SectorsLogicalOrder = [16]int{
	0x0, 0x7, 0xE, 0x6, 0xD, 0x5, 0xC, 0x4,
	0xB, 0x3, 0xA, 0x2, 0x9, 0x1, 0x8, 0xF,
}

// See Beneath Apple ProDOS, figure 3.1
var prodosSectorsLogicalOrder = [16]int{
	0x0, 0x8, 0x1, 0x9, 0x2, 0xA, 0x3, 0xB,
	0x4, 0xC, 0x5, 0xD, 0x6, 0xE, 0x7, 0xF,
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

var sixAndTwoUntranslateTable = [256]int16{
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, 0, 1, -1, -1, 2, 3, -1, 4, 5, 6,
	-1, -1, -1, -1, -1, -1, 7, 8, -1, -1, -1, 9, 10, 11, 12, 13,
	-1, -1, 14, 15, 16, 17, 18, 19, -1, 20, 21, 22, 23, 24, 25, 26,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, 27, -1, 28, 29, 30,
	-1, -1, -1, 31, -1, -1, 32, 33, -1, 34, 35, 36, 37, 38, 39, 40,
	-1, -1, -1, -1, -1, 41, 42, 43, -1, 44, 45, 46, 47, 48, 49, 50,
	-1, -1, 51, 52, 53, 54, 55, 56, -1, 57, 58, 59, 60, 61, 62, 63,
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
		   result[0]:  1-D7- 1-D5- 1-D3- 1-D1
		   result[1]:  1-D6- 1-D4- 1-D2- 1-D0
	*/
	e := make([]byte, 2)
	e[0] = ((b >> 1) & 0x55) | 0xaa
	e[1] = (b & 0x55) | 0xaa
	return e
}

func oddEvenDecodeByte(b0, b1 byte) byte {
	/*
		A byte is encoded in two bytes to make sure the bytes start with 1 and
		does not have two consecutive zeros.
		   b0:      1-D7- 1-D5- 1-D3- 1-D1
		   b1:      1-D6- 1-D4- 1-D2- 1-D0
		   result: D7-D6-D5-D4-D3-D2-D1-D0
	*/
	return ((b0 & 0x55) << 1) | (b1 & 0x55)
}

const (
	diskPrologByte1        = uint8(0xd5)
	diskPrologByte2        = uint8(0xaa)
	diskPrologByte3Address = uint8(0x96)
	diskPrologByte3Data    = uint8(0xad)
)

func nibEncodeTrack(data []byte, volume byte, track byte, logicalOrder *[16]int) []byte {
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
		files they are in physical order.
		*/
		logicalSector := logicalOrder[physicalSector]
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

func findProlog(diskPrologByte3 uint8, data []byte, position int) int {
	l := len(data)
	for i := position; i < l; i++ {
		if (data[i] == diskPrologByte1) &&
			(data[(i+1)%l] == diskPrologByte2) &&
			(data[(i+2)%l] == diskPrologByte3) {

			return (i + 3) % l
		}
	}

	return -1
}

func nibDecodeTrack(data []byte, logicalOrder *[16]int) ([]byte, error) {
	b := make([]byte, bytesPerTrack) // Buffer slice with enough capacity

	i := int(0)
	l := len(data)

	for {
		// Find address field prolog
		i = findProlog(diskPrologByte3Address, data, i)
		if i == -1 {
			break
		}

		// We just want the sector from the address field, we ignore the rest, no error detection
		sector := oddEvenDecodeByte(data[(i+4)%l], data[(i+5)%l])
		logicalSector := logicalOrder[sector]
		dst := int(logicalSector) * bytesPerSector

		// Find data prolog
		i = (i + 8 + 3) % l // We skip the four two byte fields and the epilog
		i = findProlog(diskPrologByte3Data, data, i)

		// Read secondary buffer
		prevV := byte(0)
		for j := 0; j < secondaryBufferSize; j++ {
			w := sixAndTwoUntranslateTable[data[i%l]]
			if w == -1 {
				return nil, errors.New("Invalid byte from nib data")
			}
			v := byte(w) ^ prevV
			prevV = v
			for k := 0; k < 3; k++ {
				// The elements of the secondary buffer add two bits to three bytes
				offset := j + k*secondaryBufferSize
				if offset < bytesPerSector {
					b[dst+offset] |= ((v & 0x02) >> 1) | ((v & 0x01) << 1)
				}
				v >>= 2
			}
			i++
		}

		// Read primary buffer
		for j := 0; j < primaryBufferSize; j++ {
			w := sixAndTwoUntranslateTable[data[i%l]]
			if w == -1 {
				return nil, errors.New("Invalid byte from nib data")
			}
			v := byte(w) ^ prevV
			b[dst+j] |= v << 2 // The elements of the secondary buffer are the 6 MSB bits
			prevV = v
			i++
		}
	}

	return b, nil
}
