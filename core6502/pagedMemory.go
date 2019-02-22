package core6502

// MemoryPage is a data page of 256 bytes
type MemoryPage interface {
	Peek(uint8) uint8
	Poke(uint8, uint8)
}

// PagedMemory represents the addressable space of the processor
type PagedMemory struct {
	data [256]MemoryPage
}

// Peek returns the data on the given address
func (m *PagedMemory) Peek(address uint16) uint8 {
	hi := uint8(address >> 8)
	lo := uint8(address)
	return m.data[hi].Peek(lo)
}

// Poke sets the data at the given address
func (m *PagedMemory) Poke(address uint16, value uint8) {
	hi := uint8(address >> 8)
	lo := uint8(address)
	//fmt.Println(hi)
	m.data[hi].Poke(lo, value)
}

// SetPage assigns a MemoryPage implementation on the page given
func (m *PagedMemory) SetPage(index uint8, page MemoryPage) {
	m.data[index] = page
}
