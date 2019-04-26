package apple2

// See https://fabiensanglard.net/fd_proxy/prince_of_persia/Inside%20the%20Apple%20IIe.pdf
// See https://i.stack.imgur.com/yn21s.gif

type memoryManager struct {
	apple2 *Apple2
	// Map of assigned pages
	activeMemory [256]memoryPage

	// Pages prepared to be paged in and out
	physicalMainRAM        []ramPage        // 0x0000 to 0xbfff, Up to 48 Kb
	physicalROM            []romPage        // 0xd000 to 0xffff, 12 Kb
	physicalROMe           []romPage        // 0xc000 to 0xcfff, Zero or 4bk in the Apple2e
	unassignedExpansionROM []unassignedPage // 0xc000 to 0xcfff
}

// memoryPage is a data page of 256 bytes
type memoryPage interface {
	Peek(uint8) uint8
	Poke(uint8, uint8)
	internalPeek(uint8) uint8
	all() []uint8
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
	lo := uint8(address)
	return mmu.activeMemory[hi].Peek(lo)
}

func (mmu *memoryManager) internalPeek(address uint16) uint8 {
	hi := uint8(address >> 8)
	lo := uint8(address)
	return mmu.activeMemory[hi].internalPeek(lo)
}

func (mmu *memoryManager) internalPage(hi uint8) []uint8 {
	return mmu.activeMemory[hi].all()
}

// Poke sets the data at the given address
func (mmu *memoryManager) Poke(address uint16, value uint8) {
	if address == ioC8Off {
		mmu.resetSlotExpansionRoms()
	}
	hi := uint8(address >> 8)
	lo := uint8(address)
	mmu.activeMemory[hi].Poke(lo, value)
}

// SetPage assigns a MemoryPage implementation on the page given
func (mmu *memoryManager) setPage(index uint8, page memoryPage) {
	//fmt.Printf("Assigning page 0x%02x type %s\n", index, reflect.TypeOf(page))
	mmu.activeMemory[index] = page

}

// When 0xcfff is accessed the card expansion rom is unassigned
func (mmu *memoryManager) resetSlotExpansionRoms() {
	if mmu.apple2.io.isSoftSwitchActive(ioFlagIntCxRom) {
		// Ignore if the Apple2 shadow ROM is active
		return
	}
	for i := 8; i < 16; i++ {
		p := mmu.unassignedExpansionROM[i]
		mmu.setPage(uint8(i+0xc0), &p)
	}
}

func newMemoryManager(a *Apple2) *memoryManager {
	var mmu memoryManager
	mmu.apple2 = a

	// Assign RAM from 0x0000 to 0xbfff, 48kb
	mmu.physicalMainRAM = make([]ramPage, 0xc0)
	for i := 0; i <= 0xbf; i++ {
		mmu.setPage(uint8(i), &(mmu.physicalMainRAM[i]))
	}

	// Set the 0xc100 to 0xcfff as unasigned, 4kb. It wil be taken by slot cards.
	mmu.unassignedExpansionROM = make([]unassignedPage, 0x10)
	for i := 1; i < 0x10; i++ {
		page := uint8(i + 0xc0)
		p := &mmu.unassignedExpansionROM[i]
		p.page = page
		mmu.setPage(page, p)
	}
	return &mmu
}

func (mmu *memoryManager) resetPaging() {
	// Assign the first 12kb of ROM from 0xd000 to 0xfff
	for i := 0xd0; i <= 0xff; i++ {
		mmu.setPage(uint8(i), &(mmu.physicalROM[i-0xd0]))
	}
}
