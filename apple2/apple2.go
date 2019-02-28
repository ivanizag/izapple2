package apple2

import "go6502/core6502"

// Run instantiates an apple2 and start emulation
func Run(romFile string, log bool) {
	mmu := newAddressSpace(romFile)
	s := core6502.NewNMOS6502(mmu)
	fe := newAnsiConsoleFrontend(mmu)
	mmu.ioPage.setKeyboardProvider(fe)

	go fe.textModeGoRoutine()

	// Start the processor
	s.Reset()
	for {
		s.ExecuteInstruction(log)
	}
}
