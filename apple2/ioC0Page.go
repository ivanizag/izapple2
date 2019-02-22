package apple2

import (
	"fmt"
)

type ioC0Page struct {
	ioFlags      uint64
	data         [1]uint8
	keyboard     keyboardProvider
	addressSpace *addressSpace
}

type keyboardProvider interface {
	getKey() (key uint8, ok bool)
}

// See https://www.kreativekorp.com/miscpages/a2info/iomemory.shtml
// See https://stason.org/TULARC/pc/apple2/programmer/004-I-d-like-to-do-some-serious-Apple-II-programming-Whe.html

const (
	ioFlagNone         uint8 = 0
	ioFlagGraphics     uint8 = 3
	ioFlagMixed        uint8 = 8
	ioFlagSecondPage   uint8 = 1
	ioFlagHiRes        uint8 = 2
	ioFlagAnnunciator0 uint8 = 4
	ioFlagAnnunciator1 uint8 = 5
	ioFlagAnnunciator2 uint8 = 6
	ioFlagAnnunciator3 uint8 = 7
)

const (
	ioDataKeyboard uint8 = 0
)

type softSwitch struct {
	ioFlag      uint8
	value       bool
	onWriteOnly bool
}

var softSwitches = [256]softSwitch{
	0x50: softSwitch{ioFlagGraphics, false, false},
	0x51: softSwitch{ioFlagGraphics, true, false},
	0x52: softSwitch{ioFlagMixed, false, false},
	0x53: softSwitch{ioFlagMixed, true, false},
	0x54: softSwitch{ioFlagSecondPage, false, false},
	0x55: softSwitch{ioFlagSecondPage, true, false},
	0x56: softSwitch{ioFlagHiRes, false, false},
	0x57: softSwitch{ioFlagHiRes, true, false},
	0x58: softSwitch{ioFlagAnnunciator0, false, false},
	0x59: softSwitch{ioFlagAnnunciator0, true, false},
	0x5a: softSwitch{ioFlagAnnunciator1, false, false},
	0x5b: softSwitch{ioFlagAnnunciator1, true, false},
	0x5c: softSwitch{ioFlagAnnunciator2, false, false},
	0x5d: softSwitch{ioFlagAnnunciator2, true, false},
	0x5e: softSwitch{ioFlagAnnunciator3, false, false},
	0x5f: softSwitch{ioFlagAnnunciator3, true, false},
}

func (p *ioC0Page) Peek(address uint8) uint8 {
	//fmt.Printf("Peek on $C0%02x ", address)
	return p.access(address, false, 0)
}

func (p *ioC0Page) Poke(address uint8, value uint8) {
	//fmt.Printf("Poke on $C0%02x with %02x ", address, value)
	p.access(address, true, value)
}

func (p *ioC0Page) access(address uint8, isWrite bool, value uint8) uint8 {

	ss := softSwitches[address]
	if ss.ioFlag != ioFlagNone {
		if !isWrite || !!ss.onWriteOnly {
			if ss.value {
				p.ioFlags |= 1 << ss.ioFlag
			} else {
				p.ioFlags &^= 1 << ss.ioFlag
			}
		}
	} else {
		switch address {
		case 0x00: // keyboard (Is this the full range 0x0?)
			return p.getKey()
		case 0x10: // strobe (Is this the full range 0x1?)
			return p.strobeKeyboard()
		case 0x30: // spkr
			// TODO: Support sound
		default:
			panic(fmt.Sprintf("Unknown softswitch 0xC0%02x", address))
		}
	}
	return 0
}

func (p *ioC0Page) setKeyboardProvider(kb keyboardProvider) {
	p.keyboard = kb
}

func (p *ioC0Page) getKey() uint8 {
	strobed := (p.data[ioDataKeyboard] & (1 << 7)) == 0
	if strobed && p.keyboard != nil {
		if key, ok := p.keyboard.getKey(); ok {
			p.data[ioDataKeyboard] = key + (1 << 7)
		}
	}
	return p.data[ioDataKeyboard]
}

func (p *ioC0Page) strobeKeyboard() uint8 {
	result := p.data[ioDataKeyboard]
	p.data[ioDataKeyboard] &^= 1 << 7
	return result
}
