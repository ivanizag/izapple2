package core6502

import "fmt"

const (
	modeImplicit = iota + 1
	modeImplicitX
	modeImplicitY
	modeAccumulator
	modeImmediate
	modeZeroPage
	modeZeroPageX
	modeZeroPageY
	modeRelative
	modeAbsolute
	modeAbsoluteX
	modeAbsoluteY
	modeIndirect
	modeIndexedIndirectX
	modeIndirectIndexedY
	// Added on the 65c02
	modeIndirectZeroPage
	modeAbsoluteIndexedIndirectX
	modeZeroPageAndRelative
)

func getWordInLine(line []uint8) uint16 {
	return uint16(line[1]) + 0x100*uint16(line[2])
}

func resolveValue(s *State, line []uint8, opcode opcode) uint8 {
	switch opcode.addressMode {
	case modeAccumulator:
		return s.reg.getA()
	case modeImplicitX:
		return s.reg.getX()
	case modeImplicitY:
		return s.reg.getY()
	case modeImmediate:
		return line[1]
	}

	// The value is in memory
	address := resolveAddress(s, line, opcode)
	return s.mem.Peek(address)
}

func resolveSetValue(s *State, line []uint8, opcode opcode, value uint8) {
	switch opcode.addressMode {
	case modeAccumulator:
		s.reg.setA(value)
		return
	case modeImplicitX:
		s.reg.setX(value)
		return
	case modeImplicitY:
		s.reg.setY(value)
		return
	}

	// The value is in memory
	address := resolveAddress(s, line, opcode)
	s.mem.Poke(address, value)
}

func resolveAddress(s *State, line []uint8, opcode opcode) uint16 {
	var address uint16

	switch opcode.addressMode {
	case modeZeroPage:
		address = uint16(line[1])
	case modeZeroPageX:
		address = uint16(line[1] + s.reg.getX())
	case modeZeroPageY:
		address = uint16(line[1] + s.reg.getY())
	case modeAbsolute:
		address = getWordInLine(line)
	case modeAbsoluteX:
		address = getWordInLine(line) + uint16(s.reg.getX())
	case modeAbsoluteY:
		address = getWordInLine(line) + uint16(s.reg.getY())
	case modeIndexedIndirectX:
		addressAddress := uint8(line[1] + s.reg.getX())
		address = getZeroPageWord(s.mem, addressAddress)
	case modeIndirect:
		addressAddress := getWordInLine(line)
		address = getWord(s.mem, addressAddress)
	case modeIndirectIndexedY:
		address = getZeroPageWord(s.mem, line[1]) +
			uint16(s.reg.getY())
	// 65c02 additions
	case modeIndirectZeroPage:
		address = getZeroPageWord(s.mem, line[1])
	case modeAbsoluteIndexedIndirectX:
		addressAddress := getWordInLine(line) + uint16(s.reg.getX())
		address = getWord(s.mem, addressAddress)
	case modeRelative:
		// This assumes that PC is already pointing to the next instruction
		address = s.reg.getPC() + uint16(int8(line[1])) // Note: line[1] is signed
	case modeZeroPageAndRelative:
		// Two addressing modes combined. We refer to the second one, relative,
		// placed one byte after the zeropage reference
		address = s.reg.getPC() + uint16(int8(line[2])) // Note: line[2] is signed
	default:
		panic("Assert failed. Missing addressing mode")
	}
	return address
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
	// 65c02 additions:
	case modeIndirectZeroPage:
		t += fmt.Sprintf(" ($%02x)", line[1])
	case modeAbsoluteIndexedIndirectX:
		t += fmt.Sprintf(" ($%04x,X)", getWordInLine(line))
	case modeZeroPageAndRelative:
		t += fmt.Sprintf(" $%02x *%+x", line[1], int8(line[2]))
	default:
		t += "UNKNOWN MODE"
	}
	return t
}
