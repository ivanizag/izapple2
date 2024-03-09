package izapple2

import (
	"fmt"
)

type ioC0Page struct {
	softSwitchesR      [256]softSwitchR
	softSwitchesW      [256]softSwitchW
	softSwitchesRName  [256]string
	softSwitchesWName  [256]string
	softSwitchesData   [128]uint8
	keyboard           KeyboardProvider
	speaker            SpeakerProvider
	paddlesStrobeCycle uint64
	joysticks          JoysticksProvider
	mouse              MouseProvider
	apple2             *Apple2
	traceMask          uint16 // A bit for each 16 softswitches
	panicMask          uint16 // A bit for each 16 softswitches
	traceRegistrations bool
}

type softSwitchR func() uint8
type softSwitchW func(value uint8)

// SpeakerProvider provides a speaker implementation
type SpeakerProvider interface {
	// Click receives a speaker click. The argument is the CPU cycle when it is generated
	Click(cycle uint64)
}

// JoysticksProvider abstracts the joysticks
type JoysticksProvider interface {
	ReadButton(i int) bool
	ReadPaddle(i int) (uint8, bool)
}

// MouseProvider abstracts the mouse
type MouseProvider interface {
	ReadMouse() (x uint16, y uint16, pressed bool)
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
	return &io
}

func (p *ioC0Page) setTrace(trace bool) {
	if trace {
		p.traceMask = 0xffff
	} else {
		p.traceMask = 0x0000
	}
}

func (p *ioC0Page) traceSlot(slot int) {
	p.traceMask |= 1 << (8 + slot)
}

func (p *ioC0Page) panicNotImplementedSlot(slot int) {
	p.panicMask |= 1 << (8 + slot)
}

func (p *ioC0Page) setTraceRegistrations(traceRegistrations bool) {
	p.traceRegistrations = traceRegistrations
}

func (p *ioC0Page) setPanicNotImplemented(value bool) {
	if value {
		p.panicMask = 0xffff
	} else {
		p.panicMask = 0x0000
	}
}

func (p *ioC0Page) addSoftSwitchRW(address uint8, ss softSwitchR, name string) {
	p.addSoftSwitchR(address, ss, name)
	p.addSoftSwitchW(address, func(uint8) {
		ss()
	}, name)
}

func (p *ioC0Page) addSoftSwitchR(address uint8, ss softSwitchR, name string) {
	if p.traceRegistrations {
		fmt.Printf("Softswitch registered in $c0%02x for reads as %s\n", address, name)
	}
	p.softSwitchesR[address] = ss
	p.softSwitchesRName[address] = name
}

func (p *ioC0Page) addSoftSwitchW(address uint8, ss softSwitchW, name string) {
	if p.traceRegistrations {
		fmt.Printf("Softswitch registered in $c0%02x for writes as %s\n", address, name)
	}
	p.softSwitchesW[address] = ss
	p.softSwitchesWName[address] = name
}

func (p *ioC0Page) isSoftSwitchActive(ioFlag uint8) bool {
	return (p.softSwitchesData[ioFlag] & ssOn) == ssOn
}

func (p *ioC0Page) setKeyboardProvider(kb KeyboardProvider) {
	p.keyboard = kb
}

func (p *ioC0Page) setSpeakerProvider(s SpeakerProvider) {
	p.speaker = s
}

func (p *ioC0Page) setJoysticksProvider(j JoysticksProvider) {
	p.joysticks = j
}

func (p *ioC0Page) setMouseProvider(m MouseProvider) {
	p.mouse = m
}

func (p *ioC0Page) isTraced(address uint16) bool {
	ss := address & 0xff
	return ss != 0xc000 && // Do not trace the spammy keyboard softswitch
		(p.traceMask&(1<<(ss>>4))) != 0
}

func (p *ioC0Page) isPanicNotImplemented(address uint16) bool {
	ss := address & 0xff
	return ss != 0xc068 && // Ignore known IIGS softswitch
		(p.panicMask&(1<<(ss>>4))) != 0
}

func (p *ioC0Page) peek(address uint16) uint8 {
	pageAddress := uint8(address)
	ss := p.softSwitchesR[pageAddress]
	if ss == nil {
		if p.isTraced(address) {
			fmt.Printf("Unknown softswitch on read to $%04x\n", address)
		}
		if p.isPanicNotImplemented(address) {
			panic(fmt.Sprintf("Unknown softswitch on read to $%04x", address))
		}
		return 0
	}
	value := ss()
	if p.isTraced(address) {
		name := p.softSwitchesRName[pageAddress]
		fmt.Printf("Softswitch peek on $%04x %v: $%02x\n", address, name, value)
	}
	return value
}

func (p *ioC0Page) poke(address uint16, value uint8) {
	pageAddress := uint8(address)
	ss := p.softSwitchesW[pageAddress]
	if ss == nil {
		if p.isTraced(address) {
			fmt.Printf("Unknown softswitch on write $%02x to $%04x\n", value, address)
		}
		if p.isPanicNotImplemented(address) {
			panic(fmt.Sprintf("Unknown softswitch on write to $%04x", address))
		}
		return
	}
	if p.isTraced(address) {
		name := p.softSwitchesWName[pageAddress]
		fmt.Printf("Softswitch poke on $%04x %v with $%02x\n", address, name, value)
	}
	ss(value)
}

func ssFromBool(value bool) uint8 {
	if value {
		return ssOn
	}
	return ssOff
}
