package apple2

import "go6502/core6502"

// Run instantiates an apple2 and start emulation
func Run(romFile string, log bool) {
	mmu := newAddressSpace(romFile)

	var s core6502.State
	s.Mem = mmu

	var fe ansiConsoleFrontend
	mmu.ioPage.setKeyboardProvider(&fe)
	go fe.textModeGoRoutine(mmu.textPages1)

	// Start the processor
	core6502.Reset(&s)
	for {
		core6502.ExecuteInstruction(&s, log)
	}
}
