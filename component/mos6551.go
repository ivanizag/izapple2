package component

/*
MOS 6551 Asynchronous Communication Interface Adapter (ACIA)
See:

	"Apple IIc Technical Reference Manual", Chapter 11, Serial I/O, with the
	Synertek 6551 register figures, and Appendix E, Interrupts.
	https://archive.org/details/Apple_IIc_Technical_Reference_Manual
	https://westerndesigncenter.com/documentation/w65c51n.pdf

Used by the Super Serial Card and by the two serial ports of the Apple IIc.

Registers:

	0: Transmit/receive data   1: Status (writing it is a program reset)
	2: Command                 3: Control

The bytes are moved whole, not bit by bit. A byte takes on the line the time
of its start, data, parity and stop bits at the baud rate of the control
register. The byte received waits on the line until the receive data register
is free, there are no overruns.

The DCD and DSR inputs are levels, true is low. A change on any of them
interrupts when DTR is on, the receiver interrupt disable bit does not mask
it: the IIc firmware takes the keyboard interrupts on DSR with that bit set.
The status shows the level after the change, and the next change waits for
the status to be read.

CTS is always low, the transmitter is never stopped by it. The break and
the parity bits are not sent, and there are no parity or framing errors.
*/
type MOS6551 struct {
	line       SerialLine
	cyclesBaud float64 // CPU cycles per second divided by the baud rate unit

	control uint8
	command uint8
	status  uint8

	tdr       uint8
	shift     uint8
	shifting  bool
	shiftDone uint64 // Cycle when the byte in the shift register is sent
	rdr       uint8
	rxNext    uint64 // Cycle when the line is looked at for a byte again

	dcd, dsr     bool // Levels of the inputs, true is high
	linesPending bool // A change is on the status, waiting to be read
	cycles       uint64
}

// SerialLine is what is on the other side of the ACIA
type SerialLine interface {
	// Send gets a byte the ACIA has sent
	Send(value uint8)
	// Receive returns a byte for the ACIA, if there is any
	Receive() (value uint8, ok bool)
}

const (
	mos6551StatusIRQ  uint8 = 1 << 7
	mos6551StatusDSR  uint8 = 1 << 6
	mos6551StatusDCD  uint8 = 1 << 5
	mos6551StatusTDRE uint8 = 1 << 4
	mos6551StatusRDRF uint8 = 1 << 3
	mos6551StatusOVRN uint8 = 1 << 2

	mos6551CommandDTR uint8 = 1 << 0
	mos6551CommandIRD uint8 = 1 << 1 // Receiver interrupt disabled
	mos6551CommandTIC uint8 = 3 << 2 // Transmitter control
	mos6551CommandREM uint8 = 1 << 4 // Receiver echo mode

	mos6551TransmitterOff uint8 = 0 << 2
	mos6551TransmitterIRQ uint8 = 1 << 2

	mos6551ControlStopBits uint8 = 1 << 7
	mos6551ControlRate     uint8 = 0x0f

	// The 16x external clock is the 1.8432 MHz crystal, as the internal
	// generator uses. Divided by 16 it is 115200 baud.
	mos6551ExternalClockBaud = 115200
)

// Baud rates of the internal generator, 0 is the external clock
var mos6551BaudRates = [16]float64{
	mos6551ExternalClockBaud, 50, 75, 109.92, 134.58, 150, 300, 600,
	1200, 1800, 2400, 3600, 4800, 7200, 9600, 19200,
}

// NewMOS6551 returns an ACIA connected to a line. The CPU clock is needed
// for the timing of the bytes.
func NewMOS6551(cpuClockHz float64, line SerialLine) *MOS6551 {
	var a MOS6551
	a.line = line
	a.cyclesBaud = cpuClockHz
	a.dcd = true
	a.dsr = true
	a.Reset()
	return &a
}

// Reset is the hardware reset
func (a *MOS6551) Reset() {
	a.control = 0
	a.command = 0
	a.status = mos6551StatusTDRE
	a.shifting = false
	a.linesPending = false
	a.latchLines()
}

// Read returns a register. Reading the status clears the interrupt.
func (a *MOS6551) Read(reg uint8) uint8 {
	switch reg & 0x03 {
	case 0:
		a.status &^= mos6551StatusRDRF | mos6551StatusOVRN
		return a.rdr
	case 1:
		value := a.status
		a.status &^= mos6551StatusIRQ
		if a.linesPending {
			// The changes that came after the one reported interrupt now
			a.linesPending = false
			a.updateLines()
		}
		return value
	case 2:
		return a.command
	default:
		return a.control
	}
}

