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

func TestStore(t *testing.T) {
	var s state
	s.registers.setA(0x10)
	s.registers.setX(0x40)
	s.registers.setY(0x80)

	executeLine(&s, []uint8{0x85, 0x50})
	if s.memory[0x0050] != 0x10 {
		t.Error("Error in STA zpg")
	}

	executeLine(&s, []uint8{0x86, 0x51})
	if s.memory[0x0051] != 0x40 {
		t.Error("Error in STX zpg")
	}

	executeLine(&s, []uint8{0x84, 0x52})
	if s.memory[0x0052] != 0x80 {
		t.Error("Error in STY zpg")
	}

	executeLine(&s, []uint8{0x8D, 0x20, 0xC0})
	if s.memory[0xC020] != 0x10 {
		t.Error("Error in STA abs")
	}

	executeLine(&s, []uint8{0x9D, 0x08, 0x10})
	if s.memory[0x1048] != 0x10 {
		t.Error("Error in STA abs, X")
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

func TestShiftRotate(t *testing.T) {
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

	s.registers.setFlag(flagC)
	s.registers.setA(0x81)
	executeLine(&s, []uint8{0x0A})
	if s.registers.getA() != 0x02 {
		t.Errorf("Error in ASL. %v", s.registers)
	}
	if !s.registers.getFlag(flagC) {
		t.Errorf("Error in ASL carry")
	}

	s.registers.setFlag(flagC)
	s.registers.setA(0x02)
	executeLine(&s, []uint8{0x4A})
	if s.registers.getA() != 0x01 {
		t.Errorf("Error in LSR. %v", s.registers)
	}
	if s.registers.getFlag(flagC) {
		t.Errorf("Error in LSR carry")
	}
}

func TestClearSetFlag(t *testing.T) {
	var s state
	s.registers.setP(0x00)

	executeLine(&s, []uint8{0xF8})
	if !s.registers.getFlag(flagD) {
		t.Errorf("Error in SED. %v", s.registers)
	}

	executeLine(&s, []uint8{0xD8})
	if s.registers.getFlag(flagD) {
		t.Errorf("Error in CLD. %v", s.registers)
	}

}

func TestLogic(t *testing.T) {
	var s state

	s.registers.setA(0xF0)
	executeLine(&s, []uint8{0x29, 0x1C})
	if s.registers.getA() != 0x10 {
		t.Errorf("Error in AND <. %v", s.registers)
	}

	s.registers.setA(0xF0)
	executeLine(&s, []uint8{0x49, 0x1C})
	if s.registers.getA() != 0xEC {
		t.Errorf("Error in EOR <. %v", s.registers)
	}

	s.registers.setA(0xF0)
	executeLine(&s, []uint8{0x09, 0x0C})
	if s.registers.getA() != 0xFC {
		t.Errorf("Error in ORA <. %v", s.registers)
	}
}

func TestAdd(t *testing.T) {
	var s state

	s.registers.setA(0xA0)
	s.registers.clearFlag(flagC)
	executeLine(&s, []uint8{0x69, 0x0B})
	if s.registers.getA() != 0xAB {
		t.Errorf("Error in ADC A0 + 0B. %v", s.registers)
	}
	if s.registers.getFlag(flagC) {
		t.Errorf("Error in carry ADC A0 + 0B. %v", s.registers)
	}

	s.registers.setA(0xFF)
	s.registers.clearFlag(flagC)
	executeLine(&s, []uint8{0x69, 0x02})
	if s.registers.getA() != 0x01 {
		t.Errorf("Error in ADC A0 + 0B with carry. %v", s.registers)
	}
	if !s.registers.getFlag(flagC) {
		t.Errorf("Error in carry ADC A0 + 0B with carry. %v", s.registers)
	}

	s.registers.setA(0xA0)
	s.registers.setFlag(flagC)
	executeLine(&s, []uint8{0x69, 0x01})
	if s.registers.getA() != 0xA2 {
		t.Errorf("Error in ADC C + A0 + 0B with carry. %v", s.registers)
	}
	if s.registers.getFlag(flagC) {
		t.Errorf("Error in carry ADC C + A0 + 0B with carry. %v", s.registers)
	}
}

func TestSub(t *testing.T) {
	var s state

	s.registers.setA(0x09)
	s.registers.clearFlag(flagC)
	executeLine(&s, []uint8{0xE9, 0x05})
	if s.registers.getA() != 0x04 {
		t.Errorf("Error in SBC A0 + 0B. %v", s.registers)
	}
	if s.registers.getFlag(flagC) {
		t.Errorf("Error in carry SBC A0 + 0B. %v", s.registers)
	}

	s.registers.setA(0x01)
	s.registers.clearFlag(flagC)
	executeLine(&s, []uint8{0xE9, 0x02})
	if s.registers.getA() != 0xFF {
		t.Errorf("Error in SBC A0 + 0B with carry. %v", s.registers)
	}
	if !s.registers.getFlag(flagC) {
		t.Errorf("Error in carry SBC A0 + 0B with carry. %v", s.registers)
	}

	s.registers.setA(0x08)
	s.registers.setFlag(flagC)
	executeLine(&s, []uint8{0xE9, 0x02})
	if s.registers.getA() != 0x05 {
		t.Errorf("Error in SBC C + A0 + 0B with carry. %v", s.registers)
	}
	if s.registers.getFlag(flagC) {
		t.Errorf("Error in carry SBC C + A0 + 0B with carry. %v", s.registers)
	}

}

func TestCompare(t *testing.T) {
	var s state

	s.registers.setA(0x02)
	executeLine(&s, []uint8{0xC9, 0x01})
	if s.registers.getP() != 0x01 {
		t.Errorf("Error in CMP <. %v", s.registers)
	}

	executeLine(&s, []uint8{0xC9, 0x02})
	if s.registers.getP() != 0x03 {
		t.Errorf("Error in CMP =. %v", s.registers)
	}

	executeLine(&s, []uint8{0xC9, 0x03})
	if s.registers.getP() != 0x80 {
		t.Errorf("Error in CMP >. %v", s.registers)
	}

	s.registers.setX(0x04)
	executeLine(&s, []uint8{0xE0, 0x05})
	if s.registers.getP() != 0x80 {
		t.Errorf("Error in CPX >. %v", s.registers)
	}

	s.registers.setY(0x08)
	executeLine(&s, []uint8{0xC0, 0x09})
	if s.registers.getP() != 0x80 {
		t.Errorf("Error in CPY >. %v", s.registers)
	}

}
func TestBit(t *testing.T) {
	var s state

	s.registers.setA(0x0F)
	s.memory[0x0040] = 0xF0
	executeLine(&s, []uint8{0x24, 0x40})
	if s.registers.getP() != 0xC2 {
		t.Errorf("Error in BIT. %v", s.registers)
	}

	s.registers.setA(0xF0)
	s.memory[0x0040] = 0xF0
	executeLine(&s, []uint8{0x24, 0x40})
	if s.registers.getP() != 0xC0 {
		t.Errorf("Error in BIT, 2. %v", s.registers)
	}

	s.registers.setA(0xF0)
	s.memory[0x01240] = 0x80
	executeLine(&s, []uint8{0x2C, 0x40, 0x12})
	if s.registers.getP() != 0x80 {
		t.Errorf("Error in BIT, 2. %v", s.registers)
	}
}

func TestBranch(t *testing.T) {
	var s state
	s.registers.setPC(0xC600)
	s.registers.setFlag(flagV)
	executeLine(&s, []uint8{0x50, 0x20})
	if s.registers.getPC() != 0xC600 {
		t.Errorf("Error in BVC, %v", s.registers)
	}

	executeLine(&s, []uint8{0x70, 0x20})
	if s.registers.getPC() != 0xC620 {
		t.Errorf("Error in BVS, %v", s.registers)
	}

	s.registers.setPC(0xD600)
	s.registers.clearFlag(flagC)
	executeLine(&s, []uint8{0x90, 0xA0})
	if s.registers.getPC() != 0xD5A0 {
		t.Errorf("Error in BCC, %v", s.registers)
	}
}

func TestStack(t *testing.T) {
	var s state

	s.registers.setSP(0xF0)
	s.registers.setA(0xA0)
	s.registers.setP(0x0A)
	executeLine(&s, []uint8{0x48})
	if s.registers.getSP() != 0xEF {
		t.Errorf("Error in PHA stack pointer, %v", s.registers)
	}
	if s.memory[0x01F0] != 0xA0 {
		t.Errorf("Error in PHA, %v", s.registers)
	}

	executeLine(&s, []uint8{0x08})
	if s.registers.getSP() != 0xEE {
		t.Errorf("Error in PHP stack pointer, %v", s.registers)
	}
	if s.memory[0x01EF] != 0x3A {
		t.Errorf("Error in PHP, %v", s.registers)
	}

	executeLine(&s, []uint8{0x68})
	if s.registers.getSP() != 0xEF {
		t.Errorf("Error in PLA stack pointer, %v", s.registers)
	}
	if s.registers.getA() != 0x3A {
		t.Errorf("Error in PLA, %v", s.registers)
	}

	executeLine(&s, []uint8{0x28})
	if s.registers.getSP() != 0xF0 {
		t.Errorf("Error in PLP stack pointer, %v", s.registers)
	}
	if s.registers.getP() != 0xA0 {
		t.Errorf("Error in PLP, %v", s.registers)
	}
}

// TODO: Tests for BRK, JMP, JSR, RTI, RTS
