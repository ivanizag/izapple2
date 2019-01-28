package main

import (
	"testing"
)

func TestLoad(t *testing.T) {
	var s state

	executeLine(&s, []uint8{0xA9, 0x42})
	if s.registers.getA() != 0x42 {
		t.Error("Error in LDA #")
	}

	executeLine(&s, []uint8{0xA9, 0x00})
	if s.registers.getP() != flagZ {
		t.Error("Error in flags for LDA $0")
	}

	executeLine(&s, []uint8{0xA9, 0xF0})
	if s.registers.getP() != flagN {
		t.Error("Error in flags for LDA $F0")
	}

	executeLine(&s, []uint8{0xA0, 0xFE})
	if s.registers.getY() != 0xFE {
		t.Error("Error in LDY #")
	}

	s.memory[0x38] = 0x87
	executeLine(&s, []uint8{0xA5, 0x38})
	if s.registers.getA() != 0x87 {
		t.Error("Error in LDA zpg")
	}

	s.memory[0x57] = 0x90
	s.registers.setX(0x10)
	executeLine(&s, []uint8{0xB5, 0x47})
	if s.registers.getA() != 0x90 {
		t.Error("Error in LDA zpg, X")
	}

	s.memory[0x38] = 0x12
	s.registers.setX(0x89)
	executeLine(&s, []uint8{0xB5, 0xAF})
	if s.registers.getA() != 0x12 {
		t.Error("Error in LDA zpgX with sero page overflow")
	}

	s.memory[0x1234] = 0x67
	executeLine(&s, []uint8{0xAD, 0x34, 0x12})
	if s.registers.getA() != 0x67 {
		t.Error("Error in LDA abs")
	}

	s.memory[0xC057] = 0x7E
	s.registers.setX(0x57)
	executeLine(&s, []uint8{0xBD, 0x00, 0xC0})
	if s.registers.getA() != 0x7E {
		t.Error("Error in LDA abs, X")
	}

	s.memory[0xD059] = 0x7A
	s.registers.setY(0x59)
	executeLine(&s, []uint8{0xB9, 0x00, 0xD0})
	if s.registers.getA() != 0x7A {
		t.Error("Error in LDA abs, Y")
	}

	s.memory[0x24] = 0x74
	s.memory[0x25] = 0x20
	s.registers.setX(0x04)
	s.memory[0x2074] = 0x66
	executeLine(&s, []uint8{0xA1, 0x20})
	if s.registers.getA() != 0x66 {
		t.Error("Error in LDA (oper,X)")
	}

	s.memory[0x86] = 0x28
	s.memory[0x87] = 0x40
	s.registers.setY(0x10)
	s.memory[0x4038] = 0x99
	executeLine(&s, []uint8{0xB1, 0x86})
	if s.registers.getA() != 0x99 {
		t.Error("Error in LDA (oper),Y")
	}
}

func TestTransfer(t *testing.T) {
	var s state

	s.registers.setA(0xB0)
	executeLine(&s, []uint8{0xAA})
	if s.registers.getX() != 0xB0 {
		t.Error("Error in TAX")
	}
	if s.registers.getP() != flagN {
		t.Error("Error in TAX flags")
	}

	s.registers.setA(0xB1)
	executeLine(&s, []uint8{0xA8})
	if s.registers.getY() != 0xB1 {
		t.Error("Error in TAY")
	}

	s.registers.setSP(0xB2)
	executeLine(&s, []uint8{0xBA})
	if s.registers.getX() != 0xB2 {
		t.Error("Error in TSX")
	}

	s.registers.setX(0xB3)
	executeLine(&s, []uint8{0x8A})
	if s.registers.getA() != 0xB3 {
		t.Error("Error in TXA")
	}

	s.registers.setX(0xB4)
	executeLine(&s, []uint8{0x9A})
	if s.registers.getSP() != 0xB4 {
		t.Error("Error in TXS")
	}
	s.registers.setY(0xB5)
	executeLine(&s, []uint8{0x98})
	if s.registers.getA() != 0xB5 {
		t.Error("Error in TYA")
	}
}

func TestIncDec(t *testing.T) {
	var s state

	s.registers.setX(0x7E)
	executeLine(&s, []uint8{0xE8})
	if s.registers.getX() != 0x7F {
		t.Errorf("Error in INX")
	}

	s.registers.setY(0xFC)
	executeLine(&s, []uint8{0x88})
	if s.registers.getY() != 0xFB {
		t.Error("Error in DEY")
	}
	if s.registers.getP() != flagN {
		t.Error("Error in DEY flags")
	}
}

func TestRotate(t *testing.T) {
	var s state

	s.registers.setA(0xF0)
	executeLine(&s, []uint8{0x2A})
	if s.registers.getA() != 0xE0 {
		t.Errorf("Error in ROL")
	}
	if !s.registers.getFlag(flagC) {
		t.Errorf("Error in ROL carry. %v", s.registers)
	}

	s.registers.setFlag(flagC)
	s.registers.setA(0x0F)
	executeLine(&s, []uint8{0x6A})
	if s.registers.getA() != 0x87 {
		t.Errorf("Error in ROR. %v", s.registers)
	}
	if !s.registers.getFlag(flagC) {
		t.Errorf("Error in ROR carry")
	}

}
