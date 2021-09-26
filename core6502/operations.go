package core6502

func buildOpTransfer(regSrc int, regDst int) opFunc {
	return func(s *State, line []uint8, opcode opcode) {
		value := s.reg.getRegister(regSrc)
		s.reg.setRegister(regDst, value)
		if regDst != regSP {
			s.reg.updateFlagZN(value)
		}
	}
}

func buildOpIncDec(inc bool) opFunc {
	return func(s *State, line []uint8, opcode opcode) {
		value := resolveValue(s, line, opcode)
		if opcode.addressMode == modeAbsoluteX || opcode.addressMode == modeAbsoluteY {
			// Double read, needed to pass A2Audit for the Language Card
			value = resolveValue(s, line, opcode)
		}
		if inc {
			value++
		} else {
			value--
		}
		s.reg.updateFlagZN(value)
		resolveSetValue(s, line, opcode, value)
	}
}

func buildOpShift(isLeft bool, isRotate bool) opFunc {
	return func(s *State, line []uint8, opcode opcode) {
		value := resolveValue(s, line, opcode)

		oldCarry := s.reg.getFlagBit(flagC)
		var carry bool
		if isLeft {
			carry = (value & 0x80) != 0
			value <<= 1
			if isRotate {
				value += oldCarry
			}
		} else {
			carry = (value & 0x01) != 0
			value >>= 1
			if isRotate {
				value += oldCarry << 7
			}
		}
		s.reg.updateFlag(flagC, carry)
		s.reg.updateFlagZN(value)
		resolveSetValue(s, line, opcode, value)
	}
}

func buildOpLoad(regDst int) opFunc {
	return func(s *State, line []uint8, opcode opcode) {
		value := resolveValue(s, line, opcode)
		s.reg.setRegister(regDst, value)
		s.reg.updateFlagZN(value)
	}
}

func buildOpStore(regSrc int) opFunc {
	return func(s *State, line []uint8, opcode opcode) {
		value := s.reg.getRegister(regSrc)
		resolveSetValue(s, line, opcode, value)
	}
}

func buildOpUpdateFlag(flag uint8, value bool) opFunc {
	return func(s *State, line []uint8, opcode opcode) {
		s.reg.updateFlag(flag, value)
	}
}

func buildOpBranch(flag uint8, test bool) opFunc {
	return func(s *State, line []uint8, opcode opcode) {
		if s.reg.getFlag(flag) == test {
			s.extraCycleBranchTaken = true
			address := resolveAddress(s, line, opcode)
			s.reg.setPC(address)
		}
	}
}

func buildOpBranchOnBit(bit uint8, test bool) opFunc {
	return func(s *State, line []uint8, opcode opcode) {
		// Note that those operations have two addressing modes:
		// one for the zero page value, another for the relative jump.
		// We will have to resolve the first one here.
		value := s.mem.Peek(uint16(line[1]))
		bitValue := ((value >> bit) & 1) == 1

		if bitValue == test {
			address := resolveAddress(s, line, opcode)
			s.reg.setPC(address)
		}
	}
}

func buildOpSetBit(bit uint8, set bool) opFunc {
	return func(s *State, line []uint8, opcode opcode) {
		value := resolveValue(s, line, opcode)
		if set {
			value = value | (1 << bit)
		} else {
			value = value &^ (1 << bit)
		}
		resolveSetValue(s, line, opcode, value)
	}
}

func opBIT(s *State, line []uint8, opcode opcode) {
	value := resolveValue(s, line, opcode)
	acc := s.reg.getA()
	s.reg.updateFlag(flagZ, value&acc == 0)
	// The immediate addressing mode (65C02 or 65816 only) does not affect N & V.
	if opcode.addressMode != modeImmediate {
		s.reg.updateFlag(flagN, value&(1<<7) != 0)
		s.reg.updateFlag(flagV, value&(1<<6) != 0)
	}
}

