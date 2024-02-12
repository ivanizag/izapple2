package izapple2

import (
	"fmt"
)

/*
Mouse card implementation. Does not emulate a real card, only the behaviour. Idea taken
  from aiie (https://hackaday.io/project/19925-aiie-an-embedded-apple-e-emulator/log/188017-entry-23-here-mousie-mousie-mousie)

See:
	https://www.apple.asimov.net/documentation/hardware/io/AppleMouse%20II%20User%27s%20Manual.pdf
	https://mirrors.apple2.org.za/Apple%20II%20Documentation%20Project/Interface%20Cards/Digitizers/Apple%20Mouse%20Interface%20Card/Documentation/Apple%20II%20Mouse%20Technical%20Notes.pdf
	http://www.1000bit.it/support/manuali/apple/technotes/mous/tn.mous.2.html

	The management of IN# and PR# is copied from cardInOut

	Not compatible with A2OSX that needs interrupts on VBL

*/

// CardMouse represents a SmartPort card
type CardMouse struct {
	cardBase

	lastX, lastY uint16
	lastPressed  bool

	minX, minY, maxX, maxY uint16
	mode                   uint8

	response string
	iOut     int
	iIn      int
}

func newCardMouseBuilder() *cardBuilder {
	return &cardBuilder{
		name:        "Mouse Card",
		description: "Mouse card implementation, does not emulate a real card, only the firmware behaviour",
		buildFunc: func(params map[string]string) (Card, error) {
			return &CardMouse{
				maxX: 0x3ff,
				maxY: 0x3ff,
			}, nil
		},
	}
}

const (
	mouseXLo    = uint16(0x478)
	mouseYLo    = uint16(0x4f8)
	mouseXHi    = uint16(0x578)
	mouseYHi    = uint16(0x5f8)
	mouseStatus = uint16(0x778)
	mouseMode   = uint16(0x7f8)
)

const (
	mouseModeEnabled          = uint8(1)
	mouseModeIntMoveEnabled   = uint8(2)
	mouseModeIntButtonEnabled = uint8(4)
	mouseModeIntVBlankEnabled = uint8(8)
)

func (c *CardMouse) set(field uint16, value uint8) {
	// Update the card screen-holes
	c.a.mmu.Poke(field+uint16(c.slot), value)
}

func (c *CardMouse) get(field uint16) uint8 {
	// Read from the card screen-holes
	return c.a.mmu.Peek(field /*+ uint16(c.slot)*/)
}

func (c *CardMouse) setMode(mode uint8) {
	c.mode = mode
	enabled := mode&mouseModeEnabled == 1
	moveInts := mode&mouseModeIntMoveEnabled == 1
	buttonInts := mode&mouseModeIntButtonEnabled == 1
	vBlankInts := mode&mouseModeIntVBlankEnabled == 1

	c.tracef("Mode set to 0x%02x. Enabled %v. Interrups: move=%v, button=%v, vblank=%v.\n",
		mode, enabled, moveInts, buttonInts, vBlankInts)

	if moveInts || buttonInts || vBlankInts {
		panic("Mouse interrupts not implemented")
	}
}

func (c *CardMouse) checkFromFirmware() {
	pc, _ := c.a.cpu.GetPCAndSP()
	if (pc >> 8) != 0xc0+uint16(c.slot) {
		c.tracef("Softswitch access from outside the firmware. It will not work.\n")
	}
}

func (c *CardMouse) readMouse() (uint16, uint16, bool) {
	x, y, pressed := c.a.io.mouse.ReadMouse()
	xTrans := uint16(uint64(c.maxX-c.minX) * uint64(x) / 65536)
	yTrans := uint16(uint64(c.maxY-c.minY) * uint64(y) / 65536)
	return xTrans, yTrans, pressed
}

