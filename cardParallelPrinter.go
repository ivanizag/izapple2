package izapple2

import (
	"os"
)

/*
Apple II Parallel Printer Interface card.

See:
	https://mirrors.apple2.org.za/Apple%20II%20Documentation%20Project/Interface%20Cards/Parallel/Apple%20II%20Parallel%20Printer%20Interface%20Card/

The card tells whether the printer is ready by the byte at $CnC1 of its ROM:
the acknowledge of the printer drives the address line A6 of the PROM, so
that while the printer is busy $CnC1 reads $90, a BCC on itself where the
firmware waits, and once it is ready $10, a BPL out of the loop. Programs
with drivers of their own, as The Print Shop, poll bit 7 of $CnC1. A file is
always ready, so the ROM has $10 there.

*/

// CardParallelPrinter represents a Parallel Printer Interface card
type CardParallelPrinter struct {
	cardBase
	file  *os.File
	ascii bool
}

func newCardParallelPrinterBuilder() *cardBuilder {
	return &cardBuilder{
		name:        "Parallel Printer Interface",
		description: "Card to dump to a file what would be printed to a parallel printer",
		defaultParams: &[]paramSpec{
			{"file", "File to store the printed code", "printer.out"},
			{"ascii", "Remove the 7 bit. Useful for normal text printing, but breaks graphics printing ", "false"},
		},
		buildFunc: func(params map[string]string) (Card, error) {
			var c CardParallelPrinter
			c.ascii = paramsGetBool(params, "ascii")
			filepath := paramsGetPath(params, "file")
			f, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return nil, err
			}
			c.file = f
			err = c.loadRomFromResource("<internal>/Apple II Parallel Printer Interface Card ROM fixed.bin", cardRomSimple)
			if err != nil {
				return nil, err
			}
			return &c, nil
		},
	}
}

func (c *CardParallelPrinter) assign(a *Apple2, slot int) {
	c.addCardSoftSwitchW(0, func(value uint8) {
		c.printByte(value)
	}, "PARALLELDEVW")

	c.addCardSoftSwitchR(4, func() uint8 {
		return 0xff // TODO: What are the bit values?
	}, "PARALLELSTATUSR")

	c.cardBase.assign(a, slot)
}

func (c *CardParallelPrinter) printByte(value uint8) {
	if c.ascii {
		// As text the MSB has to be removed, but if done, graphics modes won't work
		value &= 0x7f // Remove the MSB bit
	}
	c.file.Write([]byte{value})
}
