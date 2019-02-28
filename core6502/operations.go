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
		value, setValue := resolveGetSetValue(s, line, opcode)
		if inc {
			value++
		} else {
			value--
		}
		s.reg.updateFlagZN(value)
		setValue(value)
	}
}

func buildOpShift(isLeft bool, isRotate bool) opFunc {
	return func(s *State, line []uint8, opcode opcode) {
		value, setValue := resolveGetSetValue(s, line, opcode)

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
		setValue(value)
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
		setValue := resolveSetValue(s, line, opcode)
		value := s.reg.getRegister(regSrc)
		setValue(value)
	}
}

func buildOpUpdateFlag(flag uint8, value bool) opFunc {
	return func(s *State, line []uint8, opcode opcode) {
		s.reg.updateFlag(flag, value)
	}
}

func buildOpBranch(flag uint8, value bool) opFunc {
	return func(s *State, line []uint8, opcode opcode) {
		if s.reg.getFlag(flag) == value {
			// This assumes that PC is already pointing to the next instruction
			pc := s.reg.getPC()
			pc += uint16(int8(line[1]))
			s.reg.setPC(pc)
		}
	}
}

func opBIT(s *State, line []uint8, opcode opcode) {
	value := resolveValue(s, line, opcode)
	acc := s.reg.getA()
	// Future note: The immediate addressing mode (65C02 or 65816 only) does not affect V.
	s.reg.updateFlag(flagZ, value&acc == 0)
	s.reg.updateFlag(flagN, value&(1<<7) != 0)
	s.reg.updateFlag(flagV, value&(1<<6) != 0)
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
		totalBcdLo := int(aValue&0x0f) + int(value&0x0f) + int(carry)
		totalBcdHi := int(aValue>>4) + int(value>>4)
		if totalBcdLo >= 10 {
			totalBcdHi++
		}
		totalBcd := (totalBcdHi%10)<<4 + (totalBcdLo % 10)
		s.reg.setA(uint8(totalBcd))
		s.reg.updateFlag(flagC, totalBcdHi > 9)
	} else {
		s.reg.setA(truncated)
		s.reg.updateFlag(flagC, total > 0xFF)
	}

	// ZNV flags behave for BCD as if the operation was binary?
	s.reg.updateFlagZN(truncated)
	s.reg.updateFlag(flagV, signedTotal < -128 || signedTotal > 127)
}

func opSBC(s *State, line []uint8, opcode opcode) {
	value := resolveValue(s, line, opcode)
	aValue := s.reg.getA()
	carry := s.reg.getFlagBit(flagC)

	total := 0x100 + uint16(aValue) - uint16(value) + uint16(carry) - 1
	signedTotal := int16(int8(aValue)) - int16(int8(value)) + int16(carry) - 1
	truncated := uint8(total)

	if s.reg.getFlag(flagD) {
		totalBcdLo := 10 + int(aValue&0x0f) - int(value&0x0f) + int(carry) - 1
		totalBcdHi := 10 + int(aValue>>4) - int(value>>4)
		if totalBcdLo < 10 {
			totalBcdHi--
		}
		totalBcd := (totalBcdHi%10)<<4 + (totalBcdLo % 10)
		s.reg.setA(uint8(totalBcd))
		s.reg.updateFlag(flagC, totalBcdHi >= 10)
	} else {
		s.reg.setA(truncated)
		s.reg.updateFlag(flagC, total > 0xFF)
	}

	// ZNV flags behave for SBC as if the operation was binary
	s.reg.updateFlagZN(truncated)
	s.reg.updateFlag(flagV, signedTotal < -128 || signedTotal > 127)
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

func opPLA(s *State, line []uint8, opcode opcode) {
	value := pullByte(s)
	s.reg.setA(value)
	s.reg.updateFlagZN(value)
}

func opPLP(s *State, line []uint8, opcode opcode) {
	value := pullByte(s)
	s.reg.setP(value)
}

func opPHA(s *State, line []uint8, opcode opcode) {
	pushByte(s, s.reg.getA())
}

func opPHP(s *State, line []uint8, opcode opcode) {
	pushByte(s, s.reg.getP()|(flagB+flag5))
}

func opJMP(s *State, line []uint8, opcode opcode) {
	address := resolveAddress(s, line, opcode)
	s.reg.setPC(address)
}

func opNOP(s *State, line []uint8, opcode opcode) {}

func opJSR(s *State, line []uint8, opcode opcode) {
	pushWord(s, s.reg.getPC()-1)
	address := resolveAddress(s, line, opcode)
	s.reg.setPC(address)
}

func opRTI(s *State, line []uint8, opcode opcode) {
	s.reg.setP(pullByte(s))
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
