package apple2

import "io/ioutil"

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

func (mmu *memoryManager) setPage(index uint8, mh memoryHandler) {
	mmu.setPageRead(index, mh)
	mmu.setPageWrite(index, mh)
}

func (mmu *memoryManager) setPageRead(index uint8, mh memoryHandler) {
	mmu.activeMemoryRead[index] = mh
}

func (mmu *memoryManager) setPageWrite(index uint8, mh memoryHandler) {
	mmu.activeMemoryWrite[index] = mh
}

// When 0xcfff is accessed the card expansion rom is unassigned
func (mmu *memoryManager) resetSlotExpansionRoms() {
	if mmu.apple2.io.isSoftSwitchActive(ioFlagIntCxRom) {
		// Ignore if the Apple2 shadow ROM is active
		return
	}
	for i := uint8(0xc8); i < 0xd0; i++ {
		mmu.setPage(i, nil)
	}
}

func newMemoryManager(a *Apple2) *memoryManager {
	var mmu memoryManager
	mmu.apple2 = a

	// Assign RAM from 0x0000 to 0xbfff, 48kb
	ram := make([]uint8, 0xc000)
	mmu.physicalMainRAM = newMemoryRange(0, ram)
	for i := 0; i < 0xc000; i = i + 0x100 {
		mmu.setPage(uint8(i>>8), mmu.physicalMainRAM)
	}

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
		mmu.setPageRead(uint8(0xd0+(i>>8)), mmu.physicalROM)
	}
}
