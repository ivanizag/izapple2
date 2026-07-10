package component

import (
	"testing"
)

func TestMOS6522Timer1OneShot(t *testing.T) {
	var v MOS6522
	v.Write(14, 0x80|0x40) // Enable the T1 interrupt
	v.Write(4, 100)        // T1 latch low
	v.Write(5, 0)          // T1 counter high, starts the timer

	v.Tick(100)
	if v.InterruptAsserted() {
		t.Error("The interrupt should not fire before the timer underflows")
	}

	v.Tick(2)
	if !v.InterruptAsserted() {
		t.Error("The interrupt should fire when the timer underflows")
	}

	// In one shot mode it must fire only once
	v.Write(13, 0x40) // Clear the T1 flag
	v.Tick(1000)
	if v.InterruptAsserted() {
		t.Error("The one shot timer should not fire again")
	}
}

func TestMOS6522Timer1FreeRunning(t *testing.T) {
	var v MOS6522
	v.Write(11, 0x40)      // ACR: T1 free running
	v.Write(14, 0x80|0x40) // Enable the T1 interrupt
	v.Write(4, 100)
	v.Write(5, 0)

	v.Tick(102)
	if !v.InterruptAsserted() {
		t.Error("The interrupt should fire on the first underflow")
	}

	v.Write(13, 0x40) // Clear the T1 flag
	v.Tick(102)       // A full period without reprogramming the timer
	if !v.InterruptAsserted() {
		t.Error("The interrupt should fire again in free running mode")
	}

	// Catching up with many periods at once must not hang and must
	// leave the counter in range
	v.Write(13, 0x40)
	v.Tick(1_000_000)
	if !v.InterruptAsserted() {
		t.Error("The interrupt should fire after catching up many periods")
	}
	counter := uint16(v.Read(4)) | uint16(v.Read(5))<<8
	if counter > 101 {
		t.Errorf("The counter should stay in the period range, got %v", counter)
	}
}

func TestMOS6522Timer1ReadClearsFlag(t *testing.T) {
	var v MOS6522
	v.Write(14, 0x80|0x40)
	v.Write(4, 10)
	v.Write(5, 0)
	v.Tick(50)

	if !v.InterruptAsserted() {
		t.Error("The interrupt should be asserted")
	}
	v.Read(4) // Reading T1C-L clears the flag
	if v.InterruptAsserted() {
		t.Error("Reading T1C-L should clear the T1 flag")
	}
}

func TestMOS6522Timer2(t *testing.T) {
	var v MOS6522
	v.Write(14, 0x80|0x20) // Enable the T2 interrupt
	v.Write(8, 50)         // T2 latch low
	v.Write(9, 0)          // T2 counter high, starts the timer

	v.Tick(50)
	if v.InterruptAsserted() {
		t.Error("The interrupt should not fire before the timer underflows")
	}
	v.Tick(2)
	if !v.InterruptAsserted() {
		t.Error("The interrupt should fire when the timer underflows")
	}

	v.Write(13, 0x20)
	v.Tick(1000)
	if v.InterruptAsserted() {
		t.Error("T2 should not fire again")
	}
}

func TestMOS6522InterruptEnable(t *testing.T) {
	var v MOS6522
	v.Write(4, 10)
	v.Write(5, 0)
	v.Tick(50) // T1 underflows but the interrupt is not enabled

	if v.InterruptAsserted() {
		t.Error("The interrupt should be masked by IER")
	}
	if v.Read(13)&0x40 == 0 {
		t.Error("The T1 flag should be set on IFR even when masked")
	}
	if v.Read(13)&0x80 != 0 {
		t.Error("The IFR bit 7 should be clear when masked")
	}

	v.Write(14, 0x80|0x40) // Enable the T1 interrupt afterwards
	if !v.InterruptAsserted() {
		t.Error("The interrupt should be asserted once enabled")
	}
	if v.Read(13)&0x80 == 0 {
		t.Error("The IFR bit 7 should be set when interrupting")
	}

	v.Write(14, 0x40) // Disable the T1 interrupt
	if v.InterruptAsserted() {
		t.Error("The interrupt should be masked again")
	}
	if v.Read(14) != 0x80 {
		t.Errorf("IER should read as 0x80, got 0x%02x", v.Read(14))
	}
}

func TestMOS6522Ports(t *testing.T) {
	var v MOS6522
	v.Write(3, 0x0f) // DDRA: half output, half input
	v.Write(1, 0xa5) // ORA
	v.SetInputA(0x5a)

	if port := v.GetPortA(); port != 0x55 {
		t.Errorf("Port A should mix output and input bits, got 0x%02x", port)
	}
	if value := v.Read(1); value != 0x55 {
		t.Errorf("Reading port A should mix output and input bits, got 0x%02x", value)
	}

	v.Write(2, 0xff) // DDRB: all output
	v.Write(0, 0x07)
	if port := v.GetPortB(); port != 0x07 {
		t.Errorf("Port B should return the output register, got 0x%02x", port)
	}
}

func TestMOS6522Reset(t *testing.T) {
	var v MOS6522
	v.Write(14, 0x80|0x40)
	v.Write(4, 10)
	v.Write(5, 0)
	v.Tick(50)

	v.Reset()
	if v.InterruptAsserted() {
		t.Error("Reset should clear the interrupt state")
	}
	if v.Read(2) != 0 || v.Read(3) != 0 {
		t.Error("Reset should clear the data direction registers")
	}
}
