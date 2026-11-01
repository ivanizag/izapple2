package izapple2

import (
	"fmt"
	"slices"
)

/*
Apple II Super Serial Card, a 6551 ACIA with the firmware 341-0065.

See:

	"Super Serial Card Installation and Operating Manual", with the ROM
	listing disassembled at https://6502disassembly.com/a2-rom/SSC.html
	https://www.mrynet.com/FTP/os/AppleII/apple-super-serial-setup.txt
	https://websites.umich.edu/~archive/apple2/misc/hardware/card.ssc.txt

The ROM has the $C800 space, the last page is also on $Cn00.

The two banks of DIP switches are read at $C0n1 and $C0n2, a switch on
reads as 0. The bits, as the firmware listing names them:

	$C0n1: SW1-1 SW1-2 SW1-3 SW1-4   Z     Z     SW1-5 SW1-6
	$C0n2: SW2-1   Z   SW2-2   Z   SW2-3 SW2-4 SW2-5  CTS

SW1-1 to SW1-4 have the baud rate code of the 6551, SW1-5 and SW1-6 the
mode of the firmware. The meaning of SW2-1 to SW2-4 depends on the mode.
SW2-5 adds a line feed after the carriage returns. SW2-6, not readable,
wires the interrupt of the ACIA to the slot. SW1-7 and SW2-7 choose the CTS
pin, not emulated. The Z bits are read as 1 and CTS as 0, active.

The DCD and DSR inputs are low when the device is ready, the firmware sends
only then.
*/
type CardSuperSerial struct {
	cardBase
	port *serialPort
	sw1  uint8
	sw2  uint8
	irq  bool
}

var sscBaudRates = []string{"115200", "50", "75", "110", "135", "150", "300", "600",
	"1200", "1800", "2400", "3600", "4800", "7200", "9600", "19200"}

var sscModes = map[string]uint8{
	"communications": 0x00, // SW1-5 on, SW1-6 on
	"printer":        0x02, // SW1-5 off, SW1-6 on
	"sicp8":          0x01, // SW1-5 on, SW1-6 off
	"sicp8a":         0x03, // SW1-5 off, SW1-6 off
}

var sscParities = map[string]uint8{
	"none": 0x08, // SW2-3 off, SW2-4 on
	"odd":  0x04, // SW2-3 on, SW2-4 off
	"even": 0x0c, // SW2-3 off, SW2-4 off
}

var sscWidths = map[string]uint8{
	"40":  0x00, // SW2-3 on, SW2-4 on, with the video on
	"72":  0x04,
	"80":  0x08,
	"132": 0x0c,
}

const (
	sscSwitchesZ1 uint8 = 0x0c
	sscSwitchesZ2 uint8 = 0x50
)

func newCardSuperSerialBuilder() *cardBuilder {
	params := []paramSpec{
		{"baud", "Baud rate, SW1-1 to SW1-4: 50, 75, 110, 135, 150, 300, 600, 1200, 1800, 2400, 3600, 4800, 7200, 9600 or 19200", "9600"},
		{"mode", "Firmware mode, SW1-5 and SW1-6: communications, printer, sicp8 or sicp8a", "communications"},
		{"stopbits", "Stop bits, SW2-1: 1 or 2", "1"},
		{"databits", "Data bits in communications mode, SW2-2: 7 or 8", "8"},
		{"parity", "Parity in communications mode, SW2-3 and SW2-4: none, odd or even", "none"},
		{"crdelay", "Delay after a carriage return in printer mode, SW2-2", "false"},
		{"width", "Line width in printer mode, SW2-3 and SW2-4: 40, 72, 80 or 132", "80"},
		{"linefeed", "Add a line feed after the carriage returns, SW2-5", "false"},
		{"interrupts", "Wire the interrupts of the ACIA to the slot, SW2-6", "true"},
	}
	params = append(params, serialDeviceParams...)
	return &cardBuilder{
		name:          "Super Serial Card",
		description:   "Serial port with a 6551 ACIA connected to the console, a file, a TCP port or a modem to the internet",
		defaultParams: &params,
		buildFunc: func(params map[string]string) (Card, error) {
			var c CardSuperSerial
			err := c.setSwitches(params)
			if err != nil {
				return nil, err
			}
			err = c.loadRomFromResource("<internal>/Apple II Super Serial Card ROM 341-0065-A.bin", cardRomUpperEnd)
			if err != nil {
				return nil, err
			}
			c.irq = paramsGetBool(params, "interrupts")

			// The device is created last, a TCP port is not opened if
			// something else fails
			device, err := buildSerialDevice(params, paramsGetSlot(params))
			if err != nil {
				return nil, err
			}
			c.port = newSerialPort(device)
			return &c, nil
		},
	}
}

// setSwitches encodes the parameters as the DIP switches read by the firmware
func (c *CardSuperSerial) setSwitches(params map[string]string) error {
	baud := slices.Index(sscBaudRates, paramsGetString(params, "baud"))
	if baud < 1 {
		return fmt.Errorf("invalid baud rate '%s' for the super serial card", paramsGetString(params, "baud"))
	}
	mode, ok := sscModes[paramsGetString(params, "mode")]
	if !ok {
		return fmt.Errorf("invalid mode '%s' for the super serial card", paramsGetString(params, "mode"))
	}
	c.sw1 = uint8(baud)<<4 | sscSwitchesZ1 | mode

	c.sw2 = sscSwitchesZ2
	switch paramsGetString(params, "stopbits") {
	case "1":
	case "2":
		c.sw2 |= 0x80 // SW2-1 off
	default:
		return fmt.Errorf("invalid stop bits '%s' for the super serial card", paramsGetString(params, "stopbits"))
	}

	if paramsGetString(params, "mode") == "printer" {
		if !paramsGetBool(params, "crdelay") {
			c.sw2 |= 0x20 // SW2-2 off
		}
		width, ok := sscWidths[paramsGetString(params, "width")]
		if !ok {
			return fmt.Errorf("invalid width '%s' for the super serial card", paramsGetString(params, "width"))
		}
		c.sw2 |= width
	} else {
		switch paramsGetString(params, "databits") {
		case "8":
		case "7":
			c.sw2 |= 0x20 // SW2-2 off
		default:
			return fmt.Errorf("invalid data bits '%s' for the super serial card", paramsGetString(params, "databits"))
		}
		parity, ok := sscParities[paramsGetString(params, "parity")]
		if !ok {
			return fmt.Errorf("invalid parity '%s' for the super serial card", paramsGetString(params, "parity"))
		}
		c.sw2 |= parity
	}

	if !paramsGetBool(params, "linefeed") {
		c.sw2 |= 0x02 // SW2-5 off
	}
	return nil
}

func (c *CardSuperSerial) assign(a *Apple2, slot int) {
	c.addCardSoftSwitchR(1, func() uint8 {
		return c.sw1
	}, "SSCDIPSW1")
	c.addCardSoftSwitchR(2, func() uint8 {
		return c.sw2
	}, "SSCDIPSW2")

	c.port.assign(a, &c.cardBase, slot, c.irq, func() bool {
		return !c.port.isReady()
	})
	c.cardBase.assign(a, slot)
}

func (c *CardSuperSerial) reset() {
	c.port.reset()
}
