package main

func main() {
	var s state
	var t textPages
	var io ioC0Page

	/*
		for c := uint8(0); c < 40; c++ {
			for l := uint8(0); l < 24; l++ {
				t.write(c, l, '0'+(c+l)%10)
				t.dump()
			}
		}
	*/
	//s.memory.initWithRomAndText("../roms/APPLE2.ROM", &t)
	s.memory.initWithRomAndText("../roms/apple.rom", &t)
	//s.memory.initWithRomAndText("../roms/apple2o.rom", &t)
	s.memory.data[0xc0] = &io

	startAddress := s.memory.getWord(0xfffc)
	s.registers.setPC(startAddress)
	for true {
		log := false
		pc := s.registers.getPC()
		executeInstruction(&s, log)
		if pc == s.registers.getPC() {
			//s.memory.printPage(0x00)
			//s.memory.printPage(0x01)
			panic("No change in PC")
		}
		t.dumpIfDirty()
	}
}
