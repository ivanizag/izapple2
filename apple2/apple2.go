package apple2

import "go6502/core6502"

// Run instantiates an apple2 and start emulation
func Run(romFile string, log bool) {

	// Setup the Apple ][ address space
	var s core6502.State
	s.Mem.InitWithRAM()
	s.Mem.LoadRom(romFile)
	var io ioC0Page
	s.Mem.SetPage(0xc0, &io)
	var t textPages
	for j := 0; j < 4; j++ {
		s.Mem.SetPage(uint8(4+j), &(t.pages[j]))
	}

	var fe ansiConsoleFrontend
	io.setKeyboardProvider(&fe)
	go fe.textModeGoRoutine(&t)

	// Start the processor
	core6502.Reset(&s)
	for {
		core6502.ExecuteInstruction(&s, log)
	}
}
