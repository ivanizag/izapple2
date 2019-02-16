package main

import "fmt"

type ioC0Page struct {
	ioFlags uint64
	data    [1]uint8
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

func (p *ioC0Page) peek(address uint8) uint8 {
	//fmt.Printf("Peek on $C0%02x ", address)
	return p.access(address, false, 0)
}

func (p *ioC0Page) poke(address uint8, value uint8) {
	//fmt.Printf("Poke on $C0%02x with %02x ", address, value)
	p.access(address, true, value)
}

func (p *ioC0Page) getData() *[256]uint8 {
	var blankPage [256]uint8
	return &blankPage
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
		case 0x00: // keyboard
			// TODO: keyboard suppport
			return 0 //128 + 'A'
		case 0x10: // strobe
			result := p.data[ioDataKeyboard]
			p.data[ioDataKeyboard] &^= 1 << 7
			return result
		case 0x30: // spkr
			// TODO: Support sound
		default:
			panic(fmt.Sprintf("Unknown softswitch 0xC0%02x", address))
		}
	}
	return 0
}
