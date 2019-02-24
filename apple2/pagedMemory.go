package apple2

// memoryPage is a data page of 256 bytes
type memoryPage interface {
	Peek(uint8) uint8
	Poke(uint8, uint8)
}

// pagedMemory represents the addressable space of the processor
type pagedMemory struct {
	data [256]memoryPage
}

// Peek returns the data on the given address
func (m *pagedMemory) Peek(address uint16) uint8 {
	hi := uint8(address >> 8)
	lo := uint8(address)
	return m.data[hi].Peek(lo)
}

// Poke sets the data at the given address
func (m *pagedMemory) Poke(address uint16, value uint8) {
	hi := uint8(address >> 8)
	lo := uint8(address)
	m.data[hi].Poke(lo, value)
}

// SetPage assigns a MemoryPage implementation on the page given
func (m *pagedMemory) SetPage(index uint8, page memoryPage) {
	//fmt.Printf("Assigning page 0x%02x type %s\n", index, reflect.TypeOf(page))
	m.data[index] = page
}