func opTRB(s *State, line []uint8, opcode opcode) {
	value := resolveValue(s, line, opcode)
	a := s.reg.getA()
	s.reg.updateFlag(flagZ, (value&a) == 0)
	resolveSetValue(s, line, opcode, value&^a)
}

func opTSB(s *State, line []uint8, opcode opcode) {
	value := resolveValue(s, line, opcode)
	a := s.reg.getA()
	s.reg.updateFlag(flagZ, (value&a) == 0)
	resolveSetValue(s, line, opcode, value|a)
}

func buildOpCompare(reg int) opFunc {
	return func(s *State, line []uint8, opcode opcode) {
		value := resolveValue(s, line, opcode)
		reference := s.reg.getRegister(reg)
		s.reg.updateFlagZN(reference - value)
		s.reg.updateFlag(flagC, reference >= value)
	}
}

func operationAnd(a uint8, b uint8) uint8 { return a & b }
func operationOr(a uint8, b uint8) uint8  { return a | b }
func operationXor(a uint8, b uint8) uint8 { return a ^ b }

func buildOpLogic(operation func(uint8, uint8) uint8) opFunc {
	return func(s *State, line []uint8, opcode opcode) {
		value := resolveValue(s, line, opcode)
		result := operation(value, s.reg.getA())
		s.reg.setA(result)
		s.reg.updateFlagZN(result)
	}
}

func opADC(s *State, line []uint8, opcode opcode) {
	value := resolveValue(s, line, opcode)
	aValue := s.reg.getA()
	carry := s.reg.getFlagBit(flagC)

	total := uint16(aValue) + uint16(value) + uint16(carry)
	signedTotal := int16(int8(aValue)) + int16(int8(value)) + int16(carry)
	truncated := uint8(total)

	if s.reg.getFlag(flagD) {
		totalBcdLo := uint(aValue&0x0f) + uint(value&0x0f) + uint(carry)
		totalBcdHi := uint(aValue>>4) + uint(value>>4)
		if totalBcdLo >= 10 {
			totalBcdLo -= 10
			totalBcdHi++
		}
		totalBcdHiPrenormalised := uint8(totalBcdHi & 0xf)
		newCarry := false
		if totalBcdHi >= 10 {
			totalBcdHi -= 10
			newCarry = true
		}
		totalBcd := uint8(totalBcdHi)<<4 + (uint8(totalBcdLo) & 0xf)
		s.reg.setA(uint8(totalBcd))
		s.reg.updateFlag(flagC, newCarry)
		s.reg.updateFlag(flagV, (value>>7 == aValue>>7) &&
			(value>>7 != totalBcdHiPrenormalised>>3))
	} else {
		s.reg.setA(truncated)
		s.reg.updateFlag(flagC, total > 0xFF)
		s.reg.updateFlag(flagV, signedTotal < -128 || signedTotal > 127)
		// Effectively the same as the less clear:
		// s.reg.updateFlag(flagV, (value>>7 == aValue>>7) && (value>>7 != truncated>>7))
		// See http://www.6502.org/tutorials/vflag.html
	}

	// ZN flags behave for BCD as if the operation was binary?
	s.reg.updateFlagZN(truncated)
}

func opADCAlt(s *State, line []uint8, opcode opcode) {
	opADC(s, line, opcode)
	if s.reg.getFlag(flagD) {
		s.extraCycleBCD = true
	}

	// The Z and N flags on BCD are fixed in 65c02.
	s.reg.updateFlagZN(s.reg.getA())
}

