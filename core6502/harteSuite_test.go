package core6502

/*
	Tests from https://github.com/TomHarte/ProcessorTests

	Know issues:
		- Test 6502/v1/20_55_13 (Note 1)
		- Not implemented undocumented opcodes for NMOS (Note 2)
		- Errors on flag N for ADC in BCD mode (Note 3)

	The tests are disabled by defaut because they take long to run
	and require a huge download.
	To enable them, clone the repo https://github.com/TomHarte/ProcessorTests
	and change the variables ProcessorTestsEnable and ProcessorTestsPath.
*/

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
)

var ProcessorTestsEnable = false
var ProcessorTestsPath = "/home/casa/code/ProcessorTests/"

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

func TestHarteNMOS6502(t *testing.T) {
	if !ProcessorTestsEnable {
		t.Skip("TomHarte/ProcessorTests are not enabled")
	}

	s := NewNMOS6502(nil) // Use to get the opcodes names

	path := ProcessorTestsPath + "6502/v1/"
	for i := 0x00; i <= 0xff; i++ {
		mnemonic := s.opcodes[i].name
		if mnemonic != "" { // Note 2
			opcode := fmt.Sprintf("%02x", i)
			t.Run(opcode+mnemonic, func(t *testing.T) {
				t.Parallel()
				m := new(FlatMemory)
				s := NewNMOS6502(m)
				testOpcode(t, s, path, opcode, mnemonic)
			})
			//} else {
			//	opcode := fmt.Sprintf("%02x", i)
			//	t.Run(opcode+mnemonic, func(t *testing.T) {
			//		t.Error("Opcode not implemented")
			//	})
		}
	}
}

func TestHarteCMOS65c02(t *testing.T) {
	if !ProcessorTestsEnable {
		t.Skip("TomHarte/ProcessorTests are not enabled")
	}

	s := NewCMOS65c02(nil) // Use to get the opcodes names

	path := ProcessorTestsPath + "wdc65c02/v1/"
	for i := 0x00; i <= 0xff; i++ {
		mnemonic := s.opcodes[i].name
		opcode := fmt.Sprintf("%02x", i)
		t.Run(opcode+mnemonic, func(t *testing.T) {
			t.Parallel()
			m := new(FlatMemory)
			s := NewCMOS65c02(m)
			testOpcode(t, s, path, opcode, mnemonic)
		})
	}
}

func testOpcode(t *testing.T, s *State, path string, opcode string, mnemonic string) {
	data, err := ioutil.ReadFile(path + opcode + ".json")
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
		if scenario.Name != "20 55 13" { // Note 1
			t.Run(scenario.Name, func(t *testing.T) {
				testScenario(t, s, &scenario, mnemonic)
			})
		}
	}
}

func testScenario(t *testing.T, s *State, sc *scenario, mnemonic string) {
	// Setup CPU
	start := s.GetCycles()
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
	if s.reg.getFlag(flagD) && (mnemonic == "ADC") {
		// Note 3
		assertFlags(t, sc, sc.Initial.P, s.reg.getP()&0x7f, sc.Final.P&0x7f)
	} else {
		assertFlags(t, sc, sc.Initial.P, s.reg.getP(), sc.Final.P)
	}
	assertReg8(t, sc, "SP", s.reg.getSP(), sc.Final.S)
	assertReg16(t, sc, "PC", s.reg.getPC(), sc.Final.Pc)

	cycles := s.GetCycles() - start
	if cycles != uint64(len(sc.Cycles)) {
		t.Errorf("Took %v cycles, it should be %v for %+v", cycles, len(sc.Cycles), sc)
	}
}

func assertReg8(t *testing.T, sc *scenario, name string, actual uint8, wanted uint8) {
	if actual != wanted {
		t.Errorf("Register %s is $%02x and should be $%02x for %+v", name, actual, wanted, sc)
	}
}

func assertReg16(t *testing.T, sc *scenario, name string, actual uint16, wanted uint16) {
	if actual != wanted {
		t.Errorf("Register %s is $%04x and should be $%04x for %+v", name, actual, wanted, sc)
	}
}

func assertFlags(t *testing.T, sc *scenario, initial uint8, actual uint8, wanted uint8) {
	if actual != wanted {
		t.Errorf("%08b flag diffs, they are %08b and should be %08b, initial %08b for %+v", actual^wanted, actual, wanted, initial, sc)
	}
}
