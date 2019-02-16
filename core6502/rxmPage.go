package core6502

import (
	"bufio"
	"os"
)

type ramPage struct {
	data [256]uint8
}

type romPage struct {
	data [256]uint8
}

func (p *ramPage) Peek(address uint8) uint8 {
	return p.data[address]
}

func (p *ramPage) Poke(address uint8, value uint8) {
	p.data[address] = value
}

func (p *romPage) Peek(address uint8) uint8 {
	return p.data[address]
}

func (p *romPage) Poke(address uint8, value uint8) {
	// Do nothing
}

func (p *romPage) burn(address uint8, value uint8) {
	p.data[address] = value
}

// InitWithRAM adds RAM memory to all the memory pages
func (m *Memory) InitWithRAM() {
	var ramPages [256]ramPage
	for i := 0; i < 256; i++ {
		m.SetPage(uint8(i), &ramPages[i])
	}
}

func (m *Memory) transformToRom(page uint8) {
	var romPage romPage
	address := uint16(page) << 8
	for i := 0; i < 256; i++ {
		romPage.burn(uint8(i), m.Peek(address))
		address++
	}
	m.SetPage(page, &romPage)
}

// LoadRom loads a binary file to the top of the memory and makes those pages read only.
func (m *Memory) LoadRom(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	stats, statsErr := f.Stat()
	if statsErr != nil {
		panic(err)
	}

	size := stats.Size()
	bytes := make([]byte, size)

	buf := bufio.NewReader(f)
	buf.Read(bytes)

	romStart := uint16(0xFFFF - size + 1)
	for i, v := range bytes {
		m.Poke(uint16(i)+romStart, uint8(v))
	}

	for i := uint8(romStart >> 8); i != 0; i++ {
		m.transformToRom(i)
	}
}
