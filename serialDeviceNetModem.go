package izapple2

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

/*
serialDeviceNetModem is a Hayes modem that dials to the internet instead of
to the phone network, as the WiFi modems for the retro computers do. It
starts in command mode, taking AT commands, and the dial command connects
to a TCP port:

	ATDT host:port   Dials with telnet, the port defaults to 23
	ATD host:port    Dials without telnet, ATDP too
	+++              With a second of silence before and after, back to
	                 command mode keeping the connection
	ATO              Back online
	ATH              Hangs up

The host and port can also be quoted, ATDT"host:port". Telnet is as the T
modifier of Zimodem: the options are refused but echo and suppress go ahead,
the 0xFF bytes are doubled and a carriage return is sent as CR NUL.

The other commands:

	A/       Repeats the last command, without AT or Return
	En       Echo of the commands, on with 1
	In       Information
	Qn       Quiet, no results with 1
	Vn       Results in words with 1, as numbers with 0
	Z, &F    Hangs up and goes back to the defaults
	Sn=v     Sets a register, Sn? shows it
	Ln, Mn, Xn, Bn, Nn, Wn, &Cn, &Dn, &Kn, &W and the rest of &: accepted
	         and ignored

The registers used are S2, the escape character, S3, S4 and S5, the
carriage return, line feed and backspace, S7, the seconds to wait for the
connection and S12, the guard time of the escape in 1/50 of a second.

There is no answering of calls, ATA is an error. The modem is always ready,
as with &C0, the carrier is not on DCD: the Apple IIc firmware sends only
with it low, and it has to send the AT commands.

See:

	https://en.wikipedia.org/wiki/Hayes_command_set
	https://github.com/bozimmerman/Zimodem/blob/master/README
	https://www.rfc-editor.org/rfc/rfc854
*/
type serialDeviceNetModem struct {
	mutex    sync.Mutex
	toApple  chan uint8
	state    int
	line     []uint8
	lastLine string

	echo      bool
	verbose   bool
	quiet     bool
	registers [netModemRegisters]uint8

	conn        net.Conn
	telnet      bool
	dialAborted bool
	lastSend    time.Time // To time the guard of the escape
	escapeCount int
}

const (
	netModemCommand = iota
	netModemDialing
	netModemOnline

	netModemRegisters   = 13
	netModemDefaultPort = "23"
)

// Result codes, the number is the index
var netModemResults = []string{"OK", "CONNECT", "RING", "NO CARRIER", "ERROR"}

const (
	netModemOK = iota
	netModemConnect
	netModemRing
	netModemNoCarrier
	netModemError
)

var netModemDefaultRegisters = [netModemRegisters]uint8{
	0,   // S0: rings to answer
	0,   // S1: rings counted
	'+', // S2: escape character
	'\r',
	'\n',
	8,  // S5: backspace
	2,  // S6: seconds to wait for the dial tone
	50, // S7: seconds to wait for the carrier
	2,  // S8: seconds of the comma pause
	6,  // S9
	14, // S10
	95, // S11
	50, // S12: escape guard time, in 1/50 s
}

const (
	telnetIAC  = 255
	telnetDONT = 254
	telnetDO   = 253
	telnetWONT = 252
	telnetWILL = 251
	telnetSB   = 250
	telnetSE   = 240

	telnetOptionEcho = 1
	telnetOptionSGA  = 3
)

func newSerialDeviceNetModem() *serialDeviceNetModem {
	var m serialDeviceNetModem
	m.toApple = make(chan uint8, serialReceiveBuffer)
	m.setDefaults()
	return &m
}

func (m *serialDeviceNetModem) setDefaults() {
	m.echo = true
	m.verbose = true
	m.quiet = false
	m.registers = netModemDefaultRegisters
}

func (m *serialDeviceNetModem) isReady() bool {
	return true
}

func (m *serialDeviceNetModem) Receive() (uint8, bool) {
	select {
	case value := <-m.toApple:
		return value, true
	default:
		return 0, false
	}
}

func (m *serialDeviceNetModem) Send(value uint8) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	switch m.state {
	case netModemOnline:
		m.sendOnline(value)
	case netModemDialing:
		// Any character cancels the dial
		m.dialAborted = true
		m.state = netModemCommand
		m.result(netModemNoCarrier)
	default:
		m.sendCommand(value & 0x7f)
	}
}

