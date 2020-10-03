package izapple2

/*
RAMWorks style card on the Apple IIe aus slot.
	https://patents.google.com/patent/US4601018
	https://ae.applearchives.com/apple_e/ramworks_iii/ramworks_iii_basic_manual_1.pdf

Diagnostics disks:
	https://ae.applearchives.com/apple_e/ramworks_iii/ramworks_diagnostics.zip

It's is like the extra 64kb on an Apple IIe 80col 64kb card, but with up to 256 banks
*/

func setupRAMWorksCard(a *Apple2, banks int) {
	a.mmu.initExtendedRAM(banks)

	ssr := func(_ *ioC0Page) uint8 {
		return a.mmu.extendedRAMBlock
	}

	ssw := func(_ *ioC0Page, value uint8) {
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
}
