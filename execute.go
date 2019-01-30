package main

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

// https://www.masswerk.at/6502/6502_instruction_set.html
// http://www.emulator101.com/reference/6502-reference.html
// https://www.csh.rit.edu/~moffitt/docs/6502.html#FLAGS
// https://ia800509.us.archive.org/18/items/Programming_the_6502/Programming_the_6502.pdf

func getWordInLine(line []uint8) uint16 {
	return uint16(line[1]) + 0x100*uint16(line[2])
}

func resolveWithAddressMode(s *state, line []uint8, addressMode int) (value uint8, setValue func(uint8)) {
	var address uint16
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
	bytes  int
	cycles int
	action opFunc
}

type opFunc func(s *state, line []uint8, opcode opcode)

func opNOP(s *state, line []uint8, opcode opcode) {}

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
		value, setValue := resolveWithAddressMode(s, line, addressMode)
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
		value, setValue := resolveWithAddressMode(s, line, addressMode)

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
		value, _ := resolveWithAddressMode(s, line, addressMode)
		s.registers.setRegister(regDst, value)
		s.registers.updateFlagZN(value)
	}
}

func buildOpStore(addressMode int, regSrc int) opFunc {
	return func(s *state, line []uint8, opcode opcode) {
		_, setValue := resolveWithAddressMode(s, line, addressMode)
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

/*
TODO:

ADC
SBC

AND
ORA
EOR

BIT
CMP
CPX
CPY

BRK

JMP
JSR
RTI
RTS

PHA
PHP
PLA
PLP

*/

var opcodes = [256]opcode{
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

	0xA2: opcode{"LDX", 2, 2, buildOpLoad(modeImmediate, regX)},
	0xA6: opcode{"LDX", 2, 3, buildOpLoad(modeZeroPage, regX)},
	0xB6: opcode{"LDX", 2, 4, buildOpLoad(modeZeroPageY, regX)},
	0xAE: opcode{"LDX", 3, 4, buildOpLoad(modeAbsolute, regX)},
	0xBE: opcode{"LDX", 3, 4, buildOpLoad(modeAbsoluteY, regX)}, // Extra cycles
	0xA1: opcode{"LDX", 2, 6, buildOpLoad(modeIndexedIndirectX, regA)},
	0xB1: opcode{"LDX", 2, 5, buildOpLoad(modeIndirectIndexedY, regA)}, // Extra cycles

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
