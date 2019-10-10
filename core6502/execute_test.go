package core6502

import (
	"testing"
)

func TestLoad(t *testing.T) {
	s := NewNMOS6502(new(FlatMemory))

	s.executeLine([]uint8{0xA9, 0x42})
	if s.reg.getA() != 0x42 {
		t.Error("Error in LDA #")
	}

	s.executeLine([]uint8{0xA9, 0x00})
	if s.reg.getP() != flagZ {
		t.Error("Error in flags for LDA $0")
	}

	s.executeLine([]uint8{0xA9, 0xF0})
	if s.reg.getP() != flagN {
		t.Error("Error in flags for LDA $F0")
	}

	s.executeLine([]uint8{0xA0, 0xFE})
	if s.reg.getY() != 0xFE {
		t.Error("Error in LDY #")
	}

	s.mem.Poke(0x38, 0x87)
	s.executeLine([]uint8{0xA5, 0x38})
	if s.reg.getA() != 0x87 {
		t.Error("Error in LDA zpg")
	}

	s.mem.Poke(0x57, 0x90)
	s.reg.setX(0x10)
	s.executeLine([]uint8{0xB5, 0x47})
	if s.reg.getA() != 0x90 {
		t.Error("Error in LDA zpg, X")
	}

	s.mem.Poke(0x38, 0x12)
	s.reg.setX(0x89)
	s.executeLine([]uint8{0xB5, 0xAF})
	if s.reg.getA() != 0x12 {
		t.Error("Error in LDA zpgX with sero page overflow")
	}

	s.mem.Poke(0x1234, 0x67)
	s.executeLine([]uint8{0xAD, 0x34, 0x12})
	if s.reg.getA() != 0x67 {
		t.Error("Error in LDA abs")
	}

	s.mem.Poke(0xC057, 0x7E)
	s.reg.setX(0x57)
	s.executeLine([]uint8{0xBD, 0x00, 0xC0})
	if s.reg.getA() != 0x7E {
		t.Error("Error in LDA abs, X")
	}

	s.mem.Poke(0xD059, 0x7A)
	s.reg.setY(0x59)
	s.executeLine([]uint8{0xB9, 0x00, 0xD0})
	if s.reg.getA() != 0x7A {
		t.Error("Error in LDA abs, Y")
	}

	s.mem.Poke(0x24, 0x74)
	s.mem.Poke(0x25, 0x20)
	s.reg.setX(0x04)
	s.mem.Poke(0x2074, 0x66)
	s.executeLine([]uint8{0xA1, 0x20})
	if s.reg.getA() != 0x66 {
		t.Error("Error in LDA (oper,X)")
	}

	s.mem.Poke(0x86, 0x28)
	s.mem.Poke(0x87, 0x40)
	s.reg.setY(0x10)
	s.mem.Poke(0x4038, 0x99)
	s.executeLine([]uint8{0xB1, 0x86})
	if s.reg.getA() != 0x99 {
		t.Error("Error in LDA (oper),Y")
	}
}

func TestStore(t *testing.T) {
	s := NewNMOS6502(new(FlatMemory))
	s.reg.setA(0x10)
	s.reg.setX(0x40)
	s.reg.setY(0x80)

	s.executeLine([]uint8{0x85, 0x50})
	if s.mem.Peek(0x0050) != 0x10 {
		t.Error("Error in STA zpg")
	}

	s.executeLine([]uint8{0x86, 0x51})
	if s.mem.Peek(0x0051) != 0x40 {
		t.Error("Error in STX zpg")
	}

	s.executeLine([]uint8{0x84, 0x52})
	if s.mem.Peek(0x0052) != 0x80 {
		t.Error("Error in STY zpg")
	}

	s.executeLine([]uint8{0x8D, 0x20, 0xC0})
	if s.mem.Peek(0xC020) != 0x10 {
		t.Error("Error in STA abs")
	}

	s.executeLine([]uint8{0x9D, 0x08, 0x10})
	if s.mem.Peek(0x1048) != 0x10 {
		t.Error("Error in STA abs, X")
	}
}

