package main

import (
	"bufio"
	"fmt"
	"os"
)

type memoryPage interface {
	peek(uint8) uint8
	poke(uint8, uint8)
	getData() *[256]uint8
}

type ramPage struct {
	data [256]uint8
}

type romPage struct {
	data [256]uint8
}

type memory struct {
	data [256]memoryPage
}

func (p *ramPage) peek(address uint8) uint8 {
	return p.data[address]
}

func (p *ramPage) poke(address uint8, value uint8) {
	p.data[address] = value
}

func (p *ramPage) getData() *[256]uint8 {
	return &p.data
}

func (p *romPage) peek(address uint8) uint8 {
	return p.data[address]
}

func (p *romPage) poke(address uint8, value uint8) {
	// Do nothing
}

func (p *romPage) getData() *[256]uint8 {
	return &p.data
}

func (m *memory) peek(address uint16) uint8 {
	hi := uint8(address >> 8)
	lo := uint8(address)
	return m.data[hi].peek(lo)
}

func (m *memory) poke(address uint16, value uint8) {
	hi := uint8(address >> 8)
	lo := uint8(address)
	//fmt.Println(hi)
	m.data[hi].poke(lo, value)
}

func (m *memory) getWord(address uint16) uint16 {
	return uint16(m.peek(address)) + 0x100*uint16(m.peek(address+1))
}

func (m *memory) getZeroPageWord(address uint8) uint16 {
	return uint16(m.peek(uint16(address))) + 0x100*uint16(m.peek(uint16(address+1)))
}

func (m *memory) initWithRam() {
	var ramPages [256]ramPage
	for i := 0; i < 256; i++ {
		m.data[i] = &ramPages[i]
	}
}

func (m *memory) transformToRom(page uint8) {
	var romPage romPage
	ramPage := m.data[page]
	romPage.data = *ramPage.getData()
	m.data[page] = &romPage
}

func (m *memory) initWithRomAndText(filename string, textPages *textPages) {
	// Valid for ROMs with size 20480 bytes = 20 KB = 80 pages
	// from $B000 to $F000
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
	if size != 20480 {
		panic("Invalid ROM file size. It must be 20480 bytes")
	}
	bytes := make([]byte, size)

	buf := bufio.NewReader(f)
	buf.Read(bytes)

	m.initWithRam()
	for i, v := range bytes {
		m.poke(uint16(i)+0xB000, uint8(v))
	}

	var i uint8
	for i = 217; i != 0; i++ {
		m.transformToRom(i)
	}

	for j := 0; j < 4; j++ {
		m.data[4+i] = &textPages.pages[i]
	}
}

func (m *memory) loadBinary(filename string) {
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

	m.initWithRam()
	for i, v := range bytes {
		m.poke(uint16(i), uint8(v))
	}
}

func (m *memory) printPage(page uint8) {
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
