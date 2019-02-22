package core6502

import (
	"testing"
)

func TestLoad(t *testing.T) {
	var s State
	s.Mem = new(FlatMemory)

	executeLine(&s, []uint8{0xA9, 0x42})
	if s.Reg.getA() != 0x42 {
		t.Error("Error in LDA #")
	}

	executeLine(&s, []uint8{0xA9, 0x00})
	if s.Reg.getP() != flagZ {
		t.Error("Error in flags for LDA $0")
	}

	executeLine(&s, []uint8{0xA9, 0xF0})
	if s.Reg.getP() != flagN {
		t.Error("Error in flags for LDA $F0")
	}

	executeLine(&s, []uint8{0xA0, 0xFE})
	if s.Reg.getY() != 0xFE {
		t.Error("Error in LDY #")
	}

	s.Mem.Poke(0x38, 0x87)
	executeLine(&s, []uint8{0xA5, 0x38})
	if s.Reg.getA() != 0x87 {
		t.Error("Error in LDA zpg")
	}

	s.Mem.Poke(0x57, 0x90)
	s.Reg.setX(0x10)
	executeLine(&s, []uint8{0xB5, 0x47})
	if s.Reg.getA() != 0x90 {
		t.Error("Error in LDA zpg, X")
	}

	s.Mem.Poke(0x38, 0x12)
	s.Reg.setX(0x89)
	executeLine(&s, []uint8{0xB5, 0xAF})
	if s.Reg.getA() != 0x12 {
		t.Error("Error in LDA zpgX with sero page overflow")
	}

	s.Mem.Poke(0x1234, 0x67)
	executeLine(&s, []uint8{0xAD, 0x34, 0x12})
	if s.Reg.getA() != 0x67 {
		t.Error("Error in LDA abs")
	}

	s.Mem.Poke(0xC057, 0x7E)
	s.Reg.setX(0x57)
	executeLine(&s, []uint8{0xBD, 0x00, 0xC0})
	if s.Reg.getA() != 0x7E {
		t.Error("Error in LDA abs, X")
	}

	s.Mem.Poke(0xD059, 0x7A)
	s.Reg.setY(0x59)
	executeLine(&s, []uint8{0xB9, 0x00, 0xD0})
	if s.Reg.getA() != 0x7A {
		t.Error("Error in LDA abs, Y")
	}

	s.Mem.Poke(0x24, 0x74)
	s.Mem.Poke(0x25, 0x20)
	s.Reg.setX(0x04)
	s.Mem.Poke(0x2074, 0x66)
	executeLine(&s, []uint8{0xA1, 0x20})
	if s.Reg.getA() != 0x66 {
		t.Error("Error in LDA (oper,X)")
	}

	s.Mem.Poke(0x86, 0x28)
	s.Mem.Poke(0x87, 0x40)
	s.Reg.setY(0x10)
	s.Mem.Poke(0x4038, 0x99)
	executeLine(&s, []uint8{0xB1, 0x86})
	if s.Reg.getA() != 0x99 {
		t.Error("Error in LDA (oper),Y")
	}
}

func TestStore(t *testing.T) {
	var s State
	s.Mem = new(FlatMemory)
	s.Reg.setA(0x10)
	s.Reg.setX(0x40)
	s.Reg.setY(0x80)

	executeLine(&s, []uint8{0x85, 0x50})
	if s.Mem.Peek(0x0050) != 0x10 {
		t.Error("Error in STA zpg")
	}

	executeLine(&s, []uint8{0x86, 0x51})
	if s.Mem.Peek(0x0051) != 0x40 {
		t.Error("Error in STX zpg")
	}

	executeLine(&s, []uint8{0x84, 0x52})
	if s.Mem.Peek(0x0052) != 0x80 {
		t.Error("Error in STY zpg")
	}

	executeLine(&s, []uint8{0x8D, 0x20, 0xC0})
	if s.Mem.Peek(0xC020) != 0x10 {
		t.Error("Error in STA abs")
	}

	executeLine(&s, []uint8{0x9D, 0x08, 0x10})
	if s.Mem.Peek(0x1048) != 0x10 {
		t.Error("Error in STA abs, X")
	}
}

