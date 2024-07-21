package izapple2

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

/*
No slot clock with DS1216 Phantom Time Chip for ROM

See:
	- http://ctrl.pomme.reset.free.fr/index.php/hardware/no-slot-clock-ds1216e/
	- http://ctrl.pomme.reset.free.fr/wp-content/uploads/NSC/DS1216.pdf
	- https://mirrors.apple2.org.za/Apple%20II%20Documentation%20Project/Chips/SMT%20No-Slot%20Clock/
	- https://www.digchip.com/datasheets/parts/datasheet/000/DS1215-pdf.php

Test software:
	- http://ctrl.pomme.reset.free.fr/wp-content/uploads/NSC/NSC_UTILITIES_V14.dsk

Following the spirit of a phantom chip, the DS1215 will replace a memoryHandler do its things and
delegate to the replaced memoryHandler when needed.

On the Apple IIe it is usually installed under the ROM CD (under CF on later models). Similar for the Apple IIc.
It is usually not compatible with the ROMs of the Apple II+, but could be installed in a card with 28 pins ROM,
working on different addresses. We will install it under the main ROM for all models or under a card ROM.

Actually software looks like it uses: 0xCs00, 0xCs01 and 0xCs04, usually slot 3
*/

type noSlotClockDS1216 struct {
	memory      memoryHandler
	state       uint8
	index       uint8
	timeCapture uint64
}

var nscBitPattern = [64]bool{
	true, false, true, false, false, false, true, true, //C5
	false, true, false, true, true, true, false, false, //3A
	true, true, false, false, false, true, false, true, //A3
	false, false, true, true, true, false, true, false, //5C
	true, false, true, false, false, false, true, true, //C5
	false, true, false, true, true, true, false, false, //3A
	true, true, false, false, false, true, false, true, //A3
	false, false, true, true, true, false, true, false, //5C
}

const (
	nscStateDisabled = uint8(0)
	nscStatePattern  = uint8(1)
	nscStateEnabled  = uint8(2)
)

func newNoSlotClockDS1216(_ *Apple2, memory memoryHandler) *noSlotClockDS1216 {
	var nsc noSlotClockDS1216
	nsc.memory = memory
	nsc.state = nscStateDisabled
	nsc.index = 0
	return &nsc
}

func (nsc *noSlotClockDS1216) peek(address uint16) uint8 {
	read := (address & 0x04) != 0  // Bit A2 of the address bus
	value := (address & 0x01) != 0 // Bit A0 of the address bus

	var data uint8
	switch nsc.state {
	case nscStateDisabled:
		if read {
			// Prior to executing the first of 64 write cycles, a read cycle should be executed
			// by holding A2 high. The read cycle will reset the comparison register pointer
			// within the SmartWatch, ensuring the pattern recognition starts with the first
			// bit of the sequence.
			nsc.state = nscStatePattern
			nsc.index = 0
		}
		data = nsc.memory.peek(address)

	case nscStatePattern:
		// Communication with the SmartWatch is established by pattern recognition of a serial
		// bit stream of 64 bits that must be matched by executing 64 consecutive write cycles,
		// placing address bit A2 low with the proper data on address bit A0. The 64 write cycles
		// are used only to gain access to the SmartWatch.
		if read {
			// If a read cycle occurs at any time during pattern recognition, the present
			// sequence is aborted and the comparison register pointer is reset.
			nsc.index = 0
		} else {
			// When the first write cycle is executed, it is compared to bit 0 of the 64-bit
			// comparison register. Pattern recognition continues for a total of 64 write cycles
			// until all the bits in the comparison register have been matched.
			if value == nscBitPattern[nsc.index] {
				// If a match is found, the pointer increments to the next location of the
				// comparison register and awaits the next write cycle.
				nsc.index++
				if nsc.index == 64 {
					// With a correct match for 64 bits, the SmartWatch is enabled and data transfer to or
					// from the timekeeping registers can proceed.
					nsc.state = nscStateEnabled
					nsc.index = 0
					nsc.loadTime()
				}
			} else {
				// If a match is not found, the pointer does not advance and all subsequent write
				// cycles are ignored.
				nsc.state = nscStateDisabled
			}
		}
		data = nsc.memory.peek(address)

	case nscStateEnabled:
		// The next 64 cycles will cause the SmartWatch to either receive data on data in (A0) or
		// transmit data on data out (DQ0), depending on the level of /WRITE READ (A2).
		if read {
			// Get info
			data = uint8(nsc.timeCapture>>nsc.index) & 1
			// The info is set on the LSB. The rest of bits are zero. Should they be the value in ROM?
		} else {
			// Store info
			if value {
				nsc.timeCapture |= (1 << nsc.index)
			} else {
				nsc.timeCapture &= ^(1 << nsc.index)
			}
			data = 0 // What is returned on write?
		}
		nsc.index++
		if nsc.index == 64 {
			// Is this right?
			nsc.state = nscStateDisabled
			nsc.index = 0
		}
	}

	return data
}

func (nsc *noSlotClockDS1216) poke(address uint16, value uint8) {
	nsc.memory.poke(address, value)
}

func (nsc *noSlotClockDS1216) loadTime() {
	now := time.Now()

	var register uint64

	year := uint64(now.Year()) % 100
	register = year / 10
	register <<= 4
	register += year % 10
	register <<= 4

	month := uint64(now.Month())
	register += month / 10
	register <<= 4
	register += month % 10
	register <<= 4

	day := uint64(now.Day())
	register += day / 10
	register <<= 4
	register += day % 10
	register <<= 4

	// Bits 4 and 5 of the day register are used to control the RST and oscillator
	// functions. These bits are shipped from the factory set to logic 1.
	register += 0x0 //0x3, but zero on read.
	register <<= 4
	register += uint64(now.Weekday()) + 1
	register <<= 4

	hour := uint64(now.Hour())
	register += 0x0 // 0x8 for 24 hour mode, but zero on read.
	register += hour / 10
	register <<= 4
	register += hour % 10
	register <<= 4

	minute := uint64(now.Minute())
	register += minute / 10
	register <<= 4
	register += minute % 10
	register <<= 4

	second := uint64(now.Second())
	register += second / 10
	register <<= 4
	register += second % 10
	register <<= 4

	centisecond := uint64(now.Nanosecond() / 10000000)
	register += centisecond / 10
	register <<= 4
	register += centisecond % 10

	nsc.timeCapture = register
}

func setupNoSlotClock(a *Apple2, arg string) error {
	if arg == "main" {
		nsc := newNoSlotClockDS1216(a, a.mmu.physicalROM)
		a.mmu.physicalROM = nsc
	} else {
		slot, err := strconv.ParseUint(arg, 10, 8)
		if err != nil || slot < 1 || slot > 7 {
			return errors.New("invalid slot for the no slot clock, use 'none', 'main' or a slot number from 1 to 7")
		}
		cardRom := a.mmu.cardsROM[slot]
		if cardRom == nil {
			return fmt.Errorf("no ROM available on slot %d to add a no slot clock", slot)
		}
		nsc := newNoSlotClockDS1216(a, cardRom)
		a.mmu.cardsROM[slot] = nsc
	}
	return nil
}
