package core6502

import (
	"fmt"
	"testing"
)

func TestNMOS6502(t *testing.T) {
	m := new(FlatMemory)
	s := NewNMOS6502(m)

	m.loadBinary("testdata/6502_functional_test.bin")
	executeSuite(t, s, 0x200, 240, false, 255)
}

// To execute test suites from https://github.com/Klaus2m5/6502_65C02_functional_tests
func executeSuite(t *testing.T, s *State, stepAddress uint16, steps uint8, showStep bool, traceCPUStep uint8) {
	s.reg.setPC(0x0400)
	currentStep := uint8(255)
	for {
		testCase := s.mem.Peek(stepAddress)
		if testCase != currentStep {
			currentStep = testCase
			if showStep {
				fmt.Printf("[ Step %d ]\n", testCase)
			}
			s.SetTrace(testCase == traceCPUStep)
		}
		if testCase >= steps {
			break
		}
		pc := s.reg.getPC()
		s.ExecuteInstruction()
		if pc == s.reg.getPC() {
			t.Fatalf("Failure in test %v.", testCase)
		}
	}
}
