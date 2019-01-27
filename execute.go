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
const modeAbsolute = 2
const modeAbsoluteX = 4
const modeAbsoluteY = 5

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
	case modeAbsolute:
		address := getWordInLine(line)
		value = s.memory[address]
	case modeAbsoluteX:
		address := getWordInLine(line) + uint16(s.registers.getX())
		value = s.memory[address]
	case modeAbsoluteY:
		address := getWordInLine(line) + uint16(s.registers.getY())
		value = s.memory[address]
	}

	s.registers.setRegister(opcode.reg, value)
}

var opcodes = [256]opcode{
	0x00: opcode{"BRK", 1, 7, modeImmediate, regNone, opNOP},
	0xA0: opcode{"LDY", 2, 2, modeImmediate, regY, opLDR},
	0xA5: opcode{"LDA", 2, 3, modeZeroPage, regA, opLDR},
	0xB5: opcode{"LDA", 2, 4, modeZeroPageX, regA, opLDR},
	0xA9: opcode{"LDA", 2, 2, modeImmediate, regA, opLDR},
	0xAD: opcode{"LDA", 3, 4, modeAbsolute, regA, opLDR},
	0xB9: opcode{"LDA", 3, 4, modeAbsoluteY, regA, opLDR}, // Extra cycles
	0xBD: opcode{"LDA", 3, 4, modeAbsoluteX, regA, opLDR}, // Extra cycles
}

func executeLine(s *state, line []uint8) {
	opcode := opcodes[line[0]]
	opcode.action(s, line, opcode)
}
