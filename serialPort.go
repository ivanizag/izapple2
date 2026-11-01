package izapple2

import (
	"github.com/ivanizag/izapple2/component"
)

/*
serialPort is a 6551 ACIA with a device of the host on its line, as used by
the Super Serial Card and the serial ports of the Apple IIc. The ACIA is on
$C0n8 to $C0nB of the slot.

DCD is low when the device is ready. DSR is wired by the card, dsrHigh
returns its level.
*/
type serialPort struct {
	a         *Apple2
	acia      *component.MOS6551
	device    serialDevice
	irqSource int
	irqWired  bool
	dsrHigh   func() bool

	lastCycle  uint64 // The ACIA is caught up to this cycle
	nextTick   uint64 // Cycle for the ticker to catch up again
	nextLevels uint64 // Cycle to look again at the device readiness
	ready      bool
}

const (
	serialPortRegisters  = 0x08
	serialPortReadyCheck = 1000 // Cycles between looks at the device readiness

	// Cycles between catch ups of the ticker. The accesses to the registers
	// catch up on their own, this is for the bytes received, the inputs and
	// the interrupts. A byte is more than 500 cycles at 19200 baud.
	serialPortTickCycles = 64
)

func newSerialPort(device serialDevice) *serialPort {
	var p serialPort
	p.device = device
	p.acia = component.NewMOS6551(CPUClockMhz*1_000_000, device)
	return &p
}

// assign wires the port to the IRQ source and the DSR input, and registers
// its ticker
func (p *serialPort) assign(a *Apple2, c *cardBase, irqSource int, irqWired bool, dsrHigh func() bool) {
	p.a = a
	p.dsrHigh = dsrHigh
	p.irqSource = irqSource
	p.irqWired = irqWired
	p.lastCycle = a.GetCycles()
	a.registerTicker(p)

	names := []string{"ACIADATA", "ACIASTATUS", "ACIACOMMAND", "ACIACONTROL"}
	for reg := range uint8(4) {
		c.addCardSoftSwitchR(serialPortRegisters+reg, func() uint8 {
			p.advance()
			value := p.acia.Read(reg)
			p.updateIRQ()
			return value
		}, names[reg]+"R")
		c.addCardSoftSwitchW(serialPortRegisters+reg, func(value uint8) {
			p.advance()
			p.acia.Write(reg, value)
			p.updateIRQ()
		}, names[reg]+"W")
	}
}

func (p *serialPort) reset() {
	p.acia.Reset()
	p.updateIRQ()
}

func (p *serialPort) tick() {
	if p.a.GetCycles() >= p.nextTick {
		p.catchUp()
		p.nextTick = p.lastCycle + serialPortTickCycles
	}
}

/*
advance moves the time of the ACIA before an access to its registers. What
comes from outside, the bytes received and the changes of DCD and DSR, waits
for the ticker: it has to come between instructions, for the interrupt to be
taken before the program reads the status. If not, the polling loops of the
firmware read the status first, clearing the interrupt, and take the byte
out of order with the ones its interrupt handler keeps in a buffer.
*/
func (p *serialPort) advance() {
	current := p.a.GetCycles()
	if current > p.lastCycle {
		p.acia.Advance(current - p.lastCycle)
		p.lastCycle = current
	}
}

// catchUp moves the time, receives and updates the inputs and the interrupt
func (p *serialPort) catchUp() {
	current := p.a.GetCycles()
	if current < p.lastCycle {
		return
	}
	p.acia.Tick(current - p.lastCycle)
	p.lastCycle = current
	if current >= p.nextLevels {
		p.ready = p.device.isReady()
		p.nextLevels = current + serialPortReadyCheck
	}
	p.acia.SetDCD(!p.ready)
	p.acia.SetDSR(p.dsrHigh())
	p.updateIRQ()
}

// isReady returns whether the device can take what is sent
func (p *serialPort) isReady() bool {
	return p.ready
}

func (p *serialPort) updateIRQ() {
	p.a.requestIRQ(p.irqSource, p.irqWired && p.acia.InterruptAsserted())
}
