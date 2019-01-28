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

// https://www.masswerk.at/6502/6502_instruction_set.html
// http://www.emulator101.com/reference/6502-reference.html
// https://www.csh.rit.edu/~moffitt/docs/6502.html#FLAGS

func getWordInLine(line []uint8) uint16 {
	return uint16(line[1]) + 0x100*uint16(line[2])
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

func buildOpIncDecRegister(reg int, inc bool) opFunc {
	return func(s *state, line []uint8, opcode opcode) {
		value := s.registers.getRegister(reg)
		if inc {
			value++
		} else {
			value--
		}
		s.registers.setRegister(reg, value)
		s.registers.updateFlagZN(value)
	}
}

func resolveWithAddressMode(s *state, line []uint8, addressMode int) (value uint8, hasAddress bool, address uint16) {
	hasAddress = true
	switch addressMode {
	case modeAccumulator:
		value = s.registers.getA()
		hasAddress = false
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
	return
}

func buildRotate(addressMode int, isLeft bool) opFunc {
	return func(s *state, line []uint8, opcode opcode) {
		value, hasAddress, address := resolveWithAddressMode(s, line, addressMode)

		oldCarry := s.registers.getFlagBit(flagC)
		var carry bool
		if isLeft {
			carry = (value & 0x80) != 0
			value <<= 1
			value += oldCarry
		} else {
			carry = (value & 0x01) != 0
			value >>= 1
			value += oldCarry << 7
		}
		s.registers.updateFlag(flagC, carry)
		s.registers.updateFlagZN(value)

		if hasAddress {
			s.memory[address] = value
		} else {
			s.registers.setA(value)
		}
	}
}

func buildOpLoad(addressMode int, regDst int) opFunc {
	return func(s *state, line []uint8, opcode opcode) {
		value, _, _ := resolveWithAddressMode(s, line, addressMode)
		s.registers.setRegister(regDst, value)
		s.registers.updateFlagZN(value)
	}
}

var opcodes = [256]opcode{
	0x26: opcode{"ROL", 2, 5, buildRotate(modeZeroPage, true)},
	0x2A: opcode{"ROL", 1, 2, buildRotate(modeAccumulator, true)},
	0x2E: opcode{"ROL", 3, 6, buildRotate(modeAbsolute, true)},
	0x36: opcode{"ROL", 2, 6, buildRotate(modeZeroPageX, true)},
	0x3E: opcode{"ROL", 3, 7, buildRotate(modeAbsoluteX, true)},

	0x66: opcode{"ROR", 2, 5, buildRotate(modeZeroPage, false)},
	0x6A: opcode{"ROR", 1, 2, buildRotate(modeAccumulator, false)},
	0x6E: opcode{"ROR", 3, 6, buildRotate(modeAbsolute, false)},
	0x76: opcode{"ROR", 2, 6, buildRotate(modeZeroPageX, false)},
	0x7E: opcode{"ROR", 3, 7, buildRotate(modeAbsoluteX, false)},

	0x88: opcode{"DEY", 1, 2, buildOpIncDecRegister(regY, false)},
	0x8A: opcode{"TXA", 1, 2, buildOpTransfer(regX, regA)},
	0x98: opcode{"TYA", 1, 2, buildOpTransfer(regY, regA)},
	0x9A: opcode{"TXS", 1, 2, buildOpTransfer(regX, regSP)},
	0xA0: opcode{"LDY", 2, 2, buildOpLoad(modeImmediate, regY)},
	0xA1: opcode{"LDX", 2, 6, buildOpLoad(modeIndexedIndirectX, regA)},
	0xA2: opcode{"LDX", 2, 2, buildOpLoad(modeImmediate, regX)},
	0xA4: opcode{"LDY", 2, 3, buildOpLoad(modeZeroPage, regY)},
	0xA5: opcode{"LDA", 2, 3, buildOpLoad(modeZeroPage, regA)},
	0xA6: opcode{"LDX", 2, 3, buildOpLoad(modeZeroPage, regX)},
	0xA8: opcode{"TAY", 1, 2, buildOpTransfer(regA, regY)},
	0xA9: opcode{"LDA", 2, 2, buildOpLoad(modeImmediate, regA)},
	0xAA: opcode{"TAX", 1, 2, buildOpTransfer(regA, regX)},
	0xAC: opcode{"LDY", 3, 4, buildOpLoad(modeAbsolute, regY)},
	0xAD: opcode{"LDA", 3, 4, buildOpLoad(modeAbsolute, regA)},
	0xAE: opcode{"LDX", 3, 4, buildOpLoad(modeAbsolute, regX)},
	0xB1: opcode{"LDX", 2, 5, buildOpLoad(modeIndirectIndexedY, regA)}, // Extra cycles
	0xB4: opcode{"LDY", 2, 4, buildOpLoad(modeZeroPageX, regY)},
	0xB5: opcode{"LDA", 2, 4, buildOpLoad(modeZeroPageX, regA)},
	0xB6: opcode{"LDX", 2, 4, buildOpLoad(modeZeroPageY, regX)},
	0xB9: opcode{"LDA", 3, 4, buildOpLoad(modeAbsoluteY, regA)}, // Extra cycles
	0xBA: opcode{"TSX", 1, 2, buildOpTransfer(regSP, regX)},
	0xBC: opcode{"LDY", 3, 4, buildOpLoad(modeAbsoluteX, regY)}, // Extra cycles
	0xBD: opcode{"LDA", 3, 4, buildOpLoad(modeAbsoluteX, regA)}, // Extra cycles
	0xBE: opcode{"LDX", 3, 4, buildOpLoad(modeAbsoluteY, regX)}, // Extra cycles
	0xC8: opcode{"INY", 1, 2, buildOpIncDecRegister(regY, true)},
	0xCA: opcode{"DEX", 1, 2, buildOpIncDecRegister(regX, false)},
	0xE8: opcode{"INX", 1, 2, buildOpIncDecRegister(regX, true)},
	0xEA: opcode{"NOP", 1, 2, opNOP},
}

func executeLine(s *state, line []uint8) {
	opcode := opcodes[line[0]]
	opcode.action(s, line, opcode)
}
