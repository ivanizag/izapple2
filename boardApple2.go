package izapple2

const (
	ioDataKeyboard uint8 = 0x10

	ioFlagText         uint8 = 0x50
	ioFlagMixed        uint8 = 0x52
	ioFlagSecondPage   uint8 = 0x54
	ioFlagHiRes        uint8 = 0x56
	ioFlagAnnunciator0 uint8 = 0x58
	ioFlagAnnunciator1 uint8 = 0x5a
	ioFlagAnnunciator2 uint8 = 0x5c
	ioFlagAnnunciator3 uint8 = 0x5e

	// ioDataCassette uint8 = 0x60
	// ioFlagButton0  uint8 = 0x61
	// ioFlagButton1  uint8 = 0x62
	// ioFlagButton2  uint8 = 0x63
	// ioDataPaddle0  uint8 = 0x64
	// ioDataPaddle1  uint8 = 0x65
	// ioDataPaddle2  uint8 = 0x66
	// ioDataPaddle3  uint8 = 0x67

	// Not real softSwitches. Using the numbers to store the flags somewhere.
	ioFlagRGBCardActive uint8 = 0x7d
	ioFlag1RGBCard      uint8 = 0x7e
	ioFlag2RGBCard      uint8 = 0x7f
)

func addApple2SoftSwitches(io *ioC0Page) {

	io.addSoftSwitchRW(0x00, buildKeySoftSwitch(io), "KEYBOARD")           // Keyboard
	io.addSoftSwitchRW(0x10, buildStrobeKeyboardSoftSwitch(io), "AKD")     // Keyboard Strobe
	io.addSoftSwitchR(0x20, buildNotImplementedSoftSwitchR(io), "TAPEOUT") // Cassette Output
	io.addSoftSwitchRW(0x30, buildSpeakerSoftSwitch(io), "SPEAKER")        // Speaker
	io.addSoftSwitchR(0x40, buildNotImplementedSoftSwitchR(io), "STROBE")  // Game connector Strobe
	// Note: Some sources indicate that all these cover 16 positions
	// for read and write. But the Apple2e takes over some of them, with
	// the prevention on acting only on writes.

	io.addSoftSwitchRW(0x50, getSoftSwitch(io, ioFlagText, false), "TEXTOFF")
	io.addSoftSwitchRW(0x51, getSoftSwitch(io, ioFlagText, true), "TEXTON")
	io.addSoftSwitchRW(0x52, getSoftSwitch(io, ioFlagMixed, false), "MIXEDOFF")
	io.addSoftSwitchRW(0x53, getSoftSwitch(io, ioFlagMixed, true), "MIXEDON")
	io.addSoftSwitchRW(0x54, getSoftSwitch(io, ioFlagSecondPage, false), "PAGE2OFF")
	io.addSoftSwitchRW(0x55, getSoftSwitch(io, ioFlagSecondPage, true), "PAGE2ON")
	io.addSoftSwitchRW(0x56, getSoftSwitch(io, ioFlagHiRes, false), "HIRESOFF")
	io.addSoftSwitchRW(0x57, getSoftSwitch(io, ioFlagHiRes, true), "HIRESON")

	io.addSoftSwitchRW(0x58, getSoftSwitch(io, ioFlagAnnunciator0, false), "ANN0OFF")
	io.addSoftSwitchRW(0x59, getSoftSwitch(io, ioFlagAnnunciator0, true), "ANN0ON")
	io.addSoftSwitchRW(0x5a, getSoftSwitch(io, ioFlagAnnunciator1, false), "ANN1OFF")
	io.addSoftSwitchRW(0x5b, getSoftSwitch(io, ioFlagAnnunciator1, true), "ANN1ON")
	io.addSoftSwitchRW(0x5c, getSoftSwitch(io, ioFlagAnnunciator2, false), "ANN2OFF")
	io.addSoftSwitchRW(0x5d, getSoftSwitch(io, ioFlagAnnunciator2, true), "ANN2ON")
	io.addSoftSwitchRW(0x5e, getSoftSwitch(io, ioFlagAnnunciator3, false), "ANN3OFF")
	io.addSoftSwitchRW(0x5f, getSoftSwitch(io, ioFlagAnnunciator3, true), "ANN3ON")

	io.addSoftSwitchR(0x60, buildNotImplementedSoftSwitchR(io), "CASSETTE") // Cassette Input
	io.addSoftSwitchR(0x61, buildButtonSoftSwitch(io, 0), "PB0")
	io.addSoftSwitchR(0x62, buildButtonSoftSwitch(io, 1), "PB1")
	io.addSoftSwitchR(0x63, buildButtonSoftSwitch(io, 2), "PB2")
	io.addSoftSwitchR(0x64, buildPaddleSoftSwitch(io, 0), "PDL0")
	io.addSoftSwitchR(0x65, buildPaddleSoftSwitch(io, 1), "PDL1")
	io.addSoftSwitchR(0x66, buildPaddleSoftSwitch(io, 2), "PDL2")
	io.addSoftSwitchR(0x67, buildPaddleSoftSwitch(io, 3), "PDL3")

	// The previous 8 softswitches are repeated
	io.addSoftSwitchR(0x68, buildNotImplementedSoftSwitchR(io), "CASSETTE") // Cassette Input
	io.addSoftSwitchR(0x69, buildButtonSoftSwitch(io, 0), "PB0")
	io.addSoftSwitchR(0x6A, buildButtonSoftSwitch(io, 1), "PB1")
	io.addSoftSwitchR(0x6B, buildButtonSoftSwitch(io, 2), "PB2")
	io.addSoftSwitchR(0x6C, buildPaddleSoftSwitch(io, 0), "PDL0")
	io.addSoftSwitchR(0x6D, buildPaddleSoftSwitch(io, 1), "PDL1")
	io.addSoftSwitchR(0x6E, buildPaddleSoftSwitch(io, 2), "PDL2")
	io.addSoftSwitchR(0x6F, buildPaddleSoftSwitch(io, 3), "PDL3")

	io.addSoftSwitchR(0x70, buildStrobePaddlesSoftSwitch(io), "RESETPDL") // Game controllers reset

	// For RGB screen modes. Default to NTSC artifacts
	io.softSwitchesData[ioFlag1RGBCard] = ssOn
	io.softSwitchesData[ioFlag2RGBCard] = ssOn
}