func (c *CardMouse) assign(a *Apple2, slot int) {
	c.addCardSoftSwitchR(0, func() uint8 {
		c.checkFromFirmware()
		if c.iOut == 0 {
			// Create a new response
			x, y, pressed := c.readMouse()

			button := 1
			if !pressed {
				button += 2
			}
			if !c.lastPressed {
				button++
			}

			keyboard := "+"
			strobed := (c.a.io.softSwitchesData[ioDataKeyboard] & (1 << 7)) == 0
			if !strobed {
				keyboard = "-"
			}

			c.response = fmt.Sprintf("%v,%v,%v%v\r", x, y, keyboard, button)
		}
		value := uint8(c.response[c.iOut])
		c.iOut++
		if c.iOut == len(c.response) {
			c.iOut = 0
		}

		value += 0x80
		c.tracef("IN#%v -> %02x.\n", slot, value)
		return value
	}, "MOUSEOUT")

	c.addCardSoftSwitchW(1, func(value uint8) {
		c.checkFromFirmware()
		c.tracef("PR#%v <- %02x\n", slot, value)
		if c.iIn == 0 {
			// We care only about the first byte
			c.setMode(value & 0x0f)
		}
		c.iIn++
		if value == 13 {
			c.iIn = 0 // Ready for the next command
		}
	}, "MOUSEIN")

	c.addCardSoftSwitchW(2, func(value uint8) {
		c.checkFromFirmware()
		c.tracef("SetMouse(0x%02v)\n", value)
		c.setMode(value & 0x0f)
	}, "SETMOUSE")

	c.addCardSoftSwitchW(3, func(value uint8) {
		c.checkFromFirmware()
		c.tracef("ServeMouse() NOT IMPLEMENTED\n")
		panic("Mouse interrupts not implemented")
	}, "SERVEMOUSE")

	c.addCardSoftSwitchW(4, func(value uint8) {
		c.checkFromFirmware()
		if c.mode&mouseModeEnabled == 1 {
			x, y, pressed := c.readMouse()

			status := uint8(0)
			if pressed {
				status |= 1 << 7
			}
			if c.lastPressed {
				status |= 1 << 6
			}
			if (x != c.lastX) || (y != c.lastY) {
				status |= 1 << 5
			}

			c.set(mouseXHi, uint8(x>>8))
			c.set(mouseYHi, uint8(y>>8))
			c.set(mouseXLo, uint8(x))
			c.set(mouseYLo, uint8(y))
			c.set(mouseStatus, status)
			c.set(mouseMode, c.mode)
			if (status&(1<<5) != 0) || (pressed != c.lastPressed) {
				c.tracef("ReadMouse(): x: %v, y: %v, pressed: %v\n",
					x, y, pressed)
			}

			c.lastX = x
			c.lastY = y
			c.lastPressed = pressed
		}
	}, "READMOUSE")

	c.addCardSoftSwitchW(5, func(value uint8) {
		c.checkFromFirmware()
		c.tracef("ClearMouse() NOT IMPLEMENTED\n")
		c.set(mouseXHi, 0)
		c.set(mouseYHi, 0)
		c.set(mouseXLo, 0)
		c.set(mouseYLo, 0)
	}, "CLEARMOUSE")
	c.addCardSoftSwitchW(6, func(value uint8) {
		c.checkFromFirmware()
		c.tracef("PosMouse() NOT IMPLEMENTED\n")
	}, "POSMOUSE")

	c.addCardSoftSwitchW(7, func(value uint8) {
		c.checkFromFirmware()
		c.tracef("ClampMouse(%v)\n", value)

		if value == 0 {
			c.minX = uint16(c.get(mouseXLo)) + uint16(c.get(mouseXHi))<<8
			c.maxX = uint16(c.get(mouseYLo)) + uint16(c.get(mouseYHi))<<8
		} else if value == 1 {
			c.minY = uint16(c.get(mouseXLo)) + uint16(c.get(mouseXHi))<<8
			c.maxY = uint16(c.get(mouseYLo)) + uint16(c.get(mouseYHi))<<8
		}

		c.tracef("Current bounds: X[%v-%v], Y[%v-%v],\n", c.minX, c.maxX, c.minY, c.maxY)
	}, "CLAMPMOUSE")

	c.addCardSoftSwitchW(8, func(value uint8) {
		c.checkFromFirmware()
		c.tracef("HomeMouse() NOT IMPLEMENTED\n")
	}, "HOMEMOUSE")

	c.addCardSoftSwitchW(9, func(value uint8) {
		c.checkFromFirmware()
		c.tracef("InitMouse()\n")
		c.minX = 0
		c.minY = 0
		c.maxX = 0x3ff
		c.maxY = 0x3ff
		c.mode = 0
	}, "INITMOUSE")

	c.addCardSoftSwitchW(0xc, func(value uint8) {
		c.checkFromFirmware()
		// See http://www.1000bit.it/support/manuali/apple/technotes/mous/tn.mous.2.html
		c.tracef("TimeData(%v) NOT IMPLEMENTED\n", value)
	}, "TIMEDATEMOUSE")

	data := buildBaseInOutRom(slot)
	c.romCsxx = newMemoryRangeROM(0xC200, data[:], "Mouse card")

	// Identification as a mouse card
	// From Technical Note Misc #8, "Pascal 1.1 Firmware Protocol ID Bytes":
	data[0x05] = 0x38
	data[0x07] = 0x18
	data[0x0b] = 0x01
	data[0x0c] = 0x20
	// From "AppleMouse // User's Manual", Appendix B:
	//data[0x0c] = 0x20
	data[0xfb] = 0xd6

	// Set 8 entrypoints to sofstwitches 2 to 1f
	for i := uint8(0); i < 14; i++ {
		base := 0x60 + 0x05*i
		data[0x12+i] = base
		data[base+0] = 0x8D // STA $C0x2
		data[base+1] = 0x82 + i + uint8(slot<<4)
		data[base+2] = 0xC0
		data[base+3] = 0x18 // CLC ;no error
		data[base+4] = 0x60 // RTS
	}

	c.cardBase.assign(a, slot)
}
