package apple2

import (
	"encoding/binary"
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
	physicalMainRAM *memoryRange  // 0x0000 to 0xbfff, Up to 48 Kb
	physicalROM     memoryHandler // 0xd000 to 0xffff, 12 Kb
	physicalROMe    memoryHandler // 0xc000 to 0xcfff, Zero or 4bk in the Apple2e

	// Pages prapared for optional card ROM banks
	activeSlot    uint8
	cardsROMExtra [8]memoryHandler // 0xc800 to 0xcfff. for each card
}

type memoryHandler interface {
	peek(uint16) uint8
	poke(uint16, uint8)
}

const (
	ioC8Off uint16 = 0xCFFF
)

func (mmu *memoryManager) access(address uint16, activeMemory [256]memoryHandler) memoryHandler {
	if address == ioC8Off {
		mmu.resetSlotExpansionRoms()
	}

	hi := uint8(address >> 8)
	if hi >= 0xC1 && hi <= 0xC7 {
		slot := hi - 0xC0
		if slot != mmu.activeSlot {
			mmu.activateCardRomExtra(slot)
		}
	}
	mh := activeMemory[hi]
	if mh == nil {
		return nil
	}
	return mh
}

// Peek returns the data on the given address
func (mmu *memoryManager) Peek(address uint16) uint8 {
	mh := mmu.access(address, mmu.activeMemoryRead)
	if mh == nil {
		return 0xf4 // Or some random number
	}
	return mh.peek(address)
}

// Poke sets the data at the given address
func (mmu *memoryManager) Poke(address uint16, value uint8) {
	mh := mmu.access(address, mmu.activeMemoryWrite)
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

func (mmu *memoryManager) prepareCardExtraRom(slot int, mh memoryHandler) {
	mmu.cardsROMExtra[slot] = mh
}

// When 0xcfff is accessed the card expansion rom is unassigned
func (mmu *memoryManager) resetSlotExpansionRoms() {
	if mmu.apple2.io.isSoftSwitchActive(ioFlagIntCxRom) {
		// Ignore if the Apple2 shadow ROM is active
		return
	}
	mmu.activeSlot = 0
	mmu.setPagesRead(0xc8, 0xcf, nil)
}

// When a card base ROM is accesed the extra rom is assigned if available
func (mmu *memoryManager) activateCardRomExtra(slot uint8) {
	//fmt.Printf("Activate slot %v\n", slot)
	if mmu.cardsROMExtra[slot] != nil {
		mmu.setPagesRead(0xC8, 0xCF, mmu.cardsROMExtra[slot])
	}
	mmu.activeSlot = slot
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

func (mmu *memoryManager) save(w io.Writer) error {
	err := mmu.physicalMainRAM.save(w)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, mmu.activeSlot)
	if err != nil {
		return err
	}
	return nil
}

func (mmu *memoryManager) load(r io.Reader) error {
	err := mmu.physicalMainRAM.load(r)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.BigEndian, &mmu.activeSlot)
	if err != nil {
		return err
	}
	mmu.activateCardRomExtra(mmu.activeSlot)

	mmu.resetBaseRamPaging()
	return nil
}
