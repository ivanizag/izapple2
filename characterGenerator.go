package izapple2

import (
	"errors"
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
	pageSize  int
}

type charColumnMap func(column int) int

func charGenColumnsMap2Plus(column int) int {
	return 6 - column
}

func charGenColumnsMap2e(column int) int {
	return column
}

const (
	charGenPageSize2Plus = 2048
	charGenPageSize2E    = 2048 * 2
)

// NewCharacterGenerator instantiates a new Character Generator with the rom on the file given
func newCharacterGenerator(filename string, order charColumnMap, isApple2e bool) (*CharacterGenerator, error) {
	var cg CharacterGenerator
	cg.columnMap = order
	cg.pageSize = charGenPageSize2Plus
	if isApple2e {
		cg.pageSize = charGenPageSize2E
	}

	err := cg.load(filename)
	if err != nil {
		return nil, err
	}

	return &cg, nil
}

func (cg *CharacterGenerator) load(filename string) error {
	bytes, _, err := LoadResource(filename)
	if err != nil {
		return err
	}
	size := len(bytes)
	if size < cg.pageSize {
		return errors.New("character ROM size not supported")
	}
	cg.data = bytes
	return nil
}

func (cg *CharacterGenerator) setPage(page int) {
	// Some clones had a switch to change codepage with extra characters
	pages := len(cg.data) / cg.pageSize
	cg.page = page % pages
}

func (cg *CharacterGenerator) getPage() int {
	return cg.page
}

func (cg *CharacterGenerator) nextPage() {
	cg.setPage(cg.page + 1)
}

func (cg *CharacterGenerator) getPixel(char uint8, row int, column int) bool {
	bits := cg.data[int(char)*8+row+cg.page*cg.pageSize]
	bit := cg.columnMap(column)
	value := bits >> uint(bit) & 1
	return value == 1
}
