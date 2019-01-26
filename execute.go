package main

type state struct {
	register r,
	memory m
}

func step(state *S) {

}

const modeNone = -1
const modeImmediate = 0
const modeZeroPage = 1
const modeAbsolute = 2 

type opcode struct {
	name string
	code int8
	bytes int
	cycles int
	mode int
}

// https://www.masswerk.at/6502/6502_instruction_set.html

func opA1LDA(state *, opcode) {
	value := s->memory
	s->register.setRegister(regA, value)

}

func opLDA(state *s, reg, mode, arg) {
}

opcodes := []Opcode{
	0: opcode('BRK', 0x0, 1, 7, modeImmediate)
	1:
}