// sendOnline passes the byte to the connection, looking for the escape
func (m *serialDeviceNetModem) sendOnline(value uint8) {
	now := time.Now()
	guard := m.guardTime()
	if value&0x7f == m.registers[2] && (m.escapeCount > 0 || now.Sub(m.lastSend) >= guard) {
		m.escapeCount++
		if m.escapeCount == 3 {
			m.startEscapeTimer(now, guard)
		}
	} else {
		m.escapeCount = 0
	}
	m.lastSend = now

	if m.conn == nil {
		return
	}
	data := []uint8{value}
	if m.telnet {
		if value == telnetIAC {
			data = []uint8{telnetIAC, telnetIAC}
		} else if value&0x7f == '\r' {
			data = []uint8{value, 0}
		}
	}
	m.conn.Write(data)
}

// startEscapeTimer goes to command mode after the guard time, if nothing is
// sent after the third escape character
func (m *serialDeviceNetModem) startEscapeTimer(third time.Time, guard time.Duration) {
	time.AfterFunc(guard, func() {
		m.mutex.Lock()
		defer m.mutex.Unlock()
		if m.state == netModemOnline && m.escapeCount == 3 && m.lastSend.Equal(third) {
			m.escapeCount = 0
			m.state = netModemCommand
			m.result(netModemOK)
		}
	})
}

func (m *serialDeviceNetModem) guardTime() time.Duration {
	return time.Duration(m.registers[12]) * time.Second / 50
}

// sendCommand collects the line of a command
func (m *serialDeviceNetModem) sendCommand(value uint8) {
	if m.echo {
		m.reply([]uint8{value})
	}

	switch value {
	case m.registers[3]:
		line := strings.ToUpper(string(m.line))
		m.line = m.line[:0]
		if at := strings.Index(line, "AT"); at >= 0 {
			m.lastLine = line[at:]
			m.execute(line[at+2:])
		}
	case m.registers[5]:
		if len(m.line) > 0 {
			m.line = m.line[:len(m.line)-1]
		}
	case '/':
		if strings.HasSuffix(strings.ToUpper(string(m.line)), "A") {
			m.line = m.line[:0]
			if m.lastLine != "" {
				m.execute(m.lastLine[2:])
			}
			return
		}
		fallthrough
	default:
		if len(m.line) < 255 {
			m.line = append(m.line, value)
		}
	}
}

// execute runs the commands of a line, what is after AT
func (m *serialDeviceNetModem) execute(commands string) {
	i := 0
	number := func() int {
		start := i
		for i < len(commands) && commands[i] >= '0' && commands[i] <= '9' {
			i++
		}
		n, _ := strconv.Atoi(commands[start:i])
		return n
	}

	for i < len(commands) {
		command := commands[i]
		i++
		switch command {
		case ' ':
		case 'D':
			m.dial(commands[i:])
			return
		case 'O':
			number()
			if m.conn == nil {
				m.result(netModemNoCarrier)
				return
			}
			m.state = netModemOnline
			m.escapeCount = 0
			m.result(netModemConnect)
			return
		case 'H':
			if number() == 0 {
				m.hangUp()
			}
		case 'Z':
			number()
			m.hangUp()
			m.setDefaults()
		case 'E':
			m.echo = number() != 0
		case 'V':
			m.verbose = number() != 0
		case 'Q':
			m.quiet = number() != 0
		case 'I':
			number()
			m.replyLine("izapple2 netmodem, dials host:port")
		case 'S':
			register := number()
			if register >= netModemRegisters || i >= len(commands) {
				m.result(netModemError)
				return
			}
			if commands[i] == '?' {
				i++
				m.replyLine(fmt.Sprintf("%03d", m.registers[register]))
			} else if commands[i] == '=' {
				i++
				m.registers[register] = uint8(number())
			} else {
				m.result(netModemError)
				return
			}
		case '&':
			if i >= len(commands) {
				m.result(netModemError)
				return
			}
			if commands[i] == 'F' {
				m.setDefaults()
			}
			i++
			number()
		case 'B', 'C', 'L', 'M', 'N', 'P', 'T', 'W', 'X', 'Y':
			number()
		default:
			m.result(netModemError)
			return
		}
	}
	m.result(netModemOK)
}

