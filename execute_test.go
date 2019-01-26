package main

import (
	"testing"
)

func TestLDA(t *testing.T) {
	var s state

	opLDAimm(&s, []uint8{0xA9, 0x42})
	if s.registers.getA() != 0x42 {
		t.Error("Error in LDA #")
	}

	opLDYimm(&s, []uint8{0xA0, 0xFE})
	if s.registers.getY() != 0xFE {
		t.Error("Error in LDY #")
	}

	s.memory[0x38] = 0x87
	opLDAzpg(&s, []uint8{0xA5, 0x38})
	if s.registers.getA() != 0x87 {
		t.Error("Error in LDA zpg")
	}

	s.memory[0x57] = 0x90
	s.registers.setX(0x10)
	opLDAzpgX(&s, []uint8{0xB5, 0x47})
	if s.registers.getA() != 0x90 {
		t.Error("Error in LDA zpgX")
	}

	s.memory[0x38] = 0x12
	s.registers.setX(0x89)
	opLDAzpgX(&s, []uint8{0xB5, 0xAF})
	if s.registers.getA() != 0x12 {
		t.Error("Error in LDA zpgX with sero page overflow")
	}

	s.memory[0x1234] = 0x67
	opLDAabs(&s, []uint8{0xAD, 0x34, 0x12})
	if s.registers.getA() != 0x67 {
		t.Error("Error in LDA abs")
	}

}

func TestLDA2(t *testing.T) {
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
		t.Error("Error in LDA zpgX")
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

}
