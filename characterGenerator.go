package izapple2

import (
	"errors"
	"fmt"

	"github.com/ivanizag/izapple2/storage"
)

/*
 See:
 https://mirrors.apple2.org.za/Apple%20II%20Documentation%20Project/Companies/Apple/Documentation/Apple%20Technical%20Information%20Library/a2til041.txt
*/

// CharacterGenerator represents the ROM wth the characters bitmaps
type CharacterGenerator struct {
	data      []uint8
	columnMap charColumnMap
	page      int
}

type charColumnMap func(column int) int

func charGenColumnsMap2Plus(column int) int {
	return 6 - column
}

func charGenColumnsMap2e(column int) int {
	return column
}

const (
	charGenPageSize = 2048
)

// NewCharacterGenerator instantiates a new Character Generator with the rom on the file given
func newCharacterGenerator(filename string, order charColumnMap) (*CharacterGenerator, error) {
	var cg CharacterGenerator
	err := cg.load(filename)
	if err != nil {
		return nil, err
	}
	cg.columnMap = order
	return &cg, nil
}

func (cg *CharacterGenerator) load(filename string) error {
	bytes, _, err := storage.LoadResource(filename)
	if err != nil {
		return err
	}
	size := len(bytes)
	if size < charGenPageSize {
		return errors.New("Character ROM size not supported")
	}
	cg.data = bytes
	return nil
}

func (cg *CharacterGenerator) setPage(page int) {
	// Some clones had a switch to change codepage with extra characters
	pages := len(cg.data) / charGenPageSize
	cg.page = page % pages
}

func (cg *CharacterGenerator) nextPage() {
	cg.setPage(cg.page + 1)
}

func (cg *CharacterGenerator) getPixel(char uint8, row int, column int) bool {
	bits := cg.data[int(char)*8+row+cg.page*charGenPageSize]
	bit := cg.columnMap(column)
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
	pages := len(cg.data) / charGenPageSize
	for p := 0; p < pages; p++ {
		cg.setPage(p)
		for i := 0; i < 256; i++ {
			cg.dumpChar(uint8(i))
			//cg.dumpCharRaw(int(i))
		}
	}
}
