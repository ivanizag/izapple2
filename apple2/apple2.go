package apple2

import "go6502/core6502"

// Run instantiates an apple2 and start emulation
func Run(romFile string, log bool) {

	// Setup the Apple ][ address space
	var m core6502.PagedMemory
	m.InitWithRAM()
	m.LoadRom(romFile)
	var io ioC0Page
	m.SetPage(0xc0, &io)
	var t textPages
	for j := 0; j < 4; j++ {
		m.SetPage(uint8(4+j), &(t.pages[j]))
	}

	for j := uint8(0xc1); j < 0xd0; j++ {
		var p tracePage
		p.page = j
		m.SetPage(j, &p)
	}

	var s core6502.State
	s.Mem = &m

	var fe ansiConsoleFrontend
	io.setKeyboardProvider(&fe)
	go fe.textModeGoRoutine(&t)

	// Start the processor
	core6502.Reset(&s)
	for {
		core6502.ExecuteInstruction(&s, log)
	}
}
