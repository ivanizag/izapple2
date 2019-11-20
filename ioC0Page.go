package apple2

import (
	"encoding/binary"
	"fmt"
	"io"
)

type ioC0Page struct {
	softSwitchesR       [256]softSwitchR
	softSwitchesW       [256]softSwitchW
	softSwitchesRName   [256]string
	softSwitchesWName   [256]string
	softSwitchesData    [128]uint8
	keyboard            KeyboardProvider
	speaker             SpeakerProvider
	paddlesStrobeCycle  uint64
	joysticks           JoysticksProvider
	apple2              *Apple2
	trace               bool
	panicNotImplemented bool
}

type softSwitchR func(io *ioC0Page) uint8
type softSwitchW func(io *ioC0Page, value uint8)

// KeyboardProvider provides a keyboard implementation
type KeyboardProvider interface {
	GetKey(strobe bool) (key uint8, ok bool)
}

// SpeakerProvider provides a speaker implementation
type SpeakerProvider interface {
	// Click receives a speaker click. The argument is the CPU cycle when it is generated
	Click(cycle uint64)
}

// JoysticksProvider declares de the joysticks
type JoysticksProvider interface {
	ReadButton(i int) bool
	ReadPaddle(i int) (uint8, bool)
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
	p.trace = trace
}

func (p *ioC0Page) setPanicNotImplemented(value bool) {
	p.panicNotImplemented = value
}

func (p *ioC0Page) save(w io.Writer) error {
	return binary.Write(w, binary.BigEndian, p.softSwitchesData)
}

func (p *ioC0Page) load(r io.Reader) error {
	return binary.Read(r, binary.BigEndian, &p.softSwitchesData)
}

func (p *ioC0Page) addSoftSwitchRW(address uint8, ss softSwitchR, name string) {
	p.addSoftSwitchR(address, ss, name)
	p.addSoftSwitchW(address, func(p *ioC0Page, _ uint8) {
		ss(p)
	}, name)
}

func (p *ioC0Page) addSoftSwitchR(address uint8, ss softSwitchR, name string) {
	//if p.softSwitchesR[address] != nil {
	//	fmt.Printf("Addresss 0x0c%02x is already assigned for read\n", address)
	//}
	p.softSwitchesR[address] = ss
	p.softSwitchesRName[address] = name
}

func (p *ioC0Page) addSoftSwitchW(address uint8, ss softSwitchW, name string) {
	//if p.softSwitchesW[address] != nil {
	//	fmt.Printf("Addresss 0x0c%02x is already assigned for write\n", address)
	//}
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

func (p *ioC0Page) peek(address uint16) uint8 {
	pageAddress := uint8(address)
	ss := p.softSwitchesR[pageAddress]
	if ss == nil {
		if p.trace {
			fmt.Printf("Unknown softswitch on read to $%04x\n", address)
		}
		if p.panicNotImplemented {
			panic(fmt.Sprintf("Unknown softswitch on read to $%04x", address))
		}
		return 0
	}
	value := ss(p)
	if p.trace && address != 0xc000 {
		name := p.softSwitchesRName[pageAddress]
		fmt.Printf("Softswitch peek on $%04x %v: $%02x\n", address, name, value)
	}
	return value
}

func (p *ioC0Page) poke(address uint16, value uint8) {
	pageAddress := uint8(address)
	ss := p.softSwitchesW[pageAddress]
	if ss == nil {
		if p.trace {
			fmt.Printf("Unknown softswitch on write to $%04x\n", address)
		}
		if p.panicNotImplemented {
			panic(fmt.Sprintf("Unknown softswitch on write to $%04x", address))
		}
		return
	}
	if p.trace && address != 0xc000 {
		name := p.softSwitchesWName[pageAddress]
		fmt.Printf("Softswitch poke on $%04x %v with $%02x\n", address, name, value)
	}
	ss(p, value)
}

func ssFromBool(value bool) uint8 {
	if value {
		return ssOn
	}
	return ssOff
}
