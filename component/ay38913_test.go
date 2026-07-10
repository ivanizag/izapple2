package component

import (
	"testing"
)

func ay38913SetRegister(ay *AY38913, reg uint8, value uint8) {
	ay.LatchAddress(reg)
	ay.WriteData(value)
}

func TestAY38913RegisterMasks(t *testing.T) {
	var ay AY38913

	ay38913SetRegister(&ay, 1, 0xff)
	if v := ay.ReadData(); v != 0x0f {
		t.Errorf("The coarse tone period should be masked to 4 bits, got 0x%02x", v)
	}

	ay38913SetRegister(&ay, 6, 0xff)
	if v := ay.ReadData(); v != 0x1f {
		t.Errorf("The noise period should be masked to 5 bits, got 0x%02x", v)
	}

	ay38913SetRegister(&ay, 8, 0xff)
	if v := ay.ReadData(); v != 0x1f {
		t.Errorf("The amplitude should be masked to 5 bits, got 0x%02x", v)
	}
}

func TestAY38913AddressLatch(t *testing.T) {
	var ay AY38913

	ay38913SetRegister(&ay, 2, 0x12)
	ay.LatchAddress(0x1f) // Invalid, high nibble not zero
	ay.LatchAddress(0x02)
	if v := ay.ReadData(); v != 0x12 {
		t.Errorf("The invalid address should be ignored, got 0x%02x", v)
	}
}

func TestAY38913Tone(t *testing.T) {
	var ay AY38913
	ay38913SetRegister(&ay, 0, 5)    // Channel A tone period 5
	ay38913SetRegister(&ay, 7, 0x3e) // Only tone A enabled
	ay38913SetRegister(&ay, 8, 15)   // Channel A at full fixed volume

	// The output must toggle between 0 and 1 every 5 steps
	last := ay.Step()
	changes := []int{}
	for i := range 50 {
		level := ay.Step()
		if level != last {
			if level != 0.0 && level != 1.0 {
				t.Fatalf("The level should toggle between 0 and 1, got %v", level)
			}
			changes = append(changes, i)
			last = level
		}
	}
	if len(changes) < 8 {
		t.Fatalf("The tone should toggle, got %v changes", len(changes))
	}
	for i := 1; i < len(changes); i++ {
		if changes[i]-changes[i-1] != 5 {
			t.Errorf("The half period should be 5 steps, got %v", changes[i]-changes[i-1])
		}
	}
}

func TestAY38913ConstantLevel(t *testing.T) {
	var ay AY38913
	ay38913SetRegister(&ay, 7, 0x3f) // Everything disabled
	ay38913SetRegister(&ay, 8, 15)   // Channel A at full fixed volume

	// With tone and noise disabled the channel stays high, this is how
	// the volume register is used to play digitized samples
	for range 20 {
		if level := ay.Step(); level != 1.0 {
			t.Fatalf("The channel should stay at full level, got %v", level)
		}
	}
}

func TestAY38913EnvelopeDecay(t *testing.T) {
	var ay AY38913
	ay38913SetRegister(&ay, 7, 0x3f)  // Everything disabled, channels high
	ay38913SetRegister(&ay, 8, 0x10)  // Channel A on envelope
	ay38913SetRegister(&ay, 11, 1)    // Envelope period 1
	ay38913SetRegister(&ay, 13, 0x00) // Decay then off

	first := ay.Step()
	if first != 1.0 {
		t.Errorf("The envelope should start at full level, got %v", first)
	}

	last := first
	for range 100 {
		level := ay.Step()
		if level > last {
			t.Fatalf("The envelope should never raise on decay, got %v after %v", level, last)
		}
		last = level
	}
	if last != 0.0 {
		t.Errorf("The envelope should end at zero, got %v", last)
	}
}

func TestAY38913EnvelopeAttackHold(t *testing.T) {
	var ay AY38913
	ay38913SetRegister(&ay, 7, 0x3f)
	ay38913SetRegister(&ay, 8, 0x10)
	ay38913SetRegister(&ay, 11, 1)
	ay38913SetRegister(&ay, 13, 0x0d) // Attack then hold at full level

	var last float32
	for range 100 {
		last = ay.Step()
	}
	if last != 1.0 {
		t.Errorf("The envelope should hold at full level, got %v", last)
	}
}

func TestAY38913Noise(t *testing.T) {
	var ay AY38913
	ay38913SetRegister(&ay, 6, 1)    // Fastest noise
	ay38913SetRegister(&ay, 7, 0x37) // Only noise A enabled
	ay38913SetRegister(&ay, 8, 15)

	changes := 0
	last := ay.Step()
	for range 1000 {
		level := ay.Step()
		if level != last {
			changes++
			last = level
		}
	}
	if changes < 100 {
		t.Errorf("The noise should toggle the output often, got %v changes", changes)
	}
}
