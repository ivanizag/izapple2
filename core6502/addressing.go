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
	modeAbsoluteX65c02
	modeAbsoluteY
	modeIndirect
	modeIndexedIndirectX
	modeIndirectIndexedY
	// Added on the 65c02
	modeIndirect65c02Fix
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

	// On writes, the possible extra cycle crossing page boundaries is
	// added and already accounted for on NMOS
	if opcode.addressMode != modeAbsoluteX65c02 {
		s.extraCycleCrossingBoundaries = false
	}
}

func resolveAddress(s *State, line []uint8, opcode opcode) uint16 {
	var address uint16
	extraCycle := false

	switch opcode.addressMode {
	case modeZeroPage:
		address = uint16(line[1])
	case modeZeroPageX:
		address = uint16(line[1] + s.reg.getX())
	case modeZeroPageY:
		address = uint16(line[1] + s.reg.getY())
	case modeAbsolute:
		address = getWordInLine(line)
	case modeAbsoluteX65c02:
		fallthrough
	case modeAbsoluteX:
		base := getWordInLine(line)
		address, extraCycle = addOffset(base, s.reg.getX())
	case modeAbsoluteY:
		base := getWordInLine(line)
		address, extraCycle = addOffset(base, s.reg.getY())
	case modeIndexedIndirectX:
		addressAddress := line[1] + s.reg.getX()
		address = getZeroPageWord(s.mem, addressAddress)
	case modeIndirect:
		addressAddress := getWordInLine(line)
		address = getWordNoCrossPage(s.mem, addressAddress)
	case modeIndirect65c02Fix:
		addressAddress := getWordInLine(line)
		address = getWord(s.mem, addressAddress)
	case modeIndirectIndexedY:
		base := getZeroPageWord(s.mem, line[1])
		address, extraCycle = addOffset(base, s.reg.getY())
	// 65c02 additions
	case modeIndirectZeroPage:
		address = getZeroPageWord(s.mem, line[1])
	case modeAbsoluteIndexedIndirectX:
		addressAddress := getWordInLine(line) + uint16(s.reg.getX())
		address = getWord(s.mem, addressAddress)
	case modeRelative:
		// This assumes that PC is already pointing to the next instruction
		base := s.reg.getPC()
		address, extraCycle = addOffsetRelative(base, line[1])
	case modeZeroPageAndRelative:
		// Two addressing modes combined. We refer to the second one, relative,
		// placed one byte after the zeropage reference
		base := s.reg.getPC()
		address, _ = addOffsetRelative(base, line[2])
	default:
		panic("Assert failed. Missing addressing mode")
	}

	if extraCycle {
		s.extraCycleCrossingBoundaries = true
	}

	return address
}

/*
Note: extra cycle on reads when crossing page boundaries.

Only for:
	modeAbsoluteX
	modeAbsoluteY
	modeIndirectIndexedY
	modeRelative
	modeZeroPageAndRelative
That is when we add a 8 bit offset to a 16 bit base. The reason is
that if don't have a page crossing the CPU optimizes one cycle assuming
that the MSB addition won't change. If it does we spend this extra cycle.

Note that for writes we don't add a cycle in this case. There is no
optimization that could make a double write. The regular cycle count
is alwaus the same with no optimization.
*/
func addOffset(base uint16, offset uint8) (uint16, bool) {
	dest := base + uint16(offset)
	if (base & 0xff00) != (dest & 0xff00) {
		return dest, true
	}
	return dest, false
}

func addOffsetRelative(base uint16, offset uint8) (uint16, bool) {
	dest := base + uint16(int8(offset))
	if (base & 0xff00) != (dest & 0xff00) {
		return dest, true
	}
	return dest, false
}

func lineString(line []uint8, opcode opcode) string {
	t := opcode.name
	switch opcode.addressMode {
	case modeImplicit:
	case modeImplicitX:
	case modeImplicitY:
		//Nothing
	case modeAccumulator:
		t += " A"
	case modeImmediate:
		t += fmt.Sprintf(" #$%02x", line[1])
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
	case modeAbsoluteX65c02:
		fallthrough
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
		t += fmt.Sprintf(" $%02x %+x", line[1], int8(line[2]))
	default:
		t += "UNKNOWN MODE"
	}
	return t
}