// Write sets a register. Writing the status is a program reset.
func (a *MOS6551) Write(reg uint8, value uint8) {
	switch reg & 0x03 {
	case 0:
		a.tdr = value
		a.status &^= mos6551StatusTDRE
		a.loadShift()
	case 1:
		a.status &^= mos6551StatusOVRN
		a.command &= 0xe0
	case 2:
		a.command = value
		a.loadShift()
	default:
		a.control = value
	}
}

// Tick advances the time, sending and receiving the bytes that are due
func (a *MOS6551) Tick(elapsedCycles uint64) {
	a.Advance(elapsedCycles)
	a.receive()
}

/*
Advance moves the time and the transmitter, without receiving. To be called
before an access to the registers: a byte received there would be seen by
the access before the CPU takes its interrupt, it has to come between the
instructions, with Tick.
*/
func (a *MOS6551) Advance(elapsedCycles uint64) {
	a.cycles += elapsedCycles

	if a.shifting && a.cycles >= a.shiftDone {
		a.shifting = false
		a.line.Send(a.shift & a.dataMask())
		a.loadShift()
	}
}

// receive takes a byte from the line when the receiver is free
func (a *MOS6551) receive() {
	if a.command&mos6551CommandDTR != 0 && a.status&mos6551StatusRDRF == 0 && a.cycles >= a.rxNext {
		if value, ok := a.line.Receive(); ok {
			a.rdr = value & a.dataMask()
			a.status |= mos6551StatusRDRF
			if a.command&mos6551CommandIRD == 0 {
				a.status |= mos6551StatusIRQ
			}
			if a.command&mos6551CommandREM != 0 {
				a.line.Send(a.rdr)
			}
			a.rxNext = a.cycles + a.byteCycles()
		} else {
			// Look again in a bit time
			a.rxNext = a.cycles + a.byteCycles()/10
		}
	}
}

// InterruptAsserted returns whether the IRQ output is active
func (a *MOS6551) InterruptAsserted() bool {
	return a.status&mos6551StatusIRQ != 0
}

// SetDCD sets the level of the Data Carrier Detect input, true is high
func (a *MOS6551) SetDCD(high bool) {
	if a.dcd != high {
		a.dcd = high
		a.updateLines()
	}
}

// SetDSR sets the level of the Data Set Ready input, true is high
func (a *MOS6551) SetDSR(high bool) {
	if a.dsr != high {
		a.dsr = high
		a.updateLines()
	}
}

// IsDTR returns whether the Data Terminal Ready output is active, the
// receiver and the interrupts are enabled
func (a *MOS6551) IsDTR() bool {
	return a.command&mos6551CommandDTR != 0
}

// updateLines takes a change of the inputs to the status, unless the last
// change has not been read yet
func (a *MOS6551) updateLines() {
	if a.linesPending {
		return
	}
	previous := a.status & (mos6551StatusDCD | mos6551StatusDSR)
	a.latchLines()
	if a.status&(mos6551StatusDCD|mos6551StatusDSR) != previous && a.IsDTR() {
		a.status |= mos6551StatusIRQ
		a.linesPending = true
	}
}

func (a *MOS6551) latchLines() {
	a.status &^= mos6551StatusDCD | mos6551StatusDSR
	if a.dcd {
		a.status |= mos6551StatusDCD
	}
	if a.dsr {
		a.status |= mos6551StatusDSR
	}
}

// loadShift moves the transmit data register to the shift register when the
// transmitter is on and free
func (a *MOS6551) loadShift() {
	tic := a.command & mos6551CommandTIC
	if a.shifting || a.status&mos6551StatusTDRE != 0 || tic == mos6551TransmitterOff {
		return
	}
	a.shift = a.tdr
	a.shifting = true
	a.shiftDone = a.cycles + a.byteCycles()
	a.status |= mos6551StatusTDRE
	if tic == mos6551TransmitterIRQ && a.IsDTR() {
		a.status |= mos6551StatusIRQ
	}
}

// dataMask keeps the bits of the word length
func (a *MOS6551) dataMask() uint8 {
	return 0xff >> ((a.control >> 5) & 0x03)
}

// byteCycles is the time of a byte on the line: the start bit, the data
// bits, the parity bit and the stop bits
func (a *MOS6551) byteCycles() uint64 {
	bits := 1 + 8 - int((a.control>>5)&0x03)
	if a.command&0x20 != 0 {
		bits++ // Parity
	}
	if a.control&mos6551ControlStopBits != 0 {
		bits += 2
	} else {
		bits++
	}
	baud := mos6551BaudRates[a.control&mos6551ControlRate]
	return uint64(float64(bits) * a.cyclesBaud / baud)
}
