package main

import (
	"bufio"
	"fmt"
	"os"
)

type memory [65536]uint8

func (m *memory) getWord(address uint16) uint16 {
	return uint16(m[address]) + 0x100*uint16(m[address+1])
}

func (m *memory) getZeroPageWord(address uint8) uint16 {
	return uint16(m[address]) + 0x100*uint16(m[address+1])
	// TODO: Does address + 1 wraps around the zero page?
}

func (m *memory) loadBinary(filename string) {
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

	for i, v := range bytes {
		m[i] = uint8(v)
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
