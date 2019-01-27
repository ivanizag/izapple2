package main

type memory [65536]uint8

func (m *memory) getWord(address uint16) uint16 {
	return uint16(m[address]) + 0x100*uint16(m[address+1])
}

func (m *memory) getZeroPageWord(address uint8) uint16 {
	return uint16(m[address]) + 0x100*uint16(m[address+1])
	// TODO: Does address + 1 wraps around the zero page?
}
