package core6502

/*
For the diffrences with NMOS6502 see:
	http://6502.org/tutorials/65c02opcodes.html
	http://wilsonminesco.com/NMOS-CMOSdif/
	http://www.obelisk.me.uk/65C02/reference.html
	http://www.obelisk.me.uk/65C02/addressing.html
	http://anyplatform.net/media/guides/cpus/65xx%20Processor%20Data.txt
*/

// NewCMOS65c02 returns an initialized 65c02
func NewCMOS65c02(m Memory) *State {
	var s State
	s.mem = m

	var opcodes [256]opcode
	for i := 0; i < 256; i++ {
		opcodes[i] = opcodesNMOS6502[i]
		if opcodes65c02Delta[i].cycles != 0 {
			opcodes[i] = opcodes65c02Delta[i]
		}
	}
	add65c02NOPs(&opcodes)
	s.opcodes = &opcodes
	return &s
}

func add65c02NOPs(opcodes *[256]opcode) {
	nop11 := opcode{"NOP", 1, 1, modeImplicit, opNOP}
	nop22 := opcode{"NOP", 2, 2, modeImmediate, opNOP}
	nop23 := opcode{"NOP", 2, 3, modeImmediate, opNOP}
	nop24 := opcode{"NOP", 2, 4, modeImmediate, opNOP}
	nop34 := opcode{"NOP", 3, 4, modeAbsolute, opNOP}

	opcodes[0x02] = nop22
	opcodes[0x22] = nop22
	opcodes[0x42] = nop22
	opcodes[0x62] = nop22
	opcodes[0x82] = nop22
	opcodes[0xc2] = nop22
	opcodes[0xe2] = nop22

	opcodes[0x44] = nop23
	opcodes[0x54] = nop24
	opcodes[0xD4] = nop24
	opcodes[0xF4] = nop24

	opcodes[0x5c] = nop34
	opcodes[0xdc] = nop34
	opcodes[0xfc] = nop34

	for i := 0; i < 0x100; i = i + 0x10 {
		opcodes[i+0x03] = nop11
		// RMB and SMB; opcodes[i+0x07] = nop11
		opcodes[i+0x0b] = nop11
		// BBR and BBS: opcodes[i+0x0f] = nop11
	}

	/* Changes for Rockwell65c02
	nop12 := opcode{"NOP", 1, 2, modeImplicit, opNOP}
	opcodes[0xcb] = nop12
	opcodes[0xdb] = nop24
	*/

	// Detection of 65c816
	opcodes[0xbf].name = "XCE"
}