func TestTransfer(t *testing.T) {
	s := NewNMOS6502(new(FlatMemory))

	s.reg.setA(0xB0)
	s.executeLine([]uint8{0xAA})
	if s.reg.getX() != 0xB0 {
		t.Error("Error in TAX")
	}
	if s.reg.getP() != flagN {
		t.Error("Error in TAX flags")
	}

	s.reg.setA(0xB1)
	s.executeLine([]uint8{0xA8})
	if s.reg.getY() != 0xB1 {
		t.Error("Error in TAY")
	}

	s.reg.setSP(0xB2)
	s.executeLine([]uint8{0xBA})
	if s.reg.getX() != 0xB2 {
		t.Error("Error in TSX")
	}

	s.reg.setX(0xB3)
	s.executeLine([]uint8{0x8A})
	if s.reg.getA() != 0xB3 {
		t.Error("Error in TXA")
	}

	s.reg.setX(0xB4)
	s.executeLine([]uint8{0x9A})
	if s.reg.getSP() != 0xB4 {
		t.Error("Error in TXS")
	}

	s.reg.setY(0xB5)
	s.executeLine([]uint8{0x98})
	if s.reg.getA() != 0xB5 {
		t.Error("Error in TYA")
	}
}

func TestIncDec(t *testing.T) {
	s := NewNMOS6502(new(FlatMemory))

	s.reg.setX(0x7E)
	s.executeLine([]uint8{0xE8})
	if s.reg.getX() != 0x7F {
		t.Errorf("Error in INX")
	}

	s.reg.setY(0xFC)
	s.executeLine([]uint8{0x88})
	if s.reg.getY() != 0xFB {
		t.Error("Error in DEY")
	}
	if s.reg.getP() != flagN {
		t.Error("Error in DEY flags")
	}
}

func TestShiftRotate(t *testing.T) {
	s := NewNMOS6502(new(FlatMemory))

	s.reg.setA(0xF0)
	s.executeLine([]uint8{0x2A})
	if s.reg.getA() != 0xE0 {
		t.Errorf("Error in ROL")
	}
	if !s.reg.getFlag(flagC) {
		t.Errorf("Error in ROL carry. %v", s.reg)
	}

	s.reg.setFlag(flagC)
	s.reg.setA(0x0F)
	s.executeLine([]uint8{0x6A})
	if s.reg.getA() != 0x87 {
		t.Errorf("Error in ROR. %v", s.reg)
	}
	if !s.reg.getFlag(flagC) {
		t.Errorf("Error in ROR carry")
	}

	s.reg.setFlag(flagC)
	s.reg.setA(0x81)
	s.executeLine([]uint8{0x0A})
	if s.reg.getA() != 0x02 {
		t.Errorf("Error in ASL. %v", s.reg)
	}
	if !s.reg.getFlag(flagC) {
		t.Errorf("Error in ASL carry")
	}

	s.reg.setFlag(flagC)
	s.reg.setA(0x02)
	s.executeLine([]uint8{0x4A})
	if s.reg.getA() != 0x01 {
		t.Errorf("Error in LSR. %v", s.reg)
	}
	if s.reg.getFlag(flagC) {
		t.Errorf("Error in LSR carry")
	}
}

func TestClearSetFlag(t *testing.T) {
	s := NewNMOS6502(new(FlatMemory))
	s.reg.setP(0x00)

	s.executeLine([]uint8{0xF8})
	if !s.reg.getFlag(flagD) {
		t.Errorf("Error in SED. %v", s.reg)
	}

	s.executeLine([]uint8{0xD8})
	if s.reg.getFlag(flagD) {
		t.Errorf("Error in CLD. %v", s.reg)
	}

}

func TestLogic(t *testing.T) {
	s := NewNMOS6502(new(FlatMemory))

	s.reg.setA(0xF0)
	s.executeLine([]uint8{0x29, 0x1C})
	if s.reg.getA() != 0x10 {
		t.Errorf("Error in AND <. %v", s.reg)
	}

	s.reg.setA(0xF0)
	s.executeLine([]uint8{0x49, 0x1C})
	if s.reg.getA() != 0xEC {
		t.Errorf("Error in EOR <. %v", s.reg)
	}

	s.reg.setA(0xF0)
	s.executeLine([]uint8{0x09, 0x0C})
	if s.reg.getA() != 0xFC {
		t.Errorf("Error in ORA <. %v", s.reg)
	}
}

