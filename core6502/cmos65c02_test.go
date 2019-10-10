package core6502

import (
	"testing"
)

func TestCMOs65c02NoUndocumented(t *testing.T) {
	m := new(FlatMemory)
	s := NewCMOS65c02(m)

	for i := 0; i < 256; i++ {
		if s.opcodes[i].cycles == 0 {
			t.Errorf("Opcode missing for $%02x.", i)
		}
	}
}

func TestCMOS65c02asNMOS(t *testing.T) {
	m := new(FlatMemory)
	s := NewCMOS65c02(m)

	m.loadBinary("testdata/6502_functional_test.bin")
	executeSuite(t, s, 0x200, 240, false, 255)
}

func TestCMOS65c02(t *testing.T) {
	m := new(FlatMemory)
	s := NewCMOS65c02(m)

	m.loadBinary("testdata/65C02_extended_opcodes_test.bin")
	executeSuite(t, s, 0x202, 240, false, 255)
}
