package main

import "fmt"

type state struct {
	registers registers
	memory    memory
}

func step(s *state) {

}

const (
	modeImplicit = iota + 1
	modeImplicitX
	modeImplicitY
	modeAccumulator
	modeImmediate
	modeZeroPage
	modeZeroPageX
	modeZeroPageY
	modeRelative
	modeAbsolute
	modeAbsoluteX
	modeAbsoluteY
	modeIndirect
	modeIndexedIndirectX
	modeIndirectIndexedY
)

// https://www.masswerk.at/6502/6502_instruction_set.html
// http://www.emulator101.com/reference/6502-reference.html
// https://www.csh.rit.edu/~moffitt/docs/6502.html#FLAGS
// https://ia800509.us.archive.org/18/items/Programming_the_6502/Programming_the_6502.pdf

func getWordInLine(line []uint8) uint16 {
	return uint16(line[1]) + 0x100*uint16(line[2])
}

func resolve(s *state, line []uint8, opcode opcode) (value uint8, address uint16, setValue func(uint8)) {
	hasAddress := true
	register := regNone

	switch opcode.addressMode {
	case modeAccumulator:
		value = s.registers.getA()
		hasAddress = false
		register = regA
	case modeImplicitX:
		value = s.registers.getX()
		hasAddress = false
		register = regX
	case modeImplicitY:
		value = s.registers.getY()
		hasAddress = false
		register = regY
	case modeImmediate:
		value = line[1]
		hasAddress = false
	case modeZeroPage:
		address = uint16(line[1])
	case modeZeroPageX:
		address = uint16(line[1] + s.registers.getX())
	case modeZeroPageY:
		address = uint16(line[1] + s.registers.getY())
	case modeAbsolute:
		address = getWordInLine(line)
	case modeAbsoluteX:
		address = getWordInLine(line) + uint16(s.registers.getX())
	case modeAbsoluteY:
		address = getWordInLine(line) + uint16(s.registers.getY())
	case modeIndexedIndirectX:
		addressAddress := uint8(line[1] + s.registers.getX())
		address = s.memory.getZeroPageWord(addressAddress)
	case modeIndirect:
		addressAddress := getWordInLine(line)
		address = s.memory.getWord(addressAddress)
	case modeIndirectIndexedY:
		address = s.memory.getZeroPageWord(line[1]) +
			uint16(s.registers.getY())
	}

	if hasAddress {
		value = s.memory.peek(address)
	}

	setValue = func(value uint8) {
		if hasAddress {
			s.memory.poke(address, value)
		} else if register != regNone {
			s.registers.setRegister(register, value)
		} else {
			// Todo: assert impossible
		}
	}
	return
}

type opcode struct {
	name        string
	bytes       uint8
	cycles      int
	addressMode int
	action      opFunc
}

type opFunc func(s *state, line []uint8, opcode opcode)

func buildOpTransfer(regSrc int, regDst int) opFunc {
	return func(s *state, line []uint8, opcode opcode) {
		value := s.registers.getRegister(regSrc)
		s.registers.setRegister(regDst, value)
		if regDst != regSP {
			s.registers.updateFlagZN(value)
		}
	}
}

func buildOpIncDec(inc bool) opFunc {
	return func(s *state, line []uint8, opcode opcode) {
		value, _, setValue := resolve(s, line, opcode)
		if inc {
			value++
		} else {
			value--
		}
		s.registers.updateFlagZN(value)
		setValue(value)
	}
}

func buildOpShift(isLeft bool, isRotate bool) opFunc {
	return func(s *state, line []uint8, opcode opcode) {
		value, _, setValue := resolve(s, line, opcode)

		oldCarry := s.registers.getFlagBit(flagC)
		var carry bool
		if isLeft {
			carry = (value & 0x80) != 0
			value <<= 1
			if isRotate {
				value += oldCarry
			}
		} else {
			carry = (value & 0x01) != 0
			value >>= 1
			if isRotate {
				value += oldCarry << 7
			}
		}
		s.registers.updateFlag(flagC, carry)
		s.registers.updateFlagZN(value)
		setValue(value)
	}
}

func buildOpLoad(regDst int) opFunc {
	return func(s *state, line []uint8, opcode opcode) {
		value, _, _ := resolve(s, line, opcode)
		s.registers.setRegister(regDst, value)
		s.registers.updateFlagZN(value)
	}
}