func TestAdd(t *testing.T) {
	s := NewNMOS6502(new(FlatMemory))

	s.reg.setA(0xA0)
	s.reg.clearFlag(flagC)
	s.executeLine([]uint8{0x69, 0x0B})
	if s.reg.getA() != 0xAB {
		t.Errorf("Error in ADC. %v", s.reg)
	}
	if s.reg.getFlag(flagC) {
		t.Errorf("Error in carry ADC. %v", s.reg)
	}

	s.reg.setA(0xFF)
	s.reg.clearFlag(flagC)
	s.executeLine([]uint8{0x69, 0x02})
	if s.reg.getA() != 0x01 {
		t.Errorf("Error in ADC with carry. %v", s.reg)
	}
	if !s.reg.getFlag(flagC) {
		t.Errorf("Error in carry ADC with carry. %v", s.reg)
	}

	s.reg.setA(0xA0)
	s.reg.setFlag(flagC)
	s.executeLine([]uint8{0x69, 0x01})
	if s.reg.getA() != 0xA2 {
		t.Errorf("Error in carried ADC with carry. %v", s.reg)
	}
	if s.reg.getFlag(flagC) {
		t.Errorf("Error in carry in carried ADC with carry. %v", s.reg)
	}
}

func TestAddDecimal(t *testing.T) {
	s := NewNMOS6502(new(FlatMemory))
	s.reg.setFlag(flagD)

	s.reg.setA(0x12)
	s.reg.clearFlag(flagC)
	s.executeLine([]uint8{0x69, 0x013})
	if s.reg.getA() != 0x25 {
		t.Errorf("Error in ADC decimal. %v", s.reg)
	}
	if s.reg.getFlag(flagC) {
		t.Errorf("Error in carry ADC. %v", s.reg)
	}

	s.reg.setA(0x44)
	s.reg.clearFlag(flagC)
	s.executeLine([]uint8{0x69, 0x68})
	if s.reg.getA() != 0x12 {
		t.Errorf("Error in ADC decimal  with carry. %v", s.reg)
	}
	if !s.reg.getFlag(flagC) {
		t.Errorf("Error in carry ADC decimal with carry. %v", s.reg)
	}

	s.reg.setA(0x44)
	s.reg.setFlag(flagC)
	s.executeLine([]uint8{0x69, 0x23})
	if s.reg.getA() != 0x68 {
		t.Errorf("Error in carried ADC decimal with carry. %v", s.reg)
	}
	if s.reg.getFlag(flagC) {
		t.Errorf("Error in carry in carried ADC decimal with carry. %v", s.reg)
	}
}

func TestSub(t *testing.T) {
	s := NewNMOS6502(new(FlatMemory))

	s.reg.setA(0x09)
	s.reg.clearFlag(flagC)
	s.executeLine([]uint8{0xE9, 0x05})
	if s.reg.getA() != 0x03 {
		t.Errorf("Error in SBC. %v", s.reg)
	}
	if !s.reg.getFlag(flagC) {
		t.Errorf("Error in carry SBC. %v", s.reg)
	}

	s.reg.setA(0x01)
	s.reg.clearFlag(flagC)
	s.executeLine([]uint8{0xE9, 0x02})
	if s.reg.getA() != 0xFE {
		t.Errorf("Error in SBC with carry. %v", s.reg)
	}
	if s.reg.getFlag(flagC) {
		t.Errorf("Error in carry SBC with carry. %v", s.reg)
	}

	s.reg.setA(0x08)
	s.reg.setFlag(flagC)
	s.executeLine([]uint8{0xE9, 0x02})
	if s.reg.getA() != 0x06 {
		t.Errorf("Error in carried SBC with carry. %v", s.reg)
	}
	if !s.reg.getFlag(flagC) {
		t.Errorf("Error in carry in carried SBC with carry. %v", s.reg)
	}

}

