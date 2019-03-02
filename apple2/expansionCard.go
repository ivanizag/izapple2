package apple2

import (
	"bufio"
	"os"
)

type expansionSlots struct {
	mmu   *memoryManager
	cards [7]expansionCard
}

type expansionCard interface {
	insert(es *expansionSlots, slot int) cardBase
	activateRom()
}

type cardBase struct {
	es            *expansionSlots
	rom           []romPage
	slot          int
	softSwitchesR [16]softSwitchR
	softSwitchesW [16]softSwitchW
}

/*
https://applesaucefdc.com/woz/reference2/
http://yesterbits.com/media/pubs/AppleOrchard/articles/disk-ii-part-1-1983-apr.pdf
*/

type cardDisk2 struct {
	cardBase
}

func newCardDisk2(filename string) *cardDisk2 {
	var c cardDisk2
	c.rom = loadCardRom(filename)
	return &c
}

func (c *cardBase) insert(es *expansionSlots, slot int) {
	c.es = es
	c.slot = slot
	if c.rom != nil {
		//rom = c.rom[0]
	}

}

func loadCardRom(filename string) []romPage {
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

	pages := size / 256
	if (size % 256) > 0 {
		pages++
	}

	rom := make([]romPage, pages)
	for i := int64(0); i < size; i++ {
		rom[i>>8].burn(uint8(i), bytes[i])
	}

	return rom
}
