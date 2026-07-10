package izapple2

import (
	"github.com/ivanizag/izapple2/component"
)

/*
Mockingboard sound card. Two MOS 6522 VIAs driving two AY-3-8913
programmable sound generators.

See:
	https://en.wikipedia.org/wiki/Mockingboard
	https://github.com/AppleWin/AppleWin/blob/master/source/Mockingboard.cpp

The card is addressed on its slot ROM page, there is no ROM:

	$Cn00-$Cn0F: VIA 1 registers, mirrored up to $Cn7F
	$Cn80-$Cn8F: VIA 2 registers, mirrored up to $CnFF

Each VIA drives one AY-3-8913:

	Port A: data bus
	Port B bit 0: BC1
	Port B bit 1: BDIR
	Port B bit 2: /RESET, active low

The IRQ outputs of both VIAs are wired to the 6502 IRQ line. The VIA
timers are caught up on every access and on every instruction, so the
timer reads used by the Mockingboard detection routines are cycle exact.

The mixed output level of both PSGs is reported to the frontend as an
AudioSource, stepping the chips every 8 CPU cycles.
*/
type CardMockingboard struct {
	cardBase
	via       [2]component.MOS6522
	psg       [2]component.AY38913
	lastBusOp [2]uint8

	sink      AudioSink
	lastCycle uint64 // The chips are caught up to this cycle
	psgCycle  uint64 // Start of the next PSG synthesis step
	lastLevel float32
}

const (
	mockingboardBusRead  uint8 = 1
	mockingboardBusWrite uint8 = 2
	mockingboardBusLatch uint8 = 3

	mockingboardPsgStepCycles = 8
)

func newCardMockingboardBuilder() *cardBuilder {
	return &cardBuilder{
		name:        "Mockingboard",
		description: "Mockingboard sound card with two AY-3-8913 sound generators",
		buildFunc: func(params map[string]string) (Card, error) {
			var c CardMockingboard
			return &c, nil
		},
	}
}

func (c *CardMockingboard) assign(a *Apple2, slot int) {
	c.cardBase.assign(a, slot)
	if slot != 0 {
		a.mmu.setCardROM(slot, traceMemory(c, c.name, c.traceMemory))
	}
	a.registerTickerCard(c)
	c.lastCycle = a.GetCycles()
	c.psgCycle = c.lastCycle
	c.psg[0].Reset()
	c.psg[1].Reset()
}

func (c *CardMockingboard) reset() {
	for i := range 2 {
		c.via[i].Reset()
		c.psg[i].Reset()
		c.lastBusOp[i] = 0
	}
	if c.a != nil {
		c.a.requestIRQ(c.slot, false)
	}
}

// peek and poke serve the $Cn00-$CnFF page, implementing memoryHandler
func (c *CardMockingboard) peek(address uint16) uint8 {
	c.catchUp()
	n := int(address>>7) & 1
	value := c.via[n].Read(uint8(address & 0x0f))
	c.updateIRQ() // Some register reads clear interrupt flags
	return value
}

func (c *CardMockingboard) poke(address uint16, value uint8) {
	c.catchUp()
	n := int(address>>7) & 1
	reg := uint8(address & 0x0f)
	c.via[n].Write(reg, value)
	if reg == 0 || reg == 2 {
		// ORB or DDRB writes can change the PSG bus control lines
		c.updatePsgBus(n)
	}
	c.updateIRQ()
}

// updatePsgBus runs the AY-3-8913 bus protocol wired to the VIA port B
func (c *CardMockingboard) updatePsgBus(n int) {
	portB := c.via[n].GetPortB()
	if portB&0x04 == 0 {
		// /RESET is active low
		c.psg[n].Reset()
		c.lastBusOp[n] = 0
		return
	}

	// The operations execute when BDIR-BC1 change
	busOp := portB & 0x03
	if busOp != c.lastBusOp[n] {
		switch busOp {
		case mockingboardBusRead:
			c.via[n].SetInputA(c.psg[n].ReadData())
		case mockingboardBusWrite:
			c.psg[n].WriteData(c.via[n].GetPortA())
		case mockingboardBusLatch:
			c.psg[n].LatchAddress(c.via[n].GetPortA())
		}
		c.lastBusOp[n] = busOp
	}
}

// tick is called on every instruction to keep the timers, the interrupt
// line and the sound synthesis up to date
func (c *CardMockingboard) tick() {
	c.catchUp()
}

func (c *CardMockingboard) catchUp() {
	current := c.a.GetCycles()
	if current <= c.lastCycle {
		return
	}
	elapsed := current - c.lastCycle
	c.lastCycle = current

	c.via[0].Tick(elapsed)
	c.via[1].Tick(elapsed)
	c.updateIRQ()
	c.synthesize(current)
}

func (c *CardMockingboard) updateIRQ() {
	c.a.requestIRQ(c.slot, c.via[0].InterruptAsserted() || c.via[1].InterruptAsserted())
}

// synthesize advances both PSGs up to the given cycle in steps of 8 CPU
// cycles, reporting the changes of the mixed level to the audio sink
func (c *CardMockingboard) synthesize(toCycle uint64) {
	if c.sink == nil {
		c.psgCycle = toCycle
		return
	}

	for ; c.psgCycle+mockingboardPsgStepCycles <= toCycle; c.psgCycle += mockingboardPsgStepCycles {
		// Each PSG ranges from 0.0 to 3.0, scale the sum of both to a
		// range comparable to the speaker
		level := (c.psg[0].Step() + c.psg[1].Step()) * 0.25
		if level != c.lastLevel {
			c.sink.PushLevel(c.psgCycle, level)
			c.lastLevel = level
		}
	}
}

// GetAudioSourceName implements the AudioSource interface
func (c *CardMockingboard) GetAudioSourceName() string {
	return "mockingboard"
}

// SetAudioSink implements the AudioSource interface
func (c *CardMockingboard) SetAudioSink(sink AudioSink) {
	c.sink = sink
}
