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

type keyboardProvider interface {
	getKey() (key uint8, ok bool)
}

// See https://www.kreativekorp.com/miscpages/a2info/iomemory.shtml
// See https://stason.org/TULARC/pc/apple2/programmer/004-I-d-like-to-do-some-serious-Apple-II-programming-Whe.html

const (
	ssOn  uint8 = 0x80
	ssOff uint8 = 0x00
)

func newIoC0Page(mmu *memoryManager) *ioC0Page {
	var io ioC0Page
	io.mmu = mmu

	addApple2SoftSwitches(&io)
	if mmu.isApple2e {
		addApple2ESoftSwitches(&io)
	}

	return &io
}

func (p *ioC0Page) isSoftSwitchExtActive(ioFlag uint8) bool {
	return (p.softSwitchesData[ioFlag] & ssOn) == ssOn
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
	// The second half of the pages is reserved for slots
	if address >= 0x90 {
		// TODO reserved slots data
		return 0
	}

	ss := p.softSwitches[address]
	if ss == nil {
		panic(fmt.Sprintf("Unknown softswitch 0xC0%02x", address))
	}

	return ss(p, isWrite, value)
}
