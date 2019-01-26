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

// https://www.masswerk.at/6502/6502_instruction_set.html

func opLDAimm(s *state, line []uint8) {
	opLDR(s, line, regA, modeImmediate)
}

func opLDAzpg(s *state, line []uint8) {
	opLDR(s, line, regA, modeZeroPage)
}

func opLDYimm(s *state, line []uint8) {
	opLDR(s, line, regY, modeImmediate)
}

func opLDAzpgX(s *state, line []uint8) {
	opLDR(s, line, regA, modeZeroPageX)
}

func opLDAabs(s *state, line []uint8) {
	opLDR(s, line, regA, modeAbsolute)
}

func getWordInLine(line []uint8) uint16 {
	return uint16(line[1]) + 0x100*uint16(line[2])
}

func opLDR(s *state, line []uint8, reg int, mode int) {
	var value uint8
	switch mode {
	case modeImmediate:
		value = line[1]
	case modeZeroPage:
		address := line[1]
		value = s.memory[address]
	case modeZeroPageX:
		address := line[1] + s.registers.getRegister(regX)
		value = s.memory[address]
	case modeAbsolute:
		address := getWordInLine(line)
		value = s.memory[address]
	}

	s.registers.setRegister(reg, value)
}

type opFunc func(s *state, line []uint8, opcode opcode)
type opcode struct {
	name   string
	bytes  int
	cycles int
	mode   int
	reg    int
	action opFunc
}

func opNOP(s *state, line []uint8, opcode opcode) {}
func opLDRex(s *state, line []uint8, opcode opcode) {
	opLDR(s, line, opcode.reg, opcode.mode)
}

var opcodes = [256]opcode{
	0x00: opcode{"BRK", 1, 7, modeImmediate, regNone, opNOP},
	0xA0: opcode{"LDY", -1, -1, modeImmediate, regY, opLDRex},
	0xA5: opcode{"LDA", 2, 3, modeZeroPage, regA, opLDRex},
	0xB5: opcode{"LDA", 2, 4, modeZeroPageX, regA, opLDRex},
	0xA9: opcode{"LDA", 2, 2, modeImmediate, regA, opLDRex},
	0xAD: opcode{"LDA", 3, 4, modeAbsolute, regA, opLDRex},
}

func executeLine(s *state, line []uint8) {
	opcode := opcodes[line[0]]
	opcode.action(s, line, opcode)
}
