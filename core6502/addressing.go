package core6502

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
			// Todo: assert impossible
		}
	}
	return
}
