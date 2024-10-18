package izapple2

import (
	"fmt"
)

// Card represents an Apple II card to be inserted in a slot
type Card interface {
	configure(name string, trace bool, traceMemory bool)
	loadRom(data []uint8, layout cardRomLayout) error
	assign(a *Apple2, slot int)
	reset()
	runDMACycle()

	GetName() string
	GetInfo() map[string]string
}

type cardBase struct {
	a           *Apple2
	name        string
	trace       bool
	traceMemory bool
	romCsxx     *memoryRangeROM
	romC8xx     memoryHandler
	romCxxx     memoryHandler

	slot     int
	_ssr     [16]softSwitchR
	_ssw     [16]softSwitchW
	_ssrName [16]string
	_sswName [16]string
}

func (c *cardBase) GetName() string {
	return c.name
}

func (c *cardBase) GetInfo() map[string]string {
	return nil
}

func (c *cardBase) configure(name string, trace bool, traceMemory bool) {
	c.name = name
	c.trace = trace
	c.traceMemory = traceMemory
}

func (c *cardBase) reset() {
	// nothing
}

type cardRomLayout int

const (
	cardRomSimple       cardRomLayout = iota // The ROM is on the slot area, there can be more than one page
	cardRomUpper                             // The ROM is on the full C800 area. The slot area copies C8xx
	cardRomUpperHalfEnd                      // The ROM is on half of the C800 areas. The slot area copies CBxx
	cardRomFull                              // The ROM is on the full Cxxx area, with pages for each slot position
)

func (c *cardBase) loadRomFromResource(resource string, layout cardRomLayout) error {
	data, _, err := LoadResource(resource)
	if err != nil {
		// The resource should be internal and never fail
		return err
	}
	err = c.loadRom(data, layout)
	if err != nil {
		return err
	}
	return nil
}

func (c *cardBase) loadRom(data []uint8, layout cardRomLayout) error {
	if c.a != nil {
		return fmt.Errorf("ROM must be loaded before inserting the card in the slot")
	}
	switch layout {
	case cardRomSimple:
		if len(data) == 0x100 {
			// Just 256 bytes in Cs00
			c.romCsxx = newMemoryRangeROM(0, data, "Slot ROM")
		} else if len(data)%0x100 == 0 {
			// The ROM covers many 256 bytes pages of Csxx
			// Used on the Dan 2 controller card
			c.romCsxx = newMemoryRangePagedROM(0, data, "Slot paged ROM", uint8(len(data)/0x100))
		} else {
			return fmt.Errorf("invalid ROM size for simple layout")
		}
	case cardRomUpper:
		if len(data) == 0x800 {
			// The file has C800 to CFFF
			// The 256 bytes in Cx00 are copied from the first page in C800
			c.romCsxx = newMemoryRangeROM(0, data, "Slot ROM")
			c.romC8xx = newMemoryRangeROM(0xc800, data, "Slot C8 ROM")
		} else {
			return fmt.Errorf("invalid ROM size for upper layout")
		}
	case cardRomUpperHalfEnd:
		if len(data) == 0x400 {
			// The file has C800 to CBFF for ROM
			// The 256 bytes in Cx00 are copied from the last page in C800-CBFF
			// Used on the Videx 80 columns card
			c.romCsxx = newMemoryRangeROM(0, data[0x300:], "Slot ROM")
			c.romC8xx = newMemoryRangeROM(0xc800, data, "Slot C8 ROM")
		} else {
			return fmt.Errorf("invalid ROM size for upper half end layout")
		}
	case cardRomFull:
		if len(data) == 0x1000 {
			// The file covers the full Cxxx range. Only showing the page
			// corresponding to the slot used.
			c.romCxxx = newMemoryRangeROM(0xc000, data, "Slot ROM")
		} else if len(data)%0x1000 == 0 {
			// The ROM covers the full Cxxx range with several pages
			c.romCxxx = newMemoryRangePagedROM(0xc000, data, "Slot paged ROM", uint8(len(data)/0x1000))
		} else {
			return fmt.Errorf("invalid ROM size for full layout")
		}
	default:
		return fmt.Errorf("invalid card ROM layout")
	}

	return nil
}

func (c *cardBase) assign(a *Apple2, slot int) {
	c.a = a
	c.slot = slot
	if slot != 0 {
		if c.romCsxx != nil {
			// Relocate to the assigned slot
			c.romCsxx.setBase(uint16(0xc000 + slot*0x100))
			rom := traceMemory(c.romCsxx, c.name, c.traceMemory)
			a.mmu.setCardROM(slot, rom)
		}
		if c.romC8xx != nil {
			rom := traceMemory(c.romC8xx, c.name, c.traceMemory)
			a.mmu.setCardROMExtra(slot, rom)
		}
		if c.romCxxx != nil {
			rom := traceMemory(c.romCxxx, c.name, c.traceMemory)
			a.mmu.setCardROM(slot, c.romCxxx)
			a.mmu.setCardROMExtra(slot, rom)
		}
	}

	for i := 0; i < 0x10; i++ {
		if c._ssr[i] != nil {
			a.io.addSoftSwitchR(uint8(0xC80+slot*0x10+i), c._ssr[i], c._ssrName[i])
		}
		if c._ssw[i] != nil {
			a.io.addSoftSwitchW(uint8(0xC80+slot*0x10+i), c._ssw[i], c._sswName[i])
		}
	}
}

func (c *cardBase) runDMACycle() {
	// No DMA
}

func (c *cardBase) activateDMA() {
	if c.a.dmaActive {
		panic("DMA chain not supported")
	}
	c.a.dmaActive = true
	c.a.dmaSlot = c.slot
}

func (c *cardBase) deactivateDMA() {
	c.a.dmaActive = false
}

func (c *cardBase) addCardSoftSwitchR(address uint8, ss softSwitchR, name string) {
	c._ssr[address] = ss
	c._ssrName[address] = name
}

func (c *cardBase) addCardSoftSwitchW(address uint8, ss softSwitchW, name string) {
	c._ssw[address] = ss
	c._sswName[address] = name
}

func (c *cardBase) addCardSoftSwitchRW(address uint8, ss softSwitchR, name string) {
	c._ssr[address] = ss
	c._ssrName[address] = name

	c._ssw[address] = func(uint8) {
		ss()
	}
	c._sswName[address] = name
}

type softSwitches func(address uint8, data uint8, write bool) uint8

func (c *cardBase) addCardSoftSwitches(sss softSwitches, name string) {

	for i := uint8(0x0); i <= 0xf; i++ {
		address := i
		c.addCardSoftSwitchR(address, func() uint8 {
			return sss(address, 0, false)
		}, fmt.Sprintf("%v%XR", name, address))
		c.addCardSoftSwitchW(address, func(value uint8) {
			sss(address, value, true)
		}, fmt.Sprintf("%v%XW", name, address))
	}
}

func (c *cardBase) tracef(format string, args ...interface{}) {
	if c.trace {
		prefixedFormat := fmt.Sprintf("[%s] %v", c.name, format)
		fmt.Printf(prefixedFormat, args...)
	}
}
