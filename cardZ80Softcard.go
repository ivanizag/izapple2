package izapple2

import (
	"fmt"

	"github.com/koron-go/z80"
)

/*

Microsoft Z80 SoftCard
See:
	http://mirrors.apple2.org.za/Apple%20II%20Documentation%20Project/Interface%20Cards/Z80%20Cards/Microsoft%20SoftCard/


This card activates DMA to take control of the system. DMA is actiavted or
deactivated by writing to the Csxx area.

The emulation works on the Apple II+, but doesn't work when 80 columns are
available. It is not working then on the Apple IIe or the Apple II+ with a
Videx card.


*/

// CardVidHD represents a VidHD card
type CardZ80SoftCard struct {
	cardBase

	cpu       *z80.CPU
	z80Active bool
}

func newCardZ80SoftCardBuilder() *cardBuilder {
	return &cardBuilder{
		name:        "Microsoft Z80 SoftCard",
		description: "Microsoft Z80 SoftCard to run CP/M",
		buildFunc: func(params map[string]string) (Card, error) {
			var c CardZ80SoftCard
			c.romCxxx = &cardROMWriteTrap{
				callback: func() {
					c.flipDMA()
				},
			}

			return &c, nil
		},
	}
}

func (c *CardZ80SoftCard) assign(a *Apple2, slot int) {
	mem := &cardZ80SoftCardMMU{
		mmu: a.mmu,
	}

	c.cpu = &z80.CPU{
		Memory: mem,
	}

	c.cardBase.assign(a, slot)
}

func (c *CardZ80SoftCard) reset() {
	c.cpu.States = z80.States{}
}

func (c *CardZ80SoftCard) flipDMA() {
	c.tracef("Z80 DMA flip\n")
	c.z80Active = !c.z80Active
	if c.z80Active {
		c.activateDMA()
	} else {
		c.deactivateDMA()
	}
}

func (c *CardZ80SoftCard) runDMACycle() {
	if c.a.cpuTrace {
		fmt.Printf("Z80 PC: $%04X, A: $%02X, B: $%02X, C: $%02X, D: $%02X, E: $%02X, HL: $%04X\n",
			c.cpu.States.PC, c.cpu.States.AF.Hi, c.cpu.States.BC.Hi,
			c.cpu.States.BC.Lo, c.cpu.States.DE.Hi, c.cpu.States.DE.Lo,
			c.cpu.States.HL)
	}
	c.cpu.Step()
}

type cardROMWriteTrap struct {
	callback func()
}

func (r *cardROMWriteTrap) peek(address uint16) uint8 {
	return 0
}

func (r *cardROMWriteTrap) poke(address uint16, value uint8) {
	if address >= 0xC000 && address < 0xC800 {
		r.callback()
	}
}

type cardZ80SoftCardMMU struct {
	mmu *memoryManager
}

func (m *cardZ80SoftCardMMU) Get(addr uint16) uint8 {
	return m.mmu.Peek(z80AddressTranslation(addr))
}

func (m *cardZ80SoftCardMMU) Set(addr uint16, value uint8) {
	m.mmu.Poke(z80AddressTranslation(addr), value)
}

func z80AddressTranslation(addr uint16) uint16 {
	if addr < 0xb000 {
		return addr + 0x1000
	} else if addr < 0xe000 {
		return addr + 0x2000
	} else if addr < 0xf000 {
		return addr - 0x2000
	} else {
		return addr - 0xf000
	}
}
