package izapple2

/*
The two serial ports of the Apple IIc, a 6551 ACIA each. Their firmware is
on the ROM of the IIc, as if the ports were Super Serial Cards on slots 1
and 2, and the ACIAs are on the I/O space of those slots.

See:

	"Apple IIc Technical Reference Manual", Chapter 11, Serial I/O, and
	Appendix E, Interrupts.

The ports have no switches. Their inputs are wired as:

	DCD: pin 5 of the connector, DSR from the device. The firmware sends
	     when it is low.
	DSR: port 1, the external interrupt of the disk drive connector, pulled
	     up and high. Port 2, the keyboard strobe, high when a key is
	     pressed. The firmware uses the interrupt of its changes to keep
	     the keys in a buffer.
	CTS: tied to ground, the transmitter is always enabled.

The keyboard of the emulator gives a key when $C000 is read. While the ACIA
of port 2 has DTR on, the keys are taken as they come, for the strobe to be
high as on a real keyboard.
*/
type CardSerialIIc struct {
	cardBase
	port    *serialPort
	nextKey uint64 // Cycle to look again for a key
}

const serialIIcKeyCheck = 1000 // Cycles between looks for a key

func newCardSerialIIcBuilder() *cardBuilder {
	params := append([]paramSpec{}, serialDeviceParams...)
	return &cardBuilder{
		name:          "Apple IIc serial ports",
		description:   "Serial port of the Apple IIc, connected to the console, a file or a TCP port",
		defaultParams: &params,
		requiresIIc:   true,
		hide:          true,
		buildFunc: func(params map[string]string) (Card, error) {
			// checkApple2cConfiguration keeps them on slots 1 and 2
			device, err := buildSerialDevice(params, paramsGetSlot(params))
			if err != nil {
				return nil, err
			}
			var c CardSerialIIc
			c.port = newSerialPort(device)
			return &c, nil
		},
	}
}

func (c *CardSerialIIc) assign(a *Apple2, slot int) {
	dsrHigh := func() bool {
		return true // The pull up of the external interrupt
	}
	if slot == 2 {
		dsrHigh = c.keyboardStrobe
	}
	c.port.assign(a, &c.cardBase, slot, true, dsrHigh)
	c.cardBase.assign(a, slot)
}

func (c *CardSerialIIc) reset() {
	c.port.reset()
}

// keyboardStrobe returns whether a key is pressed, taking it from the
// keyboard when the ACIA can interrupt for it
func (c *CardSerialIIc) keyboardStrobe() bool {
	io := c.a.io
	strobe := io.softSwitchesData[ioDataKeyboard]&0x80 != 0
	if !strobe && io.keyboard != nil && c.port.acia.IsDTR() && c.a.GetCycles() >= c.nextKey {
		c.nextKey = c.a.GetCycles() + serialIIcKeyCheck
		if key, ok := io.keyboard.GetKey(true); ok {
			io.softSwitchesData[ioDataKeyboard] = key | 0x80
			strobe = true
		}
	}
	return strobe
}
