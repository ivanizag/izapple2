package core6502

import "io/ioutil"

// Memory represents the addressable space of the processor
type Memory interface {
	Peek(address uint16) uint8
	Poke(address uint16, value uint8)
}

func getWord(m Memory, address uint16) uint16 {
	return uint16(m.Peek(address)) + 0x100*uint16(m.Peek(address+1))
}

func getZeroPageWord(m Memory, address uint8) uint16 {
	return uint16(m.Peek(uint16(address))) + 0x100*uint16(m.Peek(uint16(address+1)))
}

// FlatMemory puts RAM on the 64Kb addressable by the processor
type FlatMemory struct {
	data [65536]uint8
}

// Peek returns the data on the given address
func (m *FlatMemory) Peek(address uint16) uint8 {
	return m.data[address]
}

// Poke sets the data at the given address
func (m *FlatMemory) Poke(address uint16, value uint8) {
	m.data[address] = value
}

func (m *FlatMemory) loadBinary(filename string) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	for i, v := range bytes {
		m.Poke(uint16(i), uint8(v))
	}
}