func buildNotImplementedSoftSwitchR(io *ioC0Page) softSwitchR {
	return func() uint8 {
		// Return random info. Some games (Serpentine) used CASSETTE and get stuck if not changing.
		return uint8(io.apple2.GetCycles())
	}
}

func buildNotImplementedSoftSwitchW(_ *ioC0Page) softSwitchW {
	return func(uint8) {
		// Do nothing
	}
}

func setStatusSoftSwitch(io *ioC0Page, ioFlag uint8) softSwitchW {
	return func(value uint8) {
		io.softSwitchesData[ioFlag] = value
	}
}

func getStatusSoftSwitch(io *ioC0Page, ioFlag uint8) softSwitchR {
	return func() uint8 {
		return io.softSwitchesData[ioFlag]
	}
}

func getSoftSwitch(io *ioC0Page, ioFlag uint8, isSet bool) softSwitchR {
	return func() uint8 {
		if isSet {
			io.softSwitchesData[ioFlag] = ssOn
		} else {
			io.softSwitchesData[ioFlag] = ssOff
		}
		return 0
	}
}

func buildSpeakerSoftSwitch(io *ioC0Page) softSwitchR {
	return func() uint8 {
		if io.speaker != nil {
			io.speaker.Click(io.apple2.GetCycles())
		}
		return 0
	}
}

func buildKeySoftSwitch(io *ioC0Page) softSwitchR {
	return func() uint8 {
		strobed := (io.softSwitchesData[ioDataKeyboard] & (1 << 7)) == 0
		if io.keyboard != nil {
			if key, ok := io.keyboard.GetKey(strobed); ok {
				io.softSwitchesData[ioDataKeyboard] = key + (1 << 7)
			}
		}
		value := io.softSwitchesData[ioDataKeyboard]
		return value
	}
}

func buildStrobeKeyboardSoftSwitch(io *ioC0Page) softSwitchR {
	return func() uint8 {
		result := io.softSwitchesData[ioDataKeyboard]
		io.softSwitchesData[ioDataKeyboard] &^= 1 << 7
		return result
	}
}

func buildButtonSoftSwitch(io *ioC0Page, i int) softSwitchR {
	return func() uint8 {
		if io.joysticks != nil && io.joysticks.ReadButton(i) {
			return 128
		}
		return 0
	}
}

/*
  Paddle values are calculated by the time taken by a current going
  through the paddle variable resistor to charge a capacitor.
  The capacitor is discharged via the strobe softswitch. The result is
  how many times a 11 cycles loop runs before the capacitor reaches
  the voltage threshold.

  See: http://www.1000bit.it/support/manuali/apple/technotes/aiie/tn.aiie.06.html
*/

const paddleToCyclesFactor = 11

func buildPaddleSoftSwitch(io *ioC0Page, i int) softSwitchR {
	return func() uint8 {
		if io.joysticks == nil {
			return 255 // Capacitors never discharge if there is not joystick
		}
		reading, hasData := io.joysticks.ReadPaddle(i)
		if !hasData {
			return 255 // Capacitors never discharge if there is not joystick
		}
		cyclesNeeded := uint64(reading) * paddleToCyclesFactor
		cyclesElapsed := io.apple2.GetCycles() - io.paddlesStrobeCycle
		if cyclesElapsed < cyclesNeeded {
			// The capacitor is not charged yet
			return 128
		}
		return 0
	}
}

func buildStrobePaddlesSoftSwitch(io *ioC0Page) softSwitchR {
	return func() uint8 {
		// On the real machine this discharges the capacitors.
		io.paddlesStrobeCycle = io.apple2.GetCycles()
		return 0
	}
}
