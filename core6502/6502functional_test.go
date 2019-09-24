package core6502

import (
	"fmt"
	"testing"
)

func TestFunctional(t *testing.T) {
	m := new(FlatMemory)
	s := NewNMOS6502(m)

	// Test suite from https://github.com/Klaus2m5/6502_65C02_functional_tests
	m.loadBinary("testdata/6502_functional_test.bin")

	s.reg.setPC(0x0400)
	for true {
		testCase := s.mem.Peek(0x0200)
		if testCase >= 240 {
			break
		}
		log := testCase > 43
		if log {
			fmt.Printf("[ %d ] ", testCase)
		}
		s.SetTrace(log)
		pc := s.reg.getPC()
		s.ExecuteInstruction()
		if pc == s.reg.getPC() {
			t.Errorf("Failure in test %v.", testCase)
		}
	}

}