var opcodes65c02Delta = [256]opcode{
	// Functional difference
	0x00: {"BRK", 1, 7, modeImplicit, opBRKAlt},
	0x24: {"BIT", 2, 3, modeZeroPage, opBIT},
	0x6C: {"JMP", 3, 6, modeIndirect65c02Fix, opJMP},

	// Fixed BCD arithmetic flags
	0x69: {"ADC", 2, 2, modeImmediate, opADCAlt},
	0x65: {"ADC", 2, 3, modeZeroPage, opADCAlt},
	0x75: {"ADC", 2, 4, modeZeroPageX, opADCAlt},
	0x6D: {"ADC", 3, 4, modeAbsolute, opADCAlt},
	0x7D: {"ADC", 3, 4, modeAbsoluteX, opADCAlt}, // Extra cycles
	0x79: {"ADC", 3, 4, modeAbsoluteY, opADCAlt}, // Extra cycles
	0x61: {"ADC", 2, 6, modeIndexedIndirectX, opADCAlt},
	0x71: {"ADC", 2, 5, modeIndirectIndexedY, opADCAlt}, // Extra cycles
	0xE9: {"SBC", 2, 2, modeImmediate, opSBCAlt},
	0xE5: {"SBC", 2, 3, modeZeroPage, opSBCAlt},
	0xF5: {"SBC", 2, 4, modeZeroPageX, opSBCAlt},
	0xED: {"SBC", 3, 4, modeAbsolute, opSBCAlt},
	0xFD: {"SBC", 3, 4, modeAbsoluteX, opSBCAlt}, // Extra cycles
	0xF9: {"SBC", 3, 4, modeAbsoluteY, opSBCAlt}, // Extra cycles
	0xE1: {"SBC", 2, 6, modeIndexedIndirectX, opSBCAlt},
	0xF1: {"SBC", 2, 5, modeIndirectIndexedY, opSBCAlt}, // Extra cycles

	// Different cycle count
	0x1e: {"ASL", 3, 6, modeAbsoluteX65c02, buildOpShift(true, false)},
	0x3e: {"ROL", 3, 6, modeAbsoluteX65c02, buildOpShift(true, true)},
	0x5e: {"LSR", 3, 6, modeAbsoluteX65c02, buildOpShift(false, false)},
	0x7e: {"ROR", 3, 6, modeAbsoluteX65c02, buildOpShift(false, true)},

	// New indirect zero page addresssing mode
	0x12: {"ORA", 2, 5, modeIndirectZeroPage, buildOpLogic(operationOr)},
	0x32: {"AND", 2, 5, modeIndirectZeroPage, buildOpLogic(operationAnd)},
	0x52: {"EOR", 2, 5, modeIndirectZeroPage, buildOpLogic(operationXor)},
	0x72: {"ADC", 2, 5, modeIndirectZeroPage, opADCAlt},
	0x92: {"STA", 2, 5, modeIndirectZeroPage, buildOpStore(regA)},
	0xb2: {"LDA", 2, 5, modeIndirectZeroPage, buildOpLoad(regA)},
	0xd2: {"CMP", 2, 5, modeIndirectZeroPage, buildOpCompare(regA)},
	0xf2: {"SBC", 2, 5, modeIndirectZeroPage, opSBCAlt},

	// New addressing options
	0x89: {"BIT", 2, 2, modeImmediate, opBIT},
	0x34: {"BIT", 2, 4, modeZeroPageX, opBIT},
	0x3c: {"BIT", 3, 4, modeAbsoluteX, opBIT}, // Extra cycles
	0x1a: {"INC", 1, 2, modeAccumulator, buildOpIncDec(true)},
	0x3a: {"DEC", 1, 2, modeAccumulator, buildOpIncDec(false)},
	0x7c: {"JMP", 3, 6, modeAbsoluteIndexedIndirectX, opJMP},

	// Additional instructions: BRA, PHX, PHY, PLX, PLY, STZ, TRB, TSB
	0xda: {"PHX", 1, 3, modeImplicit, buildOpPush(regX)},
	0x5a: {"PHY", 1, 3, modeImplicit, buildOpPush(regY)},
	0xfa: {"PLX", 1, 4, modeImplicit, buildOpPull(regX)},
	0x7a: {"PLY", 1, 4, modeImplicit, buildOpPull(regY)},
	0x80: {"BRA", 2, 3, modeRelative, opJMP}, // Extra cycles

	0x64: {"STZ", 2, 3, modeZeroPage, opSTZ},
	0x74: {"STZ", 2, 4, modeZeroPageX, opSTZ},
	0x9c: {"STZ", 3, 4, modeAbsolute, opSTZ},
	0x9e: {"STZ", 3, 5, modeAbsoluteX, opSTZ},

	0x14: {"TRB", 2, 5, modeZeroPage, opTRB},
	0x1c: {"TRB", 3, 6, modeAbsolute, opTRB},

	0x04: {"TSB", 2, 5, modeZeroPage, opTSB},
	0x0c: {"TSB", 3, 6, modeAbsolute, opTSB},

	// Additional in Rockwell 65c02 and WDC 65c02?
	// They have a double addressing mode: zeropage and relative.
	0x0f: {"BBR0", 3, 6, modeZeroPageAndRelative, buildOpBranchOnBit(0, false)}, // Extra cycles
	0x1f: {"BBR1", 3, 6, modeZeroPageAndRelative, buildOpBranchOnBit(1, false)}, // Extra cycles
	0x2f: {"BBR2", 3, 6, modeZeroPageAndRelative, buildOpBranchOnBit(2, false)}, // Extra cycles
	0x3f: {"BBR3", 3, 6, modeZeroPageAndRelative, buildOpBranchOnBit(3, false)}, // Extra cycles
	0x4f: {"BBR4", 3, 6, modeZeroPageAndRelative, buildOpBranchOnBit(4, false)}, // Extra cycles
	0x5f: {"BBR5", 3, 6, modeZeroPageAndRelative, buildOpBranchOnBit(5, false)}, // Extra cycles
	0x6f: {"BBR6", 3, 6, modeZeroPageAndRelative, buildOpBranchOnBit(6, false)}, // Extra cycles
	0x7f: {"BBR7", 3, 6, modeZeroPageAndRelative, buildOpBranchOnBit(7, false)}, // Extra cycles
	0x8f: {"BBS0", 3, 6, modeZeroPageAndRelative, buildOpBranchOnBit(0, true)},  // Extra cycles
	0x9f: {"BBS1", 3, 6, modeZeroPageAndRelative, buildOpBranchOnBit(1, true)},  // Extra cycles
	0xaf: {"BBS2", 3, 6, modeZeroPageAndRelative, buildOpBranchOnBit(2, true)},  // Extra cycles
	0xbf: {"BBS3", 3, 6, modeZeroPageAndRelative, buildOpBranchOnBit(3, true)},  // Extra cycles
	0xcf: {"BBS4", 3, 6, modeZeroPageAndRelative, buildOpBranchOnBit(4, true)},  // Extra cycles
	0xdf: {"BBS5", 3, 6, modeZeroPageAndRelative, buildOpBranchOnBit(5, true)},  // Extra cycles
	0xef: {"BBS6", 3, 6, modeZeroPageAndRelative, buildOpBranchOnBit(6, true)},  // Extra cycles
	0xff: {"BBS7", 3, 6, modeZeroPageAndRelative, buildOpBranchOnBit(7, true)},  // Extra cycles

	0x07: {"RMB0", 2, 5, modeZeroPage, buildOpSetBit(0, false)},
	0x17: {"RMB1", 2, 5, modeZeroPage, buildOpSetBit(1, false)},
	0x27: {"RMB2", 2, 5, modeZeroPage, buildOpSetBit(2, false)},
	0x37: {"RMB3", 2, 5, modeZeroPage, buildOpSetBit(3, false)},
	0x47: {"RMB4", 2, 5, modeZeroPage, buildOpSetBit(4, false)},
	0x57: {"RMB5", 2, 5, modeZeroPage, buildOpSetBit(5, false)},
	0x67: {"RMB6", 2, 5, modeZeroPage, buildOpSetBit(6, false)},
	0x77: {"RMB7", 2, 5, modeZeroPage, buildOpSetBit(7, false)},
	0x87: {"SMB0", 2, 5, modeZeroPage, buildOpSetBit(0, true)},
	0x97: {"SMB1", 2, 5, modeZeroPage, buildOpSetBit(1, true)},
	0xa7: {"SMB2", 2, 5, modeZeroPage, buildOpSetBit(2, true)},
	0xb7: {"SMB3", 2, 5, modeZeroPage, buildOpSetBit(3, true)},
	0xc7: {"SMB4", 2, 5, modeZeroPage, buildOpSetBit(4, true)},
	0xd7: {"SMB5", 2, 5, modeZeroPage, buildOpSetBit(5, true)},
	0xe7: {"SMB6", 2, 5, modeZeroPage, buildOpSetBit(6, true)},
	0xf7: {"SMB7", 2, 5, modeZeroPage, buildOpSetBit(7, true)},

	// Maybe additional Rockwell: STP, WAI
}
