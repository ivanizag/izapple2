package main

import "fmt"

func main() {
	var s state
	s.memory.loadBinary("6502_65C02_functional_tests/bin_files/6502_functional_test.bin")

	s.registers.setPC(0x0400)
	for true {
		for i := 0; i < 20; i++ {
			testCase := s.memory[0x0200]
			fmt.Printf("[ %d ] ", testCase)
			pc := s.registers.getPC()
			executeInstruction(&s)
			if pc == s.registers.getPC() {
				s.memory.printPage(0x01)
				panic("No change in PC")
			}
		}
	}
}