func buildOpStore(regSrc int) opFunc {
	return func(s *state, line []uint8, opcode opcode) {
		_, _, setValue := resolve(s, line, opcode)
		value := s.registers.getRegister(regSrc)
		setValue(value)
	}
}

func buildOpUpdateFlag(flag uint8, value bool) opFunc {
	return func(s *state, line []uint8, opcode opcode) {
		s.registers.updateFlag(flag, value)
	}
}

func buildOpBranch(flag uint8, value bool) opFunc {
	return func(s *state, line []uint8, opcode opcode) {
		if s.registers.getFlag(flag) == value {
			// This assumes that PC is already pointing to the next instruction
			pc := s.registers.getPC()
			pc += uint16(int8(line[1]))
			s.registers.setPC(pc)
		}
	}
}

func opBIT(s *state, line []uint8, opcode opcode) {
	value, _, _ := resolve(s, line, opcode)
	acc := s.registers.getA()
	// Future note: The immediate addressing mode (65C02 or 65816 only) does not affect V.
	s.registers.updateFlag(flagZ, value&acc == 0)
	s.registers.updateFlag(flagN, value&(1<<7) != 0)
	s.registers.updateFlag(flagV, value&(1<<6) != 0)
}

func buildOpCompare(reg int) opFunc {
	return func(s *state, line []uint8, opcode opcode) {
		value, _, _ := resolve(s, line, opcode)
		reference := s.registers.getRegister(reg)
		s.registers.updateFlagZN(reference - value)
		s.registers.updateFlag(flagC, reference >= value)
	}
}

func operationAnd(a uint8, b uint8) uint8 { return a & b }
func operationOr(a uint8, b uint8) uint8  { return a | b }
func operationXor(a uint8, b uint8) uint8 { return a ^ b }

func buildOpLogic(operation func(uint8, uint8) uint8) opFunc {
	return func(s *state, line []uint8, opcode opcode) {
		value, _, _ := resolve(s, line, opcode)
		result := operation(value, s.registers.getA())
		s.registers.setA(result)
		s.registers.updateFlagZN(result)
	}
}

func opADC(s *state, line []uint8, opcode opcode) {
	value, _, _ := resolve(s, line, opcode)
	aValue := s.registers.getA()
	carry := s.registers.getFlagBit(flagC)

	total := uint16(aValue) + uint16(value) + uint16(carry)
	signedTotal := int16(int8(aValue)) + int16(int8(value)) + int16(carry)
	truncated := uint8(total)

	if s.registers.getFlag(flagD) {
		totalBcdLo := int(aValue&0x0f) + int(value&0x0f) + int(carry)
		totalBcdHi := int(aValue>>4) + int(value>>4)
		if totalBcdLo >= 10 {
			totalBcdHi++
		}
		totalBcd := (totalBcdHi%10)<<4 + (totalBcdLo % 10)
		s.registers.setA(uint8(totalBcd))
		s.registers.updateFlag(flagC, totalBcdHi > 9)
	} else {
		s.registers.setA(truncated)
		s.registers.updateFlag(flagC, total > 0xFF)
	}

	// ZNV flags behave for BCD as if the operation was binary?
	s.registers.updateFlagZN(truncated)
	s.registers.updateFlag(flagV, signedTotal < -128 || signedTotal > 127)
}

func opSBC(s *state, line []uint8, opcode opcode) {
	value, _, _ := resolve(s, line, opcode)
	aValue := s.registers.getA()
	carry := s.registers.getFlagBit(flagC)

	total := 0x100 + uint16(aValue) - uint16(value) + uint16(carry) - 1
	signedTotal := int16(int8(aValue)) - int16(int8(value)) + int16(carry) - 1
	truncated := uint8(total)

	if s.registers.getFlag(flagD) {
		totalBcdLo := 10 + int(aValue&0x0f) - int(value&0x0f) + int(carry) - 1
		totalBcdHi := 10 + int(aValue>>4) - int(value>>4)
		if totalBcdLo < 10 {
			totalBcdHi--
		}
		totalBcd := (totalBcdHi%10)<<4 + (totalBcdLo % 10)
		s.registers.setA(uint8(totalBcd))
		s.registers.updateFlag(flagC, totalBcdHi >= 10)
	} else {
		s.registers.setA(truncated)
		s.registers.updateFlag(flagC, total > 0xFF)
	}

	// ZNV flags behave for SBC as if the operation was binary
	s.registers.updateFlagZN(truncated)
	s.registers.updateFlag(flagV, signedTotal < -128 || signedTotal > 127)
}

