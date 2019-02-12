package main

import (
	"bufio"
	"fmt"
	"os"
)

type memoryPage interface {
	peek(uint8) uint8
	poke(uint8, uint8)
}

type ramPage [256]uint8
type romPage [256]uint8

type memory [256]memoryPage

func (p *ramPage) peek(address uint8) uint8 {
	return p[address]
}

func (p *ramPage) poke(address uint8, value uint8) {
	p[address] = value
}

func (p *romPage) peek(address uint8) uint8 {
	return p[address]
}

func (p *romPage) poke(address uint8, value uint8) {
	// Do nothing
}

func (m *memory) peek(address uint16) uint8 {
	hi := uint8(address >> 8)
	lo := uint8(address)
	return m[hi].peek(lo)
}

func (m *memory) poke(address uint16, value uint8) {
	hi := uint8(address >> 8)
	lo := uint8(address)
	//fmt.Println(hi)
	m[hi].poke(lo, value)
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
		m[i] = &ramPages[i]
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
			fmt.Printf("%02x ", m[address])
			address++
		}
		fmt.Printf("\n")
	}
}
