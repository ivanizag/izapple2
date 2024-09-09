package storage

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

/*
See:
	https://applesaucefdc.com/woz/
*/

type FileWoz struct {
	version  int
	Info     woz2Info
	trackMap []uint8
	tracks   [wozMaxTrack]disketteTrackWoz
	meta     map[string]string
}

type disketteTrackWoz struct {
	bitCount uint32
	data     []uint8
}

// Structures from the WOZ Disk Image Reference for deserialization
type wozChunkHeader struct {
	ID   [4]byte
	Size uint32
}

type woz1Info struct {
	Version        uint8
	DiskType       uint8
	WriteProtected uint8
	Synchronized   uint8
	Cleaned        uint8
	Creator        [32]byte
}

type woz2Info struct {
	woz1Info
	DiskSides          uint8
	BootSectorFormat   uint8
	OptimalBitTiming   uint8
	CompatibleHardware uint16
	RequiredRAM        uint16
	LargestTrack       uint16
}

type woz1TrackFooter struct {
	BytesUsed      uint16
	BitCount       uint16
	SplicePoint    uint16
	SpliceNibble   uint8
	SpliceBitCount uint8
	Reserved       uint16
}

type woz2TrackHeader struct {
	StartingBlock uint16
	BlockCount    uint16
	BitCount      uint32
}

const (
	wozFirstChunkPos      = 12
	wozChunkHeaderLen     = 8
	wozMaxTrack           = 160
	woz1TrackDataSize     = 6656
	woz1TrackFooterOffset = 6646
	woz2TrackBlockSize    = 512
	woz2FirstTrackBlock   = 3 // The bits on the TRKS block start on 3*512
	woz2TrackBitsOffset   = 1280
)

var headerWoz1 = []uint8{0x57, 0x4f, 0x5A, 0x31, 0xFF, 0x0A, 0x0D, 0x0A}
var headerWoz2 = []uint8{0x57, 0x4f, 0x5A, 0x32, 0xFF, 0x0A, 0x0D, 0x0A}

func (f *FileWoz) GetNextBitAndPosition(position uint32, positionMax uint32, quarterTrack int) (bool, uint32, uint32) {
	if positionMax == 0 {
		// First unitialised use
		positionMax = ^uint32(0) // MaxUint32
	}

	position++
	position %= positionMax

	trackIndex := f.trackMap[quarterTrack]
	if trackIndex == 0xff {
		// No track defined
		// TODO: return random value
		return false, position, positionMax
	}
	trackWoz := f.tracks[trackIndex]

	if trackWoz.bitCount != positionMax {
		// Adjust position as tracks have different length
		position = position * trackWoz.bitCount / positionMax
		positionMax = trackWoz.bitCount
	}

	value := (trackWoz.data[position/8] >> (7 - position%8) & 1) == 1
	return value, position, positionMax
}

func (f *FileWoz) SetBit(value bool, position uint32, positionMax uint32, quarterTrack int) {
	// The position is not moved, GetNextBitAndPosition() would have been called previously.

	trackIndex := f.trackMap[quarterTrack]
	if trackIndex == 0xff {
		// No track defined. Nothing is saved
		return
	}
	trackWoz := f.tracks[trackIndex]

	mask := uint8(1) << (7 - position%8)
	if value {
		trackWoz.data[position/8] |= mask
	} else {
		trackWoz.data[position/8] &= ^mask
	}

	fmt.Printf("Saving %v at %v\n", value, position)
}

func isFileWoz(data []uint8) bool {
	header := data[:len(headerWoz2)]
	if bytes.Equal(headerWoz1, header) {
		return true
	}
	if bytes.Equal(headerWoz2, header) {
		return true
	}
	return false
}

