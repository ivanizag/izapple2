package apple2

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

type disketteWoz struct {
	version  int
	info     woz2Info
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
	Version        uint8 // 2
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

func loadDisquetteWoz(filename string) (*disketteWoz, error) {
	var d disketteWoz

	data, err := loadResource(filename)
	if err != nil {
		return nil, err
	}

	// Verify header. Note, the CRC is not verified
	header := data[:len(headerWoz2)]
	if bytes.Equal(headerWoz1, header) {
		d.version = 1
	} else if bytes.Equal(headerWoz2, header) {
		d.version = 2
	} else {
		return nil, errors.New("Invalid WOZ header")
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
			return nil, errors.New("Invalid chunk in WOZ file")
		}

		id := string(chunkHeader.ID[:])
		chunks[id] = data[i:iNext]
		i = iNext

		//fmt.Printf("Chunk %v, size %v - %v\n", id, chunkHeader.Size, len(chunks[id]))
	}

	// Read the INFO chunk
	infoData, ok := chunks["INFO"]
	if !ok {
		return nil, errors.New("Chunk INFO missing from WOZ file")
	}
	switch d.version {
	case 1:
		binary.Read(bytes.NewReader(infoData), binary.LittleEndian, &d.info.woz1Info)
	case 2:
		binary.Read(bytes.NewReader(infoData), binary.LittleEndian, &d.info)
	}

	// Read the optional META chunk
	metaData, ok := chunks["META"]
	if ok {
		d.meta = make(map[string]string)
		text := string(metaData)
		entries := strings.Split(text, "\n")
		for _, entry := range entries {
			parts := strings.Split(entry, "\t")
			if len(parts) >= 2 {
				d.meta[parts[0]] = parts[1]
				//fmt.Printf("*** %v: %v\n", parts[0], parts[1])
			}
		}
	}

	// Read the TMAP chunk
	trackMap, ok := chunks["TMAP"]
	if !ok {
		return nil, errors.New("Chunk INFO missing from WOZ file")
	}
	d.trackMap = trackMap

	// Read the TRKS chunk
	tracksData, ok := chunks["TRKS"]
	if d.version == 1 {
		i := 0
		track := 0
		for i+woz1TrackDataSize <= len(tracksData) {
			var trackFooter woz1TrackFooter
			binary.Read(bytes.NewReader(tracksData[i+woz1TrackFooterOffset:]), binary.LittleEndian, &trackFooter)
			d.tracks[track].bitCount = uint32(trackFooter.BitCount)
			d.tracks[track].data = tracksData[i : i+int(trackFooter.BytesUsed)]
			i += woz1TrackDataSize
			track++
		}
	} else if d.version == 2 {
		reader := bytes.NewReader(tracksData)
		for i := 0; i < wozMaxTrack; i++ {
			var trackHeader woz2TrackHeader
			binary.Read(reader, binary.LittleEndian, &trackHeader)
			if trackHeader.BitCount != 0 {
				d.tracks[i].bitCount = trackHeader.BitCount

				dataPos := woz2TrackBlockSize*(int(trackHeader.StartingBlock)-woz2FirstTrackBlock) + woz2TrackBitsOffset
				dataSize := woz2TrackBlockSize * int(trackHeader.BlockCount)
				//fmt.Printf("@%v %v:%v (%v) of %v\n", trackHeader.StartingBlock, dataPos, dataPos+dataSize, dataSize, len(tracksData))
				d.tracks[i].data = tracksData[dataPos : dataPos+dataSize]
			}
		}
	} else {
		return nil, errors.New("Woz version not supported")
	}

	return &d, nil
}

func (d *disketteWoz) dump() {
	fmt.Printf("Woz image:\n")
	fmt.Printf("  Version: %v\n", d.info.Version)
	fmt.Printf("  Disk type: %v\n", d.info.DiskType)
	fmt.Printf("  Write protected: %v\n", d.info.WriteProtected)
	fmt.Printf("  Synchronized: %v\n", d.info.Synchronized)
	fmt.Printf("  Cleaned: %v\n", d.info.Cleaned)
	fmt.Printf("  Creator: %v\n", string(d.info.Creator[:]))
	if d.info.Version >= 2 {
		fmt.Printf("  Disk sides: %v\n", d.info.DiskSides)
		fmt.Printf("  Boot sector format: %v\n", d.info.BootSectorFormat)
		fmt.Printf("  Optimal bit timing: %v ns\n", 125*int(d.info.OptimalBitTiming))
		fmt.Printf("  Compatible hardware: 0x%x\n", d.info.CompatibleHardware)
		fmt.Printf("  Required RAM: %vKB\n", d.info.RequiredRAM)
		fmt.Printf("  Largest track: %v blocks\n", d.info.LargestTrack)
	}
	if d.meta != nil {
		fmt.Printf("  Metadata:\n")
		for k, v := range d.meta {
			fmt.Printf("    %v: %v\n", k, v)
		}
	}
	fmt.Printf("   Tracks:\n")
	for i, track := range d.trackMap {
		if track != 255 {
			fmt.Printf("    Track %.2f: %v (%v bits, %v bytes)\n",
				0.25*float32(i), track, d.tracks[track].bitCount, len(d.tracks[track].data))
		}
	}
}
