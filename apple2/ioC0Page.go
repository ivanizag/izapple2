package apple2

import (
	"fmt"
)

type ioC0Page struct {
	softSwitchesR    [256]softSwitchR
	softSwitchesW    [256]softSwitchW
	softSwitchesData [128]uint8
	keyboard         KeyboardProvider
	apple2           *Apple2
}

type softSwitchR func(io *ioC0Page) uint8
type softSwitchW func(io *ioC0Page, value uint8)

// KeyboardProvider declares the keyboard implementation requirements
type KeyboardProvider interface {
	GetKey(strobe bool) (key uint8, ok bool)
}

// See https://www.kreativekorp.com/miscpages/a2info/iomemory.shtml
// See https://stason.org/TULARC/pc/apple2/programmer/004-I-d-like-to-do-some-serious-Apple-II-programming-Whe.html

const (
	ssOn  uint8 = 0x80
	ssOff uint8 = 0x00
)

func newIoC0Page(a *Apple2) *ioC0Page {
	var io ioC0Page
	io.apple2 = a

	addApple2SoftSwitches(&io)
	if a.isApple2e {
		addApple2ESoftSwitches(&io)
	}

	return &io
}

func (p *ioC0Page) addSoftSwitchRW(address uint8, ss softSwitchR) {
	p.addSoftSwitchR(address, ss)
	p.addSoftSwitchW(address, func(p *ioC0Page, _ uint8) {
		ss(p)
	})
}

func (p *ioC0Page) addSoftSwitchR(address uint8, ss softSwitchR) {
	if p.softSwitchesR[address] != nil {
		fmt.Printf("Addresss 0x0c%02x is already assigned for read", address)
	}
	p.softSwitchesR[address] = ss
}

func (p *ioC0Page) addSoftSwitchW(address uint8, ss softSwitchW) {
	if p.softSwitchesW[address] != nil {
		fmt.Printf("Addresss 0x0c%02x is already assigned for write", address)
	}
	p.softSwitchesW[address] = ss
}

func (p *ioC0Page) isSoftSwitchActive(ioFlag uint8) bool {
	return (p.softSwitchesData[ioFlag] & ssOn) == ssOn
}

func (p *ioC0Page) setKeyboardProvider(kb KeyboardProvider) {
	p.keyboard = kb
}

func (p *ioC0Page) Peek(address uint8) uint8 {
	//fmt.Printf("Peek on $C0%02x ", address)
	ss := p.softSwitchesR[address]
	if ss == nil {
		if p.apple2.panicSS {
			panic(fmt.Sprintf("Unknown softswitch on read to 0xC0%02x", address))
		}
		return 0
	}
	return ss(p)
}

func (p *ioC0Page) internalPeek(address uint8) uint8 {
	return 0
}

func (p *ioC0Page) Poke(address uint8, value uint8) {
	//fmt.Printf("Poke on $C0%02x with %02x ", address, value)
	ss := p.softSwitchesW[address]
	if ss == nil {
		if p.apple2.panicSS {
			panic(fmt.Sprintf("Unknown softswitch on write to 0xC0%02x", address))
		}
		return
	}
	ss(p, value)
}
