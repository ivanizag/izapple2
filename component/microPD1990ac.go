package component

import (
	"time"
)

/*
microPD1990ac Serial I/O Calendar Clock IC
See:

	https://www.semiee.com/file/backup/NEC-D1990.pdf

Used by the ThunderClock+ real time clock card.

The 40 bit register has 5 bytes (10 nibbles):

	byte 4:
		month, binary from 1 to 12
		day of week, BCD 0 to 6
	byte 3: day of month, BCD 1 to 31
	byte 2: hour, BCD 0 to 23
	byte 1: minute, BCD 0 to 59
	byte 0: seconds, BCD 0 to 59
*/
type MicroPD1990ac struct {
	clock    bool   // CLK state
	strobe   bool   // STB state
	command  uint8  // C0, C1, C2 command. From 0 to 7
	register uint64 // 40 bit shift register
}

const (
	mpd1990commandRegHold  = 0
	mpd1990commandRegShift = 1
	mpd1990commandTimeSet  = 2
	mpd1990commandTimeRead = 3
)

func (m *MicroPD1990ac) In(clock bool, strobe bool, command uint8, dataIn bool) {
	// Detect signal raise
	clockRaise := clock && !m.clock
	strobeRaise := strobe && !m.strobe

	// Update signal status
	m.clock = clock
	m.strobe = strobe

	// On strobe raise, update command and execute if needed
	if strobeRaise {
		m.command = command

		switch m.command {
		case mpd1990commandRegShift:
			// Nothing to do
		case mpd1990commandTimeRead:
			m.loadTime()
		default:
			// Ignore unknown commands (like set time)
			// panic(fmt.Sprintf("PD1990ac command %v not implemented.", m.command))
		}
	}

	// On clock raise, with shift enable, shift the register
	if clockRaise && m.command == mpd1990commandRegShift {
		// Rotate right the 40 bits of the shift register
		lsb := m.register & 1
		m.register >>= 1
		m.register += lsb << 39
	}
}

func (m *MicroPD1990ac) Out() bool {
	/*
		if m.command == mpd1990commandRegHold {
			panic("Output on RegHold should be a 1Hz signal. Not implemented.")
		}

		if m.command == mpd1990commandTimeRead {
			panic("Output on RegHold should be a 512Hz signal with LSB. Not implemented.")
		}
	*/

	// Return the LSB of the register shift
	return (m.register & 1) == 1
}

func (m *MicroPD1990ac) loadTime() {
	now := time.Now()

	var register uint64

	register = uint64(now.Month())
	register <<= 4
	register += uint64(now.Weekday())

	day := uint64(now.Day())
	register <<= 4
	register += day / 10
	register <<= 4
	register += day % 10

	hour := uint64(now.Hour())
	register <<= 4
	register += hour / 10
	register <<= 4
	register += hour % 10

	minute := uint64(now.Minute())
	register <<= 4
	register += minute / 10
	register <<= 4
	register += minute % 10

	second := uint64(now.Second())
	register <<= 4
	register += second / 10
	register <<= 4
	register += second % 10

	m.register = register
}
