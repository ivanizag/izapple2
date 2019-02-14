package main

func main() {
	var s state
	var t textPages

	s.memory.initWithRomAndText("../roms/APPLE2.ROM", &t)
	startAddress := s.memory.getWord(0xfffc)
	s.registers.setPC(startAddress)
	for true {
		log := true
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
