package apple2

import (
	"bufio"
	"fmt"
	"os"
)

/*
 See:
 hhttps://mirrors.apple2.org.za/Apple%20II%20Documentation%20Project/Companies/Apple/Documentation/Apple%20Technical%20Information%20Library/a2til041.txt
*/

// CharacterGenerator represents the ROM wth the characters bitmaps
type CharacterGenerator struct {
	data []uint8
}

const (
	rev7CharGenSize = 2048
)

// NewCharacterGenerator instantiates a new Character Generator with the rom on the file given
func NewCharacterGenerator(filename string) *CharacterGenerator {
	var cg CharacterGenerator
	cg.load(filename)
	return &cg
}

func (cg *CharacterGenerator) load(filename string) {
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
	if size != rev7CharGenSize {
		panic("Character ROM size not supported")
	}
	cg.data = make([]uint8, size)
	buf := bufio.NewReader(f)
	buf.Read(cg.data)
}

func (cg *CharacterGenerator) getPixel(char uint8, row int, column int) bool {
	bits := cg.data[int(char)*8+row]
	bit := bits >> (uint(6 - column)) & 1
	return bit == 1
}

func (cg *CharacterGenerator) dumpCharFast(char uint8) {
	base := int(char) * 8
	fmt.Printf("Char: %v\n---------\n", char)
	for i := 0; i < 8; i++ {
		fmt.Print("|")
		b := cg.data[base+i]
		for j := 6; j >= 0; j-- {
			if (b>>uint(j))&1 == 1 {
				fmt.Print("#")
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Println("|")
	}
	fmt.Println("---------")
}

func (cg *CharacterGenerator) dumpChar(char uint8) {
	fmt.Printf("Char: %v\n---------\n", char)
	for row := 0; row < 8; row++ {
		fmt.Print("|")
		for col := 0; col < 7; col++ {
			if cg.getPixel(char, row, col) {
				fmt.Print("#")
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Println("|")
	}
	fmt.Println("---------")
}

// Dump to sdtout all the character maps
func (cg *CharacterGenerator) Dump() {
	for i := 0; i < 256; i++ {
		cg.dumpChar(uint8(i))
	}
}