func TestTransfer(t *testing.T) {
	var s State

	s.Reg.setA(0xB0)
	executeLine(&s, []uint8{0xAA})
	if s.Reg.getX() != 0xB0 {
		t.Error("Error in TAX")
	}
	if s.Reg.getP() != flagN {
		t.Error("Error in TAX flags")
	}

	s.Reg.setA(0xB1)
	executeLine(&s, []uint8{0xA8})
	if s.Reg.getY() != 0xB1 {
		t.Error("Error in TAY")
	}

	s.Reg.setSP(0xB2)
	executeLine(&s, []uint8{0xBA})
	if s.Reg.getX() != 0xB2 {
		t.Error("Error in TSX")
	}

	s.Reg.setX(0xB3)
	executeLine(&s, []uint8{0x8A})
	if s.Reg.getA() != 0xB3 {
		t.Error("Error in TXA")
	}

	s.Reg.setX(0xB4)
	executeLine(&s, []uint8{0x9A})
	if s.Reg.getSP() != 0xB4 {
		t.Error("Error in TXS")
	}

	s.Reg.setY(0xB5)
	executeLine(&s, []uint8{0x98})
	if s.Reg.getA() != 0xB5 {
		t.Error("Error in TYA")
	}
}

func TestIncDec(t *testing.T) {
	var s State

	s.Reg.setX(0x7E)
	executeLine(&s, []uint8{0xE8})
	if s.Reg.getX() != 0x7F {
		t.Errorf("Error in INX")
	}

	s.Reg.setY(0xFC)
	executeLine(&s, []uint8{0x88})
	if s.Reg.getY() != 0xFB {
		t.Error("Error in DEY")
	}
	if s.Reg.getP() != flagN {
		t.Error("Error in DEY flags")
	}
}

func TestShiftRotate(t *testing.T) {
	var s State

	s.Reg.setA(0xF0)
	executeLine(&s, []uint8{0x2A})
	if s.Reg.getA() != 0xE0 {
		t.Errorf("Error in ROL")
	}
	if !s.Reg.getFlag(flagC) {
		t.Errorf("Error in ROL carry. %v", s.Reg)
	}

	s.Reg.setFlag(flagC)
	s.Reg.setA(0x0F)
	executeLine(&s, []uint8{0x6A})
	if s.Reg.getA() != 0x87 {
		t.Errorf("Error in ROR. %v", s.Reg)
	}
	if !s.Reg.getFlag(flagC) {
		t.Errorf("Error in ROR carry")
	}

	s.Reg.setFlag(flagC)
	s.Reg.setA(0x81)
	executeLine(&s, []uint8{0x0A})
	if s.Reg.getA() != 0x02 {
		t.Errorf("Error in ASL. %v", s.Reg)
	}
	if !s.Reg.getFlag(flagC) {
		t.Errorf("Error in ASL carry")
	}

	s.Reg.setFlag(flagC)
	s.Reg.setA(0x02)
	executeLine(&s, []uint8{0x4A})
	if s.Reg.getA() != 0x01 {
		t.Errorf("Error in LSR. %v", s.Reg)
	}
	if s.Reg.getFlag(flagC) {
		t.Errorf("Error in LSR carry")
	}
}

func TestClearSetFlag(t *testing.T) {
	var s State
	s.Reg.setP(0x00)

	executeLine(&s, []uint8{0xF8})
	if !s.Reg.getFlag(flagD) {
		t.Errorf("Error in SED. %v", s.Reg)
	}

	executeLine(&s, []uint8{0xD8})
	if s.Reg.getFlag(flagD) {
		t.Errorf("Error in CLD. %v", s.Reg)
	}

}

func TestLogic(t *testing.T) {
	var s State

	s.Reg.setA(0xF0)
	executeLine(&s, []uint8{0x29, 0x1C})
	if s.Reg.getA() != 0x10 {
		t.Errorf("Error in AND <. %v", s.Reg)
	}

	s.Reg.setA(0xF0)
	executeLine(&s, []uint8{0x49, 0x1C})
	if s.Reg.getA() != 0xEC {
		t.Errorf("Error in EOR <. %v", s.Reg)
	}

	s.Reg.setA(0xF0)
	executeLine(&s, []uint8{0x09, 0x0C})
	if s.Reg.getA() != 0xFC {
		t.Errorf("Error in ORA <. %v", s.Reg)
	}
}

