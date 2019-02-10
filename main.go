package main

import "fmt"

func main() {
	var s state
	s.memory.loadBinary("6502_65C02_functional_tests/bin_files/6502_functional_test.bin")

	s.registers.setPC(0x0400)
	for true {
		testCase := s.memory[0x0200]
		if testCase >= 240 {
			break
		}
		log := testCase > 43
		if log {
			fmt.Printf("[ %d ] ", testCase)
		}
		pc := s.registers.getPC()
		executeInstruction(&s, log)
		if pc == s.registers.getPC() {
			//s.memory.printPage(0x00)
			//s.memory.printPage(0x01)
			panic("No change in PC")
		}
	}

	fmt.Printf("Test completed\n")
}
