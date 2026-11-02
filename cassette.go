package izapple2

import (
	"errors"
	"fmt"

	"github.com/ivanizag/izapple2/storage"
)

/*
Emulation of the cassette tape input.

Reads of $C060 return the tape comparator output on bit 7, and the
Monitor ROM measures the time between bit flips to decode the
recorded data: a 770 Hz header tone, a short sync cycle, then 2 kHz
cycles for "0" bits and 1 kHz cycles for "1" bits.

The tape, loaded by the storage package, is the list of CPU cycles,
from the start of the tape, at which the signal crosses zero.

The tape starts playing on the first read of $C060 and pauses when
the program stops polling, so no play or rewind controls are needed.
Fast mode is requested while the tape is playing.

LoadTape puts another tape in the recorder, from its start, while the
machine runs.

See:
  - "Apple II Reference Manual", cassette interface
  - https://retrocomputing.stackexchange.com/questions/143/what-format-is-used-for-apple-ii-cassette-tapes
*/

// Cycles without reads of $C060 before the tape is paused. It must be
// longer than the 3.5 seconds delay of the Monitor ROM READ routine so
// that the tape plays through the delay as it does on real hardware,
// skipping the start of the tape and any noise before the header tone.
const cassetteAutoPauseCycles = 5_000_000

type cassette struct {
	a           *Apple2
	transitions []uint64 // Cycles from the tape start with a zero crossing
	cursor      int
	playing     bool
	position    uint64 // Tape position in cycles while paused
	startCycle  uint64 // Cycle of the tape start position while playing
	lastRead    uint64
}

// newCassette loads a WAV recording of a tape and prepares it to be
// read on the cassette input softswitch
func newCassette(a *Apple2, data []uint8) (*cassette, error) {
	transitions, err := storage.MakeTape(data, CPUClockMhz*1_000_000)
	if err != nil {
		return nil, err
	}

	var c cassette
	c.a = a
	c.transitions = transitions

	a.registerTicker(&c)
	return &c, nil
}

// insert replaces the tape with another one, stopped at its start
func (c *cassette) insert(transitions []uint64) {
	if c.playing {
		c.playing = false
		c.a.ReleaseFastMode()
	}
	c.transitions = transitions
	c.cursor = 0
	c.position = 0
}

/*
LoadTape puts a tape in the cassette recorder, a WAV recording from a file,
an URL or inside a zip, as with -tape. It replaces the tape there was,
and starts at its beginning. The Apple IIc has no cassette input.
*/
func (a *Apple2) LoadTape(path string) error {
	if a.isApple2c {
		return errors.New("the Apple IIc has no cassette input")
	}
	data, _, err := LoadResource(path)
	if err != nil {
		return err
	}
	if a.io.cassette == nil {
		c, err := newCassette(a, data)
		if err != nil {
			return fmt.Errorf("could not load the tape %s: %w", path, err)
		}
		a.io.setCassette(c)
		return nil
	}

	transitions, err := storage.MakeTape(data, CPUClockMhz*1_000_000)
	if err != nil {
		return fmt.Errorf("could not load the tape %s: %w", path, err)
	}
	a.io.cassette.insert(transitions)
	return nil
}

// read returns the comparator output on bit 7 for a read of $C060
func (c *cassette) read(cycle uint64) uint8 {
	if c.cursor < len(c.transitions) {
		if !c.playing {
			// Auto play on read, continuing from the last position
			c.playing = true
			c.startCycle = cycle - c.position
			c.a.RequestFastMode()
			c.a.requestTickAt(cycle + cassetteAutoPauseCycles + 1)
		}
		c.lastRead = cycle

		elapsed := cycle - c.startCycle
		for c.cursor < len(c.transitions) && c.transitions[c.cursor] <= elapsed {
			c.cursor++
		}
		if c.cursor == len(c.transitions) {
			// End of the tape
			c.pause(cycle)
		}
	}

	return uint8(c.cursor&1) << 7
}

// tick pauses the tape when the program stops polling the cassette
// input, keeping the position and releasing fast mode
func (c *cassette) tick() uint64 {
	if !c.playing {
		return tickNever
	}
	cycle := c.a.GetCycles()
	if cycle-c.lastRead > cassetteAutoPauseCycles {
		c.pause(cycle)
		return tickNever
	}
	return c.lastRead + cassetteAutoPauseCycles + 1
}

func (c *cassette) pause(cycle uint64) {
	// Rewind to the last transition delivered so that on resume the
	// next transition arrives a full half-cycle later. Resuming at an
	// arbitrary point could deliver a short half-cycle that the ROM
	// would mistake for the sync cycle.
	c.position = 0
	if c.cursor > 0 {
		c.position = c.transitions[c.cursor-1]
	}
	c.playing = false
	c.a.ReleaseFastMode()
}
