package core6502

import (
	"encoding/binary"
	"fmt"
	"io"
)

// https://www.masswerk.at/6502/6502_instruction_set.html
// http://www.emulator101.com/reference/6502-reference.html
// https://www.csh.rit.edu/~moffitt/docs/6502.html#FLAGS
// https://ia800509.us.archive.org/18/items/Programming_the_6502/Programming_the_6502.pdf

const (
	maxInstructionSize = 3
)

// State represents the state of the simulated device
type State struct {
	reg       registers
	mem       Memory
	cycles    uint64
	opcodes   *[256]opcode
	trace     bool
	lineCache []uint8
	// We cache the allocation of a line to avoid a malloc per instruction. To be used only
	// by ExecuteInstruction(). 2x speedup on the emulation!!
}

const (
	vectorNMI   uint16 = 0xfffa
	vectorReset uint16 = 0xfffc
	vectorBreak uint16 = 0xfffe
)

type opcode struct {
	name        string
	bytes       uint16
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
func (s *State) ExecuteInstruction() {
	pc := s.reg.getPC()
	opcodeID := s.mem.Peek(pc)
	opcode := s.opcodes[opcodeID]

	if opcode.cycles == 0 {
		panic(fmt.Sprintf("Unknown opcode 0x%02x\n", opcodeID))
	}

	if s.lineCache == nil {
		s.lineCache = make([]uint8, maxInstructionSize)
	}
	for i := uint16(0); i < opcode.bytes; i++ {
		s.lineCache[i] = s.mem.Peek(pc)
		pc++
	}
	s.reg.setPC(pc)

	if s.trace {
		//fmt.Printf("%#04x %#02x\n", pc-opcode.bytes, opcodeID)
		fmt.Printf("%#04x %-13s: ", pc-opcode.bytes, lineString(s.lineCache, opcode))
	}
	opcode.action(s, s.lineCache, opcode)
	s.cycles += uint64(opcode.cycles)
	if s.trace {
		fmt.Printf("%v, [%02x]\n", s.reg, s.lineCache[0:opcode.bytes])
	}
}

// Reset resets the processor. Moves the program counter to the vector in 0cfffc.
func (s *State) Reset() {
	startAddress := getWord(s.mem, vectorReset)
	s.cycles += 6
	s.reg.setPC(startAddress)
}

// GetCycles returns the count of CPU cycles since last reset.
func (s *State) GetCycles() uint64 {
	return s.cycles
}

// SetTrace activates tracing of the cpu execution
func (s *State) SetTrace(trace bool) {
	s.trace = trace
}

// GetTrace gets trhe tracing state of the cpu execution
func (s *State) GetTrace() bool {
	return s.trace
}

// Save saves the CPU state (registers and cycle counter)
func (s *State) Save(w io.Writer) error {
	err := binary.Write(w, binary.BigEndian, s.cycles)
	if err != nil {
		return err
	}
	binary.Write(w, binary.BigEndian, s.reg.data)
	if err != nil {
		return err
	}
	return nil
}

// Load loads the CPU state (registers and cycle counter)
func (s *State) Load(r io.Reader) error {
	err := binary.Read(r, binary.BigEndian, &s.cycles)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.BigEndian, &s.reg.data)
	if err != nil {
		return err
	}
	return nil
}
