package core6502

// NewNMOS6502 returns an initialized NMOS6502
func NewNMOS6502(m Memory) *State {
	var s State
	s.mem = m
	s.opcodes = &opcodesNMOS6502
	return &s
}

var opcodesNMOS6502 = [256]opcode{
	0x00: opcode{"BRK", 1, 7, modeImplicit, opBRK},
	0x4C: opcode{"JMP", 3, 3, modeAbsolute, opJMP},
	0x6C: opcode{"JMP", 3, 3, modeIndirect, opJMP},
	0x20: opcode{"JSR", 3, 6, modeAbsolute, opJSR},
	0x40: opcode{"RTI", 1, 6, modeImplicit, opRTI},
	0x60: opcode{"RTS", 1, 6, modeImplicit, opRTS},

	0x48: opcode{"PHA", 1, 3, modeImplicit, opPHA},
	0x08: opcode{"PHP", 1, 3, modeImplicit, opPHP},
	0x68: opcode{"PLA", 1, 4, modeImplicit, opPLA},
	0x28: opcode{"PLP", 1, 4, modeImplicit, opPLP},

	0x09: opcode{"ORA", 2, 2, modeImmediate, buildOpLogic(operationOr)},
	0x05: opcode{"ORA", 2, 3, modeZeroPage, buildOpLogic(operationOr)},
	0x15: opcode{"ORA", 2, 4, modeZeroPageX, buildOpLogic(operationOr)},
	0x0D: opcode{"ORA", 3, 4, modeAbsolute, buildOpLogic(operationOr)},
	0x1D: opcode{"ORA", 3, 4, modeAbsoluteX, buildOpLogic(operationOr)}, // Extra cycles
	0x19: opcode{"ORA", 3, 4, modeAbsoluteY, buildOpLogic(operationOr)}, // Extra cycles
	0x01: opcode{"ORA", 2, 6, modeIndexedIndirectX, buildOpLogic(operationOr)},
	0x11: opcode{"ORA", 2, 5, modeIndirectIndexedY, buildOpLogic(operationOr)}, // Extra cycles

	0x29: opcode{"AND", 2, 2, modeImmediate, buildOpLogic(operationAnd)},
	0x25: opcode{"AND", 2, 3, modeZeroPage, buildOpLogic(operationAnd)},
	0x35: opcode{"AND", 2, 4, modeZeroPageX, buildOpLogic(operationAnd)},
	0x2D: opcode{"AND", 3, 4, modeAbsolute, buildOpLogic(operationAnd)},
	0x3D: opcode{"AND", 3, 4, modeAbsoluteX, buildOpLogic(operationAnd)}, // Extra cycles
	0x39: opcode{"AND", 3, 4, modeAbsoluteY, buildOpLogic(operationAnd)}, // Extra cycles
	0x21: opcode{"AND", 2, 6, modeIndexedIndirectX, buildOpLogic(operationAnd)},
	0x31: opcode{"AND", 2, 5, modeIndirectIndexedY, buildOpLogic(operationAnd)}, // Extra cycles

	0x49: opcode{"EOR", 2, 2, modeImmediate, buildOpLogic(operationXor)},
	0x45: opcode{"EOR", 2, 3, modeZeroPage, buildOpLogic(operationXor)},
	0x55: opcode{"EOR", 2, 4, modeZeroPageX, buildOpLogic(operationXor)},
	0x4D: opcode{"EOR", 3, 4, modeAbsolute, buildOpLogic(operationXor)},
	0x5D: opcode{"EOR", 3, 4, modeAbsoluteX, buildOpLogic(operationXor)}, // Extra cycles
	0x59: opcode{"EOR", 3, 4, modeAbsoluteY, buildOpLogic(operationXor)}, // Extra cycles
	0x41: opcode{"EOR", 2, 6, modeIndexedIndirectX, buildOpLogic(operationXor)},
	0x51: opcode{"EOR", 2, 5, modeIndirectIndexedY, buildOpLogic(operationXor)}, // Extra cycles

	0x69: opcode{"ADC", 2, 2, modeImmediate, opADC},
	0x65: opcode{"ADC", 2, 3, modeZeroPage, opADC},
	0x75: opcode{"ADC", 2, 4, modeZeroPageX, opADC},
	0x6D: opcode{"ADC", 3, 4, modeAbsolute, opADC},
	0x7D: opcode{"ADC", 3, 4, modeAbsoluteX, opADC}, // Extra cycles
	0x79: opcode{"ADC", 3, 4, modeAbsoluteY, opADC}, // Extra cycles
	0x61: opcode{"ADC", 2, 6, modeIndexedIndirectX, opADC},
	0x71: opcode{"ADC", 2, 5, modeIndirectIndexedY, opADC}, // Extra cycles

	0xE9: opcode{"SBC", 2, 2, modeImmediate, opSBC},
	0xE5: opcode{"SBC", 2, 3, modeZeroPage, opSBC},
	0xF5: opcode{"SBC", 2, 4, modeZeroPageX, opSBC},
	0xED: opcode{"SBC", 3, 4, modeAbsolute, opSBC},
	0xFD: opcode{"SBC", 3, 4, modeAbsoluteX, opSBC}, // Extra cycles
	0xF9: opcode{"SBC", 3, 4, modeAbsoluteY, opSBC}, // Extra cycles
	0xE1: opcode{"SBC", 2, 6, modeIndexedIndirectX, opSBC},
	0xF1: opcode{"SBC", 2, 5, modeIndirectIndexedY, opSBC}, // Extra cycles

	0x24: opcode{"BIT", 2, 3, modeZeroPage, opBIT},
	0x2C: opcode{"BIT", 3, 3, modeAbsolute, opBIT},

	0xC9: opcode{"CMP", 2, 2, modeImmediate, buildOpCompare(regA)},
	0xC5: opcode{"CMP", 2, 3, modeZeroPage, buildOpCompare(regA)},
	0xD5: opcode{"CMP", 2, 4, modeZeroPageX, buildOpCompare(regA)},
	0xCD: opcode{"CMP", 3, 4, modeAbsolute, buildOpCompare(regA)},
	0xDD: opcode{"CMP", 3, 4, modeAbsoluteX, buildOpCompare(regA)}, // Extra cycles
	0xD9: opcode{"CMP", 3, 4, modeAbsoluteY, buildOpCompare(regA)}, // Extra cycles
	0xC1: opcode{"CMP", 2, 6, modeIndexedIndirectX, buildOpCompare(regA)},
	0xD1: opcode{"CMP", 2, 5, modeIndirectIndexedY, buildOpCompare(regA)}, // Extra cycles

	0xE0: opcode{"CPX", 2, 2, modeImmediate, buildOpCompare(regX)},
	0xE4: opcode{"CPX", 2, 3, modeZeroPage, buildOpCompare(regX)},
	0xEC: opcode{"CPX", 3, 4, modeAbsolute, buildOpCompare(regX)},

	0xC0: opcode{"CPY", 2, 2, modeImmediate, buildOpCompare(regY)},
	0xC4: opcode{"CPY", 2, 3, modeZeroPage, buildOpCompare(regY)},
	0xCC: opcode{"CPY", 3, 4, modeAbsolute, buildOpCompare(regY)},

	0x2A: opcode{"ROL", 1, 2, modeAccumulator, buildOpShift(true, true)},
	0x26: opcode{"ROL", 2, 5, modeZeroPage, buildOpShift(true, true)},
	0x36: opcode{"ROL", 2, 6, modeZeroPageX, buildOpShift(true, true)},
	0x2E: opcode{"ROL", 3, 6, modeAbsolute, buildOpShift(true, true)},
	0x3E: opcode{"ROL", 3, 7, modeAbsoluteX, buildOpShift(true, true)},

	0x6A: opcode{"ROR", 1, 2, modeAccumulator, buildOpShift(false, true)},
	0x66: opcode{"ROR", 2, 5, modeZeroPage, buildOpShift(false, true)},
	0x76: opcode{"ROR", 2, 6, modeZeroPageX, buildOpShift(false, true)},
	0x6E: opcode{"ROR", 3, 6, modeAbsolute, buildOpShift(false, true)},
	0x7E: opcode{"ROR", 3, 7, modeAbsoluteX, buildOpShift(false, true)},

	0x0A: opcode{"ASL", 1, 2, modeAccumulator, buildOpShift(true, false)},
	0x06: opcode{"ASL", 2, 5, modeZeroPage, buildOpShift(true, false)},
	0x16: opcode{"ASL", 2, 6, modeZeroPageX, buildOpShift(true, false)},
	0x0E: opcode{"ASL", 3, 6, modeAbsolute, buildOpShift(true, false)},
	0x1E: opcode{"ASL", 3, 7, modeAbsoluteX, buildOpShift(true, false)},

	0x4A: opcode{"LSR", 1, 2, modeAccumulator, buildOpShift(false, false)},
	0x46: opcode{"LSR", 2, 5, modeZeroPage, buildOpShift(false, false)},
	0x56: opcode{"LSR", 2, 6, modeZeroPageX, buildOpShift(false, false)},
	0x4E: opcode{"LSR", 3, 6, modeAbsolute, buildOpShift(false, false)},
	0x5E: opcode{"LSR", 3, 7, modeAbsoluteX, buildOpShift(false, false)},

	0x38: opcode{"SEC", 1, 2, modeImplicit, buildOpUpdateFlag(flagC, true)},
	0xF8: opcode{"SED", 1, 2, modeImplicit, buildOpUpdateFlag(flagD, true)},
	0x78: opcode{"SEI", 1, 2, modeImplicit, buildOpUpdateFlag(flagI, true)},
	0x18: opcode{"CLC", 1, 2, modeImplicit, buildOpUpdateFlag(flagC, false)},
	0xD8: opcode{"CLD", 1, 2, modeImplicit, buildOpUpdateFlag(flagD, false)},
	0x58: opcode{"CLI", 1, 2, modeImplicit, buildOpUpdateFlag(flagI, false)},
	0xB8: opcode{"CLV", 1, 2, modeImplicit, buildOpUpdateFlag(flagV, false)},

	0xE6: opcode{"INC", 2, 5, modeZeroPage, buildOpIncDec(true)},
	0xF6: opcode{"INC", 2, 6, modeZeroPageX, buildOpIncDec(true)},
	0xEE: opcode{"INC", 3, 6, modeAbsolute, buildOpIncDec(true)},
	0xFE: opcode{"INC", 3, 7, modeAbsoluteX, buildOpIncDec(true)},
	0xC6: opcode{"DEC", 2, 5, modeZeroPage, buildOpIncDec(false)},
	0xD6: opcode{"DEC", 2, 6, modeZeroPageX, buildOpIncDec(false)},
	0xCE: opcode{"DEC", 3, 6, modeAbsolute, buildOpIncDec(false)},
	0xDE: opcode{"DEC", 3, 7, modeAbsoluteX, buildOpIncDec(false)},
	0xE8: opcode{"INX", 1, 2, modeImplicitX, buildOpIncDec(true)},
	0xC8: opcode{"INY", 1, 2, modeImplicitY, buildOpIncDec(true)},
	0xCA: opcode{"DEX", 1, 2, modeImplicitX, buildOpIncDec(false)},
	0x88: opcode{"DEY", 1, 2, modeImplicitY, buildOpIncDec(false)},

	0xAA: opcode{"TAX", 1, 2, modeImplicit, buildOpTransfer(regA, regX)},
	0xA8: opcode{"TAY", 1, 2, modeImplicit, buildOpTransfer(regA, regY)},
	0x8A: opcode{"TXA", 1, 2, modeImplicit, buildOpTransfer(regX, regA)},
	0x98: opcode{"TYA", 1, 2, modeImplicit, buildOpTransfer(regY, regA)},
	0x9A: opcode{"TXS", 1, 2, modeImplicit, buildOpTransfer(regX, regSP)},
	0xBA: opcode{"TSX", 1, 2, modeImplicit, buildOpTransfer(regSP, regX)},

	0xA9: opcode{"LDA", 2, 2, modeImmediate, buildOpLoad(regA)},
	0xA5: opcode{"LDA", 2, 3, modeZeroPage, buildOpLoad(regA)},
	0xB5: opcode{"LDA", 2, 4, modeZeroPageX, buildOpLoad(regA)},
	0xAD: opcode{"LDA", 3, 4, modeAbsolute, buildOpLoad(regA)},
	0xBD: opcode{"LDA", 3, 4, modeAbsoluteX, buildOpLoad(regA)}, // Extra cycles
	0xB9: opcode{"LDA", 3, 4, modeAbsoluteY, buildOpLoad(regA)}, // Extra cycles
	0xA1: opcode{"LDA", 2, 6, modeIndexedIndirectX, buildOpLoad(regA)},
	0xB1: opcode{"LDA", 2, 5, modeIndirectIndexedY, buildOpLoad(regA)}, // Extra cycles
	0xA2: opcode{"LDX", 2, 2, modeImmediate, buildOpLoad(regX)},
	0xA6: opcode{"LDX", 2, 3, modeZeroPage, buildOpLoad(regX)},
	0xB6: opcode{"LDX", 2, 4, modeZeroPageY, buildOpLoad(regX)},
	0xAE: opcode{"LDX", 3, 4, modeAbsolute, buildOpLoad(regX)},
	0xBE: opcode{"LDX", 3, 4, modeAbsoluteY, buildOpLoad(regX)}, // Extra cycles
	0xA0: opcode{"LDY", 2, 2, modeImmediate, buildOpLoad(regY)},
	0xA4: opcode{"LDY", 2, 3, modeZeroPage, buildOpLoad(regY)},
	0xB4: opcode{"LDY", 2, 4, modeZeroPageX, buildOpLoad(regY)},
	0xAC: opcode{"LDY", 3, 4, modeAbsolute, buildOpLoad(regY)},
	0xBC: opcode{"LDY", 3, 4, modeAbsoluteX, buildOpLoad(regY)}, // Extra cycles

	0x85: opcode{"STA", 2, 3, modeZeroPage, buildOpStore(regA)},
	0x95: opcode{"STA", 2, 4, modeZeroPageX, buildOpStore(regA)},
	0x8D: opcode{"STA", 3, 4, modeAbsolute, buildOpStore(regA)},
	0x9D: opcode{"STA", 3, 5, modeAbsoluteX, buildOpStore(regA)},
	0x99: opcode{"STA", 3, 5, modeAbsoluteY, buildOpStore(regA)},
	0x81: opcode{"STA", 2, 6, modeIndexedIndirectX, buildOpStore(regA)},
	0x91: opcode{"STA", 2, 6, modeIndirectIndexedY, buildOpStore(regA)},
	0x86: opcode{"STX", 2, 3, modeZeroPage, buildOpStore(regX)},
	0x96: opcode{"STX", 2, 4, modeZeroPageY, buildOpStore(regX)},
	0x8E: opcode{"STX", 3, 4, modeAbsolute, buildOpStore(regX)},
	0x84: opcode{"STY", 2, 3, modeZeroPage, buildOpStore(regY)},
	0x94: opcode{"STY", 2, 4, modeZeroPageX, buildOpStore(regY)},
	0x8C: opcode{"STY", 3, 4, modeAbsolute, buildOpStore(regY)},

	0x90: opcode{"BCC", 2, 2, modeRelative, buildOpBranch(flagC, false)}, // Extra cycles
	0xB0: opcode{"BCS", 2, 2, modeRelative, buildOpBranch(flagC, true)},  // Extra cycles
	0xD0: opcode{"BNE", 2, 2, modeRelative, buildOpBranch(flagZ, false)}, // Extra cycles
	0xF0: opcode{"BEQ", 2, 2, modeRelative, buildOpBranch(flagZ, true)},  // Extra cycles
	0x10: opcode{"BPL", 2, 2, modeRelative, buildOpBranch(flagN, false)}, // Extra cycles
	0x30: opcode{"BMI", 2, 2, modeRelative, buildOpBranch(flagN, true)},  // Extra cycles
	0x50: opcode{"BVC", 2, 2, modeRelative, buildOpBranch(flagV, false)}, // Extra cycles
	0x70: opcode{"BVS", 2, 2, modeRelative, buildOpBranch(flagV, true)},  // Extra cycles

	0xEA: opcode{"NOP", 1, 2, modeImplicit, opNOP},
}
