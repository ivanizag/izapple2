package izapple2

/*
	The Basis 108 clone has 128kb of RAM plus 2KB of static RAM at $0400 for 80 columns text.
*/

type memoryRangeBasis108 struct {
	dataMain   []uint8
	dataAux    []uint8
	dataStatic []uint8
	name       string

	staticRam bool
	auxRam    bool
}

func newMemoryRangeBasis108() *memoryRangeBasis108 {
	var m memoryRangeBasis108
	m.dataMain = make([]uint8, 48*1024)
	m.dataAux = make([]uint8, 48*1024)
	m.dataStatic = make([]uint8, 0xc000-0x0400)

	// How is the static RAM initialized?
	for i := 0; i < len(m.dataStatic); i++ {
		m.dataStatic[i] = ' ' + 0x80
	}

	m.name = "Basis 108 RAM"
	return &m
}

func (m *memoryRangeBasis108) peek(address uint16) uint8 {
	if m.staticRam && address >= 0x0400 && address < 0x0c00 {
		return m.dataStatic[address-0x0400]
	}

	if m.auxRam {
		return m.dataAux[address]
	}
	return m.dataMain[address]
}

func (m *memoryRangeBasis108) poke(address uint16, value uint8) {
	if m.staticRam && address >= 0x0400 && address < 0x0c00 {
		m.dataStatic[address-0x0400] = value
	} else if m.auxRam {
		m.dataAux[address] = value
	} else {
		m.dataMain[address] = value
	}
}

func (m *memoryRangeBasis108) subRange(a, b uint16) []uint8 {
	if m.staticRam && a >= 0x0400 && b < 0x0c00 {
		return m.dataStatic[a-0x0400 : b-0x0400]
	}
	if m.auxRam {
		return m.dataAux[a:b]
	}
	return m.dataMain[a:b]
}

func (m *memoryRangeBasis108) getTextMemory(secondPage bool, ext bool) []uint8 {
	addressStart := textPage1Address
	if secondPage {
		addressStart = textPage2Address
	}

	if ext {
		return m.dataStatic[addressStart-0x0400 : addressStart-0x0400+textPageSize]
	}
	return m.dataMain[addressStart : addressStart+textPageSize]
}
