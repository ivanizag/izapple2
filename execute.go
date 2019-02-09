package main

import "fmt"

type state struct {
	registers registers
	memory    memory
}

func step(s *state) {

}

const modeNone = -1
const modeImmediate = 0
const modeZeroPage = 1
const modeZeroPageX = 3
const modeZeroPageY = 6
const modeAbsolute = 2
const modeAbsoluteX = 4
const modeAbsoluteY = 5
const modeIndexedIndirectX = 7
const modeIndirectIndexedY = 8
const modeAccumulator = 9
const modeRegisterX = 10
const modeRegisterY = 11
const modeIndirect = 12

// https://www.masswerk.at/6502/6502_instruction_set.html
// http://www.emulator101.com/reference/6502-reference.html
// https://www.csh.rit.edu/~moffitt/docs/6502.html#FLAGS
// https://ia800509.us.archive.org/18/items/Programming_the_6502/Programming_the_6502.pdf

func getWordInLine(line []uint8) uint16 {
	return uint16(line[1]) + 0x100*uint16(line[2])
}

func resolveWithAddressMode(s *state, line []uint8, addressMode int) (value uint8, address uint16, setValue func(uint8)) {
	hasAddress := true
	register := regNone

	switch addressMode {
	case modeAccumulator:
		value = s.registers.getA()
		hasAddress = false
		register = regA
	case modeRegisterX:
		value = s.registers.getX()
		hasAddress = false
		register = regX
	case modeRegisterY:
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
		value = s.memory[address]
	}

	setValue = func(value uint8) {
		if hasAddress {
			s.memory[address] = value
		} else if register != regNone {
			s.registers.setRegister(register, value)
		} else {
			// Todo: assert impossible
		}
	}
	return
}

