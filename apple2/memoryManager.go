package apple2

import (
	"io/ioutil"
)

// See https://fabiensanglard.net/fd_proxy/prince_of_persia/Inside%20the%20Apple%20IIe.pdf
// See https://i.stack.imgur.com/yn21s.gif

type memoryManager struct {
	apple2 *Apple2
	// Map of assigned pages
	activeMemoryRead  [256]memoryHandler
	activeMemoryWrite [256]memoryHandler

	// Pages prepared to be paged in and out
	physicalMainRAM *memoryRange // 0x0000 to 0xbfff, Up to 48 Kb
	physicalROM     *memoryRange // 0xd000 to 0xffff, 12 Kb
	physicalROMe    *memoryRange // 0xc000 to 0xcfff, Zero or 4bk in the Apple2e
}

type memoryHandler interface {
	peek(uint16) uint8
	poke(uint16, uint8)
}

const (
	ioC8Off uint16 = 0xCFFF
)

// Peek returns the data on the given address
func (mmu *memoryManager) Peek(address uint16) uint8 {
	if address == ioC8Off {
		mmu.resetSlotExpansionRoms()
	}

	hi := uint8(address >> 8)
	mh := mmu.activeMemoryRead[hi]
	if mh == nil {
		return 0xf4 // Or some random number
	}
	return mh.peek(address)
}

// Poke sets the data at the given address
func (mmu *memoryManager) Poke(address uint16, value uint8) {
	if address == ioC8Off {
		mmu.resetSlotExpansionRoms()
	}
	hi := uint8(address >> 8)
	mh := mmu.activeMemoryWrite[hi]
	if mh == nil {
		return
	}
	mh.poke(address, value)
}

func (mmu *memoryManager) setPages(begin uint8, end uint8, mh memoryHandler) {
	mmu.setPagesRead(begin, end, mh)
	mmu.setPagesWrite(begin, end, mh)
}

func (mmu *memoryManager) setPagesRead(begin uint8, end uint8, mh memoryHandler) {
	i := begin
	for {
		mmu.activeMemoryRead[i] = mh
		if i == end {
			break
		}
		i++
	}
}

func (mmu *memoryManager) setPagesWrite(begin uint8, end uint8, mh memoryHandler) {
	i := begin
	for {
		mmu.activeMemoryWrite[i] = mh
		if i == end {
			break
		}
		i++
	}
}

// When 0xcfff is accessed the card expansion rom is unassigned
func (mmu *memoryManager) resetSlotExpansionRoms() {
	if mmu.apple2.io.isSoftSwitchActive(ioFlagIntCxRom) {
		// Ignore if the Apple2 shadow ROM is active
		return
	}
	mmu.setPagesRead(0xc8, 0xcf, nil)
}

func newMemoryManager(a *Apple2) *memoryManager {
	var mmu memoryManager
	mmu.apple2 = a

	// Assign RAM from 0x0000 to 0xbfff, 48kb
	ram := make([]uint8, 0xc000)
	mmu.physicalMainRAM = newMemoryRange(0, ram)
	mmu.setPages(0x00, 0xc0, mmu.physicalMainRAM)

	return &mmu
}

func (mmu *memoryManager) loadRom(filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	size := len(data)
	if size != apple2RomSize && size != apple2eRomSize {
		panic("Rom size not supported")
	}

	a := mmu.apple2
	romStart := 0
	if size == apple2eRomSize {
		// The extra 4kb ROM is first in the rom file.
		// It starts with 256 unused bytes not mapped to 0xc000.
		a.isApple2e = true
		extraRomSize := apple2eRomSize - apple2RomSize
		a.mmu.physicalROMe = newMemoryRange(0xc000, data[0:extraRomSize])
		romStart = extraRomSize
	}

	a.mmu.physicalROM = newMemoryRange(0xd000, data[romStart:])
	mmu.resetRomPaging()
}

func (mmu *memoryManager) resetRomPaging() {
	// Assign the first 12kb of ROM from 0xd000 to 0xfff
	for i := 0x0000; i < 0x3000; i = i + 0x100 {
		mmu.setPagesRead(0xd0, 0xff, mmu.physicalROM)
	}
}