func TestCompare(t *testing.T) {
	s := NewNMOS6502(new(FlatMemory))

	s.reg.setA(0x02)
	s.executeLine([]uint8{0xC9, 0x01})
	if s.reg.getP() != 0x01 {
		t.Errorf("Error in CMP <. %v", s.reg)
	}

	s.executeLine([]uint8{0xC9, 0x02})
	if s.reg.getP() != 0x03 {
		t.Errorf("Error in CMP =. %v", s.reg)
	}

	s.executeLine([]uint8{0xC9, 0x03})
	if s.reg.getP() != 0x80 {
		t.Errorf("Error in CMP >. %v", s.reg)
	}

	s.reg.setX(0x04)
	s.executeLine([]uint8{0xE0, 0x05})
	if s.reg.getP() != 0x80 {
		t.Errorf("Error in CPX >. %v", s.reg)
	}

	s.reg.setY(0x08)
	s.executeLine([]uint8{0xC0, 0x09})
	if s.reg.getP() != 0x80 {
		t.Errorf("Error in CPY >. %v", s.reg)
	}

}
func TestBit(t *testing.T) {
	s := NewNMOS6502(new(FlatMemory))

	s.reg.setA(0x0F)
	s.mem.Poke(0x0040, 0xF0)
	s.executeLine([]uint8{0x24, 0x40})
	if s.reg.getP() != 0xC2 {
		t.Errorf("Error in BIT. %v", s.reg)
	}

	s.reg.setA(0xF0)
	s.mem.Poke(0x0040, 0xF0)
	s.executeLine([]uint8{0x24, 0x40})
	if s.reg.getP() != 0xC0 {
		t.Errorf("Error in BIT, 2. %v", s.reg)
	}

	s.reg.setA(0xF0)
	s.mem.Poke(0x01240, 0x80)
	s.executeLine([]uint8{0x2C, 0x40, 0x12})
	if s.reg.getP() != 0x80 {
		t.Errorf("Error in BIT, 2. %v", s.reg)
	}
}

func TestBranch(t *testing.T) {
	s := NewNMOS6502(new(FlatMemory))

	s.reg.setPC(0xC600)
	s.reg.setFlag(flagV)
	s.executeLine([]uint8{0x50, 0x20})
	if s.reg.getPC() != 0xC600 {
		t.Errorf("Error in BVC, %v", s.reg)
	}

	s.executeLine([]uint8{0x70, 0x20})
	if s.reg.getPC() != 0xC620 {
		t.Errorf("Error in BVS, %v", s.reg)
	}

	s.reg.setPC(0xD600)
	s.reg.clearFlag(flagC)
	s.executeLine([]uint8{0x90, 0xA0})
	if s.reg.getPC() != 0xD5A0 {
		t.Errorf("Error in BCC, %v", s.reg)
	}
}

func TestStack(t *testing.T) {
	s := NewNMOS6502(new(FlatMemory))

	s.reg.setSP(0xF0)
	s.reg.setA(0xA0)
	s.reg.setP(0x0A)
	s.executeLine([]uint8{0x48})
	if s.reg.getSP() != 0xEF {
		t.Errorf("Error in PHA stack pointer, %v", s.reg)
	}
	if s.mem.Peek(0x01F0) != 0xA0 {
		t.Errorf("Error in PHA, %v", s.reg)
	}

	s.executeLine([]uint8{0x08})
	if s.reg.getSP() != 0xEE {
		t.Errorf("Error in PHP stack pointer, %v", s.reg)
	}
	if s.mem.Peek(0x01EF) != 0x3A {
		t.Errorf("Error in PHP, %v", s.reg)
	}

	s.executeLine([]uint8{0x68})
	if s.reg.getSP() != 0xEF {
		t.Errorf("Error in PLA stack pointer, %v", s.reg)
	}
	if s.reg.getA() != 0x3A {
		t.Errorf("Error in PLA, %v", s.reg)
	}

	s.executeLine([]uint8{0x28})
	if s.reg.getSP() != 0xF0 {
		t.Errorf("Error in PLP stack pointer, %v", s.reg)
	}
	if s.reg.getP() != 0xA0 {
		t.Errorf("Error in PLP, %v", s.reg)
	}
}
