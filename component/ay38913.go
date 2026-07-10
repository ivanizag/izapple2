package component

/*
General Instrument AY-3-8913 Programmable Sound Generator
See:

	https://map.grauw.nl/resources/sound/generalinstrument_ay-3-8910.pdf
	https://en.wikipedia.org/wiki/General_Instrument_AY-3-8910

The 8913 is the variant of the AY-3-8910 without I/O ports. Two of them
are used by the Mockingboard card, driven by the 6522 VIAs.

Registers:

	R0-R1:   channel A tone period, 12 bits
	R2-R3:   channel B tone period, 12 bits
	R4-R5:   channel C tone period, 12 bits
	R6:      noise period, 5 bits
	R7:      mixer, bits 0-2 disable tone A-C, bits 3-5 disable noise A-C
	R8-R10:  channel A-C amplitude, bit 4 selects the envelope
	R11-R12: envelope period, 16 bits
	R13:     envelope shape: CONTINUE, ATTACK, ALTERNATE, HOLD

The bus operations selected by the BDIR and BC1 pins are exposed as the
LatchAddress, WriteData and ReadData methods.

Synthesis is done in steps of 8 chip clocks with Step(), the granularity
of the tone generators. A tone with period TP completes a full cycle
every 2*TP steps (16*TP clocks as per the datasheet).
*/
type AY38913 struct {
	regs    [16]uint8
	address uint8

	toneCounter [3]uint16
	toneOut     [3]bool

	noiseCounter uint16
	noiseShift   uint32 // 17 bit LFSR

	envCounter   uint32
	envPosition  uint8 // 0 to 15 within a ramp
	envAttack    bool  // Direction of the current ramp
	envHolding   bool
	envHeldLevel uint8
}

var ay38913RegisterMasks = [16]uint8{
	0xff, 0x0f, 0xff, 0x0f, 0xff, 0x0f, 0x1f, 0xff,
	0x1f, 0x1f, 0x1f, 0xff, 0xff, 0x0f, 0xff, 0xff,
}

/*
Volume of each of the 16 amplitude levels. The DAC is logarithmic with
steps of ~3dB per level. Normalized so that one channel at full volume
gives 1.0.
*/
var ay38913Volumes = [16]float32{
	0.0000, 0.0106, 0.0150, 0.0222, 0.0320, 0.0466, 0.0665, 0.1039,
	0.1237, 0.1986, 0.2803, 0.3548, 0.4702, 0.6030, 0.7530, 1.0000,
}

// LatchAddress selects the register for the next data operations
func (ay *AY38913) LatchAddress(value uint8) {
	// The chip compares the high nibble with its hardwired mask
	if value&0xf0 == 0 {
		ay.address = value & 0x0f
	}
}

// WriteData writes to the selected register
func (ay *AY38913) WriteData(value uint8) {
	ay.regs[ay.address] = value & ay38913RegisterMasks[ay.address]

	if ay.address == 13 {
		// Writing the envelope shape restarts the envelope
		ay.envCounter = 0
		ay.envPosition = 0
		ay.envAttack = value&0x04 != 0
		ay.envHolding = false
	}
}

// ReadData reads the selected register
func (ay *AY38913) ReadData() uint8 {
	return ay.regs[ay.address]
}

// Reset clears the registers and the generators
func (ay *AY38913) Reset() {
	for i := range ay.regs {
		ay.regs[i] = 0
	}
	ay.address = 0
	for i := range 3 {
		ay.toneCounter[i] = 0
		ay.toneOut[i] = false
	}
	ay.noiseCounter = 0
	ay.noiseShift = 1
	ay.envCounter = 0
	ay.envPosition = 0
	ay.envAttack = false
	ay.envHolding = false
}

// Step advances the generators 8 chip clocks and returns the mixed level
// of the three channels, from 0.0 to 3.0
func (ay *AY38913) Step() float32 {
	// Tone generators, a half period is TP steps
	for i := range 3 {
		period := uint16(ay.regs[2*i]) | uint16(ay.regs[2*i+1])<<8
		if period == 0 {
			period = 1
		}
		ay.toneCounter[i]++
		if ay.toneCounter[i] >= period {
			ay.toneCounter[i] = 0
			ay.toneOut[i] = !ay.toneOut[i]
		}
	}

	// Noise generator, the LFSR shifts every 2*NP steps
	noisePeriod := 2 * uint16(ay.regs[6])
	if noisePeriod == 0 {
		noisePeriod = 2
	}
	if ay.noiseShift == 0 {
		ay.noiseShift = 1 // Make the zero value usable
	}
	ay.noiseCounter++
	if ay.noiseCounter >= noisePeriod {
		ay.noiseCounter = 0
		// 17 bit LFSR with taps on bits 0 and 3
		bit := (ay.noiseShift ^ (ay.noiseShift >> 3)) & 1
		ay.noiseShift = (ay.noiseShift >> 1) | (bit << 16)
	}
	noiseOut := ay.noiseShift&1 != 0

	// Envelope generator, one of the 16 levels every 2*EP steps
	if !ay.envHolding {
		envPeriod := 2 * (uint32(ay.regs[11]) | uint32(ay.regs[12])<<8)
		if envPeriod == 0 {
			envPeriod = 2
		}
		ay.envCounter++
		if ay.envCounter >= envPeriod {
			ay.envCounter = 0
			if ay.envPosition < 15 {
				ay.envPosition++
			} else {
				ay.endOfEnvelopeRamp()
			}
		}
	}

	// Mix the channels
	level := float32(0)
	mixer := ay.regs[7]
	for i := range 3 {
		toneEnabled := mixer&(1<<i) == 0
		noiseEnabled := mixer&(1<<(3+i)) == 0
		// The channel is high unless an enabled generator drives it low
		high := (ay.toneOut[i] || !toneEnabled) && (noiseOut || !noiseEnabled)
		if high {
			amplitude := ay.regs[8+i]
			var volume uint8
			if amplitude&0x10 != 0 {
				volume = ay.envelopeLevel()
			} else {
				volume = amplitude & 0x0f
			}
			level += ay38913Volumes[volume]
		}
	}
	return level
}

func (ay *AY38913) envelopeLevel() uint8 {
	if ay.envHolding {
		return ay.envHeldLevel
	}
	if ay.envAttack {
		return ay.envPosition
	}
	return 15 - ay.envPosition
}

// endOfEnvelopeRamp applies the shape bits when a 16 level ramp completes
func (ay *AY38913) endOfEnvelopeRamp() {
	shape := ay.regs[13]
	continues := shape&0x08 != 0
	attack := shape&0x04 != 0
	alternate := shape&0x02 != 0
	hold := shape&0x01 != 0

	if !continues {
		ay.envHolding = true
		ay.envHeldLevel = 0
	} else if hold {
		ay.envHolding = true
		if attack != alternate {
			ay.envHeldLevel = 15
		} else {
			ay.envHeldLevel = 0
		}
	} else {
		// Restart the ramp, reversing the direction when alternating
		ay.envPosition = 0
		if alternate {
			ay.envAttack = !ay.envAttack
		}
	}
}
