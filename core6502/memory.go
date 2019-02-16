package core6502

import (
	"bufio"
	"fmt"
	"os"
)

// MemoryPage is a data page of 256 bytes
type MemoryPage interface {
	Peek(uint8) uint8
	Poke(uint8, uint8)
}

// Memory represents the addressable space of the processor
type Memory struct {
	data [256]MemoryPage
}

// Peek returns the data on the given address
func (m *Memory) Peek(address uint16) uint8 {
	hi := uint8(address >> 8)
	lo := uint8(address)
	return m.data[hi].Peek(lo)
}

// Poke sets the data at the given address
func (m *Memory) Poke(address uint16, value uint8) {
	hi := uint8(address >> 8)
	lo := uint8(address)
	//fmt.Println(hi)
	m.data[hi].Poke(lo, value)
}

// SetPage assigns a MemoryPage implementation on the page given
func (m *Memory) SetPage(index uint8, page MemoryPage) {
	m.data[index] = page
}

func (m *Memory) getWord(address uint16) uint16 {
	return uint16(m.Peek(address)) + 0x100*uint16(m.Peek(address+1))
}

func (m *Memory) getZeroPageWord(address uint8) uint16 {
	return uint16(m.Peek(uint16(address))) + 0x100*uint16(m.Peek(uint16(address+1)))
}

func (m *Memory) loadBinary(filename string) {
	// Load file
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

	m.InitWithRAM()
	for i, v := range bytes {
		m.Poke(uint16(i), uint8(v))
	}
}

func (m *Memory) printPage(page uint8) {
	address := uint16(page) * 0x100
	for i := 0; i < 16; i++ {
		fmt.Printf("%#04x: ", address)
		for j := 0; j < 16; j++ {
			fmt.Printf("%02x ", m.data[address])
			address++
		}
		fmt.Printf("\n")
	}
}
