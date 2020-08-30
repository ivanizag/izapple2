package apple2

// See https://fabiensanglard.net/fd_proxy/prince_of_persia/Inside%20the%20Apple%20IIe.pdf
// See https://i.stack.imgur.com/yn21s.gif

type memoryManager struct {
	apple2 *Apple2

	// Main RAM area: 0x0000 to 0xbfff
	physicalMainRAM *memoryRange // 0x0000 to 0xbfff, Up to 48 Kb

	// Slots area: 0xc000 to 0xcfff
	cardsROM      [8]memoryHandler //0xcs00 to 0xcSff. 256 bytes for each card
	cardsROMExtra [8]memoryHandler // 0xc800 to 0xcfff. 2048 bytes for each card

	// Upper area ROM: 0xc000 to 0xffff (or 0xd000 to 0xffff on the II+)
	physicalROM [4]memoryHandler // 0xc000 (or 0xd000) to 0xffff, 16 (or 12) Kb. Up to four banks

	// Language card upper area RAM: 0xd000 to 0xffff. One bank for regular LC cards, up to 8 with Saturn
	physicalLangRAM    []*memoryRange // 0xd000 to 0xffff, 12KB. Up to 8 banks.
	physicalLangAltRAM []*memoryRange // 0xd000 to 0xdfff, 4KB. Up to 8 banks.

	// Extended RAM: 0x0000 to 0xffff (with 4Kb moved from 0xc000 to 0xd000 alt). One bank for extended Apple 2e card, up to 256 with RamWorks
	physicalExtRAM    []*memoryRange // 0x0000 to 0xffff. 60Kb, 0xc000 to 0xcfff not used. Up to 256 banks
	physicalExtAltRAM []*memoryRange // 0xd000 to 0xdfff, 4Kb. Up to 256 banks.

	// Configuration switches, Language cards
	lcSelectedBlock uint8 // Language card block selected. Usually, allways 0. But Saturn has 8
	lcActiveRead    bool  // Upper RAM active for read
	lcActiveWrite   bool  // Upper RAM active for read
	lcAltBank       bool  // Alternate

	// Configuration switches, Apple //e
	altZeroPage           bool  // Use extra RAM from 0x0000 to 0x01ff. And additional language card block
	altMainRAMActiveRead  bool  // Use extra RAM from 0x0200 to 0xbfff for read
	altMainRAMActiveWrite bool  // Use extra RAM from 0x0200 to 0xbfff for write
	store80Active         bool  // Special pagination for text and graphics areas
	slotC3ROMActive       bool  // Apple2e slot 3  ROM shadow
	intCxROMActive        bool  // Apple2e slots internal ROM shadow
	activeSlot            uint8 // Active slot owner of 0xc800 to 0xcfff
	extendedRAMBlock      uint8 // Block used for entended memory for RAMWorks cards

	// Configuration switches, Base64A
	romPage uint8 // Active ROM page

	// Resolution cache
	lastAddressPage    uint16 // The first byte is the page. The second is zero when the cached is valid.
	lastAddressHandler memoryHandler
}

const (
	ioC8Off                uint16 = 0xcfff
	addressLimitZero       uint16 = 0x01ff
	addressStartText       uint16 = 0x0400
	addressLimitText       uint16 = 0x07ff
	addressStartHgr        uint16 = 0x2000
	addressLimitHgr        uint16 = 0x3fff
	addressLimitMainRAM    uint16 = 0xbfff
	addressLimitIO         uint16 = 0xc0ff
	addressLimitSlots      uint16 = 0xc7ff
	addressLimitSlotsExtra uint16 = 0xcfff
	addressLimitDArea      uint16 = 0xdfff

	invalidAddressPage uint16 = 0x0001
)

type memoryHandler interface {
	peek(uint16) uint8
	poke(uint16, uint8)
}

func newMemoryManager(a *Apple2) *memoryManager {
	var mmu memoryManager
	mmu.apple2 = a
	mmu.physicalMainRAM = newMemoryRange(0, make([]uint8, 0xc000))

	mmu.slotC3ROMActive = true // For II+, this is the default behaviour

	return &mmu
}

