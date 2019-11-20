package apple2

const (
	ioDataKeyboard uint8 = 0x10

	ioFlagText         uint8 = 0x50
	ioFlagMixed        uint8 = 0x52
	ioFlagSecondPage   uint8 = 0x54
	ioFlagHiRes        uint8 = 0x56
	ioFlagAnnunciator0 uint8 = 0x58 // On Copam Electronics Base-64A this is used to bank swith the ROM
	ioFlagAnnunciator1 uint8 = 0x5a
	ioFlagAnnunciator2 uint8 = 0x5c
	ioFlagAnnunciator3 uint8 = 0x5e

	ioDataCassette uint8 = 0x60
	ioFlagButton0  uint8 = 0x61
	ioFlagButton1  uint8 = 0x62
	ioFlagButton2  uint8 = 0x63
	ioDataPaddle0  uint8 = 0x64
	ioDataPaddle1  uint8 = 0x65
	ioDataPaddle2  uint8 = 0x66
	ioDataPaddle3  uint8 = 0x67
)

func addApple2SoftSwitches(io *ioC0Page) {

	io.addSoftSwitchRW(0x00, keySoftSwitch, "KEYBOARD")           // Keyboard
	io.addSoftSwitchRW(0x10, strobeKeyboardSoftSwitch, "AKD")     // Keyboard Strobe
	io.addSoftSwitchR(0x20, notImplementedSoftSwitchR, "TAPEOUT") // Cassette Output
	io.addSoftSwitchR(0x30, speakerSoftSwitch, "SPEAKER")         // Speaker
	io.addSoftSwitchR(0x40, notImplementedSoftSwitchR, "STROBE")  // Game connector Strobe
	// Note: Some sources indicate that all these cover 16 positions
	// for read and write. But the Apple2e take over some of them, with
	// the prevention on acting only on writes.

	io.addSoftSwitchRW(0x50, getSoftSwitch(ioFlagText, false), "TEXTOFF")
	io.addSoftSwitchRW(0x51, getSoftSwitch(ioFlagText, true), "TEXTON")
	io.addSoftSwitchRW(0x52, getSoftSwitch(ioFlagMixed, false), "MIXEDOFF")
	io.addSoftSwitchRW(0x53, getSoftSwitch(ioFlagMixed, true), "MIXEDON")
	io.addSoftSwitchRW(0x54, getSoftSwitch(ioFlagSecondPage, false), "PAGE2OFF")
	io.addSoftSwitchRW(0x55, getSoftSwitch(ioFlagSecondPage, true), "PAGE2ON")
	io.addSoftSwitchRW(0x56, getSoftSwitch(ioFlagHiRes, false), "HIRESOFF")
	io.addSoftSwitchRW(0x57, getSoftSwitch(ioFlagHiRes, true), "HIRESON")

	io.addSoftSwitchRW(0x58, getSoftSwitch(ioFlagAnnunciator0, false), "ANN0OFF")
	io.addSoftSwitchRW(0x59, getSoftSwitch(ioFlagAnnunciator0, true), "ANN0ON")
	io.addSoftSwitchRW(0x5a, getSoftSwitch(ioFlagAnnunciator1, false), "ANN1OFF")
	io.addSoftSwitchRW(0x5b, getSoftSwitch(ioFlagAnnunciator1, true), "ANN1ON")
	io.addSoftSwitchRW(0x5c, getSoftSwitch(ioFlagAnnunciator2, false), "ANN2OFF")
	io.addSoftSwitchRW(0x5d, getSoftSwitch(ioFlagAnnunciator2, true), "ANN2ON")
	io.addSoftSwitchRW(0x5e, getSoftSwitch(ioFlagAnnunciator3, false), "ANN3OFF")
	io.addSoftSwitchRW(0x5f, getSoftSwitch(ioFlagAnnunciator3, true), "ANN3ON")

	io.addSoftSwitchR(0x60, notImplementedSoftSwitchR, "CASSETTE") // Cassette Input
	io.addSoftSwitchR(0x61, getButtonSoftSwitch(0), "PB0")
	io.addSoftSwitchR(0x62, getButtonSoftSwitch(1), "PB1")
	io.addSoftSwitchR(0x63, getButtonSoftSwitch(2), "PB2")
	io.addSoftSwitchR(0x64, getPaddleSoftSwitch(0), "PDL0")
	io.addSoftSwitchR(0x65, getPaddleSoftSwitch(1), "PDL1")
	io.addSoftSwitchR(0x66, getPaddleSoftSwitch(2), "PDL2")
	io.addSoftSwitchR(0x67, getPaddleSoftSwitch(3), "PDL3")

	// The previous 8 softswitches are repeated
	io.addSoftSwitchR(0x68, notImplementedSoftSwitchR, "CASSETTE") // Cassette Input
	io.addSoftSwitchR(0x69, getButtonSoftSwitch(0), "PB0")
	io.addSoftSwitchR(0x6A, getButtonSoftSwitch(1), "PB1")
	io.addSoftSwitchR(0x6B, getButtonSoftSwitch(2), "PB2")
	io.addSoftSwitchR(0x6C, getPaddleSoftSwitch(0), "PDL0")
	io.addSoftSwitchR(0x6D, getPaddleSoftSwitch(1), "PDL1")
	io.addSoftSwitchR(0x6E, getPaddleSoftSwitch(2), "PDL2")
	io.addSoftSwitchR(0x6F, getPaddleSoftSwitch(3), "PDL3")

	io.addSoftSwitchR(0x70, strobePaddlesSoftSwitch, "RESETPDL") // Game controllers reset
}

