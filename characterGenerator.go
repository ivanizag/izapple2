package izapple2

import (
	"errors"
	"fmt"
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
	charGenPageSize2Plus    = 2048
	charGenPageSize2E       = 2048 * 2
	charGenPageSizeBasis108 = 1024
)

// NewCharacterGenerator instantiates a new Character Generator with the rom on the file given
func newCharacterGenerator(filename string, order charColumnMap, pageSize int) (*CharacterGenerator, error) {
	var cg CharacterGenerator
	cg.columnMap = order
	cg.pageSize = pageSize
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

func (cg *CharacterGenerator) getPages() int {
	return len(cg.data) / cg.pageSize
}

func (cg *CharacterGenerator) setPage(page int) {
	// Some clones had a switch to change codepage with extra characters
	pages := cg.getPages()
	cg.page = page % pages
}

func (cg *CharacterGenerator) getPage() int {
	return cg.page
}

func (cg *CharacterGenerator) nextPage() {
	cg.setPage(cg.page + 1)
}

func (cg *CharacterGenerator) getPixel(char uint8, row int, column int) bool {
	rowPos := (int(char)*8 + row) % cg.pageSize
	bits := cg.data[rowPos+cg.page*cg.pageSize]
	bit := cg.columnMap(column)
	value := bits >> uint(bit) & 1
	return value == 1
}

func setupCharactedGenerator(a *Apple2, board string, charRomFile string) error {
	var charGenMap charColumnMap
	initialCharGenPage := 0
	pageSize := charGenPageSize2Plus
	switch board {
	case "2plus":
		charGenMap = charGenColumnsMap2Plus
	case "basis108":
		charGenMap = charGenColumnsMap2Plus
		pageSize = charGenPageSizeBasis108
		initialCharGenPage = 2
	case "2e":
		charGenMap = charGenColumnsMap2e
		pageSize = charGenPageSize2E
	case "base64a":
		charGenMap = charGenColumnsMapBase64a
		initialCharGenPage = 1
	default:
		return fmt.Errorf("board %s not supported it must be '2plus', '2e', 'base64a', 'basis108", board)
	}

	cg, err := newCharacterGenerator(charRomFile, charGenMap, pageSize)
	if err != nil {
		return err
	}
	cg.setPage(initialCharGenPage)
	a.cg = cg
	return nil
}
