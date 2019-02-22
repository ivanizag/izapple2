package apple2

import "go6502/core6502"

// Run instantiates an apple2 and start emulation
func Run(romFile string, log bool) {
	a := newAddressSpace()
	a.loadRom(romFile)

	var s core6502.State
	s.Mem = a

	var fe ansiConsoleFrontend
	a.ioPage.setKeyboardProvider(&fe)
	go fe.textModeGoRoutine(a.textPages1)

	// Start the processor
	core6502.Reset(&s)
	for {
		core6502.ExecuteInstruction(&s, log)
	}
}