const stackAddress uint16 = 0x0100

func pushByte(s *state, value uint8) {
	adresss := stackAddress + uint16(s.registers.getSP())
	s.memory.poke(adresss, value)
	s.registers.setSP(s.registers.getSP() - 1)
}

func pullByte(s *state) uint8 {
	s.registers.setSP(s.registers.getSP() + 1)
	adresss := stackAddress + uint16(s.registers.getSP())
	return s.memory.peek(adresss)
}

func pushWord(s *state, value uint16) {
	pushByte(s, uint8(value>>8))
	pushByte(s, uint8(value))
}

func pullWord(s *state) uint16 {
	return uint16(pullByte(s)) +
		(uint16(pullByte(s)) << 8)

}

func opPLA(s *state, line []uint8, opcode opcode) {
	value := pullByte(s)
	s.registers.setA(value)
	s.registers.updateFlagZN(value)
}

func opPLP(s *state, line []uint8, opcode opcode) {
	value := pullByte(s)
	s.registers.setP(value)
}

func opPHA(s *state, line []uint8, opcode opcode) {
	pushByte(s, s.registers.getA())
}

func opPHP(s *state, line []uint8, opcode opcode) {
	pushByte(s, s.registers.getP()|(flagB+flag5))
}

func opJMP(s *state, line []uint8, opcode opcode) {
	_, address, _ := resolve(s, line, opcode)
	s.registers.setPC(address)
}

func opNOP(s *state, line []uint8, opcode opcode) {}

func opJSR(s *state, line []uint8, opcode opcode) {
	pushWord(s, s.registers.getPC()-1)
	_, address, _ := resolve(s, line, opcode)
	s.registers.setPC(address)
}

func opRTI(s *state, line []uint8, opcode opcode) {
	s.registers.setP(pullByte(s))
	s.registers.setPC(pullWord(s))
}

func opRTS(s *state, line []uint8, opcode opcode) {
	s.registers.setPC(pullWord(s) + 1)
}

func opBRK(s *state, line []uint8, opcode opcode) {
	pushWord(s, s.registers.getPC()+1)
	pushByte(s, s.registers.getP()|(flagB+flag5))
	s.registers.setFlag(flagI)
	s.registers.setPC(s.memory.getWord(0xFFFE))
}

