package component

import (
	"bytes"
	"testing"
)

// testIWMDrive records what the IWM does and gives the bytes queued
type testIWMDrive struct {
	phases           uint8
	enable1, enable2 bool
	written          []uint8
	toRead           []uint8
	sense            bool
}

func (d *testIWMDrive) SetLines(phases uint8, enable1 bool, enable2 bool) {
	d.phases, d.enable1, d.enable2 = phases, enable1, enable2
}

func (d *testIWMDrive) Write(value uint8) {
	d.written = append(d.written, value)
}

func (d *testIWMDrive) Read() (uint8, bool) {
	if len(d.toRead) == 0 {
		return 0, false
	}
	value := d.toRead[0]
	d.toRead = d.toRead[1:]
	return value, true
}

func (d *testIWMDrive) Sense() bool {
	return d.sense
}

// The sixteen addresses, as named on the Liron firmware
const (
	iwmTestPh0Off    = 0x0
	iwmTestPh0On     = 0x1
	iwmTestPh1On     = 0x3
	iwmTestMotorOff  = 0x8
	iwmTestMotorOn   = 0x9
	iwmTestDrive2    = 0xb
	iwmTestQ6L       = 0xc
	iwmTestQ6H       = 0xd
	iwmTestQ7L       = 0xe
	iwmTestQ7H       = 0xf
	iwmTestModeLiron = 0x07 // Latch, asynchronous, timer disabled, slow
)

// iwmTester runs accesses to the IWM with a clock
type iwmTester struct {
	iwm   *IWM
	drive *testIWMDrive
	cycle uint64
}

func newIWMTester() *iwmTester {
	drive := &testIWMDrive{sense: true}
	return &iwmTester{iwm: NewIWM(drive), drive: drive, cycle: 100}
}

func (it *iwmTester) read(address uint8) uint8 {
	return it.iwm.Access(it.cycle, address, false, 0)
}

func (it *iwmTester) write(address uint8, value uint8) {
	it.iwm.Access(it.cycle, address, true, value)
}

// setMode writes the mode register as the Liron does, with the motor off
func (it *iwmTester) setMode(mode uint8) {
	it.read(iwmTestMotorOff)
	it.read(iwmTestQ6H)
	it.write(iwmTestQ7H, mode)
	it.read(iwmTestQ7L)
}

func TestIWMAllOnesWithTheMotorOff(t *testing.T) {
	it := newIWMTester()
	if value := it.read(iwmTestQ6L); value != 0xff {
		t.Errorf("expected all ones with the motor off, got $%02x", value)
	}
}

func TestIWMModeAndStatus(t *testing.T) {
	it := newIWMTester()
	it.setMode(iwmTestModeLiron)
	if mode := it.iwm.Mode(); mode != iwmTestModeLiron {
		t.Fatalf("expected the mode written, got $%02x", mode)
	}

	// L6 set and L7 clear read the status: the mode bits and SENSE
	if status := it.read(iwmTestQ7L); status != iwmTestModeLiron|iwmStatusSense {
		t.Errorf("expected the status $%02x, got $%02x", iwmTestModeLiron|iwmStatusSense, status)
	}
	it.drive.sense = false
	it.read(iwmTestMotorOn)
	if status := it.read(iwmTestQ7L); status != iwmTestModeLiron|iwmStatusEnable {
		t.Errorf("expected the status with the enable and SENSE low, got $%02x", status)
	}

	// With the motor on the mode register is not written
	it.read(iwmTestQ6H)
	it.write(iwmTestQ7H, 0x1f)
	if mode := it.iwm.Mode(); mode != iwmTestModeLiron {
		t.Errorf("expected the mode not written with the motor on, got $%02x", mode)
	}
}

func TestIWMMotorTimer(t *testing.T) {
	it := newIWMTester()

	// The timer is enabled after reset
	it.read(iwmTestMotorOn)
	it.read(iwmTestMotorOff)
	it.cycle += iwmTimerCycles - 10
	it.read(iwmTestQ6H)
	if status := it.read(iwmTestQ7L); status&iwmStatusEnable == 0 {
		t.Error("expected the enable active while the timer runs")
	}
	it.write(iwmTestQ7H, iwmTestModeLiron)
	if it.iwm.Mode() != 0 {
		t.Error("expected the mode not written while the timer runs")
	}

	it.cycle += 20
	if status := it.read(iwmTestQ7L); status&iwmStatusEnable != 0 {
		t.Error("expected the enable inactive after the timer")
	}
	it.setMode(iwmTestModeLiron)

	// Disabled, the enable goes at once
	it.read(iwmTestMotorOn)
	it.read(iwmTestMotorOff)
	it.read(iwmTestQ6H)
	if status := it.read(iwmTestQ7L); status&iwmStatusEnable != 0 {
		t.Error("expected the enable inactive with the timer disabled")
	}
}

// startWrite enters the write state as the Liron does, writing the first
// byte. It returns the cycle the write state started.
func (it *iwmTester) startWrite(first uint8) uint64 {
	it.setMode(iwmTestModeLiron)
	it.read(iwmTestDrive2)
	it.read(iwmTestMotorOn)
	it.read(iwmTestQ7L) // Status, L6 is set
	it.write(iwmTestQ7H, first)
	return it.cycle
}

