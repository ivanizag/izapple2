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

func getWordInLine(line []uint8) uint16 {
	return uint16(line[1]) + 0x100*uint16(line[2])
}

type opcode struct {
	name   string
	bytes  int
	cycles int
	mode   int
	reg    int
	action opFunc
}

type opFunc func(s *state, line []uint8, opcode opcode)

func opNOP(s *state, line []uint8, opcode opcode) {}

func opLDR(s *state, line []uint8, opcode opcode) {
	var value uint8
	switch opcode.mode {
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

	s.registers.setRegister(opcode.reg, value)

	// TODO: Update flags (N, Z)
}

var opcodes = [256]opcode{
	0x00: opcode{"BRK", 1, 7, modeImmediate, regNone, opNOP},

	0xA0: opcode{"LDY", 2, 2, modeImmediate, regY, opLDR},

	0xA1: opcode{"LDX", 2, 6, modeIndexedIndirectX, regA, opLDR},

	0xA2: opcode{"LDX", 2, 2, modeImmediate, regX, opLDR},
	0xA4: opcode{"LDY", 2, 3, modeZeroPage, regY, opLDR},
	0xA5: opcode{"LDA", 2, 3, modeZeroPage, regA, opLDR},
	0xA6: opcode{"LDX", 2, 3, modeZeroPage, regX, opLDR},
	0xA9: opcode{"LDA", 2, 2, modeImmediate, regA, opLDR},

	0xAC: opcode{"LDY", 3, 4, modeAbsolute, regY, opLDR},
	0xAD: opcode{"LDA", 3, 4, modeAbsolute, regA, opLDR},
	0xAE: opcode{"LDX", 3, 4, modeAbsolute, regX, opLDR},

	0xB1: opcode{"LDX", 2, 5, modeIndirectIndexedY, regA, opLDR}, // Extra cycles

	0xB4: opcode{"LDY", 2, 4, modeZeroPageX, regY, opLDR},
	0xB5: opcode{"LDA", 2, 4, modeZeroPageX, regA, opLDR},
	0xB6: opcode{"LDX", 2, 4, modeZeroPageY, regX, opLDR},

	0xB9: opcode{"LDA", 3, 4, modeAbsoluteY, regA, opLDR}, // Extra cycles
	0xBC: opcode{"LDY", 3, 4, modeAbsoluteX, regY, opLDR}, // Extra cycles
	0xBD: opcode{"LDA", 3, 4, modeAbsoluteX, regA, opLDR}, // Extra cycles
	0xBE: opcode{"LDX", 3, 4, modeAbsoluteY, regX, opLDR}, // Extra cycles
}

func executeLine(s *state, line []uint8) {
	opcode := opcodes[line[0]]
	opcode.action(s, line, opcode)
}
