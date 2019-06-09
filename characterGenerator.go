package apple2

import (
	"fmt"
)

/*
 See:
 hhttps://mirrors.apple2.org.za/Apple%20II%20Documentation%20Project/Companies/Apple/Documentation/Apple%20Technical%20Information%20Library/a2til041.txt
*/

// CharacterGenerator represents the ROM wth the characters bitmaps
type CharacterGenerator struct {
	data      []uint8
	customRom bool
	columnMap charColumnMap
	page      int
}

type charColumnMap func(column int) int

const (
	rev7CharGenSize   = 2048
	defaultCharGenROM = "<internal>/Apple2rev7CharGen.rom"
)

// NewCharacterGenerator instantiates a new Character Generator with the rom on the file given
func NewCharacterGenerator(filename string) *CharacterGenerator {
	var cg CharacterGenerator
	cg.load(filename)
	return &cg
}

func (cg *CharacterGenerator) load(filename string) {
	cg.customRom = !isInternalResource(filename)
	bytes := loadResource(filename)
	size := len(bytes)
	if size < rev7CharGenSize {
		panic("Character ROM size not supported")
	}
	cg.data = bytes
}

func (cg *CharacterGenerator) setColumnMap(columnMap charColumnMap) {
	// Regular Apple II uses bits 6 to 0 but some clones have other mappings
	cg.columnMap = columnMap
}

func (cg *CharacterGenerator) setPage(page int) {
	// Some clones had a switch to change codepage with extra characters
	pages := len(cg.data) / rev7CharGenSize
	cg.page = page % pages
}

func (cg *CharacterGenerator) nextPage() {
	cg.setPage(cg.page + 1)
}

func (cg *CharacterGenerator) getPixel(char uint8, row int, column int) bool {
	bits := cg.data[int(char)*8+row+cg.page*rev7CharGenSize]
	var bit int
	if cg.columnMap != nil {
		bit = cg.columnMap(column)
	} else {
		// Standard Apple 2 mapping
		bit = 6 - column
	}
	value := bits >> uint(bit) & 1
	return value == 1
}

func (cg *CharacterGenerator) dumpCharRaw(char int) {
	base := int(char) * 8
	fmt.Printf("Char: %v\n---------\n", char)
	for i := 0; i < 8; i++ {
		fmt.Print("|")
		b := cg.data[base+i]
		for j := 0; j < 8; j++ {
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
	pages := len(cg.data) / rev7CharGenSize
	for p := 0; p < pages; p++ {
		cg.setPage(p)
		for i := 0; i < 256; i++ {
			cg.dumpChar(uint8(i))
			//cg.dumpCharRaw(int(i))
		}
	}
}
