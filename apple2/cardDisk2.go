package apple2

import (
	"bufio"
	"fmt"
	"os"
)

/*
https://applesaucefdc.com/woz/reference2/
Good explanation of the softswitches and the phases:
http://yesterbits.com/media/pubs/AppleOrchard/articles/disk-ii-part-1-1983-apr.pdf
*/

type cardDisk2 struct {
	cardBase
	phases    [4]bool
	power     [2]bool
	selected  int // Only 0 and 1 supported
	writeMode bool
}

// type softSwitchR func(io *ioC0Page) uint8

func newCardDisk2(filename string) *cardDisk2 {
	var c cardDisk2
	c.rom = loadCardRom(filename)

	// Phase control soft switches
	for i := 0; i < 4; i++ {
		c.ssr[i<<1] = func(_ *ioC0Page) uint8 {
			fmt.Printf("DISKII: Phase %v off\n", i)
			return 0
		}
		c.ssr[(i<<1)+1] = func(_ *ioC0Page) uint8 {
			fmt.Printf("DISKII: Phase %v on\n", i)
			return 0
		}
	}

	// Other soft switches
	c.ssr[0x8] = func(_ *ioC0Page) uint8 {
		c.power[c.selected] = false
		fmt.Printf("DISKII: Disk %v is off\n", c.selected)
		return 0
	}
	c.ssr[0x9] = func(_ *ioC0Page) uint8 {
		c.power[c.selected] = true
		fmt.Printf("DISKII: Disk %v is on\n", c.selected)
		return 0
	}
	c.ssr[0xA] = func(_ *ioC0Page) uint8 {
		c.selected = 0
		fmt.Printf("DISKII: Disk %v selected\n", c.selected)
		return 0
	}
	c.ssr[0xB] = func(_ *ioC0Page) uint8 {
		c.selected = 1
		fmt.Printf("DISKII: Disk %v selected\n", c.selected)
		return 0
	}

	var i uint8
	// Q6L
	c.ssr[0xC] = func(_ *ioC0Page) uint8 {
		fmt.Printf("DISKII: Reading\n")
		i++
		return i
	}

	c.ssw[0xC] = func(_ *ioC0Page, value uint8) {
		fmt.Printf("DISKII: Writing the value 0x%02x\n", value)
	}

	// Q6H
	c.ssr[0xD] = func(_ *ioC0Page) uint8 {
		c.writeMode = false
		fmt.Printf("DISKII: Sense write protection\n")
		return 0
	}

	// Q7L
	c.ssr[0xE] = func(_ *ioC0Page) uint8 {
		c.writeMode = false
		fmt.Printf("DISKII: Set read mode\n")
		return 0
	}

	// Q7H
	c.ssr[0xF] = func(_ *ioC0Page) uint8 {
		c.writeMode = true
		fmt.Printf("DISKII: Set write mode\n")
		return 0
	}
	// TODO: missing C, D, E, and F

	return &c
}

func loadCardRom(filename string) []memoryPage {
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

	memPages := make([]memoryPage, pages)
	for i := range rom {
		memPages[i] = &rom[i]
	}

	return memPages
}
