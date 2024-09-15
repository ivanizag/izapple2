package izapple2

import (
	"fmt"
)

/*
  See:
	https://github.com/davidgiven/cpm65
*/

type traceCpm struct {
	a           *Apple2
	skipConsole bool
}

const (
	cpmBdosEntrypoint uint16 = 0x0005
)

func newTraceCpm(skipConsole bool) *traceCpm {
	var t traceCpm
	t.skipConsole = skipConsole
	return &t
}

func (t *traceCpm) connect(a *Apple2) {
	t.a = a
}

func (t *traceCpm) inspect() {
	if !t.a.dmaActive {
		return // The 6502 is not running
	}

	softCard, ok := t.a.cards[t.a.dmaSlot].(*CardZ80SoftCard)
	if !ok {
		return // The DMA slot is not a Z80 SoftCard
	}

	pc := softCard.cpu.PC

	if pc == cpmBdosEntrypoint {
		command := softCard.cpu.BC.Lo
		switch command {
		case 2: // C_WRITE
			if !t.skipConsole {
				ch := softCard.cpu.DE.Lo
				fmt.Printf("CPM BDOS call %s from $%04x with \"%c\"\n",
					bdosCodeToName(command), pc, ch)
			}
		default:
			fmt.Printf("CPM BDOS call %s\n", bdosCodeToName(command))
		}
	}
}

func bdosCodeToName(code uint8) string {
	if code < uint8(len(bdosCommandNames)) {
		return fmt.Sprintf("%02v-%s", code, bdosCommandNames[code])
	}
	return fmt.Sprintf("%02v-UNKNOWN", code)
}

var bdosCommandNames = []string{
	// 0
	"P_TERMCPM", "C_READ", "C_WRITE", "A_READ", "A_WRITE",
	"L_WRITE", "C_RAWIO", "A_STATIN", "A_STATOUT", "C_WRITESTR",
	// 10
	"C_READSTR", "C_STAT", "S_BDOSVER", "DRV_ALLRESET", "DRV_SET",
	"F_OPEN", "F_CLOSE", "F_SFIRST", "F_SNEXT", "F_DELETE",
	// 20
	"F_READ", "F_WRITE", "F_MAKE", "F_RENAME", "DRV_LOGINVEC",
	"DRV_GET", "F_DMAOFF", "DRV_ALLOCVEC", "DRV_SETRO", "DRV_ROVEC",
	// 30
	"F_ATTRIB", "DRV_DPB", "F_USERNUM", "F_READRAND", "F_WRITERAND",
	"F_SIZE", "F_RANDREC", "DRV_RESET", "*", "",
	// 40
	"F_WRITEZ", "", "", "", "",
	"F_ERRMODE", "", "", "", "",
}