// dial connects to host:port, telnet if the dial modifier is T
func (m *serialDeviceNetModem) dial(target string) {
	telnet := false
	target = strings.TrimSpace(target)
	if strings.HasPrefix(target, "T") {
		telnet = true
		target = target[1:]
	} else if strings.HasPrefix(target, "P") {
		target = target[1:]
	}
	target = strings.Trim(strings.TrimSpace(target), "\"")
	if target == "" {
		m.result(netModemError)
		return
	}
	address := target
	if !strings.Contains(address, ":") {
		address += ":" + netModemDefaultPort
	}

	m.hangUp()
	m.state = netModemDialing
	m.dialAborted = false
	timeout := time.Duration(m.registers[7]) * time.Second
	go func() {
		conn, err := net.DialTimeout("tcp", address, timeout)

		m.mutex.Lock()
		defer m.mutex.Unlock()
		if m.state != netModemDialing || m.dialAborted {
			if err == nil {
				conn.Close()
			}
			return
		}
		if err != nil {
			m.state = netModemCommand
			m.result(netModemNoCarrier)
			return
		}
		m.conn = conn
		m.telnet = telnet
		m.state = netModemOnline
		m.escapeCount = 0
		m.lastSend = time.Now()
		m.result(netModemConnect)
		go m.readConnection(conn, telnet)
	}()
}

// readConnection passes what comes from the connection to the Apple, until
// it is closed
func (m *serialDeviceNetModem) readConnection(conn net.Conn, telnet bool) {
	var parser telnetParser
	buffer := make([]uint8, 256)
	for {
		n, err := conn.Read(buffer)
		for _, value := range buffer[:n] {
			if telnet {
				data, answer := parser.parse(value)
				if answer != nil {
					conn.Write(answer)
				}
				if data < 0 {
					continue
				}
				value = uint8(data)
			}
			m.toApple <- value
		}
		if err != nil {
			break
		}
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.conn == conn {
		// The other side hung up
		conn.Close()
		m.conn = nil
		m.state = netModemCommand
		m.result(netModemNoCarrier)
	}
}

func (m *serialDeviceNetModem) hangUp() {
	if m.conn != nil {
		m.conn.Close()
		m.conn = nil
	}
	if m.state == netModemDialing {
		m.dialAborted = true
	}
	m.state = netModemCommand
}

// result sends a result code, in words or as a number
func (m *serialDeviceNetModem) result(code int) {
	if m.quiet {
		return
	}
	if m.verbose {
		m.replyLine(netModemResults[code])
	} else {
		m.reply([]uint8{'0' + uint8(code), m.registers[3]})
	}
}

// replyLine sends a line of text, between CR LF as the verbose results
func (m *serialDeviceNetModem) replyLine(text string) {
	crlf := []uint8{m.registers[3], m.registers[4]}
	m.reply(crlf)
	m.reply([]uint8(text))
	m.reply(crlf)
}

// reply sends the answers of the modem to the Apple. They are lost if the
// buffer is full, Send can't wait for the Apple, it is the Apple.
func (m *serialDeviceNetModem) reply(data []uint8) {
	for _, value := range data {
		select {
		case m.toApple <- value:
		default:
		}
	}
}

/*
telnetParser takes out of the data the telnet commands, answering the
negotiation of the options: no to everything but the server doing the echo
and suppressing the go ahead.
*/
type telnetParser struct {
	state   int
	command uint8
}

const (
	telnetData = iota
	telnetCommand
	telnetOption
	telnetSubnegotiation
	telnetSubnegotiationIAC
)

// parse returns the data byte, or -1 if none, and the answer to send
func (p *telnetParser) parse(value uint8) (int, []uint8) {
	switch p.state {
	case telnetCommand:
		switch value {
		case telnetIAC:
			p.state = telnetData
			return telnetIAC, nil
		case telnetDO, telnetDONT, telnetWILL, telnetWONT:
			p.command = value
			p.state = telnetOption
		case telnetSB:
			p.state = telnetSubnegotiation
		default:
			p.state = telnetData
		}
		return -1, nil
	case telnetOption:
		p.state = telnetData
		switch p.command {
		case telnetDO:
			return -1, []uint8{telnetIAC, telnetWONT, value}
		case telnetWILL:
			if value == telnetOptionEcho || value == telnetOptionSGA {
				return -1, []uint8{telnetIAC, telnetDO, value}
			}
			return -1, []uint8{telnetIAC, telnetDONT, value}
		}
		return -1, nil
	case telnetSubnegotiation:
		if value == telnetIAC {
			p.state = telnetSubnegotiationIAC
		}
		return -1, nil
	case telnetSubnegotiationIAC:
		if value == telnetSE {
			p.state = telnetData
		} else {
			p.state = telnetSubnegotiation
		}
		return -1, nil
	default:
		if value == telnetIAC {
			p.state = telnetCommand
			return -1, nil
		}
		return int(value), nil
	}
}
