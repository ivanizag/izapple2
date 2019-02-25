package apple2

import "go6502/core6502"

// Run instantiates an apple2 and start emulation
func Run(romFile string, log bool) {
	mmu := newAddressSpace(romFile)

	var s core6502.State
	s.Mem = mmu

	fe := newAnsiConsoleFrontend(mmu)
	mmu.ioPage.setKeyboardProvider(fe)
	go fe.textModeGoRoutine()

	// Start the processor
	core6502.Reset(&s)
	for {
		core6502.ExecuteInstruction(&s, log)
	}
}
