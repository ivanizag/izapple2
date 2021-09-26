package core6502

// NewNMOS6502 returns an initialized NMOS6502
func NewNMOS6502(m Memory) *State {
	var s State
	s.mem = m
	s.opcodes = &opcodesNMOS6502
	return &s
}

var opcodesNMOS6502 = [256]opcode{
	0x00: {"BRK", 1, 7, modeImplicit, opBRK},
	0x4C: {"JMP", 3, 3, modeAbsolute, opJMP},
	0x6C: {"JMP", 3, 5, modeIndirect, opJMP},
	0x20: {"JSR", 3, 6, modeAbsolute, opJSR},
	0x40: {"RTI", 1, 6, modeImplicit, opRTI},
	0x60: {"RTS", 1, 6, modeImplicit, opRTS},

	0x48: {"PHA", 1, 3, modeImplicit, buildOpPush(regA)},
	0x08: {"PHP", 1, 3, modeImplicit, buildOpPush(regP)},
	0x68: {"PLA", 1, 4, modeImplicit, buildOpPull(regA)},
	0x28: {"PLP", 1, 4, modeImplicit, buildOpPull(regP)},

	0x09: {"ORA", 2, 2, modeImmediate, buildOpLogic(operationOr)},
	0x05: {"ORA", 2, 3, modeZeroPage, buildOpLogic(operationOr)},
	0x15: {"ORA", 2, 4, modeZeroPageX, buildOpLogic(operationOr)},
	0x0D: {"ORA", 3, 4, modeAbsolute, buildOpLogic(operationOr)},
	0x1D: {"ORA", 3, 4, modeAbsoluteX, buildOpLogic(operationOr)}, // Extra cycles
	0x19: {"ORA", 3, 4, modeAbsoluteY, buildOpLogic(operationOr)}, // Extra cycles
	0x01: {"ORA", 2, 6, modeIndexedIndirectX, buildOpLogic(operationOr)},
	0x11: {"ORA", 2, 5, modeIndirectIndexedY, buildOpLogic(operationOr)}, // Extra cycles

	0x29: {"AND", 2, 2, modeImmediate, buildOpLogic(operationAnd)},
	0x25: {"AND", 2, 3, modeZeroPage, buildOpLogic(operationAnd)},
	0x35: {"AND", 2, 4, modeZeroPageX, buildOpLogic(operationAnd)},
	0x2D: {"AND", 3, 4, modeAbsolute, buildOpLogic(operationAnd)},
	0x3D: {"AND", 3, 4, modeAbsoluteX, buildOpLogic(operationAnd)}, // Extra cycles
	0x39: {"AND", 3, 4, modeAbsoluteY, buildOpLogic(operationAnd)}, // Extra cycles
	0x21: {"AND", 2, 6, modeIndexedIndirectX, buildOpLogic(operationAnd)},
	0x31: {"AND", 2, 5, modeIndirectIndexedY, buildOpLogic(operationAnd)}, // Extra cycles

	0x49: {"EOR", 2, 2, modeImmediate, buildOpLogic(operationXor)},
	0x45: {"EOR", 2, 3, modeZeroPage, buildOpLogic(operationXor)},
	0x55: {"EOR", 2, 4, modeZeroPageX, buildOpLogic(operationXor)},
	0x4D: {"EOR", 3, 4, modeAbsolute, buildOpLogic(operationXor)},
	0x5D: {"EOR", 3, 4, modeAbsoluteX, buildOpLogic(operationXor)}, // Extra cycles
	0x59: {"EOR", 3, 4, modeAbsoluteY, buildOpLogic(operationXor)}, // Extra cycles
	0x41: {"EOR", 2, 6, modeIndexedIndirectX, buildOpLogic(operationXor)},
	0x51: {"EOR", 2, 5, modeIndirectIndexedY, buildOpLogic(operationXor)}, // Extra cycles

	0x69: {"ADC", 2, 2, modeImmediate, opADC},
	0x65: {"ADC", 2, 3, modeZeroPage, opADC},
	0x75: {"ADC", 2, 4, modeZeroPageX, opADC},
	0x6D: {"ADC", 3, 4, modeAbsolute, opADC},
	0x7D: {"ADC", 3, 4, modeAbsoluteX, opADC}, // Extra cycles
	0x79: {"ADC", 3, 4, modeAbsoluteY, opADC}, // Extra cycles
	0x61: {"ADC", 2, 6, modeIndexedIndirectX, opADC},
	0x71: {"ADC", 2, 5, modeIndirectIndexedY, opADC}, // Extra cycles

	0xE9: {"SBC", 2, 2, modeImmediate, opSBC},
	0xE5: {"SBC", 2, 3, modeZeroPage, opSBC},
	0xF5: {"SBC", 2, 4, modeZeroPageX, opSBC},
	0xED: {"SBC", 3, 4, modeAbsolute, opSBC},
	0xFD: {"SBC", 3, 4, modeAbsoluteX, opSBC}, // Extra cycles
	0xF9: {"SBC", 3, 4, modeAbsoluteY, opSBC}, // Extra cycles
	0xE1: {"SBC", 2, 6, modeIndexedIndirectX, opSBC},
	0xF1: {"SBC", 2, 5, modeIndirectIndexedY, opSBC}, // Extra cycles

	0x24: {"BIT", 2, 3, modeZeroPage, opBIT},
	0x2C: {"BIT", 3, 4, modeAbsolute, opBIT},

	0xC9: {"CMP", 2, 2, modeImmediate, buildOpCompare(regA)},
	0xC5: {"CMP", 2, 3, modeZeroPage, buildOpCompare(regA)},
	0xD5: {"CMP", 2, 4, modeZeroPageX, buildOpCompare(regA)},
	0xCD: {"CMP", 3, 4, modeAbsolute, buildOpCompare(regA)},
	0xDD: {"CMP", 3, 4, modeAbsoluteX, buildOpCompare(regA)}, // Extra cycles
	0xD9: {"CMP", 3, 4, modeAbsoluteY, buildOpCompare(regA)}, // Extra cycles
	0xC1: {"CMP", 2, 6, modeIndexedIndirectX, buildOpCompare(regA)},
	0xD1: {"CMP", 2, 5, modeIndirectIndexedY, buildOpCompare(regA)}, // Extra cycles

	0xE0: {"CPX", 2, 2, modeImmediate, buildOpCompare(regX)},
	0xE4: {"CPX", 2, 3, modeZeroPage, buildOpCompare(regX)},
	0xEC: {"CPX", 3, 4, modeAbsolute, buildOpCompare(regX)},

	0xC0: {"CPY", 2, 2, modeImmediate, buildOpCompare(regY)},
	0xC4: {"CPY", 2, 3, modeZeroPage, buildOpCompare(regY)},
	0xCC: {"CPY", 3, 4, modeAbsolute, buildOpCompare(regY)},

	0x2A: {"ROL", 1, 2, modeAccumulator, buildOpShift(true, true)},
	0x26: {"ROL", 2, 5, modeZeroPage, buildOpShift(true, true)},
	0x36: {"ROL", 2, 6, modeZeroPageX, buildOpShift(true, true)},
	0x2E: {"ROL", 3, 6, modeAbsolute, buildOpShift(true, true)},
	0x3E: {"ROL", 3, 7, modeAbsoluteX, buildOpShift(true, true)},

	0x6A: {"ROR", 1, 2, modeAccumulator, buildOpShift(false, true)},
	0x66: {"ROR", 2, 5, modeZeroPage, buildOpShift(false, true)},
	0x76: {"ROR", 2, 6, modeZeroPageX, buildOpShift(false, true)},
	0x6E: {"ROR", 3, 6, modeAbsolute, buildOpShift(false, true)},
	0x7E: {"ROR", 3, 7, modeAbsoluteX, buildOpShift(false, true)},

	0x0A: {"ASL", 1, 2, modeAccumulator, buildOpShift(true, false)},
	0x06: {"ASL", 2, 5, modeZeroPage, buildOpShift(true, false)},
	0x16: {"ASL", 2, 6, modeZeroPageX, buildOpShift(true, false)},
	0x0E: {"ASL", 3, 6, modeAbsolute, buildOpShift(true, false)},
	0x1E: {"ASL", 3, 7, modeAbsoluteX, buildOpShift(true, false)},

	0x4A: {"LSR", 1, 2, modeAccumulator, buildOpShift(false, false)},
	0x46: {"LSR", 2, 5, modeZeroPage, buildOpShift(false, false)},
	0x56: {"LSR", 2, 6, modeZeroPageX, buildOpShift(false, false)},
	0x4E: {"LSR", 3, 6, modeAbsolute, buildOpShift(false, false)},
	0x5E: {"LSR", 3, 7, modeAbsoluteX, buildOpShift(false, false)},

	0x38: {"SEC", 1, 2, modeImplicit, buildOpUpdateFlag(flagC, true)},
	0xF8: {"SED", 1, 2, modeImplicit, buildOpUpdateFlag(flagD, true)},
	0x78: {"SEI", 1, 2, modeImplicit, buildOpUpdateFlag(flagI, true)},
	0x18: {"CLC", 1, 2, modeImplicit, buildOpUpdateFlag(flagC, false)},
	0xD8: {"CLD", 1, 2, modeImplicit, buildOpUpdateFlag(flagD, false)},
	0x58: {"CLI", 1, 2, modeImplicit, buildOpUpdateFlag(flagI, false)},
	0xB8: {"CLV", 1, 2, modeImplicit, buildOpUpdateFlag(flagV, false)},

	0xE6: {"INC", 2, 5, modeZeroPage, buildOpIncDec(true)},
	0xF6: {"INC", 2, 6, modeZeroPageX, buildOpIncDec(true)},
	0xEE: {"INC", 3, 6, modeAbsolute, buildOpIncDec(true)},
	0xFE: {"INC", 3, 7, modeAbsoluteX, buildOpIncDec(true)},
	0xC6: {"DEC", 2, 5, modeZeroPage, buildOpIncDec(false)},
	0xD6: {"DEC", 2, 6, modeZeroPageX, buildOpIncDec(false)},
	0xCE: {"DEC", 3, 6, modeAbsolute, buildOpIncDec(false)},
	0xDE: {"DEC", 3, 7, modeAbsoluteX, buildOpIncDec(false)},
	0xE8: {"INX", 1, 2, modeImplicitX, buildOpIncDec(true)},
	0xC8: {"INY", 1, 2, modeImplicitY, buildOpIncDec(true)},
	0xCA: {"DEX", 1, 2, modeImplicitX, buildOpIncDec(false)},
	0x88: {"DEY", 1, 2, modeImplicitY, buildOpIncDec(false)},

	0xAA: {"TAX", 1, 2, modeImplicit, buildOpTransfer(regA, regX)},
	0xA8: {"TAY", 1, 2, modeImplicit, buildOpTransfer(regA, regY)},
	0x8A: {"TXA", 1, 2, modeImplicit, buildOpTransfer(regX, regA)},
	0x98: {"TYA", 1, 2, modeImplicit, buildOpTransfer(regY, regA)},
	0x9A: {"TXS", 1, 2, modeImplicit, buildOpTransfer(regX, regSP)},
	0xBA: {"TSX", 1, 2, modeImplicit, buildOpTransfer(regSP, regX)},

	0xA9: {"LDA", 2, 2, modeImmediate, buildOpLoad(regA)},
	0xA5: {"LDA", 2, 3, modeZeroPage, buildOpLoad(regA)},
	0xB5: {"LDA", 2, 4, modeZeroPageX, buildOpLoad(regA)},
	0xAD: {"LDA", 3, 4, modeAbsolute, buildOpLoad(regA)},
	0xBD: {"LDA", 3, 4, modeAbsoluteX, buildOpLoad(regA)}, // Extra cycles
	0xB9: {"LDA", 3, 4, modeAbsoluteY, buildOpLoad(regA)}, // Extra cycles
	0xA1: {"LDA", 2, 6, modeIndexedIndirectX, buildOpLoad(regA)},
	0xB1: {"LDA", 2, 5, modeIndirectIndexedY, buildOpLoad(regA)}, // Extra cycles
	0xA2: {"LDX", 2, 2, modeImmediate, buildOpLoad(regX)},
	0xA6: {"LDX", 2, 3, modeZeroPage, buildOpLoad(regX)},
	0xB6: {"LDX", 2, 4, modeZeroPageY, buildOpLoad(regX)},
	0xAE: {"LDX", 3, 4, modeAbsolute, buildOpLoad(regX)},
	0xBE: {"LDX", 3, 4, modeAbsoluteY, buildOpLoad(regX)}, // Extra cycles
	0xA0: {"LDY", 2, 2, modeImmediate, buildOpLoad(regY)},
	0xA4: {"LDY", 2, 3, modeZeroPage, buildOpLoad(regY)},
	0xB4: {"LDY", 2, 4, modeZeroPageX, buildOpLoad(regY)},
	0xAC: {"LDY", 3, 4, modeAbsolute, buildOpLoad(regY)},
	0xBC: {"LDY", 3, 4, modeAbsoluteX, buildOpLoad(regY)}, // Extra cycles

	0x85: {"STA", 2, 3, modeZeroPage, buildOpStore(regA)},
	0x95: {"STA", 2, 4, modeZeroPageX, buildOpStore(regA)},
	0x8D: {"STA", 3, 4, modeAbsolute, buildOpStore(regA)},
	0x9D: {"STA", 3, 5, modeAbsoluteX, buildOpStore(regA)},
	0x99: {"STA", 3, 5, modeAbsoluteY, buildOpStore(regA)},
	0x81: {"STA", 2, 6, modeIndexedIndirectX, buildOpStore(regA)},
	0x91: {"STA", 2, 6, modeIndirectIndexedY, buildOpStore(regA)},
	0x86: {"STX", 2, 3, modeZeroPage, buildOpStore(regX)},
	0x96: {"STX", 2, 4, modeZeroPageY, buildOpStore(regX)},
	0x8E: {"STX", 3, 4, modeAbsolute, buildOpStore(regX)},
	0x84: {"STY", 2, 3, modeZeroPage, buildOpStore(regY)},
	0x94: {"STY", 2, 4, modeZeroPageX, buildOpStore(regY)},
	0x8C: {"STY", 3, 4, modeAbsolute, buildOpStore(regY)},

	0x90: {"BCC", 2, 2, modeRelative, buildOpBranch(flagC, false)}, // Extra cycles
	0xB0: {"BCS", 2, 2, modeRelative, buildOpBranch(flagC, true)},  // Extra cycles
	0xD0: {"BNE", 2, 2, modeRelative, buildOpBranch(flagZ, false)}, // Extra cycles
	0xF0: {"BEQ", 2, 2, modeRelative, buildOpBranch(flagZ, true)},  // Extra cycles
	0x10: {"BPL", 2, 2, modeRelative, buildOpBranch(flagN, false)}, // Extra cycles
	0x30: {"BMI", 2, 2, modeRelative, buildOpBranch(flagN, true)},  // Extra cycles
	0x50: {"BVC", 2, 2, modeRelative, buildOpBranch(flagV, false)}, // Extra cycles
	0x70: {"BVS", 2, 2, modeRelative, buildOpBranch(flagV, true)},  // Extra cycles

	0xEA: {"NOP", 1, 2, modeImplicit, opNOP},

	/*
		Undocumented opcodes,
			see http://bbc.nvg.org/doc/6502OpList.txt
			see https://www.nesdev.com/undocumented_opcodes.txt
	*/
	0x1A: {"NOP", 1, 2, modeImplicit, opNOP},
	0x3A: {"NOP", 1, 2, modeImplicit, opNOP},
	0x5A: {"NOP", 1, 2, modeImplicit, opNOP},
	0x7A: {"NOP", 1, 2, modeImplicit, opNOP},
	0xDA: {"NOP", 1, 2, modeImplicit, opNOP},
	0xFA: {"NOP", 1, 2, modeImplicit, opNOP},

	0x04: {"DOP", 2, 3, modeImplicit, opNOP},
	0x14: {"DOP", 2, 4, modeImplicit, opNOP},
	0x34: {"DOP", 2, 4, modeImplicit, opNOP},
	0x44: {"DOP", 2, 3, modeImplicit, opNOP},
	0x54: {"DOP", 2, 4, modeImplicit, opNOP},
	0x64: {"DOP", 2, 3, modeImplicit, opNOP},
	0x74: {"DOP", 2, 4, modeImplicit, opNOP},
	0x80: {"DOP", 2, 2, modeImplicit, opNOP},
	0x82: {"DOP", 2, 2, modeImplicit, opNOP},
	0x89: {"DOP", 2, 2, modeImplicit, opNOP},
	0xC2: {"DOP", 2, 2, modeImplicit, opNOP},
	0xD4: {"DOP", 2, 4, modeImplicit, opNOP},
	0xE2: {"DOP", 2, 2, modeImplicit, opNOP},
	0xF4: {"DOP", 2, 4, modeImplicit, opNOP},

	0x0C: {"TOP", 3, 3, modeImplicit, opNOP},
	0x1C: {"TOP", 3, 4, modeImplicit, opNOP},
	0x3C: {"TOP", 3, 4, modeImplicit, opNOP},
	0x5C: {"TOP", 3, 4, modeImplicit, opNOP},
	0x7C: {"TOP", 3, 4, modeImplicit, opNOP},
	0xDC: {"TOP", 3, 4, modeImplicit, opNOP},
	0xFC: {"TOP", 3, 4, modeImplicit, opNOP},

	0x02: {"KIL", 1, 3, modeImplicit, opHALT},
	0x12: {"KIL", 1, 3, modeImplicit, opHALT},
	0x22: {"KIL", 1, 3, modeImplicit, opHALT},
	0x32: {"KIL", 1, 3, modeImplicit, opHALT},
	0x42: {"KIL", 1, 3, modeImplicit, opHALT},
	0x52: {"KIL", 1, 3, modeImplicit, opHALT},
	0x62: {"KIL", 1, 3, modeImplicit, opHALT},
	0x72: {"KIL", 1, 3, modeImplicit, opHALT},
	0x92: {"KIL", 1, 3, modeImplicit, opHALT},
	0xB2: {"KIL", 1, 3, modeImplicit, opHALT},
	0xD2: {"KIL", 1, 3, modeImplicit, opHALT},
	0xF2: {"KIL", 1, 3, modeImplicit, opHALT},
}