var opcodes = [256]opcode{
	0x00: opcode{"BRK", 1, 7, modeImplicit, opBRK},
	0x4C: opcode{"JMP", 3, 3, modeAbsolute, opJMP},
	0x6C: opcode{"JMP", 3, 3, modeIndirect, opJMP},
	0x20: opcode{"JSR", 3, 6, modeAbsolute, opJSR},
	0x40: opcode{"RTI", 1, 6, modeImplicit, opRTI},
	0x60: opcode{"RTS", 1, 6, modeImplicit, opRTS},

	0x48: opcode{"PHA", 1, 3, modeImplicit, opPHA},
	0x08: opcode{"PHP", 1, 3, modeImplicit, opPHP},
	0x68: opcode{"PLA", 1, 4, modeImplicit, opPLA},
	0x28: opcode{"PLP", 1, 4, modeImplicit, opPLP},

	0x09: opcode{"ORA", 2, 2, modeImmediate, buildOpLogic(operationOr)},
	0x05: opcode{"ORA", 2, 3, modeZeroPage, buildOpLogic(operationOr)},
	0x15: opcode{"ORA", 2, 4, modeZeroPageX, buildOpLogic(operationOr)},
	0x0D: opcode{"ORA", 3, 4, modeAbsolute, buildOpLogic(operationOr)},
	0x1D: opcode{"ORA", 3, 4, modeAbsoluteX, buildOpLogic(operationOr)}, // Extra cycles
	0x19: opcode{"ORA", 3, 4, modeAbsoluteY, buildOpLogic(operationOr)}, // Extra cycles
	0x01: opcode{"ORA", 2, 6, modeIndexedIndirectX, buildOpLogic(operationOr)},
	0x11: opcode{"ORA", 2, 5, modeIndirectIndexedY, buildOpLogic(operationOr)}, // Extra cycles

	0x29: opcode{"AND", 2, 2, modeImmediate, buildOpLogic(operationAnd)},
	0x25: opcode{"AND", 2, 3, modeZeroPage, buildOpLogic(operationAnd)},
	0x35: opcode{"AND", 2, 4, modeZeroPageX, buildOpLogic(operationAnd)},
	0x2D: opcode{"AND", 3, 4, modeAbsolute, buildOpLogic(operationAnd)},
	0x3D: opcode{"AND", 3, 4, modeAbsoluteX, buildOpLogic(operationAnd)}, // Extra cycles
	0x39: opcode{"AND", 3, 4, modeAbsoluteY, buildOpLogic(operationAnd)}, // Extra cycles
	0x21: opcode{"AND", 2, 6, modeIndexedIndirectX, buildOpLogic(operationAnd)},
	0x31: opcode{"AND", 2, 5, modeIndirectIndexedY, buildOpLogic(operationAnd)}, // Extra cycles

	0x49: opcode{"EOR", 2, 2, modeImmediate, buildOpLogic(operationXor)},
	0x45: opcode{"EOR", 2, 3, modeZeroPage, buildOpLogic(operationXor)},
	0x55: opcode{"EOR", 2, 4, modeZeroPageX, buildOpLogic(operationXor)},
	0x4D: opcode{"EOR", 3, 4, modeAbsolute, buildOpLogic(operationXor)},
	0x5D: opcode{"EOR", 3, 4, modeAbsoluteX, buildOpLogic(operationXor)}, // Extra cycles
	0x59: opcode{"EOR", 3, 4, modeAbsoluteY, buildOpLogic(operationXor)}, // Extra cycles
	0x41: opcode{"EOR", 2, 6, modeIndexedIndirectX, buildOpLogic(operationXor)},
	0x51: opcode{"EOR", 2, 5, modeIndirectIndexedY, buildOpLogic(operationXor)}, // Extra cycles

	0x69: opcode{"ADC", 2, 2, modeImmediate, opADC},
	0x65: opcode{"ADC", 2, 3, modeZeroPage, opADC},
	0x75: opcode{"ADC", 2, 4, modeZeroPageX, opADC},
	0x6D: opcode{"ADC", 3, 4, modeAbsolute, opADC},
	0x7D: opcode{"ADC", 3, 4, modeAbsoluteX, opADC}, // Extra cycles
	0x79: opcode{"ADC", 3, 4, modeAbsoluteY, opADC}, // Extra cycles
	0x61: opcode{"ADC", 2, 6, modeIndexedIndirectX, opADC},
	0x71: opcode{"ADC", 2, 5, modeIndirectIndexedY, opADC}, // Extra cycles

	0xE9: opcode{"SBC", 2, 2, modeImmediate, opSBC},
	0xE5: opcode{"SBC", 2, 3, modeZeroPage, opSBC},
	0xF5: opcode{"SBC", 2, 4, modeZeroPageX, opSBC},
	0xED: opcode{"SBC", 3, 4, modeAbsolute, opSBC},
	0xFD: opcode{"SBC", 3, 4, modeAbsoluteX, opSBC}, // Extra cycles
	0xF9: opcode{"SBC", 3, 4, modeAbsoluteY, opSBC}, // Extra cycles
	0xE1: opcode{"SBC", 2, 6, modeIndexedIndirectX, opSBC},
	0xF1: opcode{"SBC", 2, 5, modeIndirectIndexedY, opSBC}, // Extra cycles

	0x24: opcode{"BIT", 2, 3, modeZeroPage, opBIT},
	0x2C: opcode{"BIT", 3, 3, modeAbsolute, opBIT},

	0xC9: opcode{"CMP", 2, 2, modeImmediate, buildOpCompare(regA)},
	0xC5: opcode{"CMP", 2, 3, modeZeroPage, buildOpCompare(regA)},
	0xD5: opcode{"CMP", 2, 4, modeZeroPageX, buildOpCompare(regA)},
	0xCD: opcode{"CMP", 3, 4, modeAbsolute, buildOpCompare(regA)},
	0xDD: opcode{"CMP", 3, 4, modeAbsoluteX, buildOpCompare(regA)}, // Extra cycles
	0xD9: opcode{"CMP", 3, 4, modeAbsoluteY, buildOpCompare(regA)}, // Extra cycles
	0xC1: opcode{"CMP", 2, 6, modeIndexedIndirectX, buildOpCompare(regA)},
	0xD1: opcode{"CMP", 2, 5, modeIndirectIndexedY, buildOpCompare(regA)}, // Extra cycles

	0xE0: opcode{"CPX", 2, 2, modeImmediate, buildOpCompare(regX)},
	0xE4: opcode{"CPX", 2, 3, modeZeroPage, buildOpCompare(regX)},
	0xEC: opcode{"CPX", 3, 4, modeAbsolute, buildOpCompare(regX)},

	0xC0: opcode{"CPY", 2, 2, modeImmediate, buildOpCompare(regY)},
	0xC4: opcode{"CPY", 2, 3, modeZeroPage, buildOpCompare(regY)},
	0xCC: opcode{"CPY", 3, 4, modeAbsolute, buildOpCompare(regY)},

	0x2A: opcode{"ROL", 1, 2, modeAccumulator, buildOpShift(true, true)},
	0x26: opcode{"ROL", 2, 5, modeZeroPage, buildOpShift(true, true)},
	0x36: opcode{"ROL", 2, 6, modeZeroPageX, buildOpShift(true, true)},
	0x2E: opcode{"ROL", 3, 6, modeAbsolute, buildOpShift(true, true)},
	0x3E: opcode{"ROL", 3, 7, modeAbsoluteX, buildOpShift(true, true)},

	0x6A: opcode{"ROR", 1, 2, modeAccumulator, buildOpShift(false, true)},
	0x66: opcode{"ROR", 2, 5, modeZeroPage, buildOpShift(false, true)},
	0x76: opcode{"ROR", 2, 6, modeZeroPageX, buildOpShift(false, true)},
	0x6E: opcode{"ROR", 3, 6, modeAbsolute, buildOpShift(false, true)},
	0x7E: opcode{"ROR", 3, 7, modeAbsoluteX, buildOpShift(false, true)},

	0x0A: opcode{"ASL", 1, 2, modeAccumulator, buildOpShift(true, false)},
	0x06: opcode{"ASL", 2, 5, modeZeroPage, buildOpShift(true, false)},
	0x16: opcode{"ASL", 2, 6, modeZeroPageX, buildOpShift(true, false)},
	0x0E: opcode{"ASL", 3, 6, modeAbsolute, buildOpShift(true, false)},
	0x1E: opcode{"ASL", 3, 7, modeAbsoluteX, buildOpShift(true, false)},

	0x4A: opcode{"LSR", 1, 2, modeAccumulator, buildOpShift(false, false)},
	0x46: opcode{"LSR", 2, 5, modeZeroPage, buildOpShift(false, false)},
	0x56: opcode{"LSR", 2, 6, modeZeroPageX, buildOpShift(false, false)},
	0x4E: opcode{"LSR", 3, 6, modeAbsolute, buildOpShift(false, false)},
	0x5E: opcode{"LSR", 3, 7, modeAbsoluteX, buildOpShift(false, false)},

	0x38: opcode{"SEC", 1, 2, modeImplicit, buildOpUpdateFlag(flagC, true)},
	0xF8: opcode{"SED", 1, 2, modeImplicit, buildOpUpdateFlag(flagD, true)},
	0x78: opcode{"SEI", 1, 2, modeImplicit, buildOpUpdateFlag(flagI, true)},
	0x18: opcode{"CLC", 1, 2, modeImplicit, buildOpUpdateFlag(flagC, false)},
	0xD8: opcode{"CLD", 1, 2, modeImplicit, buildOpUpdateFlag(flagD, false)},
	0x58: opcode{"CLI", 1, 2, modeImplicit, buildOpUpdateFlag(flagI, false)},
	0xB8: opcode{"CLV", 1, 2, modeImplicit, buildOpUpdateFlag(flagV, false)},

	0xE6: opcode{"INC", 2, 5, modeZeroPage, buildOpIncDec(true)},
	0xF6: opcode{"INC", 2, 6, modeZeroPageX, buildOpIncDec(true)},
	0xEE: opcode{"INC", 3, 6, modeAbsolute, buildOpIncDec(true)},
	0xFE: opcode{"INC", 3, 7, modeAbsoluteX, buildOpIncDec(true)},
	0xC6: opcode{"DEC", 2, 5, modeZeroPage, buildOpIncDec(false)},
	0xD6: opcode{"DEC", 2, 6, modeZeroPageX, buildOpIncDec(false)},
	0xCE: opcode{"DEC", 3, 6, modeAbsolute, buildOpIncDec(false)},
	0xDE: opcode{"DEC", 3, 7, modeAbsoluteX, buildOpIncDec(false)},
	0xE8: opcode{"INX", 1, 2, modeImplicitX, buildOpIncDec(true)},
	0xC8: opcode{"INY", 1, 2, modeImplicitY, buildOpIncDec(true)},
	0xCA: opcode{"DEX", 1, 2, modeImplicitX, buildOpIncDec(false)},
	0x88: opcode{"DEY", 1, 2, modeImplicitY, buildOpIncDec(false)},

	0xAA: opcode{"TAX", 1, 2, modeImplicit, buildOpTransfer(regA, regX)},
	0xA8: opcode{"TAY", 1, 2, modeImplicit, buildOpTransfer(regA, regY)},
	0x8A: opcode{"TXA", 1, 2, modeImplicit, buildOpTransfer(regX, regA)},
	0x98: opcode{"TYA", 1, 2, modeImplicit, buildOpTransfer(regY, regA)},
	0x9A: opcode{"TXS", 1, 2, modeImplicit, buildOpTransfer(regX, regSP)},
	0xBA: opcode{"TSX", 1, 2, modeImplicit, buildOpTransfer(regSP, regX)},

	0xA9: opcode{"LDA", 2, 2, modeImmediate, buildOpLoad(regA)},
	0xA5: opcode{"LDA", 2, 3, modeZeroPage, buildOpLoad(regA)},
	0xB5: opcode{"LDA", 2, 4, modeZeroPageX, buildOpLoad(regA)},
	0xAD: opcode{"LDA", 3, 4, modeAbsolute, buildOpLoad(regA)},
	0xBD: opcode{"LDA", 3, 4, modeAbsoluteX, buildOpLoad(regA)}, // Extra cycles
	0xB9: opcode{"LDA", 3, 4, modeAbsoluteY, buildOpLoad(regA)}, // Extra cycles
	0xA1: opcode{"LDA", 2, 6, modeIndexedIndirectX, buildOpLoad(regA)},
	0xB1: opcode{"LDA", 2, 5, modeIndirectIndexedY, buildOpLoad(regA)}, // Extra cycles
	0xA2: opcode{"LDX", 2, 2, modeImmediate, buildOpLoad(regX)},
	0xA6: opcode{"LDX", 2, 3, modeZeroPage, buildOpLoad(regX)},
	0xB6: opcode{"LDX", 2, 4, modeZeroPageY, buildOpLoad(regX)},
	0xAE: opcode{"LDX", 3, 4, modeAbsolute, buildOpLoad(regX)},
	0xBE: opcode{"LDX", 3, 4, modeAbsoluteY, buildOpLoad(regX)}, // Extra cycles
	0xA0: opcode{"LDY", 2, 2, modeImmediate, buildOpLoad(regY)},
	0xA4: opcode{"LDY", 2, 3, modeZeroPage, buildOpLoad(regY)},
	0xB4: opcode{"LDY", 2, 4, modeZeroPageX, buildOpLoad(regY)},
	0xAC: opcode{"LDY", 3, 4, modeAbsolute, buildOpLoad(regY)},
	0xBC: opcode{"LDY", 3, 4, modeAbsoluteX, buildOpLoad(regY)}, // Extra cycles

	0x85: opcode{"STA", 2, 3, modeZeroPage, buildOpStore(regA)},
	0x95: opcode{"STA", 2, 4, modeZeroPageX, buildOpStore(regA)},
	0x8D: opcode{"STA", 3, 4, modeAbsolute, buildOpStore(regA)},
	0x9D: opcode{"STA", 3, 5, modeAbsoluteX, buildOpStore(regA)},
	0x99: opcode{"STA", 3, 5, modeAbsoluteY, buildOpStore(regA)},
	0x81: opcode{"STA", 2, 6, modeIndexedIndirectX, buildOpStore(regA)},
	0x91: opcode{"STA", 2, 6, modeIndirectIndexedY, buildOpStore(regA)},
	0x86: opcode{"STX", 2, 3, modeZeroPage, buildOpStore(regX)},
	0x96: opcode{"STX", 2, 4, modeZeroPageY, buildOpStore(regX)},
	0x8E: opcode{"STX", 3, 4, modeAbsolute, buildOpStore(regX)},
	0x84: opcode{"STY", 2, 3, modeZeroPage, buildOpStore(regY)},
	0x94: opcode{"STY", 2, 4, modeZeroPageX, buildOpStore(regY)},
	0x8C: opcode{"STY", 3, 4, modeAbsolute, buildOpStore(regY)},

	0x90: opcode{"BCC", 2, 2, modeRelative, buildOpBranch(flagC, false)}, // Extra cycles
	0xB0: opcode{"BCS", 2, 2, modeRelative, buildOpBranch(flagC, true)},  // Extra cycles
	0xD0: opcode{"BNE", 2, 2, modeRelative, buildOpBranch(flagZ, false)}, // Extra cycles
	0xF0: opcode{"BEQ", 2, 2, modeRelative, buildOpBranch(flagZ, true)},  // Extra cycles
	0x10: opcode{"BPL", 2, 2, modeRelative, buildOpBranch(flagN, false)}, // Extra cycles
	0x30: opcode{"BMI", 2, 2, modeRelative, buildOpBranch(flagN, true)},  // Extra cycles
	0x50: opcode{"BVC", 2, 2, modeRelative, buildOpBranch(flagV, false)}, // Extra cycles
	0x70: opcode{"BVS", 2, 2, modeRelative, buildOpBranch(flagV, true)},  // Extra cycles

	0xEA: opcode{"NOP", 1, 2, modeImplicit, opNOP},
}

