package component

/*
IWM, Integrated Woz Machine, the disk controller of the Liron card, the IIc
and the IIGS.

See:

	"Integrated Woz Machine (IWM) Specification", Apple, rev 19, 1982.
	http://www.brutaldeluxe.fr/documentation/iwm/apple2_IWM_Spec_Rev19_1982.pdf

The state register has eight bits, set and cleared by the accesses to the
sixteen addresses: A3 to A1 select the bit and A0 is its value.

	0-3: PHASE0 to PHASE3  4: LMotor-On  5: Drive-Sel  6: L6  7: L7

L7, L6 and Motor-On select the register of the access, after the state bit
of that same access is set:

	L7 L6 Motor-On
	0  0  0         read all ones
	0  0  1         read the data register
	0  1  x         read the status register
	1  0  x         read the write handshake register
	1  1  0         write the mode register, with A0 = 1
	1  1  1         write the data register, with A0 = 1

The registers are read on the accesses with A0 = 0.

The serial data is emulated by whole bytes, not by bits, in asynchronous
mode as the SmartPort uses it: the drive gets each byte written and gives
the bytes to read. The time of the bytes is exact: a bit cell is 4 CPU
cycles, 2 in fast mode, as the IWM clock of the Apple II is seven times the
CPU clock, and a byte is 8 bit cells.

In the write state, the byte in the buffer goes to the shift register every
byte time, starting 7 CLK periods after the write state begins. Only the
last byte written to the buffer before that is sent. If the buffer is empty
then, there is an underrun: the write state bit of the handshake register is
cleared until L7 is cleared.

In the read state, the drive is asked for a byte every byte time. The data
register keeps the last one, a byte not read is lost. In latch mode the
byte is cleared about 2 microseconds after it is read.

Not emulated: the synchronous mode timing, the bit level data for disk
mechanisms, the 8 MHz clock and the test mode.
*/
type IWM struct {
	drive  IWMDrive
	state  uint8
	mode   uint8
	cycles uint64

	motorOffCycle uint64 // Cycle LMotor-On was cleared, for the timer
	timerRunning  bool

	writing    bool
	buffer     uint8
	bufferFull bool
	underrun   bool
	nextLoad   uint64 // Cycle the buffer goes to the shift register

	reading   bool
	data      uint8
	nextRead  uint64 // Cycle the drive gives the next byte
	clearData uint64 // Cycle the latched byte is cleared, 0 for none
}

// IWMDrive is what is connected to the drive port of the IWM
type IWMDrive interface {
	// SetLines sets the outputs: the four phases, and the two enables
	SetLines(phases uint8, enable1 bool, enable2 bool)
	// Write gets a byte written by the IWM
	Write(value uint8)
	// Read returns a byte for the IWM to read, if the drive is sending
	Read() (value uint8, ok bool)
	// Sense returns the level of the SENSE input, true is high
	Sense() bool
}

const (
	iwmStateMotorOn  uint8 = 1 << 4
	iwmStateDriveSel uint8 = 1 << 5
	iwmStateL6       uint8 = 1 << 6
	iwmStateL7       uint8 = 1 << 7

	iwmModeLatch        uint8 = 1 << 0
	iwmModeTimerDisable uint8 = 1 << 2
	iwmModeFast         uint8 = 1 << 3

	iwmStatusEnable uint8 = 1 << 5
	iwmStatusSense  uint8 = 1 << 7

	iwmHandshakeWriting uint8 = 1 << 6
	iwmHandshakeReady   uint8 = 1 << 7

	// The enable stays active 2^23 FCLK periods after LMotor-On is cleared,
	// with FCLK seven times the CPU clock
	iwmTimerCycles = (1 << 23) / 7

	// 7 CLK periods, the CLK is FCLK/2 in slow mode and FCLK in fast mode
	iwmWriteStartSlow = 2
	iwmWriteStartFast = 1

	// The latched byte is cleared 14 FCLK periods after it is read
	iwmLatchClearCycles = 2
)

// NewIWM returns an IWM connected to a drive port
func NewIWM(drive IWMDrive) *IWM {
	var i IWM
	i.drive = drive
	i.Reset()
	return &i
}

// Reset clears the state and the mode, as the /RESET input
func (i *IWM) Reset() {
	i.state = 0
	i.mode = 0
	i.timerRunning = false
	i.writing = false
	i.underrun = false
	i.bufferFull = false
	i.reading = false
	i.data = 0
	i.clearData = 0
	i.updateLines()
}

