package apple2

import (
	"io"
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

func (mmu *memoryManager) resetRomPaging() {
	// Assign the first 12kb of ROM from 0xd000 to 0xffff
	mmu.setPagesRead(0xd0, 0xff, mmu.physicalROM)
}

func (mmu *memoryManager) resetBaseRamPaging() {
	// Assign the base RAM from 0x0000 to 0xbfff
	mmu.setPages(0x00, 0xbf, mmu.physicalMainRAM)
}

func newMemoryManager(a *Apple2) *memoryManager {
	var mmu memoryManager
	mmu.apple2 = a

	ram := make([]uint8, 0xc000) // Reserve 48kb
	mmu.physicalMainRAM = newMemoryRange(0, ram)
	mmu.resetBaseRamPaging()

	return &mmu
}

func (mmu *memoryManager) save(w io.Writer) {
	mmu.physicalMainRAM.save(w)
}

func (mmu *memoryManager) load(r io.Reader) {
	mmu.physicalMainRAM.load(r)
	mmu.resetBaseRamPaging()
}
