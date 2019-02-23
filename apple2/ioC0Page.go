package apple2

import (
	"fmt"
)

type ioC0Page struct {
	softSwitches     [128]softSwitch
	softSwitchesData [128]uint8
	keyboard         keyboardProvider
	mmu              *memoryManager
}

type softSwitch func(io *ioC0Page, isWrite bool, value uint8) uint8

// TODO: change interface to func
type keyboardProvider interface {
	getKey() (key uint8, ok bool)
}

// See https://www.kreativekorp.com/miscpages/a2info/iomemory.shtml
// See https://stason.org/TULARC/pc/apple2/programmer/004-I-d-like-to-do-some-serious-Apple-II-programming-Whe.html

const (
	ioDataKeyboard     uint8 = 0x10
	ioFlagGraphics     uint8 = 0x50
	ioFlagMixed        uint8 = 0x52
	ioFlagSecondPage   uint8 = 0x54
	ioFlagHiRes        uint8 = 0x56
	ioFlagAnnunciator0 uint8 = 0x58
	ioFlagAnnunciator1 uint8 = 0x5a
	ioFlagAnnunciator2 uint8 = 0x5c
	ioFlagAnnunciator3 uint8 = 0x5e
)

func (p *ioC0Page) isSoftSwitchExtActive(ioFlag uint8) bool {
	return (p.softSwitchesData[ioFlag] & 0x08) == 0x80
}

func newIoC0Page(mmu *memoryManager) *ioC0Page {
	var p ioC0Page
	p.mmu = mmu
	ss := &p.softSwitches

	ss[0x00] = getKeySoftSwitch         // Keyboard
	ss[0x10] = strobeKeyboardSoftSwitch // Keyboard Strobe
	ss[0x30] = notImplementedSoftSwitch // Speaker

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

	return &p
}

func (p *ioC0Page) setKeyboardProvider(kb keyboardProvider) {
	p.keyboard = kb
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
	// The second hals of the pages is reserved for slots
	if address >= 0x80 {
		// TODO reserved slots data
		return 0
	}

	ss := p.softSwitches[address]
	if ss == nil {
		panic(fmt.Sprintf("Unknown softswitch 0xC0%02x", address))
	}

	return ss(p, isWrite, value)
}

func getSoftSwitch(ioFlag uint8, isSet bool) softSwitch {
	return func(io *ioC0Page, isWrite bool, value uint8) uint8 {
		if isSet {
			io.softSwitchesData[ioFlag] = 0x80
		} else {
			io.softSwitchesData[ioFlag] = 0
		}
		return 0
	}
}

func notImplementedSoftSwitch(*ioC0Page, bool, uint8) uint8 {
	return 0
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
