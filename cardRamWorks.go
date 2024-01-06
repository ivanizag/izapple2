package izapple2

import (
	"fmt"
	"strconv"
)

/*
RAMWorks style card on the Apple IIe aus slot.
	https://patents.google.com/patent/US4601018
	https://ae.applearchives.com/apple_e/ramworks_iii/ramworks_iii_basic_manual_1.pdf

Diagnostics disks:
	https://ae.applearchives.com/apple_e/ramworks_iii/ramworks_diagnostics.zip

It's is like the extra 64kb on an Apple IIe 80col 64kb card, but with up to 256 banks
*/

func setupRAMWorksCard(a *Apple2, sizeArg string) error {
	size, err := strconv.Atoi(sizeArg)
	if err != nil {
		return fmt.Errorf("invalid RamWorks card RAM size: %s", sizeArg)
	}
	if size%64 != 0 {
		return fmt.Errorf("the Ramworks size must be a multiple of 64, %v is not", size)
	}

	a.mmu.initExtendedRAM(size / 64)

	ssr := func() uint8 {
		return a.mmu.extendedRAMBlock
	}

	ssw := func(value uint8) {
		a.mmu.setExtendedRAMActiveBlock(value)
	}

	// Does not have a slot assigned
	a.io.addSoftSwitchR(0x71, ssr, "RAMWORKSR")
	a.io.addSoftSwitchR(0x73, ssr, "RAMWORKSR")
	a.io.addSoftSwitchR(0x75, ssr, "RAMWORKSR")
	a.io.addSoftSwitchR(0x77, ssr, "RAMWORKSR")
	a.io.addSoftSwitchW(0x71, ssw, "RAMWORKSW")
	a.io.addSoftSwitchW(0x73, ssw, "RAMWORKSW")
	a.io.addSoftSwitchW(0x75, ssw, "RAMWORKSW")
	a.io.addSoftSwitchW(0x77, ssw, "RAMWORKSW")

	return nil
}
