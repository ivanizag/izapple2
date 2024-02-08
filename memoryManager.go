package izapple2

import "fmt"

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
	physicalROM memoryHandler // 0xc000 (or 0xd000) to 0xffff, 16 (or 12) Kb. Up to four banks

	// Language card upper area RAM: 0xd000 to 0xffff. One bank for regular LC cards, up to 8 with Saturn
	physicalLangRAM    []*memoryRange // 0xd000 to 0xffff, 12KB. Up to 8 banks.
	physicalLangAltRAM []*memoryRange // 0xd000 to 0xdfff, 4KB. Up to 8 banks.

	// Extended RAM: 0x0000 to 0xffff (with 4Kb moved from 0xc000 to 0xd000 alt). One bank for extended Apple 2e card, up to 256 with RamWorks
	physicalExtRAM    []*memoryRange // 0x0000 to 0xffff. 60Kb, 0xc000 to 0xcfff not used. Up to 256 banks
	physicalExtAltRAM []*memoryRange // 0xd000 to 0xdfff, 4Kb. Up to 256 banks.

	// Configuration switches, Language cards
	lcSelectedBlock uint8 // Language card block selected. Usually, allways 0. But Saturn has 8
	lcActiveRead    bool  // Upper RAM active for read
	lcActiveWrite   bool  // Upper RAM active for write
	lcAltBank       bool  // Alternate

	// Configuration switches, Apple //e
	altZeroPage           bool          // Use extra RAM from 0x0000 to 0x01ff. And additional language card block
	altMainRAMActiveRead  bool          // Use extra RAM from 0x0200 to 0xbfff for read
	altMainRAMActiveWrite bool          // Use extra RAM from 0x0200 to 0xbfff for write
	store80Active         bool          // Special pagination for text and graphics areas
	slotC3ROMActive       bool          // Apple2e slot 3  ROM shadow
	intCxROMActive        bool          // Apple2e slots internal ROM shadow
	intC8ROMActive        bool          // C8Rom associated to the internal slot 3. Softswitch not directly accessible. See UtA2e 5-28
	activeSlot            uint8         // Active slot owner of 0xc800 to 0xcfff
	extendedRAMBlock      uint8         // Block used for entended memory for RAMWorks cards
	mainROMinhibited      memoryHandler // Alternative ROM from 0xd000 to 0xffff provided by a card with the INH signal.

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
	setBase(uint16)
}

func newMemoryManager(a *Apple2) *memoryManager {
	var mmu memoryManager
	mmu.apple2 = a
	mmu.physicalMainRAM = newMemoryRange(0, make([]uint8, 0xc000), "Main RAM")

	mmu.slotC3ROMActive = true // For II+, this is the default behaviour

	return &mmu
}

func (mmu *memoryManager) accessCArea(address uint16) memoryHandler {
	slot := uint8((address >> 8) & 0x0f)

	// Internal IIe slot 3
	if (address <= addressLimitSlots) && !mmu.slotC3ROMActive && (slot == 3) {
		mmu.intC8ROMActive = true
		return mmu.physicalROM
	}

	// Internal IIe CxROM
	if mmu.intCxROMActive {
		return mmu.physicalROM
	}

	// First slot area
	if slot <= 7 {
		mmu.activeSlot = slot
		mmu.intC8ROMActive = false
		return mmu.cardsROM[slot]
	}

	// Extra slot area reset
	if address == ioC8Off {
		// Reset extra slot area owner
		mmu.activeSlot = 0
		mmu.intC8ROMActive = false
	}

	// Extra slot area
	if mmu.intC8ROMActive {
		return mmu.physicalROM
	}
	return mmu.cardsROMExtra[mmu.activeSlot]
}