func executeLine(s *state, line []uint8) {
	opcode := opcodes[line[0]]
	opcode.action(s, line, opcode)
}

func executeInstruction(s *state, log bool) {
	pc := s.registers.getPC()
	opcode := opcodes[s.memory.peek(pc)]

	line := make([]uint8, opcode.bytes)
	for i := uint8(0); i < opcode.bytes; i++ {
		line[i] = s.memory.peek(pc)
		pc++
	}
	s.registers.setPC(pc)

	if log {
		fmt.Printf("%#04x %-12s: ", pc, lineString(s, line, opcode))
	}
	opcode.action(s, line, opcode)
	if log {
		fmt.Printf("%v, %x\n", s.registers, line)
	}
}

func lineString(s *state, line []uint8, opcode opcode) string {
	t := opcode.name
	switch opcode.addressMode {
	case modeImplicit:
	case modeImplicitX:
	case modeImplicitY:
		//Nothing
	case modeAccumulator:
		t += fmt.Sprintf(" A")
	case modeImmediate:
		t += fmt.Sprintf(" #%02x", line[1])
	case modeZeroPage:
		t += fmt.Sprintf(" $%02x", line[1])
	case modeZeroPageX:
		t += fmt.Sprintf(" $%02x,X", line[1])
	case modeZeroPageY:
		t += fmt.Sprintf(" $%02x,Y", line[1])
	case modeRelative:
		t += fmt.Sprintf(" *%+x", int8(line[1]))
	case modeAbsolute:
		t += fmt.Sprintf(" $%04x", getWordInLine(line))
	case modeAbsoluteX:
		t += fmt.Sprintf(" $%04x,X", getWordInLine(line))
	case modeAbsoluteY:
		t += fmt.Sprintf(" $%04x,X", getWordInLine(line))
	case modeIndirect:
		t += fmt.Sprintf(" ($%04x)", getWordInLine(line))
	case modeIndexedIndirectX:
		t += fmt.Sprintf(" ($%02x,X)", line[1])
	case modeIndirectIndexedY:
		t += fmt.Sprintf(" ($%02x),Y", line[1])
	default:
		t += "UNKNOWN MODE"
	}
	return t
}
