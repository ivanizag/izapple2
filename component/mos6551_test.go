package component

import (
	"testing"
)

// testLine keeps what is sent and gives what is queued to be received
type testLine struct {
	sent     []uint8
	incoming []uint8
}

func (l *testLine) Send(value uint8) {
	l.sent = append(l.sent, value)
}

func (l *testLine) Receive() (uint8, bool) {
	if len(l.incoming) == 0 {
		return 0, false
	}
	value := l.incoming[0]
	l.incoming = l.incoming[1:]
	return value, true
}

const (
	testClock = 1_000_000
	// 9600 baud, 8 data bits, 1 stop bit, internal clock
	testControl = 0x1e
	// No parity, transmitter on with RTS low, receiver interrupts, DTR on
	testCommand = 0x09
	// 10 bits at 9600 baud
	testByteCycles = 10 * testClock / 9600
)

func newTestACIA() (*MOS6551, *testLine) {
	line := &testLine{}
	a := NewMOS6551(testClock, line)
	a.Write(3, testControl)
	a.Write(2, testCommand)
	return a, line
}

func TestMOS6551Reset(t *testing.T) {
	a, _ := newTestACIA()
	a.Reset()
	if status := a.Read(1); status != mos6551StatusTDRE|mos6551StatusDCD|mos6551StatusDSR {
		t.Errorf("expected the status after reset with TDRE and the inputs high, got $%02x", status)
	}
	if a.Read(2) != 0 || a.Read(3) != 0 {
		t.Error("expected the command and control registers cleared")
	}
}

func TestMOS6551ProgramReset(t *testing.T) {
	a, _ := newTestACIA()
	a.Write(2, 0xeb)
	a.Write(1, 0)
	if command := a.Read(2); command != 0xe0 {
		t.Errorf("expected the low bits of the command cleared, got $%02x", command)
	}
	if control := a.Read(3); control != testControl {
		t.Errorf("expected the control unchanged, got $%02x", control)
	}
}

func TestMOS6551Transmit(t *testing.T) {
	a, line := newTestACIA()
	a.Write(0, 'A')
	if a.Read(1)&mos6551StatusTDRE == 0 {
		t.Error("expected the TDR moved to the shift register and empty")
	}
	a.Write(0, 'B')
	if a.Read(1)&mos6551StatusTDRE != 0 {
		t.Error("expected the TDR full while the shift register is busy")
	}

	a.Tick(testByteCycles - 1)
	if len(line.sent) != 0 {
		t.Error("expected nothing sent before the time of a byte")
	}
	a.Tick(1)
	if string(line.sent) != "A" {
		t.Errorf("expected 'A' sent, got %q", line.sent)
	}
	if a.Read(1)&mos6551StatusTDRE == 0 {
		t.Error("expected the TDR empty, 'B' on the shift register")
	}
	a.Tick(testByteCycles)
	if string(line.sent) != "AB" {
		t.Errorf("expected 'AB' sent, got %q", line.sent)
	}
}

func TestMOS6551TransmitterOff(t *testing.T) {
	a, line := newTestACIA()
	a.Write(2, 0x01) // DTR on, transmitter off
	a.Write(0, 'A')
	a.Tick(2 * testByteCycles)
	if len(line.sent) != 0 || a.Read(1)&mos6551StatusTDRE != 0 {
		t.Error("expected the byte waiting for the transmitter")
	}
	a.Write(2, testCommand)
	a.Tick(testByteCycles)
	if string(line.sent) != "A" {
		t.Errorf("expected 'A' sent with the transmitter on, got %q", line.sent)
	}
}

func TestMOS6551SevenBits(t *testing.T) {
	a, line := newTestACIA()
	a.Write(3, testControl|0x20) // 7 data bits
	a.Write(0, 0xc1)
	a.Tick(testByteCycles)
	if len(line.sent) != 1 || line.sent[0] != 0x41 {
		t.Errorf("expected $41 sent, got %x", line.sent)
	}
}

func TestMOS6551Receive(t *testing.T) {
	a, line := newTestACIA()
	line.incoming = []uint8{'X', 'Y'}
	a.Tick(1)
	status := a.Read(1)
	if status&mos6551StatusRDRF == 0 || status&mos6551StatusIRQ == 0 {
		t.Errorf("expected a byte received with an interrupt, got status $%02x", status)
	}
	if a.InterruptAsserted() {
		t.Error("expected reading the status to clear the interrupt")
	}

	// The next byte waits for the data register to be read
	a.Tick(3 * testByteCycles)
	if value := a.Read(0); value != 'X' {
		t.Errorf("expected 'X', got %q", value)
	}
	if a.Read(1)&mos6551StatusRDRF != 0 {
		t.Error("expected reading the data to empty the register")
	}

	// And the time of a byte after the last one
	a.Tick(1)
	if a.Read(1)&mos6551StatusRDRF == 0 {
		t.Error("expected the next byte received")
	}
	if value := a.Read(0); value != 'Y' {
		t.Errorf("expected 'Y', got %q", value)
	}
}

func TestMOS6551ReceiverDisabled(t *testing.T) {
	a, line := newTestACIA()
	a.Write(2, 0x08) // DTR off
	line.incoming = []uint8{'X'}
	a.Tick(testByteCycles)
	if a.Read(1)&mos6551StatusRDRF != 0 {
		t.Error("expected nothing received with DTR off")
	}

	a.Write(2, testCommand|mos6551CommandIRD)
	a.Tick(testByteCycles)
	status := a.Read(1)
	if status&mos6551StatusRDRF == 0 || status&mos6551StatusIRQ != 0 {
		t.Errorf("expected the byte received without interrupt, got status $%02x", status)
	}
}

func TestMOS6551Echo(t *testing.T) {
	a, line := newTestACIA()
	a.Write(2, 0x11) // Echo, DTR on
	line.incoming = []uint8{'E'}
	a.Tick(1)
	if string(line.sent) != "E" {
		t.Errorf("expected the byte received echoed, got %q", line.sent)
	}
}

func TestMOS6551TransmitInterrupt(t *testing.T) {
	a, _ := newTestACIA()
	a.Write(2, 0x07) // Transmitter interrupts, receiver interrupts disabled, DTR on
	a.Write(0, 'A')
	if !a.InterruptAsserted() {
		t.Error("expected an interrupt with the TDR empty")
	}
}

func TestMOS6551LineChanges(t *testing.T) {
	a, _ := newTestACIA()
	a.Write(2, 0x0f) // DTR on, receiver interrupts disabled, as the IIc keyboard

	a.SetDSR(false)
	if !a.InterruptAsserted() {
		t.Error("expected an interrupt on the change of DSR")
	}
	// A second change waits for the status to be read
	a.SetDSR(true)
	status := a.Read(1)
	if status&mos6551StatusDSR != 0 || status&mos6551StatusIRQ == 0 {
		t.Errorf("expected the status of the first change, DSR low, got $%02x", status)
	}
	if !a.InterruptAsserted() {
		t.Error("expected another interrupt for the second change")
	}
	status = a.Read(1)
	if status&mos6551StatusDSR == 0 {
		t.Errorf("expected the status with DSR high, got $%02x", status)
	}
	if a.InterruptAsserted() {
		t.Error("expected no more interrupts")
	}

	// No interrupts with DTR off, the status follows the levels
	a.Write(2, 0x00)
	a.SetDCD(false)
	if a.InterruptAsserted() {
		t.Error("expected no interrupt with DTR off")
	}
	if a.Read(1)&mos6551StatusDCD != 0 {
		t.Error("expected DCD low on the status")
	}
}