func TestAdd(t *testing.T) {
	var s State

	s.Reg.setA(0xA0)
	s.Reg.clearFlag(flagC)
	executeLine(&s, []uint8{0x69, 0x0B})
	if s.Reg.getA() != 0xAB {
		t.Errorf("Error in ADC. %v", s.Reg)
	}
	if s.Reg.getFlag(flagC) {
		t.Errorf("Error in carry ADC. %v", s.Reg)
	}

	s.Reg.setA(0xFF)
	s.Reg.clearFlag(flagC)
	executeLine(&s, []uint8{0x69, 0x02})
	if s.Reg.getA() != 0x01 {
		t.Errorf("Error in ADC with carry. %v", s.Reg)
	}
	if !s.Reg.getFlag(flagC) {
		t.Errorf("Error in carry ADC with carry. %v", s.Reg)
	}

	s.Reg.setA(0xA0)
	s.Reg.setFlag(flagC)
	executeLine(&s, []uint8{0x69, 0x01})
	if s.Reg.getA() != 0xA2 {
		t.Errorf("Error in carried ADC with carry. %v", s.Reg)
	}
	if s.Reg.getFlag(flagC) {
		t.Errorf("Error in carry in carried ADC with carry. %v", s.Reg)
	}
}

func TestAddDecimal(t *testing.T) {
	var s State
	s.Reg.setFlag(flagD)

	s.Reg.setA(0x12)
	s.Reg.clearFlag(flagC)
	executeLine(&s, []uint8{0x69, 0x013})
	if s.Reg.getA() != 0x25 {
		t.Errorf("Error in ADC decimal. %v", s.Reg)
	}
	if s.Reg.getFlag(flagC) {
		t.Errorf("Error in carry ADC. %v", s.Reg)
	}

	s.Reg.setA(0x44)
	s.Reg.clearFlag(flagC)
	executeLine(&s, []uint8{0x69, 0x68})
	if s.Reg.getA() != 0x12 {
		t.Errorf("Error in ADC decimal  with carry. %v", s.Reg)
	}
	if !s.Reg.getFlag(flagC) {
		t.Errorf("Error in carry ADC decimal with carry. %v", s.Reg)
	}

	s.Reg.setA(0x44)
	s.Reg.setFlag(flagC)
	executeLine(&s, []uint8{0x69, 0x23})
	if s.Reg.getA() != 0x68 {
		t.Errorf("Error in carried ADC decimal with carry. %v", s.Reg)
	}
	if s.Reg.getFlag(flagC) {
		t.Errorf("Error in carry in carried ADC decimal with carry. %v", s.Reg)
	}
}

func TestSub(t *testing.T) {
	var s State

	s.Reg.setA(0x09)
	s.Reg.clearFlag(flagC)
	executeLine(&s, []uint8{0xE9, 0x05})
	if s.Reg.getA() != 0x03 {
		t.Errorf("Error in SBC. %v", s.Reg)
	}
	if !s.Reg.getFlag(flagC) {
		t.Errorf("Error in carry SBC. %v", s.Reg)
	}

	s.Reg.setA(0x01)
	s.Reg.clearFlag(flagC)
	executeLine(&s, []uint8{0xE9, 0x02})
	if s.Reg.getA() != 0xFE {
		t.Errorf("Error in SBC with carry. %v", s.Reg)
	}
	if s.Reg.getFlag(flagC) {
		t.Errorf("Error in carry SBC with carry. %v", s.Reg)
	}

	s.Reg.setA(0x08)
	s.Reg.setFlag(flagC)
	executeLine(&s, []uint8{0xE9, 0x02})
	if s.Reg.getA() != 0x06 {
		t.Errorf("Error in carried SBC with carry. %v", s.Reg)
	}
	if !s.Reg.getFlag(flagC) {
		t.Errorf("Error in carry in carried SBC with carry. %v", s.Reg)
	}

}

