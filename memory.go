package main

import (
	"bufio"
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
