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
	nop38 := opcode{"NOP", 3, 8, modeAbsolute, opNOP}

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

	opcodes[0x5c] = nop38
	opcodes[0xdc] = nop34
	opcodes[0xfc] = nop34

	for i := 0; i < 0x100; i = i + 0x10 {
		opcodes[i+0x03] = nop11
		// RMB and SMB; opcodes[i+0x07] = nop11
		opcodes[i+0x0b] = nop11
		// BBR and BBS: opcodes[i+0x0f] = nop11
	}

	// Detection of 65c816
	opcodes[0xbf].name = "XCE"
}

var opcodes65c02Delta = [256]opcode{
	// Functional difference
	0x00: opcode{"BRK", 1, 7, modeImplicit, opBRKAlt},
	0x24: opcode{"BIT", 2, 3, modeZeroPage, opBIT},
	0x2C: opcode{"BIT", 3, 3, modeAbsolute, opBIT},

	// Fixed BCD arithmetic flags
	0x69: opcode{"ADC", 2, 2, modeImmediate, opADCAlt},
	0x65: opcode{"ADC", 2, 3, modeZeroPage, opADCAlt},
	0x75: opcode{"ADC", 2, 4, modeZeroPageX, opADCAlt},
	0x6D: opcode{"ADC", 3, 4, modeAbsolute, opADCAlt},
	0x7D: opcode{"ADC", 3, 4, modeAbsoluteX, opADCAlt},
	0x79: opcode{"ADC", 3, 4, modeAbsoluteY, opADCAlt},
	0x61: opcode{"ADC", 2, 6, modeIndexedIndirectX, opADCAlt},
	0x71: opcode{"ADC", 2, 5, modeIndirectIndexedY, opADCAlt},
	0xE9: opcode{"SBC", 2, 2, modeImmediate, opSBCAlt},
	0xE5: opcode{"SBC", 2, 3, modeZeroPage, opSBCAlt},
	0xF5: opcode{"SBC", 2, 4, modeZeroPageX, opSBCAlt},
	0xED: opcode{"SBC", 3, 4, modeAbsolute, opSBCAlt},
	0xFD: opcode{"SBC", 3, 4, modeAbsoluteX, opSBCAlt},
	0xF9: opcode{"SBC", 3, 4, modeAbsoluteY, opSBCAlt},
	0xE1: opcode{"SBC", 2, 6, modeIndexedIndirectX, opSBCAlt},
	0xF1: opcode{"SBC", 2, 5, modeIndirectIndexedY, opSBCAlt},

	// Different cycle count
	0x1e: opcode{"ASL", 3, 6, modeAbsoluteX, buildOpShift(true, false)},
	0x3e: opcode{"ROL", 3, 6, modeAbsoluteX, buildOpShift(true, true)},
	0x5e: opcode{"LSR", 3, 6, modeAbsoluteX, buildOpShift(false, false)},
	0x7e: opcode{"ROR", 3, 6, modeAbsoluteX, buildOpShift(false, true)},

	// New indirect zero page addresssing mode
	0x12: opcode{"ORA", 2, 5, modeIndirectZeroPage, buildOpLogic(operationOr)},
	0x32: opcode{"AND", 2, 5, modeIndirectZeroPage, buildOpLogic(operationAnd)},
	0x52: opcode{"EOR", 2, 5, modeIndirectZeroPage, buildOpLogic(operationXor)},
	0x72: opcode{"ADC", 2, 5, modeIndirectZeroPage, opADCAlt},
	0x92: opcode{"STA", 2, 5, modeIndirectZeroPage, buildOpStore(regA)},
	0xb2: opcode{"LDA", 2, 5, modeIndirectZeroPage, buildOpLoad(regA)},
	0xd2: opcode{"CMP", 2, 5, modeIndirectZeroPage, buildOpCompare(regA)},
	0xf2: opcode{"SBC", 2, 5, modeIndirectZeroPage, opSBCAlt},

	// New addressing options
	0x89: opcode{"BIT", 2, 2, modeImmediate, opBIT},
	0x34: opcode{"BIT", 2, 4, modeZeroPageX, opBIT},
	0x3c: opcode{"BIT", 3, 4, modeAbsoluteX, opBIT},
	0x1a: opcode{"INC", 1, 2, modeAccumulator, buildOpIncDec(true)},
	0x3a: opcode{"DEC", 1, 2, modeAccumulator, buildOpIncDec(false)},
	0x7c: opcode{"JMP", 3, 6, modeAbsoluteIndexedIndirectX, opJMP},

	// Additional instructions: BRA, PHX, PHY, PLX, PLY, STZ, TRB, TSB
	0xda: opcode{"PHX", 1, 3, modeImplicit, buildOpPush(regX)},
	0x5a: opcode{"PHY", 1, 3, modeImplicit, buildOpPush(regY)},
	0xfa: opcode{"PLX", 1, 4, modeImplicit, buildOpPull(regX)},
	0x7a: opcode{"PLY", 1, 4, modeImplicit, buildOpPull(regY)},
	0x80: opcode{"BRA", 2, 4, modeRelative, opJMP},

	0x64: opcode{"STZ", 2, 3, modeZeroPage, opSTZ},
	0x74: opcode{"STZ", 2, 4, modeZeroPageX, opSTZ},
	0x9c: opcode{"STZ", 3, 4, modeAbsolute, opSTZ},
	0x9e: opcode{"STZ", 3, 5, modeAbsoluteX, opSTZ},

	0x14: opcode{"TRB", 2, 5, modeZeroPage, opTRB},
	0x1c: opcode{"TRB", 3, 6, modeAbsolute, opTRB},

	0x04: opcode{"TSB", 2, 5, modeZeroPage, opTSB},
	0x0c: opcode{"TSB", 3, 6, modeAbsolute, opTSB},

	// Additional in Rockwell 65c02 and WDC 65c02?
	// They have a double addressing mode: zeropage and relative.
	0x0f: opcode{"BBR0", 3, 2, modeZeroPageAndRelative, buildOpBranchOnBit(0, false)},
	0x1f: opcode{"BBR1", 3, 2, modeZeroPageAndRelative, buildOpBranchOnBit(1, false)},
	0x2f: opcode{"BBR2", 3, 2, modeZeroPageAndRelative, buildOpBranchOnBit(2, false)},
	0x3f: opcode{"BBR3", 3, 2, modeZeroPageAndRelative, buildOpBranchOnBit(3, false)},
	0x4f: opcode{"BBR4", 3, 2, modeZeroPageAndRelative, buildOpBranchOnBit(4, false)},
	0x5f: opcode{"BBR5", 3, 2, modeZeroPageAndRelative, buildOpBranchOnBit(5, false)},
	0x6f: opcode{"BBR6", 3, 2, modeZeroPageAndRelative, buildOpBranchOnBit(6, false)},
	0x7f: opcode{"BBR7", 3, 2, modeZeroPageAndRelative, buildOpBranchOnBit(7, false)},
	0x8f: opcode{"BBS0", 3, 2, modeZeroPageAndRelative, buildOpBranchOnBit(0, true)},
	0x9f: opcode{"BBS1", 3, 2, modeZeroPageAndRelative, buildOpBranchOnBit(1, true)},
	0xaf: opcode{"BBS2", 3, 2, modeZeroPageAndRelative, buildOpBranchOnBit(2, true)},
	0xbf: opcode{"BBS3", 3, 2, modeZeroPageAndRelative, buildOpBranchOnBit(3, true)},
	0xcf: opcode{"BBS4", 3, 2, modeZeroPageAndRelative, buildOpBranchOnBit(4, true)},
	0xdf: opcode{"BBS5", 3, 2, modeZeroPageAndRelative, buildOpBranchOnBit(5, true)},
	0xef: opcode{"BBS6", 3, 2, modeZeroPageAndRelative, buildOpBranchOnBit(6, true)},
	0xff: opcode{"BBS7", 3, 2, modeZeroPageAndRelative, buildOpBranchOnBit(7, true)},

	0x07: opcode{"RMB0", 2, 5, modeZeroPage, buildOpSetBit(0, false)},
	0x17: opcode{"RMB1", 2, 5, modeZeroPage, buildOpSetBit(1, false)},
	0x27: opcode{"RMB2", 2, 5, modeZeroPage, buildOpSetBit(2, false)},
	0x37: opcode{"RMB3", 2, 5, modeZeroPage, buildOpSetBit(3, false)},
	0x47: opcode{"RMB4", 2, 5, modeZeroPage, buildOpSetBit(4, false)},
	0x57: opcode{"RMB5", 2, 5, modeZeroPage, buildOpSetBit(5, false)},
	0x67: opcode{"RMB6", 2, 5, modeZeroPage, buildOpSetBit(6, false)},
	0x77: opcode{"RMB7", 2, 5, modeZeroPage, buildOpSetBit(7, false)},
	0x87: opcode{"SMB0", 2, 5, modeZeroPage, buildOpSetBit(0, true)},
	0x97: opcode{"SMB1", 2, 5, modeZeroPage, buildOpSetBit(1, true)},
	0xa7: opcode{"SMB2", 2, 5, modeZeroPage, buildOpSetBit(2, true)},
	0xb7: opcode{"SMB3", 2, 5, modeZeroPage, buildOpSetBit(3, true)},
	0xc7: opcode{"SMB4", 2, 5, modeZeroPage, buildOpSetBit(4, true)},
	0xd7: opcode{"SMB5", 2, 5, modeZeroPage, buildOpSetBit(5, true)},
	0xe7: opcode{"SMB6", 2, 5, modeZeroPage, buildOpSetBit(6, true)},
	0xf7: opcode{"SMB7", 2, 5, modeZeroPage, buildOpSetBit(7, true)},

	// Maybe additional Rockwell: STP, WAI
}
