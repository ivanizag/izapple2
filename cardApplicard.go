package izapple2

import (
	"fmt"

	"github.com/koron-go/z80"
)

/*

PCPI Applicard
See:
	http://www.applelogic.org/files/Applicard%20OEM%20Manual.pdf, chapter 4
	https://github.com/hasseily/appletini-one/blob/main/README_APPLICARD.md

The card has a Z80 with 64KB of RAM and a 2KB boot ROM. It coordinates with the 6502 via
two 8 bit latches with a full flag each. There is no slot ROM and no shared memory.

The softswitches on the 6502:
	$C0x0: read the byte from the Z80
	$C0x1: write a byte for the Z80
	$C0x2: bit 7 is 1 if the byte for the Z80 has not yet been read
	$C0x3: bit 7 is 1 if there is a byte from the Z80
	$C0x5: a read or write resets the Z80
	$C0x6: a read or write interrupts the Z80 via channel 3 of the CTC
	$C0x7: a read or write causes a Z80 NMI
The Z80 ports, each one repeated on the 32 ports of its range:
	$00: write a byte for the 6502
	$20: read the byte from the 6502
	$40: bit 7 is 1 if there is a byte from the 6502, bit 0 is 1 if the byte for the
		6502 has not yet been read
	$60: bit 0 to 1 turns the ROM on, to 0 turns it off
	$80: CTC channels 0 to 3

The ROM is on at power up and reset. While on, it is repeated from $0000 to $7FFF and the
lower 32KB of RAM can't be read or written. The ROM firmware writes $5A to the 6502 and
then waits for commands.

The CTC is not emulated. Its ports read $FF, as when the ROM checks for it on a card
without it, and $C0x6 does nothing. The manual says the CTC is initialized with
interrupts disabled.

*/

// CardApplicard represents a PCPI Applicard card
type CardApplicard struct {
	cardBase

	cpu     *z80.CPU
	mem     [65536]uint8
	rom     []uint8
	romIsOn bool

	// Latches
	toZ80       uint8
	toZ80Full   bool
	fromZ80     uint8
	fromZ80Full bool
}

const applicardRomSize = 2 * 1024

// To emulate the Z80 at 6MHz and the 6502 at 1.023MHz, the Z80 runs around 2 instructions per tick
const applicardZ80InstructionsPerTick = 2

func newCardApplicardBuilder() *cardBuilder {
	return &cardBuilder{
		name:        "PCPI Applicard",
		description: "PCPI Applicard Z80 card to run CP/M",
		defaultParams: &[]paramSpec{
			{"rom", "ROM file to load", "<internal>/PCPI Applicard ROM v9.bin"},
		},
		buildFunc: func(params map[string]string) (Card, error) {
			var c CardApplicard

			romFile := paramsGetPath(params, "rom")
			data, _, err := LoadResource(romFile)
			if err != nil {
				return nil, err
			}
			if len(data) != applicardRomSize {
				return nil, fmt.Errorf("the ROM file for the Applicard must be 2Kb, %s has %v bytes", romFile, len(data))
			}
			c.rom = data
			return &c, nil
		},
	}
}

func (c *CardApplicard) assign(a *Apple2, slot int) {
	c.cpu = &z80.CPU{
		Memory: c,
		IO:     c,
	}
	c.reset()

	c.addCardSoftSwitchR(0, func() uint8 {
		c.fromZ80Full = false
		return c.fromZ80
	}, "APPLICARDDATAR")

	c.addCardSoftSwitchR(1, func() uint8 {
		return c.toZ80
	}, "APPLICARDDATAWR")
	c.addCardSoftSwitchW(1, func(value uint8) {
		c.toZ80 = value
		c.toZ80Full = true
	}, "APPLICARDDATAW")

	c.addCardSoftSwitchR(2, func() uint8 {
		return applicardFlag(c.toZ80Full, 0x80)
	}, "APPLICARDTOZ80FULL")

	c.addCardSoftSwitchR(3, func() uint8 {
		return applicardFlag(c.fromZ80Full, 0x80)
	}, "APPLICARDFROMZ80FULL")

	c.addCardSoftSwitchRW(5, func() uint8 {
		c.reset()
		return 0xff
	}, "APPLICARDRESET")

	c.addCardSoftSwitchRW(6, func() uint8 {
		// The CTC is not emulated
		return 0xff
	}, "APPLICARDINTERRUPT")

	c.addCardSoftSwitchRW(7, func() uint8 {
		c.cpu.Interrupt = z80.NMIInterrupt()
		return 0xff
	}, "APPLICARDNMI")

	a.registerTickerCard(c)
	c.cardBase.assign(a, slot)
}

// The RAM keeps its contents on reset. The ROM uses it to skip the memory test after the first boot.
func (c *CardApplicard) reset() {
	c.cpu.States = z80.States{}
	c.cpu.Interrupt = nil
	c.romIsOn = true
	c.toZ80Full = false
	c.fromZ80Full = false
}

func (c *CardApplicard) tick() {
	for i := 0; i < applicardZ80InstructionsPerTick; i++ {
		c.cpu.Step()
	}
}

// Get is the z80.Memory read
func (c *CardApplicard) Get(addr uint16) uint8 {
	if c.romIsOn && addr < 0x8000 {
		return c.rom[addr%applicardRomSize]
	}
	return c.mem[addr]
}

// Set is the z80.Memory write
func (c *CardApplicard) Set(addr uint16, value uint8) {
	if c.romIsOn && addr < 0x8000 {
		return
	}
	c.mem[addr] = value
}

// In is the z80.IO read
func (c *CardApplicard) In(port uint8) uint8 {
	switch port & 0xe0 {
	case 0x00:
		return c.fromZ80
	case 0x20:
		c.toZ80Full = false
		return c.toZ80
	case 0x40:
		return applicardFlag(c.toZ80Full, 0x80) | applicardFlag(c.fromZ80Full, 0x01)
	default:
		return 0xff
	}
}

// Out is the z80.IO write
func (c *CardApplicard) Out(port uint8, value uint8) {
	switch port & 0xe0 {
	case 0x00:
		c.fromZ80 = value
		c.fromZ80Full = true
	case 0x60:
		c.romIsOn = value&1 == 1
	}
}

func applicardFlag(isSet bool, mask uint8) uint8 {
	if isSet {
		return mask
	}
	return 0
}
