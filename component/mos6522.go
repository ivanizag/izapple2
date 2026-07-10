package component

/*
MOS 6522 Versatile Interface Adapter (VIA)
See:

	http://archive.6502.org/datasheets/mos_6522_preliminary_nov_1977.pdf
	http://archive.6502.org/datasheets/rockwell_r6522_via.pdf

Used by the Mockingboard card to drive the AY-3-8913 sound generators
and to raise interrupts with the timers.

Implemented: ports A and B, timers T1 and T2, the IFR/IER interrupt logic.
Not implemented: the shift register, the CA/CB control lines, handshaking
and the T2 pulse counting mode.

Registers:

	0: ORB/IRB   1: ORA/IRA   2: DDRB   3: DDRA
	4: T1C-L     5: T1C-H     6: T1L-L  7: T1L-H
	8: T2C-L     9: T2C-H    10: SR    11: ACR
	12: PCR     13: IFR      14: IER   15: ORA no handshake

The timers advance with Tick(elapsedCycles). Catching up with several
cycles at a time is supported, including multiple T1 underflows in free
running mode.
*/
type MOS6522 struct {
	ora, orb   uint8 // Output registers
	ira, irb   uint8 // Input registers, the values on the input pins
	ddra, ddrb uint8 // Data direction registers, 1 is output

	t1counter int64 // Signed to detect underflows when catching up
	t1latch   uint16
	t1fired   bool
	t2counter int64
	t2latchL  uint8
	t2fired   bool

	sr, acr, pcr uint8
	ifr, ier     uint8
}

const (
	mos6522IntT1 uint8 = 1 << 6
	mos6522IntT2 uint8 = 1 << 5

	mos6522AcrT1FreeRunning uint8 = 1 << 6
)

// Read returns the value of a register
func (v *MOS6522) Read(reg uint8) uint8 {
	switch reg & 0x0f {
	case 0:
		return (v.irb &^ v.ddrb) | (v.orb & v.ddrb)
	case 1, 15:
		return (v.ira &^ v.ddra) | (v.ora & v.ddra)
	case 2:
		return v.ddrb
	case 3:
		return v.ddra
	case 4:
		v.ifr &^= mos6522IntT1
		return uint8(uint16(v.t1counter))
	case 5:
		return uint8(uint16(v.t1counter) >> 8)
	case 6:
		return uint8(v.t1latch)
	case 7:
		return uint8(v.t1latch >> 8)
	case 8:
		v.ifr &^= mos6522IntT2
		return uint8(uint16(v.t2counter))
	case 9:
		return uint8(uint16(v.t2counter) >> 8)
	case 10:
		return v.sr
	case 11:
		return v.acr
	case 12:
		return v.pcr
	case 13:
		ifr := v.ifr & 0x7f
		if ifr&v.ier != 0 {
			ifr |= 0x80
		}
		return ifr
	case 14:
		return v.ier | 0x80
	}
	return 0
}

// Write sets the value of a register
func (v *MOS6522) Write(reg uint8, value uint8) {
	switch reg & 0x0f {
	case 0:
		v.orb = value
	case 1, 15:
		v.ora = value
	case 2:
		v.ddrb = value
	case 3:
		v.ddra = value
	case 4, 6:
		v.t1latch = (v.t1latch & 0xff00) | uint16(value)
	case 5:
		// Load the counter from the latch and start counting
		v.t1latch = (v.t1latch & 0x00ff) | uint16(value)<<8
		v.t1counter = int64(v.t1latch)
		v.t1fired = false
		v.ifr &^= mos6522IntT1
	case 7:
		v.t1latch = (v.t1latch & 0x00ff) | uint16(value)<<8
		v.ifr &^= mos6522IntT1
	case 8:
		v.t2latchL = value
	case 9:
		v.t2counter = int64(value)<<8 | int64(v.t2latchL)
		v.t2fired = false
		v.ifr &^= mos6522IntT2
	case 10:
		v.sr = value
	case 11:
		v.acr = value
	case 12:
		v.pcr = value
	case 13:
		// Writing ones clears the flags
		v.ifr &^= value & 0x7f
	case 14:
		if value&0x80 != 0 {
			v.ier |= value & 0x7f
		} else {
			v.ier &^= value & 0x7f
		}
	}
}

// Tick advances the timers by the elapsed CPU cycles
func (v *MOS6522) Tick(elapsedCycles uint64) {
	v.t1counter -= int64(elapsedCycles)
	if v.t1counter < 0 {
		if v.acr&mos6522AcrT1FreeRunning != 0 {
			// Reload from the latch and interrupt on each underflow
			period := int64(v.t1latch) + 2
			v.t1counter += period * (1 + (-v.t1counter-1)/period)
			v.ifr |= mos6522IntT1
		} else {
			// One shot: interrupt once, the counter rolls over
			if !v.t1fired {
				v.ifr |= mos6522IntT1
				v.t1fired = true
			}
			v.t1counter = int64(uint16(v.t1counter))
		}
	}

	v.t2counter -= int64(elapsedCycles)
	if v.t2counter < 0 {
		// One shot: interrupt once, the counter rolls over
		if !v.t2fired {
			v.ifr |= mos6522IntT2
			v.t2fired = true
		}
		v.t2counter = int64(uint16(v.t2counter))
	}
}

// InterruptAsserted returns the state of the IRQ output line
func (v *MOS6522) InterruptAsserted() bool {
	return v.ifr&v.ier&0x7f != 0
}

// Reset clears the registers as the RES pin. The timers and the shift
// register are not affected
func (v *MOS6522) Reset() {
	v.ora, v.orb = 0, 0
	v.ddra, v.ddrb = 0, 0
	v.acr, v.pcr = 0, 0
	v.ifr, v.ier = 0, 0
}

// GetPortA returns the values on the port A pins
func (v *MOS6522) GetPortA() uint8 {
	return (v.ora & v.ddra) | (v.ira &^ v.ddra)
}

// GetPortB returns the values on the port B pins
func (v *MOS6522) GetPortB() uint8 {
	return (v.orb & v.ddrb) | (v.irb &^ v.ddrb)
}

// SetInputA sets the values on the port A input pins
func (v *MOS6522) SetInputA(value uint8) {
	v.ira = value
}

// SetInputB sets the values on the port B input pins
func (v *MOS6522) SetInputB(value uint8) {
	v.irb = value
}