func TestIWMWritesOnTheByteTimes(t *testing.T) {
	it := newIWMTester()
	start := it.startWrite(0xff)

	// The first byte goes to the shift register 2 cycles later
	it.cycle = start + 1
	if handshake := it.read(iwmTestQ6L); handshake&iwmHandshakeReady != 0 {
		t.Error("expected the buffer full before the first load")
	}
	it.cycle = start + 2
	if handshake := it.read(iwmTestQ6L); handshake&iwmHandshakeReady == 0 || handshake&iwmHandshakeWriting == 0 {
		t.Errorf("expected the buffer ready and the write state, got $%02x", handshake)
	}
	if !bytes.Equal(it.drive.written, []uint8{0xff}) {
		t.Errorf("expected $ff written, got % x", it.drive.written)
	}

	// The next load is 32 cycles later, the last byte written is the one sent
	it.write(iwmTestQ6H, 0x3f)
	it.cycle = start + 20
	it.write(iwmTestQ6H, 0xcf)
	it.cycle = start + 33
	it.read(iwmTestQ6L)
	if !bytes.Equal(it.drive.written, []uint8{0xff}) {
		t.Errorf("expected nothing more before the load, got % x", it.drive.written)
	}
	it.cycle = start + 34
	it.read(iwmTestQ6L)
	if !bytes.Equal(it.drive.written, []uint8{0xff, 0xcf}) {
		t.Errorf("expected $cf written, got % x", it.drive.written)
	}

	// Nothing written for the next load is an underrun
	it.cycle = start + 66
	if handshake := it.read(iwmTestQ6L); handshake&iwmHandshakeWriting != 0 {
		t.Errorf("expected the underrun, got $%02x", handshake)
	}
	it.write(iwmTestQ6H, 0xc3)
	it.cycle = start + 200
	it.read(iwmTestQ6L)
	if len(it.drive.written) != 2 {
		t.Errorf("expected nothing written after the underrun, got % x", it.drive.written)
	}

	// Clearing L7 ends the write state and the underrun
	it.read(iwmTestQ7L)
	it.cycle += 10
	start = it.cycle
	it.read(iwmTestQ6H)
	it.write(iwmTestQ7H, 0xaa)
	it.cycle = start + 2
	if handshake := it.read(iwmTestQ6L); handshake&iwmHandshakeWriting == 0 {
		t.Errorf("expected a new write state, got $%02x", handshake)
	}
	if !bytes.Equal(it.drive.written, []uint8{0xff, 0xcf, 0xaa}) {
		t.Errorf("expected $aa written, got % x", it.drive.written)
	}
}

func TestIWMFastModeWritesEvery16Cycles(t *testing.T) {
	it := newIWMTester()
	it.setMode(iwmTestModeLiron | iwmModeFast)
	it.read(iwmTestMotorOn)
	it.read(iwmTestQ7L)
	start := it.cycle
	it.write(iwmTestQ7H, 0x01)
	it.cycle = start + 1
	it.write(iwmTestQ6H, 0x02)
	it.cycle = start + 17
	it.read(iwmTestQ6L)
	if !bytes.Equal(it.drive.written, []uint8{0x01, 0x02}) {
		t.Errorf("expected two bytes 16 cycles apart, got % x", it.drive.written)
	}
}

func TestIWMReadsTheBytesOfTheDrive(t *testing.T) {
	it := newIWMTester()
	it.setMode(iwmTestModeLiron)
	it.drive.toRead = []uint8{0xc3, 0x81, 0x82, 0x83}
	it.read(iwmTestMotorOn)
	it.read(iwmTestQ6H)
	it.read(iwmTestQ7L)
	start := it.cycle
	it.read(iwmTestQ6L) // The read state

	it.cycle = start + 31
	if value := it.read(iwmTestQ6L); value&0x80 != 0 {
		t.Errorf("expected no byte before the byte time, got $%02x", value)
	}
	it.cycle = start + 32
	if value := it.read(iwmTestQ6L); value != 0xc3 {
		t.Errorf("expected $c3, got $%02x", value)
	}

	// Latched until 2 cycles after it is read
	it.cycle = start + 33
	if value := it.read(iwmTestQ6L); value != 0xc3 {
		t.Errorf("expected $c3 still latched, got $%02x", value)
	}
	it.cycle = start + 34
	if value := it.read(iwmTestQ6L); value != 0 {
		t.Errorf("expected the latch cleared, got $%02x", value)
	}

	// A byte not read is lost
	it.cycle = start + 96
	if value := it.read(iwmTestQ6L); value != 0x82 {
		t.Errorf("expected $82, the $81 lost, got $%02x", value)
	}
}

func TestIWMLines(t *testing.T) {
	it := newIWMTester()
	it.setMode(iwmTestModeLiron)
	it.read(iwmTestPh0On)
	it.read(iwmTestPh1On)
	if it.drive.phases != 0x03 || it.drive.enable1 || it.drive.enable2 {
		t.Errorf("expected phases 0 and 1 and no enable, got %x %v %v", it.drive.phases, it.drive.enable1, it.drive.enable2)
	}
	it.read(iwmTestDrive2)
	it.read(iwmTestMotorOn)
	if it.drive.enable1 || !it.drive.enable2 {
		t.Error("expected the enable of the drive 2")
	}
	it.read(iwmTestPh0Off)
	it.read(iwmTestMotorOff)
	if it.drive.phases != 0x02 || it.drive.enable2 {
		t.Errorf("expected phase 1 and no enable, got %x %v", it.drive.phases, it.drive.enable2)
	}
}
