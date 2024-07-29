package izapple2

import (
	"bufio"
	"fmt"
	"os"
)

/*
In out card experiment to interface with the emulator host.

See:
	"Apple II Monitors peeled."
	http://mysite.du.edu/~etuttle/math/acia.htm


	PR#n stores Cn00 in CSWL and CSWH
	IN#n stores Cn00 in KSWL and KSWH
*/

// CardInOut is an experimental card to bridge with the host console
type CardInOut struct {
	cardBase
	reader *bufio.Reader
}

func newCardInOutBuilder() *cardBuilder {
	return &cardBuilder{
		name:        "InOut test card",
		description: "Card to test I/O",
		buildFunc: func(params map[string]string) (Card, error) {
			return &CardInOut{}, nil
		},
	}
}

func (c *CardInOut) assign(a *Apple2, slot int) {
	c.addCardSoftSwitchR(0, func() uint8 {
		if c.reader == nil {
			c.reader = bufio.NewReader(os.Stdin)
		}
		value, err := c.reader.ReadByte()

		if err != nil {
			panic(err)
		}
		value += 0x80
		if value&0x7f == 10 {
			value = 13 + 0x80
		}
		// fmt.Printf("[cardInOut] Read access to softswith 0x%x for slot %v, value %x.\n", 0, slot, value)
		return value
	}, "INOUTR")
	c.addCardSoftSwitchW(1, func(value uint8) {
		// fmt.Printf("[cardInOut] Write access to softswith 0x%x for slot %v, value 0x%x: %v, %v.\n", 1, slot, value, value&0x7f, string(value&0x7f))
		if value&0x7f == 13 {
			fmt.Printf("\n")
		} else {
			fmt.Printf("%v", string(value&0x7f))
		}

	}, "INOUTW")

	data := buildBaseInOutRom(slot)
	c.romCsxx = newMemoryRangeROM(0xC200, data, "InOUt card")
	c.cardBase.assign(a, slot)
}

func buildBaseInOutRom(slot int) []uint8 {
	data := [256]uint8{
		// Register
		0x4c, 0x40, 0xc2, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,

		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,

		0x48, 0xA5, 0x38, 0xD0, 0x11, 0xA9, 0xC2, 0xC5,
		0x39, 0xD0, 0x0B, 0xAD, 0x4F, 0x85, 0x38, 0x68,

		0x91, 0x28, 0xAD, 0xA0, 0xC0, 0x60, 0x68, 0x8D,
		0xA1, 0xC0, 0x60,
	}

	// Fix slot dependant addresses
	data[0x02] = uint8(0xc0 + slot)
	data[0x46] = uint8(0xc0 + slot)
	data[0x53] = uint8(0x80 + slot<<4)
	data[0x58] = uint8(0x81 + slot<<4)

	return data[:]
}

/*
The ROM code was assembled using https://www.masswerk.at/6502/assembler.html

We will have $Cn00 as the entry point for CSWL/H. But before doing
anything we have to check that we are not in $Cn00 because of an IN#.
To da that we check id KSWL/H is $Cn00, it is it we wif it to INEntry.

src:
	BASL = $28
	KSWL = $38
	KSWH = $39

	* = $C200
	Entry:
		JMP SkipHeader

	* = $C240
	SkipHeader:
		PHA
		LDA *KSWL
		BNE PREntry
		LDA #$C2
		CMP *KSWH
		BNE PREntry
	FixKSWL:
		LDA #<INEntry
		STA *KSWL
	INEntry:
		PLA
		STA (BASL),Y
		LDA $C0A0
		RTS
	PREntry:
		PLA
		STA $C0A1
		RTS



Listing:
pass 2

0000 BASL   = 0028
0000 KSWL   = 0038
0000 KSWH   = 0039

	* = $C200
	C200 ENTRY:
	C200        JMP SKIPHE      4C 40 C2

	* = $C240
	C240 SKIPHE
	C240        PHA             48
	C241        LDA *KSWL       A5 38
	C243        BNE PRENTR      D0 11
	C245        LDA #$C2        A9 C2
	C247        CMP *KSWH       C5 39
	C249        BNE PRENTR      D0 0B
	C24B        LDA #<INENTR    A9 4F
	C24D        STA *KSWL       85 38
	C24F INENTR
	C24F        PLA             68
	C250        STA (BASL),Y    91 28
	C252        LDA $C0A0       AD A0 C0
	C255        RTS             60
	C256 PRENTR
	C256        PLA             68
	C257        STA $C0A1       8D A1 C0
	C25A        RTS             60

	done.

Object Code:
c200:
4C 40 C2
c240:
48 A5 38 D0 11
A9 C2 C5 39 D0 0B A9 4F
85 38 68 91 28 AD A0 C0
60 68 8D A1 C0 60
*/