func TestCompare(t *testing.T) {
	var s State

	s.Reg.setA(0x02)
	executeLine(&s, []uint8{0xC9, 0x01})
	if s.Reg.getP() != 0x01 {
		t.Errorf("Error in CMP <. %v", s.Reg)
	}

	executeLine(&s, []uint8{0xC9, 0x02})
	if s.Reg.getP() != 0x03 {
		t.Errorf("Error in CMP =. %v", s.Reg)
	}

	executeLine(&s, []uint8{0xC9, 0x03})
	if s.Reg.getP() != 0x80 {
		t.Errorf("Error in CMP >. %v", s.Reg)
	}

	s.Reg.setX(0x04)
	executeLine(&s, []uint8{0xE0, 0x05})
	if s.Reg.getP() != 0x80 {
		t.Errorf("Error in CPX >. %v", s.Reg)
	}

	s.Reg.setY(0x08)
	executeLine(&s, []uint8{0xC0, 0x09})
	if s.Reg.getP() != 0x80 {
		t.Errorf("Error in CPY >. %v", s.Reg)
	}

}
func TestBit(t *testing.T) {
	var s State
	s.Mem = new(FlatMemory)

	s.Reg.setA(0x0F)
	s.Mem.Poke(0x0040, 0xF0)
	executeLine(&s, []uint8{0x24, 0x40})
	if s.Reg.getP() != 0xC2 {
		t.Errorf("Error in BIT. %v", s.Reg)
	}

	s.Reg.setA(0xF0)
	s.Mem.Poke(0x0040, 0xF0)
	executeLine(&s, []uint8{0x24, 0x40})
	if s.Reg.getP() != 0xC0 {
		t.Errorf("Error in BIT, 2. %v", s.Reg)
	}

	s.Reg.setA(0xF0)
	s.Mem.Poke(0x01240, 0x80)
	executeLine(&s, []uint8{0x2C, 0x40, 0x12})
	if s.Reg.getP() != 0x80 {
		t.Errorf("Error in BIT, 2. %v", s.Reg)
	}
}

func TestBranch(t *testing.T) {
	var s State

	s.Reg.setPC(0xC600)
	s.Reg.setFlag(flagV)
	executeLine(&s, []uint8{0x50, 0x20})
	if s.Reg.getPC() != 0xC600 {
		t.Errorf("Error in BVC, %v", s.Reg)
	}

	executeLine(&s, []uint8{0x70, 0x20})
	if s.Reg.getPC() != 0xC620 {
		t.Errorf("Error in BVS, %v", s.Reg)
	}

	s.Reg.setPC(0xD600)
	s.Reg.clearFlag(flagC)
	executeLine(&s, []uint8{0x90, 0xA0})
	if s.Reg.getPC() != 0xD5A0 {
		t.Errorf("Error in BCC, %v", s.Reg)
	}
}

func TestStack(t *testing.T) {
	var s State
	s.Mem = new(FlatMemory)

	s.Reg.setSP(0xF0)
	s.Reg.setA(0xA0)
	s.Reg.setP(0x0A)
	executeLine(&s, []uint8{0x48})
	if s.Reg.getSP() != 0xEF {
		t.Errorf("Error in PHA stack pointer, %v", s.Reg)
	}
	if s.Mem.Peek(0x01F0) != 0xA0 {
		t.Errorf("Error in PHA, %v", s.Reg)
	}

	executeLine(&s, []uint8{0x08})
	if s.Reg.getSP() != 0xEE {
		t.Errorf("Error in PHP stack pointer, %v", s.Reg)
	}
	if s.Mem.Peek(0x01EF) != 0x3A {
		t.Errorf("Error in PHP, %v", s.Reg)
	}

	executeLine(&s, []uint8{0x68})
	if s.Reg.getSP() != 0xEF {
		t.Errorf("Error in PLA stack pointer, %v", s.Reg)
	}
	if s.Reg.getA() != 0x3A {
		t.Errorf("Error in PLA, %v", s.Reg)
	}

	executeLine(&s, []uint8{0x28})
	if s.Reg.getSP() != 0xF0 {
		t.Errorf("Error in PLP stack pointer, %v", s.Reg)
	}
	if s.Reg.getP() != 0xA0 {
		t.Errorf("Error in PLP, %v", s.Reg)
	}
}

// TODO: Tests for BRK, JMP, JSR, RTI, RTS
