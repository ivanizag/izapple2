package izapple2

import (
	"github.com/koron-go/z80"
)

/*

Microsoft Z80 SoftCard II
See:
	http://mirrors.apple2.org.za/Apple%20II%20Documentation%20Project/Interface%20Cards/Z80%20Cards/Microsoft%20SoftCard/
	https://www.applefritter.com/files/2023/04/22/AM2950.pdf
	https://github.com/sbabcoc/SoftCard2/tree/main/src/main/java/com/nordstrom/softcard (that I think comes from Claude)

This card Z80 has its own memory and coordinates with the 6502 via an AMD Am2950 (8-bit bidirectional port)

All the softswitches on the 6502 read and write on the Am2950:
	Read:
		$C0x0: read from Am2950
		$C0x1: read status from Am2950
	Write:
		$C0x0: write to the Am2950
		$C0x3: resets the card
The z80 in and out ports do the same
	In:
		bit0=0: read from Am2950
		bit0=1: read status from Am2950
	Out:
		bit0=0: write to the Am2950

The card start in boot mode. All memory reads below $8000 come from the Am2950 and the
writes below $8000 go to the memory and the Am2950. It exits boot mode on the first OUT. On boot mode
the Z80 may need to wait if the data is not available.

*/

// CardZ80SoftCard2 represents a Microsoft Z80 SoftCard II card
type CardZ80SoftCard2 struct {
	cardBase

	cpu       *z80.CPU
	mem       [65536]uint8
	isBooting bool

	// Am2950 latches
	toZ80       uint8
	toZ80Full   bool
	fromZ80     uint8
	fromZ80Full bool

	// Boot mode: an instruction that runs out of data is undone and retried
	stalled  bool
	fetched  []uint8 // bytes taken from the latch by the instruction being retried
	fetchPos int
}

func newCardZ80SoftCard2Builder() *cardBuilder {
	return &cardBuilder{
		name:        "Microsoft Z80 SoftCard II",
		description: "Microsoft Z80 SoftCard II to run CP/M",
		buildFunc: func(params map[string]string) (Card, error) {
			var c CardZ80SoftCard2
			return &c, nil
		},
	}
}

func (c *CardZ80SoftCard2) assign(a *Apple2, slot int) {
	c.cpu = &z80.CPU{
		Memory: c,
		IO:     c,
	}
	c.reset()

	c.addCardSoftSwitchR(0, func() uint8 {
		c.fromZ80Full = false
		return c.fromZ80
	}, "SOFTCARD2DATAR")

	c.addCardSoftSwitchR(1, func() uint8 {
		return cardZ80SoftCard2Status(c.fromZ80Full, c.toZ80Full)
	}, "SOFTCARD2DATASTATUSR")

	c.addCardSoftSwitchW(0, func(value uint8) {
		c.toZ80 = value
		c.toZ80Full = true
	}, "SOFTCARD2DATAW")

	c.addCardSoftSwitchW(3, func(value uint8) {
		c.reset()
	}, "SOFTCARD2RESET")

	a.registerTickerCard(c)
	c.cardBase.assign(a, slot)
}

func (c *CardZ80SoftCard2) reset() {
	c.cpu.States = z80.States{}
	c.isBooting = true
	c.toZ80Full = false
	c.fromZ80Full = false
	c.stalled = false
	c.fetched = c.fetched[:0]
	c.fetchPos = 0
}

// To emulate the Z80 running at 8MHz and the 6502 at 1.023MHz, the Z80 runs around3 instructions per tick
const z80to6502InstructionsPerTick = 3

func (c *CardZ80SoftCard2) tick() {
	for i := 0; i < z80to6502InstructionsPerTick; i++ {
		saved := c.cpu.States
		c.stalled = false
		c.fetchPos = 0
		c.cpu.Step()
		if c.stalled {
			// Waiting for the 6502. Undo and retry the same instruction next tick
			c.cpu.States = saved
			return
		}
		c.fetched = c.fetched[:0]
	}
}

// Get is the z80.Memory read
func (c *CardZ80SoftCard2) Get(addr uint16) uint8 {
	if !c.isBooting || addr >= 0x8000 {
		return c.mem[addr]
	}
	if c.fetchPos < len(c.fetched) {
		// Replay a byte the previous attempt already took
		v := c.fetched[c.fetchPos]
		c.fetchPos++
		return v
	}
	if c.stalled || !c.toZ80Full {
		c.stalled = true
		return 0 // dummy, the result is discarded
	}
	c.toZ80Full = false // the 6502 can send the next byte
	c.fetched = append(c.fetched, c.toZ80)
	c.fetchPos++
	return c.toZ80
}

// Set is the z80.Memory write
func (c *CardZ80SoftCard2) Set(addr uint16, value uint8) {
	if c.stalled {
		return
	}
	if c.isBooting && addr < 0x8000 {
		c.fromZ80 = value
		c.fromZ80Full = true
	}
	c.mem[addr] = value
}

// In is the z80.IO read
func (c *CardZ80SoftCard2) In(port uint8) uint8 {
	if c.stalled {
		return 0
	}
	if port&1 == 1 {
		return cardZ80SoftCard2Status(c.toZ80Full, c.fromZ80Full)
	}
	c.toZ80Full = false
	return c.toZ80
}

// Out is the z80.IO write
func (c *CardZ80SoftCard2) Out(port uint8, value uint8) {
	if c.stalled {
		return
	}
	if c.isBooting {
		c.isBooting = false
		return
	}
	c.fromZ80 = value
	c.fromZ80Full = true
}

// Support function
func cardZ80SoftCard2Status(hasReadData bool, hasWriteData bool) uint8 {
	status := uint8(0)
	if hasReadData {
		status = 0x80
	}
	if hasWriteData {
		status += 0x01
	}
	return status
}
