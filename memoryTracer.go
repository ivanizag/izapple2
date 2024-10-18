package izapple2

import (
	"fmt"
)

type memoryTracer struct {
	memory memoryHandler
	name   string
}

func traceMemory(memory memoryHandler, name string, trace bool) memoryHandler {
	if !trace {
		return memory
	}

	if memory == nil {
		return nil
	}

	return &memoryTracer{
		memory: memory,
		name:   name,
	}
}

func (m *memoryTracer) peek(address uint16) uint8 {
	value := m.memory.peek(address)
	fmt.Printf("Memory %s: peek($%04X) = $%02X\n", m.name, address, value)
	return value
}

func (m *memoryTracer) poke(address uint16, value uint8) {
	fmt.Printf("Memory %s: poke($%04X, $%02X)\n", m.name, address, value)
	m.memory.poke(address, value)
}
