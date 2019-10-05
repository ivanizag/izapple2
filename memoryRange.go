package apple2

import (
	"encoding/binary"
	"io"
)

type memoryRange struct {
	base uint16
	data []uint8
}

func newMemoryRange(base uint16, data []uint8) *memoryRange {
	var m memoryRange
	m.base = base
	m.data = data
	return &m
}

func (m *memoryRange) peek(address uint16) uint8 {
	return m.data[address-m.base]
}

func (m *memoryRange) poke(address uint16, value uint8) {
	m.data[address-m.base] = value
}

func (m *memoryRange) subRange(a, b uint16) []uint8 {
	return m.data[a-m.base : b-m.base]
}

func (m *memoryRange) save(w io.Writer) error {
	err := binary.Write(w, binary.BigEndian, m.base)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, m.data)
	if err != nil {
		return err
	}
	return nil
}

func (m *memoryRange) load(r io.Reader) error {
	err := binary.Read(r, binary.BigEndian, &m.base)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.BigEndian, &m.data)
	if err != nil {
		return err
	}
	return nil
}
