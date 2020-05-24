package apple2

/*
 See:
	Inside the Apple IIc by Gary B. Little
*/

const (
	ioFlagMouseXIntOcurred  uint8 = 0x15
	ioFlagMouseYIntOcurred  uint8 = 0x17
	ioFlagVblInt            uint8 = 0x19
	ioFlagMouseXYInterrupts uint8 = 0x58
	ioFlagVblInterrupts     uint8 = 0x5a
	ioFlagXEdge             uint8 = 0x5c
	ioFlagYEdge             uint8 = 0x5e
	ioFlagIOUDis            uint8 = 0x7e
)

func addApple2CSoftSwitches(io *ioC0Page) {
	// Disable not used IIe softswitches
	io.disableSoftSwitchesMmu(0x06, 0x07, 0x15)
	io.disableSoftSwitchesMmu(0x0a, 0x0b, 0x17)

	// Replace softswitches
	io.addSoftSwitchR(0x60, getStatusSoftSwitch(ioFlag40ColOnly), "RD80SW") // Instead og CASSETTE
	io.addSoftSwitchR(0x19, getStatusSoftSwitch(ioFlagVblInt), "VBLINT")    // Replaces the not INT based in IIe

	// Mouse interrupts
	io.addSoftSwitchR(0x15, getStatusSoftSwitch(ioFlagMouseXIntOcurred), "MOUSEX0INT")
	io.addSoftSwitchR(0x15, getStatusSoftSwitch(ioFlagMouseYIntOcurred), "MOUSEY0INT")

	// Replacement of the IOU annunciators 0x58 to 0x5f
	io.addSoftSwitchesIou(0x5f, 0x5e, 0x7f, ioFlagAnnunciator3, "DHIRES")
	io.addSoftSwitchesIou(0x7e, 0x7f, 0x7e, ioFlagIOUDis, "IOUDIS")
	addSoftSwitchesIouDis(io, 0x58, 0x59, 0x40, ioFlagMouseXYInterrupts, "MOUSEXYINTENABLED")
	addSoftSwitchesIouDis(io, 0x5a, 0x5b, 0x41, ioFlagVblInterrupts, "VBLINTENABLED")
	addSoftSwitchesIouDis(io, 0x5c, 0x5d, 0x42, ioFlagXEdge, "MOUSEXEDGE")
	addSoftSwitchesIouDis(io, 0x5e, 0x5f, 0x43, ioFlagYEdge, "MOUSEYEDGE")
	//io.addSoftSwitchR(0x70, notImplementedSoftSwitchR, "PTRIG") // TODO: use also for interrupt clear

	io.copySoftSwitchRW(0x7e, 0x78)
	io.copySoftSwitchRW(0x7f, 0x79)
	io.copySoftSwitchRW(0x7e, 0x7a)
	io.copySoftSwitchRW(0x7f, 0x7b)
	io.copySoftSwitchRW(0x7e, 0x7c)
	io.copySoftSwitchRW(0x7f, 0x7d)

	// 0x98: port 1 data register
	// 0x99: port 1 status register & reset
	// 0x9a: port 1 command register
	// 0x9b: port 1 control register

	// 0xa8: port 2 data register
	// 0xa9: port 2 status register & reset
	// 0xaa: port 2 command register, used for keyboard interrupts (InsideIIc,332)
	// 0xab: port 2 control register

	// Plus a language card in pseudo slot 0 and a disk II card in slot 6

	// Initial values
	io.softSwitchesData[ioFlag40ColOnly] = ssOn // The switch will always be off
	io.softSwitchesData[ioFlagIOUDis] = ssOn    // Verified on Apple IIc ROM FF
	mmu := io.apple2.mmu
	mmu.intCxROMActive = true
}

func addSoftSwitchesIouDis(p *ioC0Page, addressClear uint8, addressSet uint8,
	addressGet uint8, ioFlag uint8, name string) {

	prevClear := p.softSwitchesW[addressClear]
	p.addSoftSwitchW(addressClear, func(io *ioC0Page, value uint8) {
		if io.softSwitchesData[ioFlagIOUDis] == ssOn {
			io.softSwitchesData[ioFlag] = ssOff
		} else {
			prevClear(io, value)
		}
	}, name+"OFF")

	prevSet := p.softSwitchesW[addressClear]
	p.addSoftSwitchW(addressSet, func(io *ioC0Page, value uint8) {
		if io.softSwitchesData[ioFlagIOUDis] == ssOn {
			io.softSwitchesData[ioFlag] = ssOn
		} else {
			prevSet(io, value)
		}
	}, name+"ON")

	prevGet := p.softSwitchesR[addressGet]
	p.addSoftSwitchR(addressGet, func(io *ioC0Page) uint8 {
		if io.softSwitchesData[ioFlagIOUDis] == ssOn {
			return io.softSwitchesData[ioFlag]
		}
		return prevGet(io)
	}, name)

}
