package core6502

import "fmt"

// https://www.masswerk.at/6502/6502_instruction_set.html
// http://www.emulator101.com/reference/6502-reference.html
// https://www.csh.rit.edu/~moffitt/docs/6502.html#FLAGS
// https://ia800509.us.archive.org/18/items/Programming_the_6502/Programming_the_6502.pdf

// State represents the state of the simulated device
type State struct {
	reg     registers
	mem     Memory
	cycles  uint64
	opcodes *[256]opcode
}

const (
	vectorNMI   uint16 = 0xfffa
	vectorReset uint16 = 0xfffc
	vectorBreak uint16 = 0xfffe
)

type opcode struct {
	name        string
	bytes       uint8
	cycles      int
	addressMode int
	action      opFunc
}

type opFunc func(s *State, line []uint8, opcode opcode)

func (s *State) executeLine(line []uint8) {
	opcode := s.opcodes[line[0]]
	if opcode.cycles == 0 {
		panic(fmt.Sprintf("Unknown opcode 0x%02x\n", line[0]))
	}
	opcode.action(s, line, opcode)
}

// ExecuteInstruction transforms the state given after a single instruction is executed.
func (s *State) ExecuteInstruction(log bool) {
	pc := s.reg.getPC()
	opcodeID := s.mem.Peek(pc)
	opcode := s.opcodes[opcodeID]

	if opcode.cycles == 0 {
		panic(fmt.Sprintf("Unknown opcode 0x%02x\n", opcodeID))
	}

	line := make([]uint8, opcode.bytes)
	for i := uint8(0); i < opcode.bytes; i++ {
		line[i] = s.mem.Peek(pc)
		pc++
	}
	s.reg.setPC(pc)

	if log {
		fmt.Printf("%#04x %-12s: ", pc, lineString(line, opcode))
	}
	opcode.action(s, line, opcode)
	s.cycles += uint64(opcode.cycles)
	if log {
		fmt.Printf("%v, [%02x]\n", s.reg, line)
	}
}

// Reset resets the processor state. Moves the program counter to the vector in 0cfffc.
func (s *State) Reset() {
	startAddress := getWord(s.mem, vectorReset)
	s.cycles = 0
	s.reg.setPC(startAddress)
}

// GetCycles returns the count of CPU cycles since last reset.
func (s *State) GetCycles() uint64 {
	return s.cycles
}

func lineString(line []uint8, opcode opcode) string {
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
		t += fmt.Sprintf(" $%04x,Y", getWordInLine(line))
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
