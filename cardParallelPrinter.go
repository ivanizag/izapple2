package izapple2

import (
	"os"
)

/*
Apple II Parallel Printer Interface card.

See:
	https://mirrors.apple2.org.za/Apple%20II%20Documentation%20Project/Interface%20Cards/Parallel/Apple%20II%20Parallel%20Printer%20Interface%20Card/

*/

// CardParallelPrinter represents a Parallel Printer Interface card
type CardParallelPrinter struct {
	cardBase
	file *os.File
}

// NewCardParallelPrinter creates a new CardParallelPrinter
func NewCardParallelPrinter() *CardParallelPrinter {
	var c CardParallelPrinter
	c.name = "Parallel Printer Interface"
	c.loadRomFromResource("<internal>/Apple II Parallel Printer Interface Card ROM fixed.bin")
	return &c
}

func (c *CardParallelPrinter) assign(a *Apple2, slot int) {
	f, err := os.OpenFile(printerFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	c.file = f

	c.addCardSoftSwitchW(0, func(value uint8) {
		c.printByte(value)
	}, "PARALLELDEVW")

	c.cardBase.assign(a, slot)
}

const printerFile = "printer.out"

func (c *CardParallelPrinter) printByte(value uint8) {
	value = value & 0x7f // Remove the MSB bit
	c.file.Write([]byte{value})
}
