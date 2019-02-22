package apple2

import (
	"bufio"
	"os"
)

// See https://i.stack.imgur.com/yn21s.gif

type addressSpace struct {
	activeMemory *pagedMemory
	physicalRAM  [256]ramPage // up to 64 Kb
	physicalROM  [48]romPage  // up to 12 Kb
	ioPage       ioC0Page
	textPages1   *textPages
	activeSlow   int // Slot that has the addressing 0xc800 to 0ccfff
}

const (
	ioAreaMask  uint16 = 0xFF80
	ioAreaValue uint16 = 0xC000
	ioC8Off     uint16 = 0xCFFF
)

// Peek returns the data on the given address
func (a *addressSpace) Peek(address uint16) uint8 {
	if address == ioC8Off {
		a.resetSlotRoms()
	}
	if (address & ioAreaMask) == ioAreaValue {
		return a.ioPage.Peek(uint8(address))
	}
	return a.activeMemory.Peek(address)
}

// Poke sets the data at the given address
func (a *addressSpace) Poke(address uint16, value uint8) {
	if address == ioC8Off {
		a.resetSlotRoms()
	}
	if (address & ioAreaMask) == ioAreaValue {
		a.ioPage.Poke(uint8(address), value)
	}
	a.activeMemory.Poke(address, value)
}

func (a *addressSpace) resetSlotRoms() {
	// TODO
}

func newAddressSpace() *addressSpace {
	var a addressSpace

	var m pagedMemory
	a.activeMemory = &m

	// Assign RAM from 0x0000 to 0xbfff, 48kb
	for i := 0; i <= 0xbf; i++ {
		m.SetPage(uint8(i), &(a.physicalRAM[i]))
	}

	// Assign ROM from 0xd000 to 0xfff, 12 kb. The ROM is empty
	for i := 0xd0; i <= 0xff; i++ {
		m.SetPage(uint8(i), &(a.physicalROM[i-0xd0]))
	}

	// Set the 0xc000 to 0xcfff as unasigned, 4kb. It wil be taken by slot cards.
	for i := uint8(0xc0); i <= 0xcf; i++ {
		var p unassignedPage
		p.page = i
		m.SetPage(i, &p)
	}

	// Replace RAM in the TEXT1 area.
	// TODO: treat as normal ram. Add is dirty in all RAM pages
	var t textPages
	a.textPages1 = &t
	for i := 0; i < 4; i++ {
		m.SetPage(uint8(4+i), &(t.pages[i]))
	}

	return &a
}

// LoadRom loads a binary file to the top of the memory.
func (a *addressSpace) loadRom(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	stats, statsErr := f.Stat()
	if statsErr != nil {
		panic(err)
	}

	size := stats.Size()
	if size != 12288 {
		panic("Rom size not supported")
	}
	bytes := make([]byte, size)

	buf := bufio.NewReader(f)
	buf.Read(bytes)

	for i, v := range bytes {
		a.physicalROM[i>>8].burn(uint8(i), uint8(v))
	}
}