/*
Access is a read or a write of one of the sixteen addresses, at a CPU cycle.
It sets the state bit of the address and accesses the register selected. It
returns the value read, all ones when no register is read.
*/
func (i *IWM) Access(cycle uint64, address uint8, write bool, value uint8) uint8 {
	i.catchUp(cycle)
	i.setState(address & 0x0f)

	a0 := address&1 != 0
	l6 := i.state&iwmStateL6 != 0
	l7 := i.state&iwmStateL7 != 0

	if write {
		if a0 && l6 && l7 {
			if i.isMotorOn() {
				i.buffer = value
				i.bufferFull = true
			} else {
				i.mode = value
			}
		}
		return 0xff
	}

	if a0 {
		return 0xff
	}
	switch {
	case !l7 && !l6:
		if !i.isMotorOn() {
			return 0xff
		}
		return i.readData()
	case !l7 && l6:
		return i.status()
	case l7 && !l6:
		return i.handshake()
	default:
		return 0xff
	}
}

// Mode returns the mode register
func (i *IWM) Mode() uint8 {
	return i.mode
}

// State returns the state register: the phases, LMotor-On, Drive-Sel, L6
// and L7
func (i *IWM) State() uint8 {
	return i.state
}

func (i *IWM) byteCycles() uint64 {
	if i.mode&iwmModeFast != 0 {
		return 16
	}
	return 32
}

// isMotorOn is the internal Motor-On: LMotor-On, or the timer after it
func (i *IWM) isMotorOn() bool {
	return i.state&iwmStateMotorOn != 0 || i.timerRunning
}

// catchUp runs the serial data and the timer up to the cycle
func (i *IWM) catchUp(cycle uint64) {
	if cycle < i.cycles {
		cycle = i.cycles
	}
	i.cycles = cycle

	if i.timerRunning && cycle >= i.motorOffCycle+iwmTimerCycles {
		i.timerRunning = false
		i.updateLines()
		i.updateDataState()
	}

	for i.writing && !i.underrun && cycle >= i.nextLoad {
		if i.bufferFull {
			i.drive.Write(i.buffer)
			i.bufferFull = false
		} else {
			i.underrun = true
		}
		i.nextLoad += i.byteCycles()
	}

	if i.clearData != 0 && cycle >= i.clearData {
		i.data = 0
		i.clearData = 0
	}

	for i.reading && cycle >= i.nextRead {
		if value, ok := i.drive.Read(); ok {
			i.data = value
			i.clearData = 0
		}
		i.nextRead += i.byteCycles()
	}
}

// setState sets the state bit selected by the address with its A0
func (i *IWM) setState(address uint8) {
	bit := uint8(1) << (address >> 1)
	previous := i.state
	if address&1 != 0 {
		i.state |= bit
	} else {
		i.state &^= bit
	}
	if i.state == previous {
		return
	}

	if previous&iwmStateMotorOn != 0 && i.state&iwmStateMotorOn == 0 {
		i.motorOffCycle = i.cycles
		i.timerRunning = i.mode&iwmModeTimerDisable == 0
	}
	if i.state&iwmStateMotorOn != 0 {
		i.timerRunning = false
	}
	i.updateLines()
	i.updateDataState()
}

// updateDataState starts and ends the write and read states
func (i *IWM) updateDataState() {
	l7 := i.state&iwmStateL7 != 0
	l6 := i.state&iwmStateL6 != 0
	motorOn := i.isMotorOn()

	writing := l7 && motorOn
	if writing && !i.writing {
		start := uint64(iwmWriteStartSlow)
		if i.mode&iwmModeFast != 0 {
			start = iwmWriteStartFast
		}
		i.nextLoad = i.cycles + start
		i.underrun = false
		i.bufferFull = false
	}
	if !l7 {
		// Clearing L7 resets the underrun
		i.underrun = false
	}
	i.writing = writing

	reading := !l7 && !l6 && motorOn
	if reading && !i.reading {
		i.nextRead = i.cycles + i.byteCycles()
	}
	i.reading = reading
}

func (i *IWM) updateLines() {
	motorOn := i.isMotorOn()
	drive2 := i.state&iwmStateDriveSel != 0
	i.drive.SetLines(i.state&0x0f, motorOn && !drive2, motorOn && drive2)
}

func (i *IWM) readData() uint8 {
	value := i.data
	if value&0x80 != 0 && i.mode&iwmModeLatch != 0 && i.clearData == 0 {
		i.clearData = i.cycles + iwmLatchClearCycles
	}
	return value
}

func (i *IWM) status() uint8 {
	value := i.mode & 0x1f
	if i.isMotorOn() {
		value |= iwmStatusEnable
	}
	if i.drive.Sense() {
		value |= iwmStatusSense
	}
	return value
}

func (i *IWM) handshake() uint8 {
	value := uint8(0x3f)
	if i.writing && !i.underrun {
		value |= iwmHandshakeWriting
	}
	if !i.bufferFull {
		value |= iwmHandshakeReady
	}
	return value
}