func (mmu *memoryManager) accessUpperRAMArea(address uint16) memoryHandler {
	if mmu.altZeroPage && mmu.hasExtendedRAM() {
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
	if ext && mmu.hasExtendedRAM() {
		return mmu.physicalExtRAM[mmu.extendedRAMBlock]
	}
	return mmu.physicalMainRAM
}

func (mmu *memoryManager) getVideoRAM(ext bool) *memoryRange {
	if ext && mmu.hasExtendedRAM() {
		// The video memory uses the first extended RAM block, even with RAMWorks
		return mmu.physicalExtRAM[0]
	}
	return mmu.physicalMainRAM
}

func (mmu *memoryManager) inhibitROM(replacement memoryHandler) {
	// If a card INH the ROM, it replaces the ROM and the LC RAM
	mmu.mainROMinhibited = replacement
	mmu.lastAddressPage = invalidAddressPage // Invalidate cache
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
	if mmu.mainROMinhibited != nil {
		return mmu.mainROMinhibited
	}
	if mmu.lcActiveRead {
		return mmu.accessUpperRAMArea(address)
	}
	return mmu.physicalROM
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
		mmu.lastAddressPage = invalidAddressPage
		return mmu.apple2.io
	}
	if address <= addressLimitSlotsExtra {
		return mmu.accessCArea(address)
	}
	if mmu.mainROMinhibited != nil {
		return mmu.mainROMinhibited
	}
	if mmu.lcActiveWrite {
		return mmu.accessUpperRAMArea(address)
	}
	return mmu.physicalROM
}

func (mmu *memoryManager) peekWord(address uint16) uint16 {
	return uint16(mmu.Peek(address)) +
		uint16(mmu.Peek(address+1))<<8
}

// Peek returns the data on the given address
func (mmu *memoryManager) Peek(address uint16) uint8 {
	mh := mmu.accessRead(address)
	if mh == nil {
		return 0xf4 // Or some random number
	}
	value := mh.peek(address)
	//if address >= 0xc400 && address < 0xc500 {
	//	fmt.Printf("[MMU] Peek at %04x: %02x\n", address, value)
	//}

	return value
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

	value := mh.peek(address)
	//if address >= 0xc400 && address < 0xc500 {
	//	fmt.Printf("[MMU] PeekCode at %04x: %02x\n", address, value)
	//}

	return value
}

func (mmu *memoryManager) pokeRange(address uint16, data []uint8) {
	for i := 0; i < len(data); i++ {
		mmu.Poke(address+uint16(i), data[i])
	}
}

// Poke sets the data at the given address
func (mmu *memoryManager) Poke(address uint16, value uint8) {
	mh := mmu.accessWrite(address)
	if mh != nil {
		mh.poke(address, value)
	}

	//if address >= 0x0036 && address <= 0x0039 {
	//	fmt.Printf("[MMU] Poke at %04x: %02x\n", address, value)
	//}
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
		mmu.physicalLangRAM[i] = newMemoryRange(0xd000, make([]uint8, 0x3000), fmt.Sprintf("LC RAM block %v", i))
		mmu.physicalLangAltRAM[i] = newMemoryRange(0xd000, make([]uint8, 0x1000), fmt.Sprintf("LC RAM Alt block %v", i))
	}
}

func (mmu *memoryManager) initExtendedRAM(groups int) {
	// Apple IIe 80 col card with 64Kb style RAM or RAMWorks (up to 256 banks)
	mmu.physicalExtRAM = make([]*memoryRange, groups)
	mmu.physicalExtAltRAM = make([]*memoryRange, groups)
	for i := 0; i < groups; i++ {
		mmu.physicalExtRAM[i] = newMemoryRange(0, make([]uint8, 0x10000), fmt.Sprintf("Extra RAM block %v", i))
		mmu.physicalExtAltRAM[i] = newMemoryRange(0xd000, make([]uint8, 0x1000), fmt.Sprintf("Extra RAM Alt block %v", i))
	}
}

// Memory configuration
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

func (mmu *memoryManager) hasExtendedRAM() bool {
	return len(mmu.physicalExtRAM) > 0
}

func (mmu *memoryManager) reset() {
	if mmu.apple2.isApple2e {
		// MMU UtA2e 4-14, 5-22
		mmu.altZeroPage = false
		mmu.altMainRAMActiveRead = false
		mmu.altMainRAMActiveWrite = false
		mmu.store80Active = false
		mmu.slotC3ROMActive = false
		mmu.intCxROMActive = false
		mmu.intC8ROMActive = false

		// IOU UtaA2e 7-3
		// "All softswitches except KEYSTROKE, TEXT and MIXED are reset
		// when the RESET line drops low"
		mmu.apple2.io.softSwitchesData[ioFlagSecondPage] = ssOff
		mmu.apple2.io.softSwitchesData[ioFlagHiRes] = ssOff
		mmu.apple2.io.softSwitchesData[ioFlag80Col] = ssOff
		mmu.apple2.io.softSwitchesData[ioDataNewVideo] = ssOff
		// ioFlagText ?
	}
}
