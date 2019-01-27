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

// https://www.masswerk.at/6502/6502_instruction_set.html
// http://www.emulator101.com/reference/6502-reference.html
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

func buildOPTransfer(regSrc int, regDst int) opFunc {
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

func buildOpLoad(addressMode int, regDst int) opFunc {
	return func(s *state, line []uint8, opcode opcode) {
		var value uint8
		switch addressMode {
		case modeImmediate:
			value = line[1]
		case modeZeroPage:
			address := line[1]
			value = s.memory[address]
		case modeZeroPageX:
			address := line[1] + s.registers.getX()
			value = s.memory[address]
		case modeZeroPageY:
			address := line[1] + s.registers.getY()
			value = s.memory[address]
		case modeAbsolute:
			address := getWordInLine(line)
			value = s.memory[address]
		case modeAbsoluteX:
			address := getWordInLine(line) + uint16(s.registers.getX())
			value = s.memory[address]
		case modeAbsoluteY:
			address := getWordInLine(line) + uint16(s.registers.getY())
			value = s.memory[address]
		case modeIndexedIndirectX:
			addressAddress := uint8(line[1] + s.registers.getX())
			address := s.memory.getZeroPageWord(addressAddress)
			value = s.memory[address]
		case modeIndirectIndexedY:
			address := s.memory.getZeroPageWord(line[1]) +
				uint16(s.registers.getY())
			value = s.memory[address]
		}

		s.registers.setRegister(regDst, value)
		s.registers.updateFlagZN(value)
	}
}

var opcodes = [256]opcode{
	0x88: opcode{"DEY", 1, 2, buildOpIncDecRegister(regY, false)},

	0x8A: opcode{"TXA", 1, 2, buildOPTransfer(regX, regA)},

	0x98: opcode{"TYA", 1, 2, buildOPTransfer(regY, regA)},

	0x9A: opcode{"TXS", 1, 2, buildOPTransfer(regX, regSP)},

	0xA0: opcode{"LDY", 2, 2, buildOpLoad(modeImmediate, regY)},
	0xA1: opcode{"LDX", 2, 6, buildOpLoad(modeIndexedIndirectX, regA)},
	0xA2: opcode{"LDX", 2, 2, buildOpLoad(modeImmediate, regX)},

	0xA4: opcode{"LDY", 2, 3, buildOpLoad(modeZeroPage, regY)},
	0xA5: opcode{"LDA", 2, 3, buildOpLoad(modeZeroPage, regA)},
	0xA6: opcode{"LDX", 2, 3, buildOpLoad(modeZeroPage, regX)},

	0xA8: opcode{"TAY", 1, 2, buildOPTransfer(regA, regY)},
	0xA9: opcode{"LDA", 2, 2, buildOpLoad(modeImmediate, regA)},
	0xAA: opcode{"TAX", 1, 2, buildOPTransfer(regA, regX)},

	0xAC: opcode{"LDY", 3, 4, buildOpLoad(modeAbsolute, regY)},
	0xAD: opcode{"LDA", 3, 4, buildOpLoad(modeAbsolute, regA)},
	0xAE: opcode{"LDX", 3, 4, buildOpLoad(modeAbsolute, regX)},

	0xB1: opcode{"LDX", 2, 5, buildOpLoad(modeIndirectIndexedY, regA)}, // Extra cycles

	0xB4: opcode{"LDY", 2, 4, buildOpLoad(modeZeroPageX, regY)},
	0xB5: opcode{"LDA", 2, 4, buildOpLoad(modeZeroPageX, regA)},
	0xB6: opcode{"LDX", 2, 4, buildOpLoad(modeZeroPageY, regX)},

	0xB9: opcode{"LDA", 3, 4, buildOpLoad(modeAbsoluteY, regA)}, // Extra cycles
	0xBA: opcode{"TSX", 1, 2, buildOPTransfer(regSP, regX)},

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