func opSBC(s *State, line []uint8, opcode opcode) {
	value := resolveValue(s, line, opcode)
	aValue := s.reg.getA()
	carry := s.reg.getFlagBit(flagC)

	total := 0x100 + uint16(aValue) - uint16(value) + uint16(carry) - 1
	signedTotal := int16(int8(aValue)) - int16(int8(value)) + int16(carry) - 1
	truncated := uint8(total)

	if s.reg.getFlag(flagD) {
		totalBcdLo := int(aValue&0x0f) - int(value&0x0f) + int(carry) - 1
		totalBcdHi := int(aValue>>4) - int(value>>4)
		if totalBcdLo < 0 {
			totalBcdLo += 10
			totalBcdHi--
		}
		newCarry := true
		if totalBcdHi < 0 {
			totalBcdHi += 10
			newCarry = false
		}
		totalBcd := uint8(totalBcdHi)<<4 + (uint8(totalBcdLo) & 0xf)
		s.reg.setA(uint8(totalBcd))
		s.reg.updateFlag(flagC, newCarry)
	} else {
		s.reg.setA(truncated)
		s.reg.updateFlag(flagC, total > 0xFF)
	}

	// ZNV flags behave for SBC as if the operation was binary
	s.reg.updateFlagZN(truncated)
	s.reg.updateFlag(flagV, signedTotal < -128 || signedTotal > 127)
}

func opSBCAlt(s *State, line []uint8, opcode opcode) {
	opSBC(s, line, opcode)
	if s.reg.getFlag(flagD) {
		s.extraCycleBCD = true
	}
	// The Z and N flags on BCD are fixed in 65c02.
	s.reg.updateFlagZN(s.reg.getA())
}

const stackAddress uint16 = 0x0100

func pushByte(s *State, value uint8) {
	adresss := stackAddress + uint16(s.reg.getSP())
	s.mem.Poke(adresss, value)
	s.reg.setSP(s.reg.getSP() - 1)
}

func pullByte(s *State) uint8 {
	s.reg.setSP(s.reg.getSP() + 1)
	adresss := stackAddress + uint16(s.reg.getSP())
	return s.mem.Peek(adresss)
}

func pushWord(s *State, value uint16) {
	pushByte(s, uint8(value>>8))
	pushByte(s, uint8(value))
}

func pullWord(s *State) uint16 {
	return uint16(pullByte(s)) +
		(uint16(pullByte(s)) << 8)

}

func buildOpPull(regDst int) opFunc {
	return func(s *State, line []uint8, opcode opcode) {
		value := pullByte(s)
		s.reg.setRegister(regDst, value)
		if regDst == regP {
			s.reg.updateFlag5B()
		} else {
			s.reg.updateFlagZN(value)
		}
	}
}

func buildOpPush(regSrc int) opFunc {
	return func(s *State, line []uint8, opcode opcode) {
		value := s.reg.getRegister(regSrc)
		if regSrc == regP {
			value |= flagB + flag5
		}
		pushByte(s, value)
	}
}

func opJMP(s *State, line []uint8, opcode opcode) {
	address := resolveAddress(s, line, opcode)
	s.reg.setPC(address)
}

func opNOP(s *State, line []uint8, opcode opcode) {}

func opHALT(s *State, line []uint8, opcode opcode) {
	s.reg.setPC(s.reg.getPC() - 1)
}

func opJSR(s *State, line []uint8, opcode opcode) {
	pushWord(s, s.reg.getPC()-1)
	address := resolveAddress(s, line, opcode)
	s.reg.setPC(address)
}

func opRTI(s *State, line []uint8, opcode opcode) {
	s.reg.setP(pullByte(s))
	s.reg.updateFlag5B()
	s.reg.setPC(pullWord(s))
}

func opRTS(s *State, line []uint8, opcode opcode) {
	s.reg.setPC(pullWord(s) + 1)
}

func opBRK(s *State, line []uint8, opcode opcode) {
	pushWord(s, s.reg.getPC()+1)
	pushByte(s, s.reg.getP()|(flagB+flag5))
	s.reg.setFlag(flagI)
	s.reg.setPC(getWord(s.mem, vectorBreak))
}

func opBRKAlt(s *State, line []uint8, opcode opcode) {
	opBRK(s, line, opcode)
	/*
		The only difference in the BRK instruction on the 65C02 and the 6502
		is that the 65C02 clears the D (decimal) flag on the 65C02, whereas
		the D flag is not affected on the 6502.
	*/
	s.reg.clearFlag(flagD)
}

func opSTZ(s *State, line []uint8, opcode opcode) {
	resolveSetValue(s, line, opcode, 0)
}
