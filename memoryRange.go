package izapple2

import (
	"fmt"
)

type memoryRange struct {
	base uint16
	data []uint8
	name string
}

type memoryRangeROM struct {
	memoryRange
}

func newMemoryRange(base uint16, data []uint8, name string) *memoryRange {
	var m memoryRange
	m.base = base
	m.data = data
	m.setBase(base)

	m.name = name
	return &m
}

func newMemoryRangeROM(base uint16, data []uint8, name string) *memoryRangeROM {
	var m memoryRangeROM
	m.base = base
	m.data = data
	m.name = name
	return &m
}

func (m *memoryRange) setBase(base uint16) {
	m.base = base
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

func (m *memoryRangeROM) poke(address uint16, value uint8) {
	// Ignore
}

//lint:ignore U1000 this is used to write debug code
func identifyMemory(m memoryHandler) string {
	ram, ok := m.(*memoryRange)
	if ok {
		return fmt.Sprintf("RAM 0x%04x %s", ram.base, ram.name)
	}

	rom, ok := m.(*memoryRangeROM)
	if ok {
		return fmt.Sprintf("ROM 0x%04x %s", rom.base, ram.name)
	}

	return ("Unknown memory")
}
