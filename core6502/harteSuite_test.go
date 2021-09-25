package core6502

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

type scenarioState struct {
	Pc  uint16
	S   uint8
	A   uint8
	X   uint8
	Y   uint8
	P   uint8
	Ram [][]uint16
}

type scenario struct {
	Name    string
	Initial scenarioState
	Final   scenarioState
	Cycles  [][]interface{}
}

/*
	Tests from https://github.com/TomHarte/ProcessorTests/tree/main/6502/v1
	more work needed.
*/
func TestHarteNMOS6502(t *testing.T) {
	t.Skip("Not ready to be used in CI")

	s := NewNMOS6502(nil) // Use to get the opcodes names

	path := "/home/casa/code/ProcessorTests/6502/v1/"
	for i := 0x00; i <= 0xff; /*0xff*/ i++ {
		if s.opcodes[i].name != "ADC" && // Issue with ADC crossing page boundaries
			s.opcodes[i].name != "SBC" && // Issue with ADC crossing page boundaries
			s.opcodes[i].name != "" {

			opcode := fmt.Sprintf("%02x", i)
			t.Run(opcode+s.opcodes[i].name, func(t *testing.T) {
				t.Parallel()
				m := new(FlatMemory)
				s := NewNMOS6502(m)
				testOpcode(t, s, path, opcode)
			})
		}
	}
}

func TestHarteCMOS65c02(t *testing.T) {
	t.Skip("Not ready to be used in CI")

	s := NewCMOS65c02(nil) // Use to get the opcodes names

	path := "/home/casa/code/ProcessorTests/wdc65c02/v1/"
	for i := 0x00; i <= 0xff; /*0xff*/ i++ {
		if s.opcodes[i].name != "ADC" && // Issue with ADC crossing page boundaries
			s.opcodes[i].name != "SBC" && // Issue with SBC crossing page boundaries
			s.opcodes[i].name != "" {

			opcode := fmt.Sprintf("%02x", i)
			t.Run(opcode+s.opcodes[i].name, func(t *testing.T) {
				t.Parallel()
				m := new(FlatMemory)
				s := NewCMOS65c02(m)
				testOpcode(t, s, path, opcode)
			})
		}
	}
}

func testOpcode(t *testing.T, s *State, path string, opcode string) {
	data, err := os.ReadFile(path + opcode + ".json")
	if err != nil {
		t.Fatal(err)
	}

	if len(data) == 0 {
		return
	}

	var scenarios []scenario
	err = json.Unmarshal(data, &scenarios)
	if err != nil {
		t.Fatal(err)
	}

	for _, scenario := range scenarios {
		if scenario.Name != "20 55 13" { // Skip JSR on the stack being modified
			t.Run(scenario.Name, func(t *testing.T) {
				testScenario(t, s, &scenario)
			})
		}
	}
}

func testScenario(t *testing.T, s *State, sc *scenario) {
	// Setup CPU
	s.reg.setPC(sc.Initial.Pc)
	s.reg.setSP(sc.Initial.S)
	s.reg.setA(sc.Initial.A)
	s.reg.setX(sc.Initial.X)
	s.reg.setY(sc.Initial.Y)
	s.reg.setP(sc.Initial.P)
	for _, e := range sc.Initial.Ram {
		s.mem.Poke(uint16(e[0]), uint8(e[1]))
	}

	// Execute instruction
	s.ExecuteInstruction()

	// Check result
	assertReg8(t, sc, "A", s.reg.getA(), sc.Final.A)
	assertReg8(t, sc, "X", s.reg.getX(), sc.Final.X)
	assertReg8(t, sc, "Y", s.reg.getY(), sc.Final.Y)
	assertFlags(t, sc, sc.Initial.P, s.reg.getP(), sc.Final.P)
	assertReg8(t, sc, "SP", s.reg.getSP(), sc.Final.S)
	assertReg16(t, sc, "PC", s.reg.getPC(), sc.Final.Pc)
}

func assertReg8(t *testing.T, sc *scenario, name string, actual uint8, wanted uint8) {
	if actual != wanted {
		t.Errorf("Register %s is %02x and should be %02x for %+v", name, actual, wanted, sc)
	}
}

func assertReg16(t *testing.T, sc *scenario, name string, actual uint16, wanted uint16) {
	if actual != wanted {
		t.Errorf("Register %s is %04x and should be %04x for %+v", name, actual, wanted, sc)
	}
}

func assertFlags(t *testing.T, sc *scenario, initial uint8, actual uint8, wanted uint8) {
	if actual != wanted {
		t.Errorf("%08b flag diffs, they are %08b and should be %08b, initial %08b for %+v", actual^wanted, actual, wanted, initial, sc)
	}
}