type opcode struct {
	name   string
	bytes  int8
	cycles int
	action opFunc
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

func buildOpIncDec(addressMode int, inc bool) opFunc {
	return func(s *state, line []uint8, opcode opcode) {
		value, _, setValue := resolveWithAddressMode(s, line, addressMode)
		if inc {
			value++
		} else {
			value--
		}
		s.registers.updateFlagZN(value)
		setValue(value)
	}
}

func buildShift(addressMode int, isLeft bool, isRotate bool) opFunc {
	return func(s *state, line []uint8, opcode opcode) {
		value, _, setValue := resolveWithAddressMode(s, line, addressMode)

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

func buildOpLoad(addressMode int, regDst int) opFunc {
	return func(s *state, line []uint8, opcode opcode) {
		value, _, _ := resolveWithAddressMode(s, line, addressMode)
		s.registers.setRegister(regDst, value)
		s.registers.updateFlagZN(value)
	}
}

func buildOpStore(addressMode int, regSrc int) opFunc {
	return func(s *state, line []uint8, opcode opcode) {
		_, _, setValue := resolveWithAddressMode(s, line, addressMode)
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

func buildOpBit(addressMode int) opFunc {
	return func(s *state, line []uint8, opcode opcode) {
		value, _, _ := resolveWithAddressMode(s, line, addressMode)
		acc := s.registers.getA()
		s.registers.updateFlag(flagZ, value&acc == 0)
		s.registers.updateFlag(flagN, value&(1<<7) != 0)
		s.registers.updateFlag(flagV, value&(1<<6) != 0)
	}
}

func buildOpCompare(addressMode int, reg int) opFunc {
	return func(s *state, line []uint8, opcode opcode) {
		value, _, _ := resolveWithAddressMode(s, line, addressMode)
		reference := s.registers.getRegister(reg)
		s.registers.updateFlagZN(reference - value)
		s.registers.updateFlag(flagC, reference >= value)
	}
}

func operationAnd(a uint8, b uint8) uint8 { return a & b }
func operationOr(a uint8, b uint8) uint8  { return a | b }
func operationXor(a uint8, b uint8) uint8 { return a ^ b }

func buildOpLogic(addressMode int, operation func(uint8, uint8) uint8) opFunc {
	return func(s *state, line []uint8, opcode opcode) {
		value, _, _ := resolveWithAddressMode(s, line, addressMode)
		result := operation(value, s.registers.getA())
		s.registers.setA(result)
		s.registers.updateFlagZN(result)
	}
}

func buildOpAdd(addressMode int) opFunc {
	return func(s *state, line []uint8, opcode opcode) {
		value, _, _ := resolveWithAddressMode(s, line, addressMode)
		if s.registers.getFlag(flagD) {
			// TODO BCD. See http://www.6502.org/tutorials/decimal_mode.html

		} else {
			total := uint16(s.registers.getA()) +
				uint16(value) +
				uint16(s.registers.getFlagBit(flagC))
			truncated := uint8(total)
			s.registers.setA(truncated)
			s.registers.updateFlagZN(truncated)
			s.registers.updateFlag(flagC, total > 0xFF)
			// TODO: missing overflow flag
		}
	}
}

func buildOpSub(addressMode int) opFunc {
	return func(s *state, line []uint8, opcode opcode) {
		value, _, _ := resolveWithAddressMode(s, line, addressMode)
		if s.registers.getFlag(flagD) {
			// TODO BCD
		} else {
			total := 0x100 + uint16(s.registers.getA()) -
				uint16(value) -
				uint16(s.registers.getFlagBit(flagC))
			truncated := uint8(total)
			s.registers.setA(truncated)
			s.registers.updateFlagZN(truncated)
			s.registers.updateFlag(flagC, total <= 0xFF)
			// TODO: missing overflow flag
		}
	}
}

const stackAddress uint16 = 0x0100

func pushByte(s *state, value uint8) {
	adresss := stackAddress + uint16(s.registers.getSP())
	s.memory[adresss] = value
	s.registers.setSP(s.registers.getSP() - 1)
}

func pullByte(s *state) uint8 {
	s.registers.setSP(s.registers.getSP() + 1)
	adresss := stackAddress + uint16(s.registers.getSP())
	return s.memory[adresss]
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

func buildOpJump(addressMode int) opFunc {
	return func(s *state, line []uint8, opcode opcode) {
		_, address, _ := resolveWithAddressMode(s, line, addressMode)
		s.registers.setPC(address)
	}
}

func opNOP(s *state, line []uint8, opcode opcode) {}

func opJSR(s *state, line []uint8, opcode opcode) {
	pushWord(s, s.registers.getPC())
	s.registers.setPC(getWordInLine(line))
}

func opRTI(s *state, line []uint8, opcode opcode) {
	s.registers.setP(pullByte(s))
	s.registers.setPC(pullWord(s))
}

func opRTS(s *state, line []uint8, opcode opcode) {
	s.registers.setPC(pullWord(s) + 1) // TODO: Do we really need to add 1?
}

func opBRK(s *state, line []uint8, opcode opcode) {
	s.registers.setFlag(flagI)
	pushWord(s, s.registers.getPC()+1) // TODO: De we have to add 1 or 2?
	pushByte(s, s.registers.getP()|(flagB+flag5))
	s.registers.setPC(s.memory.getWord(0xFFFE))
}

var opcodes = [256]opcode{
	0x00: opcode{"BRK", 1, 7, opBRK},
	0x4C: opcode{"JMP", 3, 3, buildOpJump(modeAbsolute)},
	0x6C: opcode{"JMP", 3, 3, buildOpJump(modeIndirect)},
	0x20: opcode{"JSR", 3, 6, opJSR},
	0x40: opcode{"RTI", 1, 6, opRTI},
	0x60: opcode{"RTS", 1, 6, opRTS},

	0x48: opcode{"PHA", 1, 3, opPHA},
	0x08: opcode{"PHP", 1, 3, opPHP},
	0x68: opcode{"PLA", 1, 4, opPLA},
	0x28: opcode{"PLP", 1, 4, opPLP},

	0x09: opcode{"ORA", 2, 2, buildOpLogic(modeImmediate, operationOr)},
	0x05: opcode{"ORA", 2, 3, buildOpLogic(modeZeroPage, operationOr)},
	0x15: opcode{"ORA", 2, 4, buildOpLogic(modeZeroPageX, operationOr)},
	0x0D: opcode{"ORA", 3, 4, buildOpLogic(modeAbsolute, operationOr)},
	0x1D: opcode{"ORA", 3, 4, buildOpLogic(modeAbsoluteX, operationOr)}, // Extra cycles
	0x19: opcode{"ORA", 3, 4, buildOpLogic(modeAbsoluteY, operationOr)}, // Extra cycles
	0x01: opcode{"ORA", 2, 6, buildOpLogic(modeIndexedIndirectX, operationOr)},
	0x11: opcode{"ORA", 2, 5, buildOpLogic(modeIndirectIndexedY, operationOr)}, // Extra cycles

	0x29: opcode{"AND", 2, 2, buildOpLogic(modeImmediate, operationAnd)},
	0x25: opcode{"AND", 2, 3, buildOpLogic(modeZeroPage, operationAnd)},
	0x35: opcode{"AND", 2, 4, buildOpLogic(modeZeroPageX, operationAnd)},
	0x2D: opcode{"AND", 3, 4, buildOpLogic(modeAbsolute, operationAnd)},
	0x3D: opcode{"AND", 3, 4, buildOpLogic(modeAbsoluteX, operationAnd)}, // Extra cycles
	0x39: opcode{"AND", 3, 4, buildOpLogic(modeAbsoluteY, operationAnd)}, // Extra cycles
	0x21: opcode{"AND", 2, 6, buildOpLogic(modeIndexedIndirectX, operationAnd)},
	0x31: opcode{"AND", 2, 5, buildOpLogic(modeIndirectIndexedY, operationAnd)}, // Extra cycles

	0x49: opcode{"EOR", 2, 2, buildOpLogic(modeImmediate, operationXor)},
	0x45: opcode{"EOR", 2, 3, buildOpLogic(modeZeroPage, operationXor)},
	0x55: opcode{"EOR", 2, 4, buildOpLogic(modeZeroPageX, operationXor)},
	0x4D: opcode{"EOR", 3, 4, buildOpLogic(modeAbsolute, operationXor)},
	0x5D: opcode{"EOR", 3, 4, buildOpLogic(modeAbsoluteX, operationXor)}, // Extra cycles
	0x59: opcode{"EOR", 3, 4, buildOpLogic(modeAbsoluteY, operationXor)}, // Extra cycles
	0x41: opcode{"EOR", 2, 6, buildOpLogic(modeIndexedIndirectX, operationXor)},
	0x51: opcode{"EOR", 2, 5, buildOpLogic(modeIndirectIndexedY, operationXor)}, // Extra cycles

	0x69: opcode{"ADC", 2, 2, buildOpAdd(modeImmediate)},
	0x65: opcode{"ADC", 2, 3, buildOpAdd(modeZeroPage)},
	0x75: opcode{"ADC", 2, 4, buildOpAdd(modeZeroPageX)},
	0x6D: opcode{"ADC", 3, 4, buildOpAdd(modeAbsolute)},
	0x7D: opcode{"ADC", 3, 4, buildOpAdd(modeAbsoluteX)}, // Extra cycles
	0x79: opcode{"ADC", 3, 4, buildOpAdd(modeAbsoluteY)}, // Extra cycles
	0x61: opcode{"ADC", 2, 6, buildOpAdd(modeIndexedIndirectX)},
	0x71: opcode{"ADC", 2, 5, buildOpAdd(modeIndirectIndexedY)}, // Extra cycles

	0xE9: opcode{"SBC", 2, 2, buildOpSub(modeImmediate)},
	0xE5: opcode{"SBC", 2, 3, buildOpSub(modeZeroPage)},
	0xF5: opcode{"SBC", 2, 4, buildOpSub(modeZeroPageX)},
	0xED: opcode{"SBC", 3, 4, buildOpSub(modeAbsolute)},
	0xFD: opcode{"SBC", 3, 4, buildOpSub(modeAbsoluteX)}, // Extra cycles
	0xF9: opcode{"SBC", 3, 4, buildOpSub(modeAbsoluteY)}, // Extra cycles
	0xE1: opcode{"SBC", 2, 6, buildOpSub(modeIndexedIndirectX)},
	0xF1: opcode{"SBC", 2, 5, buildOpSub(modeIndirectIndexedY)}, // Extra cycles

	0x24: opcode{"BIT", 2, 3, buildOpBit(modeZeroPage)},
	0x2C: opcode{"BIT", 2, 3, buildOpBit(modeAbsolute)},

	0xC9: opcode{"CMP", 2, 2, buildOpCompare(modeImmediate, regA)},
	0xC5: opcode{"CMP", 2, 3, buildOpCompare(modeZeroPage, regA)},
	0xD5: opcode{"CMP", 2, 4, buildOpCompare(modeZeroPageX, regA)},
	0xCD: opcode{"CMP", 3, 4, buildOpCompare(modeAbsolute, regA)},
	0xDD: opcode{"CMP", 3, 4, buildOpCompare(modeAbsoluteX, regA)}, // Extra cycles
	0xD9: opcode{"CMP", 3, 4, buildOpCompare(modeAbsoluteY, regA)}, // Extra cycles
	0xC1: opcode{"CMP", 2, 6, buildOpCompare(modeIndexedIndirectX, regA)},
	0xD1: opcode{"CMP", 2, 5, buildOpCompare(modeIndirectIndexedY, regA)}, // Extra cycles

	0xE0: opcode{"CPX", 2, 2, buildOpCompare(modeImmediate, regX)},
	0xE4: opcode{"CPX", 2, 3, buildOpCompare(modeZeroPage, regX)},
	0xEC: opcode{"CPX", 3, 4, buildOpCompare(modeAbsolute, regX)},

	0xC0: opcode{"CPY", 2, 2, buildOpCompare(modeImmediate, regY)},
	0xC4: opcode{"CPY", 2, 3, buildOpCompare(modeZeroPage, regY)},
	0xCC: opcode{"CPY", 3, 4, buildOpCompare(modeAbsolute, regY)},

	0x2A: opcode{"ROL", 1, 2, buildShift(modeAccumulator, true, true)},
	0x26: opcode{"ROL", 2, 5, buildShift(modeZeroPage, true, true)},
	0x36: opcode{"ROL", 2, 6, buildShift(modeZeroPageX, true, true)},
	0x2E: opcode{"ROL", 3, 6, buildShift(modeAbsolute, true, true)},
	0x3E: opcode{"ROL", 3, 7, buildShift(modeAbsoluteX, true, true)},

	0x6A: opcode{"ROR", 1, 2, buildShift(modeAccumulator, false, true)},
	0x66: opcode{"ROR", 2, 5, buildShift(modeZeroPage, false, true)},
	0x76: opcode{"ROR", 2, 6, buildShift(modeZeroPageX, false, true)},
	0x6E: opcode{"ROR", 3, 6, buildShift(modeAbsolute, false, true)},
	0x7E: opcode{"ROR", 3, 7, buildShift(modeAbsoluteX, false, true)},

	0x0A: opcode{"ASL", 1, 2, buildShift(modeAccumulator, true, false)},
	0x06: opcode{"ASL", 2, 5, buildShift(modeZeroPage, true, false)},
	0x16: opcode{"ASL", 2, 6, buildShift(modeZeroPageX, true, false)},
	0x0E: opcode{"ASL", 3, 6, buildShift(modeAbsolute, true, false)},
	0x1E: opcode{"ASL", 3, 7, buildShift(modeAbsoluteX, true, false)},

	0x4A: opcode{"LSR", 1, 2, buildShift(modeAccumulator, false, false)},
	0x46: opcode{"LSR", 2, 5, buildShift(modeZeroPage, false, false)},
	0x56: opcode{"LSR", 2, 6, buildShift(modeZeroPageX, false, false)},
	0x4E: opcode{"LSR", 3, 6, buildShift(modeAbsolute, false, false)},
	0x5E: opcode{"LSR", 3, 7, buildShift(modeAbsoluteX, false, false)},

	0x38: opcode{"SEC", 1, 2, buildOpUpdateFlag(flagC, true)},
	0xF8: opcode{"SED", 1, 2, buildOpUpdateFlag(flagD, true)},
	0x78: opcode{"SEI", 1, 2, buildOpUpdateFlag(flagI, true)},

	0x18: opcode{"CLC", 1, 2, buildOpUpdateFlag(flagC, false)},
	0xD8: opcode{"CLD", 1, 2, buildOpUpdateFlag(flagD, false)},
	0x58: opcode{"CLI", 1, 2, buildOpUpdateFlag(flagI, false)},
	0xB8: opcode{"CLV", 1, 2, buildOpUpdateFlag(flagV, false)},

	0xE6: opcode{"INC", 2, 5, buildOpIncDec(modeZeroPage, true)},
	0xF6: opcode{"INC", 2, 6, buildOpIncDec(modeZeroPageX, true)},
	0xEE: opcode{"INC", 3, 6, buildOpIncDec(modeAbsolute, true)},
	0xFE: opcode{"INC", 3, 7, buildOpIncDec(modeAbsoluteX, true)},

	0xC6: opcode{"DEC", 2, 5, buildOpIncDec(modeZeroPage, false)},
	0xD6: opcode{"DEC", 2, 6, buildOpIncDec(modeZeroPageX, false)},
	0xCE: opcode{"DEC", 3, 6, buildOpIncDec(modeAbsolute, false)},
	0xDE: opcode{"DEC", 3, 7, buildOpIncDec(modeAbsoluteX, false)},

	0xE8: opcode{"INX", 1, 2, buildOpIncDec(modeRegisterX, true)},
	0xC8: opcode{"INY", 1, 2, buildOpIncDec(modeRegisterY, true)},
	0xCA: opcode{"DEX", 1, 2, buildOpIncDec(modeRegisterX, false)},
	0x88: opcode{"DEY", 1, 2, buildOpIncDec(modeRegisterY, false)},

	0xAA: opcode{"TAX", 1, 2, buildOpTransfer(regA, regX)},
	0xA8: opcode{"TAY", 1, 2, buildOpTransfer(regA, regY)},
	0x8A: opcode{"TXA", 1, 2, buildOpTransfer(regX, regA)},
	0x98: opcode{"TYA", 1, 2, buildOpTransfer(regY, regA)},
	0x9A: opcode{"TXS", 1, 2, buildOpTransfer(regX, regSP)},
	0xBA: opcode{"TSX", 1, 2, buildOpTransfer(regSP, regX)},

	0xA9: opcode{"LDA", 2, 2, buildOpLoad(modeImmediate, regA)},
	0xA5: opcode{"LDA", 2, 3, buildOpLoad(modeZeroPage, regA)},
	0xB5: opcode{"LDA", 2, 4, buildOpLoad(modeZeroPageX, regA)},
	0xAD: opcode{"LDA", 3, 4, buildOpLoad(modeAbsolute, regA)},
	0xBD: opcode{"LDA", 3, 4, buildOpLoad(modeAbsoluteX, regA)}, // Extra cycles
	0xB9: opcode{"LDA", 3, 4, buildOpLoad(modeAbsoluteY, regA)}, // Extra cycles
	0xA1: opcode{"LDA", 2, 6, buildOpLoad(modeIndexedIndirectX, regA)},
	0xB1: opcode{"LDA", 2, 5, buildOpLoad(modeIndirectIndexedY, regA)}, // Extra cycles

	0xA2: opcode{"LDX", 2, 2, buildOpLoad(modeImmediate, regX)},
	0xA6: opcode{"LDX", 2, 3, buildOpLoad(modeZeroPage, regX)},
	0xB6: opcode{"LDX", 2, 4, buildOpLoad(modeZeroPageY, regX)},
	0xAE: opcode{"LDX", 3, 4, buildOpLoad(modeAbsolute, regX)},
	0xBE: opcode{"LDX", 3, 4, buildOpLoad(modeAbsoluteY, regX)}, // Extra cycles

	0xA0: opcode{"LDY", 2, 2, buildOpLoad(modeImmediate, regY)},
	0xA4: opcode{"LDY", 2, 3, buildOpLoad(modeZeroPage, regY)},
	0xB4: opcode{"LDY", 2, 4, buildOpLoad(modeZeroPageX, regY)},
	0xAC: opcode{"LDY", 3, 4, buildOpLoad(modeAbsolute, regY)},
	0xBC: opcode{"LDY", 3, 4, buildOpLoad(modeAbsoluteX, regY)}, // Extra cycles

	0x85: opcode{"STA", 2, 3, buildOpStore(modeZeroPage, regA)},
	0x95: opcode{"STA", 2, 4, buildOpStore(modeZeroPageX, regA)},
	0x8D: opcode{"STA", 3, 4, buildOpStore(modeAbsolute, regA)},
	0x9D: opcode{"STA", 3, 5, buildOpStore(modeAbsoluteX, regA)},
	0x99: opcode{"STA", 3, 5, buildOpStore(modeAbsoluteY, regA)},
	0x81: opcode{"STA", 2, 6, buildOpStore(modeIndexedIndirectX, regA)},
	0x91: opcode{"STA", 2, 6, buildOpStore(modeIndirectIndexedY, regA)},

	0x86: opcode{"STX", 2, 3, buildOpStore(modeZeroPage, regX)},
	0x96: opcode{"STX", 2, 4, buildOpStore(modeZeroPageY, regX)},
	0x8E: opcode{"STX", 3, 4, buildOpStore(modeAbsolute, regX)},

	0x84: opcode{"STY", 2, 3, buildOpStore(modeZeroPage, regY)},
	0x94: opcode{"STY", 2, 4, buildOpStore(modeZeroPageX, regY)},
	0x8C: opcode{"STY", 3, 4, buildOpStore(modeAbsolute, regY)},

	0x90: opcode{"BCC", 2, 2, buildOpBranch(flagC, false)}, // Extra cycles
	0xB0: opcode{"BCS", 2, 2, buildOpBranch(flagC, true)},  // Extra cycles
	0xD0: opcode{"BNE", 2, 2, buildOpBranch(flagZ, false)}, // Extra cycles
	0xF0: opcode{"BEQ", 2, 2, buildOpBranch(flagZ, true)},  // Extra cycles
	0x10: opcode{"BPL", 2, 2, buildOpBranch(flagN, false)}, // Extra cycles
	0x30: opcode{"BMI", 2, 2, buildOpBranch(flagN, true)},  // Extra cycles
	0x50: opcode{"BVC", 2, 2, buildOpBranch(flagV, false)}, // Extra cycles
	0x70: opcode{"BVS", 2, 2, buildOpBranch(flagV, true)},  // Extra cycles

	0xEA: opcode{"NOP", 1, 2, opNOP},
}

func executeLine(s *state, line []uint8) {
	opcode := opcodes[line[0]]
	opcode.action(s, line, opcode)
}

func executeInstruction(s *state) {
	pc := s.registers.getPC()
	opcode := opcodes[s.memory[pc]]
	fmt.Printf("%#04x %s: ", pc, opcode.name)
	pcNext := pc + uint16(opcode.bytes)
	s.registers.setPC(pcNext)
	line := s.memory[pc:pcNext]
	opcode.action(s, line, opcode)
	fmt.Printf("%v, %v\n", s.registers, line)
}