func (mmu *memoryManager) accessCArea(address uint16) memoryHandler {
	if mmu.intCxROMActive {
		return mmu.physicalROM[mmu.romPage]
	}
	// First slot area
	if address <= addressLimitSlots {
		slot := uint8((address >> 8) & 0x07)
		mmu.activeSlot = slot
		if !mmu.slotC3ROMActive && (slot == 3) {
			return mmu.physicalROM[mmu.romPage]
		}
		return mmu.cardsROM[slot]
	}
	// Extra slot area
	if address == ioC8Off {
		// Reset extra slot area owner
		mmu.activeSlot = 0
		mmu.lastAddressPage = invalidAddressPage
	}

	if !mmu.slotC3ROMActive && (mmu.activeSlot == 3) {
		return mmu.physicalROM[mmu.romPage]
	}
	return mmu.cardsROMExtra[mmu.activeSlot]
}

func (mmu *memoryManager) accessUpperRAMArea(address uint16) memoryHandler {
	if mmu.altZeroPage {
		// Use extended RAM
		block := mmu.extendedRAMBlock
		if mmu.lcAltBank && address <= addressLimitDArea {
			return mmu.physicalExtAltRAM[block]
		}
		return mmu.physicalExtRAM[mmu.extendedRAMBlock]
	}

	// Use language card
	block := mmu.lcSelectedBlock
	if mmu.lcAltBank && address <= addressLimitDArea {
		return mmu.physicalLangAltRAM[block]
	}
	return mmu.physicalLangRAM[block]
}

func (mmu *memoryManager) getPhysicalMainRAM(ext bool) memoryHandler {
	if ext {
		return mmu.physicalExtRAM[mmu.extendedRAMBlock]
	}
	return mmu.physicalMainRAM
}

func (mmu *memoryManager) getVideoRAM(ext bool) *memoryRange {
	if ext {
		// The video memory uses the first extended RAM block, even with RAMWorks
		return mmu.physicalExtRAM[0]
	}
	return mmu.physicalMainRAM
}

func (mmu *memoryManager) accessReadCached(address uint16) memoryHandler {
	page := address & 0xff00
	if address&0xff00 == mmu.lastAddressPage {
		//fmt.Printf("    hit %v\n", mmu.apple2.cpu.GetCycles())
		return mmu.lastAddressHandler
	}

	//fmt.Printf("Not hit %v\n", mmu.apple2.cpu.GetCycles())
	mh := mmu.accessRead(address)
	if address&0xf000 != 0xc000 {
		// Do not cache 0xC area as it may reconfigure the MMU
		mmu.lastAddressPage = page
		mmu.lastAddressHandler = mh
	}
	return mh
}

func (mmu *memoryManager) accessRead(address uint16) memoryHandler {
	if address <= addressLimitZero {
		return mmu.getPhysicalMainRAM(mmu.altZeroPage)
	}
	if mmu.store80Active && address <= addressLimitHgr {
		altPage := mmu.apple2.io.isSoftSwitchActive(ioFlagSecondPage) // TODO: move flag to mmu property like the store80
		if address >= addressStartText && address <= addressLimitText {
			return mmu.getPhysicalMainRAM(altPage)
		}
		hires := mmu.apple2.io.isSoftSwitchActive(ioFlagHiRes)
		if hires && address >= addressStartHgr && address <= addressLimitHgr {
			return mmu.getPhysicalMainRAM(altPage)
		}
	}
	if address <= addressLimitMainRAM {
		return mmu.getPhysicalMainRAM(mmu.altMainRAMActiveRead)
	}
	if address <= addressLimitIO {
		mmu.lastAddressPage = invalidAddressPage
		return mmu.apple2.io
	}
	if address <= addressLimitSlotsExtra {
		return mmu.accessCArea(address)
	}
	if mmu.lcActiveRead {
		return mmu.accessUpperRAMArea(address)
	}
	return mmu.physicalROM[mmu.romPage]
}