func NewFileWoz(data []uint8) (*FileWoz, error) {
	var f FileWoz

	// Verify header. Note, the CRC is not verified
	header := data[:len(headerWoz2)]
	if bytes.Equal(headerWoz1, header) {
		f.version = 1
	} else if bytes.Equal(headerWoz2, header) {
		f.version = 2
	} else {
		return nil, errors.New("invalid WOZ header")
	}

	// Extract the chunks
	i := wozFirstChunkPos
	var chunkHeader wozChunkHeader
	chunks := make(map[string][]uint8)
	for i+wozChunkHeaderLen < len(data) {
		binary.Read(bytes.NewReader(data[i:]), binary.LittleEndian, &chunkHeader)

		i += wozChunkHeaderLen
		iNext := i + int(chunkHeader.Size)
		if i == iNext || iNext > len(data) {
			return nil, errors.New("invalid chunk in WOZ file")
		}

		id := string(chunkHeader.ID[:])
		chunks[id] = data[i:iNext]
		i = iNext

		// fmt.Printf("Chunk %v, size %v - %v\n", id, chunkHeader.Size, len(chunks[id]))
	}

	// Read the INFO chunk
	infoData, ok := chunks["INFO"]
	if !ok {
		return nil, errors.New("chunk INFO missing from WOZ file")
	}
	switch f.version {
	case 1:
		binary.Read(bytes.NewReader(infoData), binary.LittleEndian, &f.Info.woz1Info)
	case 2:
		binary.Read(bytes.NewReader(infoData), binary.LittleEndian, &f.Info)
	}

	// Read the optional META chunk
	metaData, ok := chunks["META"]
	if ok {
		f.meta = make(map[string]string)
		text := string(metaData)
		entries := strings.Split(text, "\n")
		for _, entry := range entries {
			parts := strings.Split(entry, "\t")
			if len(parts) >= 2 {
				f.meta[parts[0]] = parts[1]
				// fmt.Printf("*** %v: %v\n", parts[0], parts[1])
			}
		}
	}

	// Read the TMAP chunk
	trackMap, ok := chunks["TMAP"]
	if !ok {
		return nil, errors.New("chunk TMAP missing from WOZ file")
	}
	f.trackMap = trackMap

	// Read the TRKS chunk
	tracksData, ok := chunks["TRKS"]
	if !ok {
		return nil, errors.New("chunk TRKS missing from WOZ file")
	}
	if f.version == 1 {
		i := 0
		track := 0
		for i+woz1TrackDataSize <= len(tracksData) {
			var trackFooter woz1TrackFooter
			binary.Read(bytes.NewReader(tracksData[i+woz1TrackFooterOffset:]), binary.LittleEndian, &trackFooter)
			f.tracks[track].bitCount = uint32(trackFooter.BitCount)
			f.tracks[track].data = tracksData[i : i+int(trackFooter.BytesUsed)]
			i += woz1TrackDataSize
			track++
		}
	} else if f.version == 2 {
		reader := bytes.NewReader(tracksData)
		for i := 0; i < wozMaxTrack; i++ {
			var trackHeader woz2TrackHeader
			binary.Read(reader, binary.LittleEndian, &trackHeader)
			if trackHeader.BitCount != 0 {
				f.tracks[i].bitCount = trackHeader.BitCount

				dataPos := woz2TrackBlockSize*(int(trackHeader.StartingBlock)-woz2FirstTrackBlock) + woz2TrackBitsOffset
				dataSize := woz2TrackBlockSize * int(trackHeader.BlockCount)
				// fmt.Printf("@%v %v:%v (%v) of %v\n", trackHeader.StartingBlock, dataPos, dataPos+dataSize, dataSize, len(tracksData))
				f.tracks[i].data = tracksData[dataPos : dataPos+dataSize]
			}
		}
	} else {
		return nil, errors.New("woz version not supported")
	}

	return &f, nil
}

func (f *FileWoz) DumpTrackAsWoz(quarterTrack int) []uint8 {
	trackWoz := f.tracks[f.trackMap[quarterTrack]]
	return trackWoz.data
}

func (f *FileWoz) DumpTrackAsNib(quarterTrack int) []uint8 {
	trackWoz := f.tracks[f.trackMap[quarterTrack]]
	out := make([]uint8, 0, trackWoz.bitCount/8)
	latch := uint8(0)
	for iBit := uint32(0); iBit < trackWoz.bitCount; iBit++ {
		bit := trackWoz.data[iBit/8] >> (7 - iBit%8) & 1
		latch = (latch << 1) + bit
		if latch >= 0x80 {
			// Valid reading
			out = append(out, latch)
			latch = 0
		}
	}
	return out
}

func (f *FileWoz) Dump() {
	fmt.Printf("Woz image:\n")
	fmt.Printf("  Version: %v\n", f.Info.Version)
	fmt.Printf("  Disk type: %v\n", f.Info.DiskType)
	fmt.Printf("  Write protected: %v\n", f.Info.WriteProtected)
	fmt.Printf("  Synchronized: %v\n", f.Info.Synchronized)
	fmt.Printf("  Cleaned: %v\n", f.Info.Cleaned)
	fmt.Printf("  Creator: %v\n", string(f.Info.Creator[:]))
	if f.Info.Version >= 2 {
		fmt.Printf("  Disk sides: %v\n", f.Info.DiskSides)
		fmt.Printf("  Boot sector format: %v\n", f.Info.BootSectorFormat)
		fmt.Printf("  Optimal bit timing: %v ns\n", 125*int(f.Info.OptimalBitTiming))
		fmt.Printf("  Compatible hardware: 0x%x\n", f.Info.CompatibleHardware)
		fmt.Printf("  Required RAM: %vKB\n", f.Info.RequiredRAM)
		fmt.Printf("  Largest track: %v blocks\n", f.Info.LargestTrack)
	}
	if f.meta != nil {
		fmt.Printf("  Metadata:\n")
		for k, v := range f.meta {
			fmt.Printf("    %v: %v\n", k, v)
		}
	}
	fmt.Printf("   Tracks:\n")
	for i, track := range f.trackMap {
		if track != 255 {
			fmt.Printf("    Track %.2f: %v (%v bits, %v bytes)\n",
				0.25*float32(i), track, f.tracks[track].bitCount, len(f.tracks[track].data))
		}
	}

	// nibs := f.dumpTrackAsNib(0)
	// fmt.Printf("  Zero track: {%v} %x\n", len(nibs), nibs)
}
