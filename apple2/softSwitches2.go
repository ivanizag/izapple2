package apple2

const (
	ioDataKeyboard uint8 = 0x10

	ioFlagGraphics     uint8 = 0x50
	ioFlagMixed        uint8 = 0x52
	ioFlagSecondPage   uint8 = 0x54
	ioFlagHiRes        uint8 = 0x56
	ioFlagAnnunciator0 uint8 = 0x58
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

	io.addSoftSwitchRW(0x00, getKeySoftSwitch)         // Keyboard
	io.addSoftSwitchRW(0x10, strobeKeyboardSoftSwitch) // Keyboard Strobe
	io.addSoftSwitchR(0x20, notImplementedSoftSwitchR) // Cassette Output
	io.addSoftSwitchR(0x30, notImplementedSoftSwitchR) // Speaker
	io.addSoftSwitchR(0x40, notImplementedSoftSwitchR) // Game connector Strobe
	// Note: Some sources indicate that all these cover 16 positions
	// for read and write. But the Apple2e take over some of them, with
	// the prevention on acting only on writes.

	io.addSoftSwitchRW(0x50, getSoftSwitch(ioFlagGraphics, false))
	io.addSoftSwitchRW(0x51, getSoftSwitch(ioFlagGraphics, true))
	io.addSoftSwitchRW(0x52, getSoftSwitch(ioFlagMixed, false))
	io.addSoftSwitchRW(0x53, getSoftSwitch(ioFlagMixed, true))
	io.addSoftSwitchRW(0x54, getSoftSwitch(ioFlagSecondPage, false))
	io.addSoftSwitchRW(0x55, getSoftSwitch(ioFlagSecondPage, true))
	io.addSoftSwitchRW(0x56, getSoftSwitch(ioFlagHiRes, false))
	io.addSoftSwitchRW(0x57, getSoftSwitch(ioFlagHiRes, true))
	io.addSoftSwitchRW(0x58, getSoftSwitch(ioFlagAnnunciator0, false))
	io.addSoftSwitchRW(0x59, getSoftSwitch(ioFlagAnnunciator0, true))
	io.addSoftSwitchRW(0x5a, getSoftSwitch(ioFlagAnnunciator1, false))
	io.addSoftSwitchRW(0x5b, getSoftSwitch(ioFlagAnnunciator1, true))
	io.addSoftSwitchRW(0x5c, getSoftSwitch(ioFlagAnnunciator2, false))
	io.addSoftSwitchRW(0x5d, getSoftSwitch(ioFlagAnnunciator2, true))
	io.addSoftSwitchRW(0x5e, getSoftSwitch(ioFlagAnnunciator3, false))
	io.addSoftSwitchRW(0x5f, getSoftSwitch(ioFlagAnnunciator3, true))

	io.addSoftSwitchR(0x60, notImplementedSoftSwitchR) // Cassette Input
	io.addSoftSwitchR(0x61, getStatusSoftSwitch(ioFlagButton0))
	io.addSoftSwitchR(0x62, getStatusSoftSwitch(ioFlagButton1))
	io.addSoftSwitchR(0x63, getStatusSoftSwitch(ioFlagButton2))
	io.addSoftSwitchR(0x64, getStatusSoftSwitch(ioDataPaddle0))
	io.addSoftSwitchR(0x65, getStatusSoftSwitch(ioDataPaddle1))
	io.addSoftSwitchR(0x66, getStatusSoftSwitch(ioDataPaddle2))
	io.addSoftSwitchR(0x67, getStatusSoftSwitch(ioDataPaddle3))

	// The previous 8 softswitches are repeated
	io.addSoftSwitchR(0x68, notImplementedSoftSwitchR) // Cassette Input
	io.addSoftSwitchR(0x69, getStatusSoftSwitch(ioFlagButton0))
	io.addSoftSwitchR(0x6A, getStatusSoftSwitch(ioFlagButton1))
	io.addSoftSwitchR(0x6B, getStatusSoftSwitch(ioFlagButton2))
	io.addSoftSwitchR(0x6C, getStatusSoftSwitch(ioDataPaddle0))
	io.addSoftSwitchR(0x6D, getStatusSoftSwitch(ioDataPaddle1))
	io.addSoftSwitchR(0x6E, getStatusSoftSwitch(ioDataPaddle2))
	io.addSoftSwitchR(0x6F, getStatusSoftSwitch(ioDataPaddle3))

	io.addSoftSwitchR(0x70, notImplementedSoftSwitchR) // Game controllers reset
}

func notImplementedSoftSwitchR(*ioC0Page) uint8 {
	return 0
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

func getKeySoftSwitch(io *ioC0Page) uint8 {
	if io.keyboard != nil {
		if key, ok := io.keyboard.GetKey(); ok {
			io.softSwitchesData[ioDataKeyboard] = key + (1 << 7)
		}
	}
	return io.softSwitchesData[ioDataKeyboard]
}

func strobeKeyboardSoftSwitch(io *ioC0Page) uint8 {
	result := io.softSwitchesData[ioDataKeyboard]
	io.softSwitchesData[ioDataKeyboard] &^= 1 << 7
	return result
}