func (mmu *memoryManager) accessWrite(address uint16) memoryHandler {
	if address <= addressLimitZero {
		return mmu.getPhysicalMainRAM(mmu.altZeroPage)
	}
	if address <= addressLimitHgr && mmu.store80Active {
		altPage := mmu.apple2.io.isSoftSwitchActive(ioFlagSecondPage)
		if address >= addressStartText && address <= addressLimitText {
			return mmu.getPhysicalMainRAM(altPage)
		}
		hires := mmu.apple2.io.isSoftSwitchActive(ioFlagHiRes)
		if hires && address >= addressStartHgr && address <= addressLimitHgr {
			return mmu.getPhysicalMainRAM(altPage)
		}
	}
	if address <= addressLimitMainRAM {
		return mmu.getPhysicalMainRAM(mmu.altMainRAMActiveWrite)
	}
	if address <= addressLimitIO {
		return mmu.apple2.io
	}
	if address <= addressLimitSlotsExtra {
		return mmu.accessCArea(address)
	}
	if mmu.lcActiveWrite {
		return mmu.accessUpperRAMArea(address)
	}
	return mmu.physicalROM[mmu.romPage]
}

// Peek returns the data on the given address
func (mmu *memoryManager) Peek(address uint16) uint8 {
	mh := mmu.accessRead(address)
	if mh == nil {
		return 0xf4 // Or some random number
	}
	return mh.peek(address)
}

// Peek returns the data on the given address optimized for more local requests
func (mmu *memoryManager) PeekCode(address uint16) uint8 {
	page := address & 0xff00
	var mh memoryHandler
	if page == mmu.lastAddressPage {
		mh = mmu.lastAddressHandler
	} else {
		mh = mmu.accessRead(address)
		if address&0xf000 != 0xc000 {
			// Do not cache 0xC area as it may reconfigure the MMU
			mmu.lastAddressPage = page
			mmu.lastAddressHandler = mh
		}
	}

	if mh == nil {
		return 0xf4 // Or some random number
	}
	return mh.peek(address)
}

// Poke sets the data at the given address
func (mmu *memoryManager) Poke(address uint16, value uint8) {
	mh := mmu.accessWrite(address)
	if mh != nil {
		mh.poke(address, value)
	}
}

// Memory initialization
func (mmu *memoryManager) setCardROM(slot int, mh memoryHandler) {
	mmu.cardsROM[slot] = mh
}

func (mmu *memoryManager) setCardROMExtra(slot int, mh memoryHandler) {
	mmu.cardsROMExtra[slot] = mh
}

func (mmu *memoryManager) initLanguageRAM(groups uint8) {
	// Apple II+ language card or Saturn (up to 8 groups)
	mmu.physicalLangRAM = make([]*memoryRange, groups)
	mmu.physicalLangAltRAM = make([]*memoryRange, groups)
	for i := uint8(0); i < groups; i++ {
		mmu.physicalLangRAM[i] = newMemoryRange(0xd000, make([]uint8, 0x3000))
		mmu.physicalLangAltRAM[i] = newMemoryRange(0xd000, make([]uint8, 0x1000))
	}
}

func (mmu *memoryManager) initExtendedRAM(groups int) {
	// Apple IIe 80 col card with 64Kb style RAM or RAMWorks (up to 256 banks)
	mmu.physicalExtRAM = make([]*memoryRange, groups)
	mmu.physicalExtAltRAM = make([]*memoryRange, groups)
	for i := 0; i < groups; i++ {
		mmu.physicalExtRAM[i] = newMemoryRange(0, make([]uint8, 0x10000))
		mmu.physicalExtAltRAM[i] = newMemoryRange(0xd000, make([]uint8, 0x1000))
	}
}

// Memory configuration
func (mmu *memoryManager) setActiveROMPage(page uint8) {
	mmu.romPage = page
}

func (mmu *memoryManager) getActiveROMPage() uint8 {
	return mmu.romPage
}

func (mmu *memoryManager) setLanguageRAM(readActive bool, writeActive bool, altBank bool) {
	mmu.lcActiveRead = readActive
	mmu.lcActiveWrite = writeActive
	mmu.lcAltBank = altBank
}

func (mmu *memoryManager) setLanguageRAMActiveBlock(block uint8) {
	block = block % uint8(len(mmu.physicalLangRAM))
	mmu.lcSelectedBlock = block
}

func (mmu *memoryManager) setExtendedRAMActiveBlock(block uint8) {
	if int(block) >= len(mmu.physicalExtRAM) {
		// How does the real hardware reacts?
		block = 0
	}
	mmu.extendedRAMBlock = block
}
