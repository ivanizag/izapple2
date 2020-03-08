package apple2

import (
	"fmt"
)

/*
In out card experiment to interface with the emulator host.

See:
	"Apple II Monitors peeled."
	http://mysite.du.edu/~etuttle/math/acia.htm

*/

type cardInOut struct {
	cardBase
	i int
}

func (c *cardInOut) assign(a *Apple2, slot int) {
	for i := uint8(0x0); i <= 0xf; i++ {
		iCopy := i
		c.addCardSoftSwitchR(i, func(*ioC0Page) uint8 {
			value := []uint8{0xc1, 0xc1, 0x93, 0x0}[c.i%4]
			c.i++
			fmt.Printf("[cardInOut] Read access to softswith 0x%x for slot %v, value %x.\n", iCopy, slot, value)
			//return 0x41 + 0x80
			return []uint8{0x41, 0x41, 0x13}[i%3] + 0x80
		}, "INOUTR")
		c.addCardSoftSwitchW(i, func(_ *ioC0Page, value uint8) {
			fmt.Printf("[cardInOut] Write access to softswith 0x%x for slot %v, value 0x%x.\n", iCopy, slot, value)
		}, "INOUTW")
	}

	in := true
	out := false

	data := [256]uint8{
		// Register
		0xA9, 0xC2,
		0x85, 0x37,
		0x85, 0x39,
		0xA9, 0x10,
		0x85, 0x36,
		0xA9, 0x15,
		0x85, 0x38,
		0x60, 0xEA,

		// Out char
		0x8D, 0xA1, 0xC0,
		0x60, 0xEA,

		// Get char
		0x91, 0x28,
		0xAD, 0xA0, 0xC0,
		0x60,
	}

	if !out {
		// NOP the CSWL,H change
		for _, v := range []uint8{2, 3, 8, 9} {
			data[v] = 0xEA
		}
	}

	if !in {
		// NOP the KSWL,H change
		for _, v := range []uint8{4, 5, 12, 13} {
			data[v] = 0xEA
		}
	}

	c.romCsxx = newMemoryRange(0xC200, data[0:255])

	if slot != 2 {
		// To make ifwork on other slots, patch C2, A0 and A1
		panic("Assert failed. Only slot 2 supported for the InOut card")
	}
	c.cardBase.assign(a, slot)
}

/*
The ROM code was assembled using https://www.masswerk.at/6502/assembler.html

src:
	BASL = $28
	CSWL = $36
	CSWH = $37
	KSWL = $38
	KSWH = $39

	* = $C200
	Register:
		LDA #$C2
		STA *CSWH
		STA *KSWH
		LDA #$10
		STA *CSWL
		LDA #$15
		STA *KSWL
		RTS
		NOP
	OutChar:
		STA $C0A1
		RTS
		NOP
	GetChar:
		STA (BASL),Y
		LDA $C0A0
		RTS


assembled as:
0000 BASL   = 0028
0000 CSWL   = 0036
0000 CSWH   = 0037
0000 KSWL   = 0038
0000 KSWH   = 0039

* = $C200
C200 REGIST
C200        LDA #$C2        A9 C2
C202        STA *CSWH       85 37
C204        STA *KSWH       85 39
C206        LDA #$10        A9 10
C208        STA *CSWL       85 36
C20A        LDA #$15        A9 15
C20C        STA *KSWL       85 38
C20E        RTS             60
C20F        NOP             EA
C210 OUTCHA
C210        STA $C0A1       8D A1 C0
C213        RTS             60
C214        NOP             EA
C215 GETCHA
C215        STA (BASL),Y    91 28
C217        LDA $C0A0       AD A0 C0
C21A        RTS

object code:
A9 C2 85 37 85 39 A9 10
85 36 A9 15 85 38 60 EA
8D A1 C0 60 EA 91 28 AD
A0 C0 60

*/
