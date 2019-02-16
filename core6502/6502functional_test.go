package main

import (
	"fmt"
	"testing"
)

func TestFunctional(t *testing.T) {

	t.SkipNow()

	var s state
	// Test suite from https://github.com/Klaus2m5/6502_65C02_functional_tests
	s.memory.loadBinary("tests/6502_functional_test.bin")

	s.registers.setPC(0x0400)
	for true {
		testCase := s.memory.peek(0x0200)
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
			t.Errorf("Failuse in test %v.", testCase)
		}
	}
}
