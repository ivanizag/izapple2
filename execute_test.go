package main

import (
	"testing"
)

func TestLDA(t *testing.T) {
	var s state

	executeLine(&s, []uint8{0xA9, 0x42})
	if s.registers.getA() != 0x42 {
		t.Error("Error in LDA #")
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

}
