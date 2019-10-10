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
	getValue, _, _ := resolve(s, line, opcode)
	return getValue()
}

func resolveGetSetValue(s *State, line []uint8, opcode opcode) (value uint8, setValue func(uint8)) {
	getValue, setValue, _ := resolve(s, line, opcode)
	value = getValue()
	return
}

func resolveSetValue(s *State, line []uint8, opcode opcode) func(uint8) {
	_, setValue, _ := resolve(s, line, opcode)
	return setValue
}

func resolveAddress(s *State, line []uint8, opcode opcode) uint16 {
	_, _, address := resolve(s, line, opcode)
	return address
}

func resolve(s *State, line []uint8, opcode opcode) (getValue func() uint8, setValue func(uint8), address uint16) {
	hasAddress := true
	register := regNone

	switch opcode.addressMode {
	case modeAccumulator:
		getValue = func() uint8 { return s.reg.getA() }
		hasAddress = false
		register = regA
	case modeImplicitX:
		getValue = func() uint8 { return s.reg.getX() }
		hasAddress = false
		register = regX
	case modeImplicitY:
		getValue = func() uint8 { return s.reg.getY() }
		hasAddress = false
		register = regY
	case modeImmediate:
		getValue = func() uint8 { return line[1] }
		hasAddress = false
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

	if hasAddress {
		getValue = func() uint8 { return s.mem.Peek(address) }
	}

	setValue = func(value uint8) {
		if hasAddress {
			s.mem.Poke(address, value)
		} else if register != regNone {
			s.reg.setRegister(register, value)
		} else {
			panic("Assert failed. Should never happen")
		}
	}
	return
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
