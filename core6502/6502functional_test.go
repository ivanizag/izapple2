package core6502

import (
	"fmt"
	"testing"
)

func TestFunctional(t *testing.T) {

	//t.SkipNow()

	var s State
	var m FlatMemory
	s.Mem = &m
	// Test suite from https://github.com/Klaus2m5/6502_65C02_functional_tests
	m.loadBinary("testdata/6502_functional_test.bin")

	s.Reg.setPC(0x0400)
	for true {
		testCase := s.Mem.Peek(0x0200)
		if testCase >= 240 {
			break
		}
		log := testCase > 43
		if log {
			fmt.Printf("[ %d ] ", testCase)
		}
		pc := s.Reg.getPC()
		ExecuteInstruction(&s, log)
		if pc == s.Reg.getPC() {
			t.Errorf("Failure in test %v.", testCase)
		}
	}
}
