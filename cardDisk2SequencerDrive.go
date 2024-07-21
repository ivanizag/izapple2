package izapple2

import (
	"errors"
	"math/rand"

	"github.com/ivanizag/izapple2/component"
	"github.com/ivanizag/izapple2/storage"
)

type cardDisk2SequencerDrive struct {
	data                *storage.FileWoz
	enabled             bool
	writeProtected      bool
	currentQuarterTrack int

	position    uint32 // Current position on the track
	positionMax uint32 // As tracks may have different lengths position is related of positionMax of the las track

	mc3470Buffer uint8 // Four bit buffer to detect weak bits and to add latency
}

func (d *cardDisk2SequencerDrive) insertDiskette(filename string) error {
	data, writeable, err := LoadResource(filename)
	if err != nil {
		return err
	}
	f, err := storage.NewFileWoz(data)
	if err != nil {
		return err
	}

	// Discard not supported features
	if f.Info.DiskType != 1 {
		return errors.New("only 5.25 disks are supported")
	}

	d.data = f
	d.writeProtected = !writeable

	return nil
}

func (d *cardDisk2SequencerDrive) enable(enabled bool) {
	d.enabled = enabled
}

func (d *cardDisk2SequencerDrive) moveHead(q0, q1, q2, q3 bool, trackTracer trackTracer, slot int, driveNumber int) {
	if !d.enabled {
		return
	}

	phases := component.PinsToByte([8]bool{
		q0, q1, q2, q3,
		false, false, false, false,
	})
	d.currentQuarterTrack = moveDriveStepper(phases, d.currentQuarterTrack)

	if trackTracer != nil {
		trackTracer.traceTrack(d.currentQuarterTrack, slot, driveNumber)
	}
}

func (d *cardDisk2SequencerDrive) readPulse() bool {
	if !d.enabled || d.data == nil {
		return false
	}

	// Get next bit taking into account the MC3470 latency and weak bits
	var fluxBit bool
	fluxBit, d.position, d.positionMax = d.data.GetNextBitAndPosition(
		d.position,
		d.positionMax,
		d.currentQuarterTrack)
	d.mc3470Buffer = (d.mc3470Buffer << 1) & 0x0f
	if fluxBit {
		d.mc3470Buffer++
	}
	bit := ((d.mc3470Buffer >> 1) & 0x1) != 0 // Use the previous to last bit to add latency
	if d.mc3470Buffer == 0 && rand.Intn(100) < 30 {
		// Four consecutive zeros. It'a a fake bit.
		// Output a random value. 70% zero, 30% one
		bit = true
	}

	return bit
}

func (d *cardDisk2SequencerDrive) writePulse(value bool) {
	if d.writeProtected || !d.enabled || d.data == nil {
		return
	}

	d.data.SetBit(
		value,
		d.position,
		d.positionMax,
		d.currentQuarterTrack)
}
