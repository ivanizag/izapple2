package izapple2

import (
	"testing"
)

func makeMockingboardTester(t *testing.T) (*Apple2, *CardMockingboard) {
	overrides := newConfiguration()
	overrides.set(confS4, "mockingboard")
	at, err := makeApple2Tester("2plus", overrides)
	if err != nil {
		t.Fatal(err)
	}
	a := at.a
	card, ok := a.cards[4].(*CardMockingboard)
	if !ok {
		t.Fatal("The Mockingboard should be in slot 4")
	}
	return a, card
}

func TestCardMockingboardTimerReadIsCycleExact(t *testing.T) {
	a, _ := makeMockingboardTester(t)

	// Start VIA 1 T1 with 0x2000 cycles
	a.mmu.Poke(0xc404, 0x00)
	a.mmu.Poke(0xc405, 0x20)

	a.cycles += 100
	counter := uint16(a.mmu.Peek(0xc404)) | uint16(a.mmu.Peek(0xc405))<<8
	if counter != 0x2000-100 {
		t.Errorf("The timer should decrement with the CPU cycles, got 0x%04x", counter)
	}
}

func TestCardMockingboardTimerIRQ(t *testing.T) {
	a, card := makeMockingboardTester(t)

	// Program VIA 1 T1 free running with interrupts every 0x1000 cycles
	a.mmu.Poke(0xc40b, 0x40) // ACR: T1 free running
	a.mmu.Poke(0xc40e, 0xc0) // IER: enable T1
	a.mmu.Poke(0xc404, 0x00)
	a.mmu.Poke(0xc405, 0x10)

	a.cycles += 0x800
	card.tick()
	if a.irqRequests != 0 {
		t.Error("The IRQ should not be requested before the timer underflows")
	}

	a.cycles += 0x900
	card.tick()
	if a.irqRequests == 0 {
		t.Error("The IRQ should be requested when the timer underflows")
	}

	// Reading T1C-L clears the interrupt
	a.mmu.Peek(0xc404)
	if a.irqRequests != 0 {
		t.Error("The IRQ should be released after reading T1C-L")
	}
}

func TestCardMockingboardSecondVIA(t *testing.T) {
	a, card := makeMockingboardTester(t)

	// The second VIA is at $Cn80, enable its T2
	a.mmu.Poke(0xc48e, 0xa0) // IER: enable T2
	a.mmu.Poke(0xc488, 0x50)
	a.mmu.Poke(0xc489, 0x00)

	a.cycles += 0x100
	card.tick()
	if a.irqRequests == 0 {
		t.Error("The IRQ should be requested by the second VIA")
	}

	a.mmu.Poke(0xc48d, 0x20) // Clear the T2 flag on IFR
	if a.irqRequests != 0 {
		t.Error("The IRQ should be released after clearing IFR")
	}
}

// writePsg drives the AY bus protocol like the Mockingboard drivers do
func writePsg(a *Apple2, base uint16, reg uint8, value uint8) {
	a.mmu.Poke(base+1, reg)  // ORA: register number
	a.mmu.Poke(base, 4|3)    // Latch address
	a.mmu.Poke(base, 4)      // Inactive
	a.mmu.Poke(base+1, value)
	a.mmu.Poke(base, 4|2) // Write data
	a.mmu.Poke(base, 4)   // Inactive
}

func TestCardMockingboardPsgBus(t *testing.T) {
	a, _ := makeMockingboardTester(t)

	// Set the VIA 1 ports as the drivers do
	a.mmu.Poke(0xc403, 0xff) // DDRA: all output
	a.mmu.Poke(0xc402, 0x07) // DDRB: control lines output

	writePsg(a, 0xc400, 7, 0x3e) // Mixer: only tone A

	// Read the register back through the bus
	a.mmu.Poke(0xc401, 7)
	a.mmu.Poke(0xc400, 4|3) // Latch address
	a.mmu.Poke(0xc400, 4)
	a.mmu.Poke(0xc403, 0x00) // DDRA: all input to read
	a.mmu.Poke(0xc400, 4|1)  // Read data
	if v := a.mmu.Peek(0xc401); v != 0x3e {
		t.Errorf("The PSG register should read back 0x3e, got 0x%02x", v)
	}
}

type testAudioSink struct {
	events int
}

func (s *testAudioSink) PushLevel(cycle uint64, level float32) {
	s.events++
}

func TestCardMockingboardGeneratesSound(t *testing.T) {
	a, card := makeMockingboardTester(t)

	var card2 AudioSource
	for _, source := range a.GetAudioSources() {
		if source.GetAudioSourceName() == "mockingboard" {
			card2 = source
		}
	}
	if card2 == nil {
		t.Fatal("The card should be published as an audio source")
	}

	var sink testAudioSink
	card2.SetAudioSink(&sink)

	a.mmu.Poke(0xc403, 0xff)
	a.mmu.Poke(0xc402, 0x07)
	writePsg(a, 0xc400, 0, 100)  // Tone A period
	writePsg(a, 0xc400, 7, 0x3e) // Mixer: only tone A
	writePsg(a, 0xc400, 8, 15)   // Full volume

	// Run 10000 cycles, the tone toggles every 800 cycles
	a.cycles += 10_000
	card.tick()
	if sink.events < 10 {
		t.Errorf("The sink should receive the tone level changes, got %v", sink.events)
	}
}
