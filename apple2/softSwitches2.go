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
	ss := &io.softSwitches

	ss[0x00] = getKeySoftSwitch         // Keyboard
	ss[0x10] = strobeKeyboardSoftSwitch // Keyboard Strobe
	ss[0x20] = notImplementedSoftSwitch // Cassette Output
	ss[0x30] = notImplementedSoftSwitch // Speaker
	ss[0x40] = notImplementedSoftSwitch // Game connector Strobe
	// Note: Some sources indicate that all these cover 16 positions
	// for read and write. But the Apple2e take over some of them, with
	// the prevention on acting only on writes.

	ss[0x50] = getSoftSwitch(ioFlagGraphics, false)
	ss[0x51] = getSoftSwitch(ioFlagGraphics, true)
	ss[0x52] = getSoftSwitch(ioFlagMixed, false)
	ss[0x53] = getSoftSwitch(ioFlagMixed, true)
	ss[0x54] = getSoftSwitch(ioFlagSecondPage, false)
	ss[0x55] = getSoftSwitch(ioFlagSecondPage, true)
	ss[0x56] = getSoftSwitch(ioFlagHiRes, false)
	ss[0x57] = getSoftSwitch(ioFlagHiRes, true)
	ss[0x58] = getSoftSwitch(ioFlagAnnunciator0, false)
	ss[0x59] = getSoftSwitch(ioFlagAnnunciator0, true)
	ss[0x5a] = getSoftSwitch(ioFlagAnnunciator1, false)
	ss[0x5b] = getSoftSwitch(ioFlagAnnunciator1, true)
	ss[0x5c] = getSoftSwitch(ioFlagAnnunciator2, false)
	ss[0x5d] = getSoftSwitch(ioFlagAnnunciator2, true)
	ss[0x5e] = getSoftSwitch(ioFlagAnnunciator3, false)
	ss[0x5f] = getSoftSwitch(ioFlagAnnunciator3, true)

	ss[0x60] = notImplementedSoftSwitch // Cassetter Input
	ss[0x61] = getStatusSoftSwitch(ioFlagButton0)
	ss[0x62] = getStatusSoftSwitch(ioFlagButton1)
	ss[0x63] = getStatusSoftSwitch(ioFlagButton2)
	ss[0x64] = getStatusSoftSwitch(ioDataPaddle0)
	ss[0x65] = getStatusSoftSwitch(ioDataPaddle1)
	ss[0x66] = getStatusSoftSwitch(ioDataPaddle2)
	ss[0x67] = getStatusSoftSwitch(ioDataPaddle3)
	ss[0x68] = ss[0x60]
	ss[0x69] = ss[0x61]
	ss[0x6A] = ss[0x62]
	ss[0x6B] = ss[0x63]
	ss[0x6C] = ss[0x64]
	ss[0x6D] = ss[0x65]
	ss[0x6E] = ss[0x66]
	ss[0x6F] = ss[0x67]
	ss[0x70] = notImplementedSoftSwitch // Game controllers reset
}

func notImplementedSoftSwitch(*ioC0Page, bool, uint8) uint8 {
	return 0
}
func getStatusSoftSwitch(ioFlag uint8) softSwitch {
	return func(io *ioC0Page, isWrite bool, value uint8) uint8 {
		return io.softSwitchesData[ioFlag]
	}
}

func getSoftSwitch(ioFlag uint8, isSet bool) softSwitch {
	return func(io *ioC0Page, isWrite bool, value uint8) uint8 {
		if isSet {
			io.softSwitchesData[ioFlag] = ssOn
		} else {
			io.softSwitchesData[ioFlag] = ssOff
		}
		return 0
	}
}

func getKeySoftSwitch(p *ioC0Page, _ bool, _ uint8) uint8 {
	strobed := (p.softSwitchesData[ioDataKeyboard] & (1 << 7)) == 0
	if strobed && p.keyboard != nil {
		if key, ok := p.keyboard.getKey(); ok {
			p.softSwitchesData[ioDataKeyboard] = key + (1 << 7)
		}
	}
	return p.softSwitchesData[ioDataKeyboard]
}

func strobeKeyboardSoftSwitch(p *ioC0Page, _ bool, _ uint8) uint8 {
	result := p.softSwitchesData[ioDataKeyboard]
	p.softSwitchesData[ioDataKeyboard] &^= 1 << 7
	return result
}