func notImplementedSoftSwitchR(*ioC0Page) uint8 {
	return 0
}

func notImplementedSoftSwitchW(*ioC0Page, uint8) {
}

func setStatusSoftSwitch(ioFlag uint8) softSwitchW {
	return func(io *ioC0Page, value uint8) {
		io.softSwitchesData[ioFlag] = value
	}
}

func getStatusSoftSwitch(ioFlag uint8) softSwitchR {
	return func(io *ioC0Page) uint8 {
		return io.softSwitchesData[ioFlag]
	}
}

func getSoftSwitch(ioFlag uint8, isSet bool) softSwitchR {
	return func(io *ioC0Page) uint8 {
		if isSet {
			io.softSwitchesData[ioFlag] = ssOn
		} else {
			io.softSwitchesData[ioFlag] = ssOff
		}
		return 0
	}
}

func speakerSoftSwitch(io *ioC0Page) uint8 {
	if io.speaker != nil {
		io.speaker.Click(io.apple2.cpu.GetCycles())
	}
	return 0
}

func keySoftSwitch(io *ioC0Page) uint8 {
	strobed := (io.softSwitchesData[ioDataKeyboard] & (1 << 7)) == 0
	if io.keyboard != nil {
		if key, ok := io.keyboard.GetKey(strobed); ok {
			io.softSwitchesData[ioDataKeyboard] = key + (1 << 7)
		}
	}
	value := io.softSwitchesData[ioDataKeyboard]
	//fmt.Printf("Key $%02x, %v\n", value, strobed)
	return value

}

func strobeKeyboardSoftSwitch(io *ioC0Page) uint8 {
	result := io.softSwitchesData[ioDataKeyboard]
	//fmt.Printf("Strobe $%02x\n", result)
	io.softSwitchesData[ioDataKeyboard] &^= 1 << 7
	return result
}

func getButtonSoftSwitch(i int) softSwitchR {
	return func(io *ioC0Page) uint8 {
		if io.joysticks != nil && io.joysticks.ReadButton(i) {
			return 128
		}
		return 0
	}
}

/*
  Paddle values are calculated by the time taken by a current going
  througt the paddle variable resistor to charge a capacitor.
  The capacitor is discharged via the strobe softswitch. The result is
  hoy many times a 11 cycles loop runs before the capacitor reaches
  the voltage threshold.

  See: http://www.1000bit.it/support/manuali/apple/technotes/aiie/tn.aiie.06.html
*/

const paddleToCyclesFactor = 11

func getPaddleSoftSwitch(i int) softSwitchR {
	return func(io *ioC0Page) uint8 {
		if io.joysticks == nil {
			return 255 // Capacitors never discharge if there is not joystick
		}
		reading, hasData := io.joysticks.ReadPaddle(i)
		if !hasData {
			return 255 // Capacitors never discharge if there is not joystick
		}
		cyclesNeeded := uint64(reading) * paddleToCyclesFactor
		cyclesElapsed := io.apple2.cpu.GetCycles() - io.paddlesStrobeCycle
		if cyclesElapsed < cyclesNeeded {
			// The capacitor is not charged yet
			return 128
		}
		return 0
	}
}

func strobePaddlesSoftSwitch(io *ioC0Page) uint8 {
	// On the real machine this discharges the capacitors.
	io.paddlesStrobeCycle = io.apple2.cpu.GetCycles()
	return 0
}
