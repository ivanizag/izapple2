package apple2

import (
	"bufio"
	"os"
)

// See https://fabiensanglard.net/fd_proxy/prince_of_persia/Inside%20the%20Apple%20IIe.pdf
// See https://i.stack.imgur.com/yn21s.gif

type memoryManager struct {
	// Map of assigned pages
	activeMemory *pagedMemory
	// Pages prepared to be paged in and out
	physicalMainRAM        []ramPage        // 0x0000 to 0xbfff, Up to 48 Kb
	physicalROM            []romPage        // 0xd000 to 0xffff, 12 Kb
	physicalROMe           []romPage        // 0xc000 to 0xcfff, Zero or 4bk in the Apple2e
	unassignedExpansionROM []unassignedPage // 0xc000 to 0xcfff
	ioPage                 *ioC0Page        // 0xc000 to 0xc080
	isApple2e              bool
	activeSlot             int // Slot that has the addressing 0xc800 to 0ccfff
}

const (
	ioAreaMask  uint16 = 0xFF80
	ioAreaValue uint16 = 0xC000
	ioC8Off     uint16 = 0xCFFF
)

// Peek returns the data on the given address
func (mmu *memoryManager) Peek(address uint16) uint8 {
	if address == ioC8Off {
		mmu.resetSlotExpansionRoms()
	}
	return mmu.activeMemory.Peek(address)
}

// Poke sets the data at the given address
func (mmu *memoryManager) Poke(address uint16, value uint8) {
	if address == ioC8Off {
		mmu.resetSlotExpansionRoms()
	}
	mmu.activeMemory.Poke(address, value)
}

// When 0xcfff is accessed the card expansion rom is unassigned
func (mmu *memoryManager) resetSlotExpansionRoms() {
	if mmu.ioPage.isSoftSwitchExtActive(ioFlagIntCxRom) {
		// Ignore if the Apple2 shadow ROM is active
		return
	}
	for i := 8; i < 16; i++ {
		p := mmu.unassignedExpansionROM[i]
		mmu.activeMemory.SetPage(uint8(i+0xc0), &p)
	}
}

func newAddressSpace(romImage string) *memoryManager {
	var mmu memoryManager

	var m pagedMemory
	mmu.activeMemory = &m

	// Assign RAM from 0x0000 to 0xbfff, 48kb
	mmu.physicalMainRAM = make([]ramPage, 0xc0)
	for i := 0; i <= 0xbf; i++ {
		m.SetPage(uint8(i), &(mmu.physicalMainRAM[i]))
	}

	mmu.loadRom(romImage)
	// Assign the first 12kb of ROM from 0xd000 to 0xfff
	for i := 0xd0; i <= 0xff; i++ {
		m.SetPage(uint8(i), &(mmu.physicalROM[i-0xd0]))
	}

	// Set the io in 0xc000
	mmu.ioPage = newIoC0Page(&mmu)
	m.SetPage(0xc0, mmu.ioPage)

	// Set the 0xc100 to 0xcfff as unasigned, 4kb. It wil be taken by slot cards.
	mmu.unassignedExpansionROM = make([]unassignedPage, 0x10)
	for i := 1; i < 0x10; i++ {
		page := uint8(i + 0xc0)
		p := &mmu.unassignedExpansionROM[i]
		p.page = page
		m.SetPage(page, p)
	}

	return &mmu
}

// LoadRom loads a binary file to the top of the memory.
const (
	apple2RomSize  = 12 * 1024
	apple2eRomSize = 16 * 1024
)

func (mmu *memoryManager) loadRom(filename string) {
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
	if size != apple2RomSize && size != apple2eRomSize {
		panic("Rom size not supported")
	}
	bytes := make([]byte, size)
	buf := bufio.NewReader(f)
	buf.Read(bytes)

	romStart := 0
	if size == apple2eRomSize {
		// The extra 4kb ROM is first in the rom file.
		// It starts with 256 unused bytes not mapped to 0xc000.
		mmu.isApple2e = true
		extraRomSize := apple2eRomSize - apple2RomSize
		mmu.physicalROMe = make([]romPage, extraRomSize>>8)
		for i := 0; i < extraRomSize; i++ {
			mmu.physicalROMe[i>>8].burn(uint8(i), bytes[i])
		}
		romStart = extraRomSize
	}

	mmu.physicalROM = make([]romPage, apple2RomSize>>8)
	for i := 0; i < apple2RomSize; i++ {
		mmu.physicalROM[i>>8].burn(uint8(i), bytes[i+romStart])
	}
}
