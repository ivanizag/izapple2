package izapple2

import (
	"fmt"
)

type memoryRange struct {
	base    uint16
	data    []uint8
	name    string
	address string
	//basePtr uintptr
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
	m.address = fmt.Sprintf("%p", &m)
	return &m
}

func newMemoryRangeROM(base uint16, data []uint8, name string) *memoryRangeROM {
	var m memoryRangeROM
	m.base = base
	m.data = data
	m.setBase(base)

	m.name = name
	m.address = fmt.Sprintf("%p", &m)
	return &m
}

func (m *memoryRange) setBase(base uint16) {
	m.base = base
	//p := unsafe.Pointer(&m.data[0])
	//m.basePtr = (uintptr)(p) - (uintptr)(base)
}

func (m *memoryRange) peek(address uint16) uint8 {
	// Safe version:
	return m.data[address-m.base]

	// Really overkill
	// go-vet warns the caching of basePtr
	// This wouldn't have a warning
	//	 indexp := unsafe.Pointer((uintptr)(unsafe.Pointer(&m.data[0])) - (uintptr)(m.base) + uintptr(address))
	// But it makes sense to precalculate that
	//indexp := unsafe.Pointer(m.basePtr + uintptr(address))
	//return *(*uint8)(indexp)
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

func identifyMemory(m memoryHandler) string {
	ram, ok := m.(*memoryRange)
	if ok {
		return fmt.Sprintf("RAM 0x%04x %s at %s", ram.base, ram.name, ram.address)
	}

	rom, ok := m.(*memoryRangeROM)
	if ok {
		return fmt.Sprintf("ROM 0x%04x %s at %s", rom.base, ram.name, rom.address)
	}

	return ("Unknown memory")
}
